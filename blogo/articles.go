package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

var About struct {
	Slug string
	Data ArticleData
}

var markdown goldmark.Markdown

func InitGoldmark() {
	markdown = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			extension.GFM,
			extension.Footnote,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
}

// Loads all articles from the articles folder
func LoadArticles() error {
	InitGoldmark()
	var slugs []string
	err := filepath.Walk(path.Join(os.Getenv("CONTENT_PATH"), "/articles/"), func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			article, err := GetArticleFromFile(fpath)
			if err != nil {
				log.Error().Msgf("Could not get article from file %v", fpath)
				return err
			}

			err = LoadArticle(article)
			if err != nil {
				return err
			}

			err = GenerateArticleStatic(article)
			if err != nil {
				return err
			}
			slugs = append(slugs, strings.TrimSuffix(info.Name(), ".md"))
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Remove articles that are no longer in the articles folder from Redis
	articleSlugs, err := Badger.GetAllArticleSlugs()
	if err != nil {
		log.Err(err).Msg("Error getting articles from Redis")
		articleSlugs = []string{}
	}
	left, _ := Difference(articleSlugs, slugs)
	if len(left) > 0 {
		log.Printf("Removing %v articles from Badger. Left articles: %v", len(left), slugs)
	}
	for _, articleSlug := range left {
		if !StringInSlice(articleSlug, slugs) {
			RemoveArticle(fmt.Sprintf("%v/articles/%v.md", os.Getenv("CONTENT_PATH"), articleSlug))
		}
	}

	// Update the RSS feed
	err = UpdateFeed()
	if err != nil {
		log.Err(err).Msg("Error updating RSS feed")
	}
	return nil
}

// Parses a .md file and returns the HTML and the raw markdown
func GetArticleContent(filename string) (template.HTML, string, error) {
	log.Printf("Loading article: %v", filename)
	md, err := os.ReadFile(fmt.Sprintf("%v/articles/%v", os.Getenv("CONTENT_PATH"), filename))
	if err != nil {
		return template.HTML(""), "", err
	}

	// Remove everything in the yaml metadata block

	var htmlBuf bytes.Buffer
	err = markdown.Convert(md, &htmlBuf)
	if err != nil {
		return template.HTML(""), "", err
	}
	html := htmlBuf.Bytes()
	return template.HTML(html), string(md), nil
}

// Loads an article from a markdown file and stores it in Redis
func LoadArticle(article ArticleData) (err error) {
	switch article.Slug {
	case "about":
		About.Slug = "about"
		About.Data = article
	default:
		// Marshal the article data to JSON
		articleJson, err := json.Marshal(article)
		if err != nil {
			return fmt.Errorf("error while marshalling article to JSON: %v", err)
		}
		Badger.Set("post_"+article.Slug, articleJson)
	}

	if article.NostrUrl == "" && article.Slug != "about" {
		// Publish to Nostr
		err = PublishArticleToNostr(article)
		if err != nil {
			log.Err(err).Msg("Error publishing article to Nostr")
		}
	}
	return nil
}

// Removes an article from Redis and the articles set
func RemoveArticle(filename string) {
	log.Printf("Removing article: %v", filename)
	slug, _ := ParseFilePath(filename)
	Badger.DeleteArticle(slug)
}

// Returns an ArticleData struct from a markdown file
func GetArticleFromFile(filepath string) (ArticleData, error) {
	slug, extension := ParseFilePath(filepath)
	filename := slug + extension

	var article ArticleData
	// Read the markdown file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return article, err
	}

	var buf strings.Builder
	pContext := parser.NewContext()
	if err := markdown.Convert(content, &buf, parser.WithContext(pContext)); err != nil {
		return article, err
	}

	metadata := meta.Get(pContext)
	if filepath == fmt.Sprintf("%v/articles/about.md", os.Getenv("CONTENT_PATH")) {
		article = ArticleData{}
	} else {
		// Handle drafts
		draftValue, exists := metadata["Draft"]
		if !exists {
			log.Err(err).Msgf("Could not parse draft value for %v, considering draft", filepath)
			return article, errors.New("could not parse draft value")
		}

		var draft bool
		switch articleDraft := draftValue.(type) {
		case bool:
			draft = articleDraft
		case string:
			if isDraft, err := strconv.ParseBool(articleDraft); err == nil && isDraft {
				draft = true
			} else if err != nil {
				log.Warn().Msgf("Could not parse draft value for %v, considering draft", filepath)
				draft = true
			}
		default:
			log.Err(err).Msgf("Could not parse draft value for %v, considering draft", filepath)
			draft = true
		}

		// Parse date
		dateString := GetMapStringValue(metadata, "Date")
		date, err := time.Parse("2006-01-02 15:04", dateString)
		if err != nil {
			date, err = time.Parse("2006-01-02", dateString)
			if err != nil {
				log.Err(err).Msgf("Could not parse date for %v, using current time", filepath)
				date = time.Now()
			}
		}

		image := GetMapStringValue(metadata, "Image")
		if image != "" && strings.HasPrefix(image, "/") {
			image = fmt.Sprintf("%v%v", Blogo.Url, image)
		}

		// Fill article Data
		article = ArticleData{
			Date:     date,
			Draft:    draft,
			Image:    image,
			Title:    GetMapStringValue(metadata, "Title"),
			Author:   GetMapStringValue(metadata, "Author"),
			Summary:  GetMapStringValue(metadata, "Summary"),
			Layout:   GetMapStringValue(metadata, "Layout"),
			NostrUrl: GetMapStringValue(metadata, "NostrUrl"),
		}

		if tags, ok := metadata["Tags"].([]interface{}); ok {
			for _, tag := range tags {
				if strTag, ok := tag.(string); ok {
					article.Tags = append(article.Tags, strTag)
				} else {
					log.Warn().Msgf("Could not parse tag %v for %v", tag, filepath)
				}
			}
		}
	}

	html, md, err := GetArticleContent(fmt.Sprintf("%v", filename))
	if err != nil {
		return ArticleData{}, err
	}

	article.Html = html
	article.Md = md

	article.Slug = slug

	return article, nil
}

