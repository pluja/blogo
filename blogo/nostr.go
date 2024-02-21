package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/rs/zerolog/log"
)

var nostrSk string
var nostrPk string
var relayList []string

// Initializes the Nostr key set and relay list.
func InitNostr() error {
	nsec := os.Getenv("NOSTR_NSEC")
	var npub string
	var err error
	if nsec == "" {
		log.Warn().Msg("NOSTR_NSEC not set. Generating a new key pair.")
		nostrSk, nostrPk, nsec, npub, err = GetNewKeySet()
		if err != nil {
			return fmt.Errorf("failed to get public key: %w", err)
		}
		log.Info().Msgf("Generated new key set:\n\nnsec: %v\nnpub: %v\n\n", nsec, npub)
	} else {
		log.Info().Msg("NOSTR_NSEC set. Deriving existing key pair.")
		_, value, err := nip19.Decode(nsec)
		if err != nil {
			return fmt.Errorf("failed to decode: %w", err)
		}

		var ok bool
		nostrSk, ok = value.(string)
		if !ok {
			return fmt.Errorf("failed to nip19 decode to string: %w", err)
		}

		nostrPk, err = nostr.GetPublicKey(nostrSk)
		if err != nil {
			return fmt.Errorf("failed to get public key: %w", err)
		}
		npub, err = nip19.EncodePublicKey(nostrPk)
		if err != nil {
			return fmt.Errorf("failed to encode public key: %w", err)
		}
	}

	envRelays := os.Getenv("NOSTR_RELAYS")
	if envRelays == "" {
		log.Warn().Msg("NOSTR_RELAYS not set. Using default relays.")
		relayList = []string{"wss://nostr-pub.wellorder.net", "wss://relay.damus.io", "wss://relay.nostr.band"}
	} else {
		relayList = strings.Split(envRelays, ",")
	}

	fmt.Println("Public Key:", nostrPk)
	fmt.Println("npub: ", npub)
	return nil
}

// Publishes the article to Nostr if enabled and not yet published
func PublishArticleToNostr(article ArticleData) error {
	if article.Slug == "about" {
		log.Info().Msg("Not publishing about page to Nostr")
		return nil
	}

	if os.Getenv("PUBLISH_TO_NOSTR") == "false" {
		log.Info().Msg("PUBLISH_TO_NOSTR is set to false. Not publishing...")
		return nil
	}

	log.Printf("NostrUrl value (%v) for %v", article.NostrUrl, article.Slug)
	// We try to parse the NostrUrl field as a boolean.
	// If it's empty or set to true, we publish.
	// If it's set to false or anything that evaluates to false, we don't publish.
	publishNostr, err := strconv.ParseBool(article.NostrUrl)
	if err != nil {
		if article.NostrUrl == "" {
			publishNostr = true
		} else {
			log.Info().Msgf("NostrUrl value (%v) is set, and not False. Not publishing...", article.NostrUrl)
			return nil
		}
	}

	// If the NostrUrl field is set to something that evaluates as false, we don't publish
	if !publishNostr {
		log.Info().Msgf("NostrUrl value (%v) is set to False. Not publishing...", article.NostrUrl)
		return nil
	}

	// If the article is a draft we don't publish
	if !article.Draft {
		log.Info().Msgf("Publishing %v to Nostr", article.Slug)
		naddr, err := NostrPublish(article)
		if err != nil {
			log.Err(err).Msg("Could not publish to Nostr")
		} else {
			err = AddMetadataToFile(fmt.Sprintf("%v.md", article.Slug), "NostrUrl", fmt.Sprintf("https://habla.news/a/%v", naddr))
			if err != nil {
				log.Err(err).Msgf("Could not add %v to NostrUrl field", naddr)
			}
		}
	} else {
		log.Printf("Won't publisht this to Nostr: it's a draft")
	}
	return nil
}

// Publishes an article of type ArticleData to Nostr.
func NostrPublish(ad ArticleData) (string, error) {
	// Wipe the YAML Metadata block from the article
	sections := strings.SplitN(string(ad.Md), "---", 3)
	if len(sections) >= 3 {
		ad.Md = sections[2]
	}

	// Add the article original URL to the top of the article
	ad.Md = fmt.Sprintf("> [Read the original blog post](%v)\n\n", path.Join(Blogo.Url, "/p/", ad.Slug)) + ad.Md

	// md5 hash the title and slug to get a unique ID
	id := fmt.Sprintf("%x", md5.Sum([]byte(ad.Title+ad.Author)))

	// Create the Nostr event
	tags := nostr.Tags{
		nostr.Tag{"d", id},
		nostr.Tag{"title", ad.Title},
		nostr.Tag{"client", "blogo"},
	}

	articleTags := nostr.Tags{}
	for _, tag := range ad.Tags {
		articleTags = append(articleTags, nostr.Tag{"t", tag})
	}
	tags = append(tags, articleTags...)

	ev := nostr.Event{
		PubKey:    nostrPk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindArticle,
		Tags:      tags,
		Content:   ad.Md,
	}

	log.Debug().Msgf("Nostr event: %v", ev)

	// Sign the event
	err := ev.Sign(nostrSk)
	if err != nil {
		log.Err(err).Msg("Could not sign event")
		return "", err
	}

	// Publish the event to the relays
	ctx := context.Background()
	connected := false
	published := false
	if os.Getenv("DEV") == "true" {
		// In development mode, mock the Nostr publish
		connected = true
		published = true
	} else {
		for _, url := range relayList {
			relay, err := nostr.RelayConnect(ctx, url)
			if err != nil {
				log.Err(err).Msgf("failed to connect to relay %v:", url)
				continue
			}
			connected = true
			if err := relay.Publish(ctx, ev); err != nil {
				log.Warn().Err(err).Msgf("failed to publish to %v", url)
				continue
			}

			published = true
			fmt.Printf("published %v to %s\n", ev.ID, url)
		}
	}

	// Return an error if we were unable to publish to any relay
	if !connected || !published {
		return "", fmt.Errorf("unable to publish %v to Nostr, connected: %v, published: %v", ev.ID, connected, published)
	}

	// Encode the note ID to naddr format
	naddr, err := nip19.EncodeEntity(ev.PubKey, nostr.KindArticle, id, []string{})
	if err != nil {
		log.Err(err).Msg("Could not encode note ID")
		return ev.ID, err
	}
	return naddr, nil
}

// Returns a key set in the following order: sk, pk, nsec, npub
func GetNewKeySet() (string, string, string, string, error) {
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to get public key: %w", err)
	}
	nsec, err := nip19.EncodePrivateKey(sk)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to encode public key: %w", err)
	}
	npub, err := nip19.EncodePublicKey(pk)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to encode public key: %w", err)
	}
	return sk, pk, nsec, npub, nil
}
