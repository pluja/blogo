package main

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("p")

	pageNum, _ := strconv.Atoi(page)
	if pageNum <= 0 {
		pageNum = 1
	}

	from := (pageNum - 1) * 10
	to := from + 10

	articles := Badger.GetAllArticles()
	// Sort by date
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})

	totalPages := int(math.Ceil(float64(len(articles)) / 10.0))
	if to > len(articles) {
		to = len(articles)
	}
	pagedArticles := articles[from:to]

	varmap := map[string]interface{}{
		"Articles":   pagedArticles,
		"Blogo":      Blogo,
		"Page":       pageNum,
		"TotalPages": totalPages,
	}

	log.Debug().Msgf("From: %d - To: %d", from, to)
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
	page := r.URL.Query().Get("p")

	pageNum, _ := strconv.Atoi(page)
	pageNum = int(math.Max(0, float64(pageNum-1)))

	from := pageNum * 10
	to := from + 10

	articles := Badger.GetAllArticles()
	tagArticles := make([]ArticleData, 0, len(articles))
	for _, article := range articles {
		if StringInSlice(tag, article.Tags) {
			tagArticles = append(tagArticles, article)
		}
	}

	sort.Slice(tagArticles, func(i, j int) bool {
		return tagArticles[i].Date.After(tagArticles[j].Date)
	})

	totalPages := int(math.Ceil(float64(len(tagArticles)) / 10.0))

	if to > len(tagArticles) {
		to = len(tagArticles)
	}
	log.Debug().Msgf("From: %d - To: %d", from, to)
	pagedArticles := tagArticles[from:to]

	varmap := map[string]interface{}{
		"Articles":   pagedArticles,
		"Blogo":      Blogo,
		"Tag":        tag,
		"Page":       pageNum,
		"TotalPages": totalPages,
	}

	if err := TagTmpl.ExecuteTemplate(w, "base", varmap); err != nil {
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
