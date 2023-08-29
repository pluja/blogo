package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	articles := GetAllArticles()
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

func GetBlogPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Get article from redis
	ctx := context.Background()
	result, err := RedisDb.Get(ctx, slug).Result()
	if err != nil {
		log.Err(err).Msg("Error getting article from Redis")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Unmarshal the result into an Article struct
	var article ArticleData
	err = json.Unmarshal([]byte(result), &article)
	if err != nil {
		log.Err(err).Msg("Error unmarshalling article from Redis")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	varmap := map[string]interface{}{
		"Article": article,
		"Blogo":   Blogo,
	}

	// Execute the template from templates.go
	if err := PostTmpl.ExecuteTemplate(w, "base", varmap); err != nil {
		log.Error().Err(err).Msg("Error executing template:")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetTagPosts(w http.ResponseWriter, r *http.Request) {
	tag := chi.URLParam(r, "tag")

	// Get all article IDs from the corresponding tag set
	ctx := context.Background()
	articleIDs, err := RedisDb.SMembers(ctx, "tag:"+tag).Result()
	if err != nil {
		log.Err(err).Msg("Error getting articles from Redis")
		articleIDs = []string{}
	}

	var articles []ArticleData
	// Iterate over each article ID and fetch the article detail
	for _, id := range articleIDs {
		result, err := RedisDb.Get(ctx, id).Result()
		if err != nil {
			log.Err(err).Msg("Error getting article from Redis")
			continue
		}

		// Unmarshal the result into an Article struct
		var article ArticleData
		err = json.Unmarshal([]byte(result), &article)
		if err != nil {
			log.Err(err).Msg("Error unmarshalling article from Redis")
		}

		// Skip draft articles
		if !article.Draft {
			articles = append(articles, article)
		}
	}

	// Sort by date
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})

	varmap := map[string]interface{}{
		"Articles": articles,
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
