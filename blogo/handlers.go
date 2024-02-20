package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	articles := Badger.GetAllArticles()
	// Sort by date
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})

	varmap := map[string]interface{}{
		"Articles": articles,
		"Blogo":    Blogo,
	}
	// Execute the template from templates.go
	if err := IndexTmpl.ExecuteTemplate(w, "base", varmap); err != nil {
		log.Error().Err(err).Msg("Error executing template:")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ServeBlogPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	log.Debug().Msgf("%v", slug)
	blogPath := fmt.Sprintf("%v/content", os.Getenv("CONTENT_PATH"))
	filePath := path.Join(blogPath, fmt.Sprintf("%s.html", slug))

	log.Debug().Msgf("%v", filePath)

	http.ServeFile(w, r, filePath)
}

func GetRawMarkdown(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Get article from redis
	article, err := Badger.GetPostBySlug(slug)
	if err != nil {
		log.Err(err).Msg("Error getting article from Redis")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(article.Md))
}

func GetTagPosts(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")

	// Get all article IDs from the corresponding tag set
	articles := Badger.GetAllArticles()
	var tagArticles []ArticleData
	for _, article := range articles {
		if StringInSlice(tag, article.Tags) {
			tagArticles = append(tagArticles, article)
		}
	}

	// Sort by date
	sort.Slice(tagArticles, func(i, j int) bool {
		return tagArticles[i].Date.After(tagArticles[j].Date)
	})

	varmap := map[string]interface{}{
		"Articles": tagArticles,
		"Blogo":    Blogo,
		"Tag":      tag,
	}

	// Execute the template from templates.go
	if err := TagTmpl.ExecuteTemplate(w, "base", varmap); err != nil {
		log.Error().Err(err).Msg("Error executing template:")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetAbout(w http.ResponseWriter, r *http.Request) {
	varmap := map[string]interface{}{
		"About": About.Data,
		"Blogo": Blogo,
	}
	// Execute the template from templates.go
	if err := AboutTmpl.ExecuteTemplate(w, "base", varmap); err != nil {
		log.Error().Err(err).Msg("Error executing template:")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleRssFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/rss+xml")
	w.Write([]byte(RssFeed()))
}

func HandleAtomFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/atom+xml")
	w.Write([]byte(AtomFeed()))
}

func HandleJsonFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(JsonFeed()))
}