// Adds or modifies metadata in a markdown file.md
func AddMetadataToFile(filename, key, value string) error {
	filePath := fmt.Sprintf("%v/articles/%v", os.Getenv("CONTENT_PATH"), filename)
	// Read the markdown file
	markdown, err := os.ReadFile(filePath)
	if err != nil {
		log.Error().Msgf("Could not read markdown file %v", filePath)
		return err
	}

	// Separate the YAML block
	sections := strings.SplitN(string(markdown), "---", 3)
	if len(sections) < 3 {
		return errors.New("could not find YAML block")
	}

	// Parse the existing YAML
	metadata := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(sections[1]), &metadata)
	if err != nil {
		log.Error().Msgf("Could not unmarshal YAML data for file %v", filePath)
		return err
	}

	// Add new metadata
	metadata[key] = value

	// Rebuild the YAML
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		log.Error().Msgf("Could not marshal YAML data for file %v", filePath)
		return err
	}

	// Reassemble the markdown
	var buffer bytes.Buffer
	buffer.WriteString("---\n")
	buffer.Write(yamlData)
	buffer.WriteString("---\n")
	buffer.WriteString(sections[2])

	// Write the updated markdown to the file
	err = os.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		log.Error().Msgf("Could not write updated markdown to file %v", filePath)
		return err
	}

	article, err := GetArticleFromFile(path.Join(os.Getenv("CONTENT_PATH"), "/articles", filename))
	if err != nil {
		log.Error().Msgf("Could not get article from file %v", filePath)
		return err
	}
	LoadArticle(article)
	return nil
}

// Creates an article template skeleton
func CreateArticleTemplate(name string) error {
	// Clean from spaces using a replacer for efficiency
	replacer := strings.NewReplacer(" ", "-", "_", "-", ".", "-", ",", "-", "!", "-")
	name = replacer.Replace(name)
	name = strings.ToLower(name)

	contentPath := os.Getenv("CONTENT_PATH")
	filePath := fmt.Sprintf("%s/articles/%s.md", contentPath, name)

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file %s already exists", filePath)
	}

	metadata := map[string]interface{}{
		"Author":  "John Doe",
		"Title":   strings.Replace(name, "-", " ", -1),
		"Summary": "A brief summary of what this post is about.",
		"Tags":    []string{"tag1", "tag2"},
		"Date":    time.Now().Format("2006-01-02 15:04"),
		"Image":   "https://picsum.photos/1920/1080",
		"Layout":  "post",
		"Draft":   true,
	}

	// Build the YAML
	yamlData, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	// Assemble the markdown
	var buffer bytes.Buffer
	buffer.WriteString("---\n")
	buffer.Write(yamlData)
	buffer.WriteString("---\n")

	// Write the assembled markdown to the file
	err = os.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
