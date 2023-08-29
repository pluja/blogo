package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/feeds"
	"github.com/rs/zerolog/log"
)

func UpdateFeed() error {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       os.Getenv("BLOGO_TITLE"),
		Link:        &feeds.Link{Href: fmt.Sprintf("%v/rss", os.Getenv("BLOGO_URL"))},
		Description: os.Getenv("BLOGO_DESCRIPTION"),
		Author:      &feeds.Author{Name: os.Getenv("BLOGO_TITLE")},
		Created:     now,
	}

	articles := GetAllArticles()

	feed.Items = []*feeds.Item{}
	for _, article := range articles {
		item := &feeds.Item{
			Title:       article.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%v/p/%v", os.Getenv("BLOGO_URL"), article.Slug)},
			Description: article.Summary,
			Created:     article.Date,
		}

		feed.Items = append(feed.Items, item)
	}

	// Save feed to Redis
	ctx := context.Background()
	json, err := json.Marshal(feed)
	if err != nil {
		log.Err(err).Msg("Error marshalling feed to JSON")
		return err
	}
	err = RedisDb.Set(ctx, "feed", json, 0).Err()
	if err != nil {
		log.Err(err).Msg("Error saving feed to Redis")
	}
	return err
}

func GetFeed() feeds.Feed {
	ctx := context.Background()
	result, err := RedisDb.Get(ctx, "feed").Result()
	if err != nil {
		log.Err(err).Msg("Error getting feed from Redis")
		return feeds.Feed{}
	}

	// Unmarshal the result into an Article struct
	var feed feeds.Feed
	err = json.Unmarshal([]byte(result), &feed)
	if err != nil {
		log.Err(err).Msg("Error unmarshalling feed from Redis")
		return feeds.Feed{}
	}

	return feed
}

func RssFeed() string {
	feed := GetFeed()

	rss, err := feed.ToRss()
	if err != nil {
		log.Err(err).Msg("Error generating RSS feed")
		return ""
	}

	// return feed
	return rss
}

func AtomFeed() string {
	feed := GetFeed()

	atom, err := feed.ToAtom()
	if err != nil {
		log.Err(err).Msg("Error generating Atom feed")
		return ""
	}

	// return feed
	return atom
}

func JsonFeed() string {
	feed := GetFeed()

	json, err := feed.ToJSON()
	if err != nil {
		log.Err(err).Msg("Error generating JSON feed")
		return ""
	}

	// return feed
	return json
}
