package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var RedisDb *redis.Client

func InitRedis() {
	raddr := fmt.Sprintf("%v:%v", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	RedisDb = redis.NewClient(&redis.Options{
		Addr:     raddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func GetAllArticles() []ArticleData {
	ctx := context.Background()
	// Get all article IDs from the "articles" set
	articleIDs, err := RedisDb.SMembers(ctx, "articles").Result()
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

		articles = append(articles, article)
	}

	// Sort by date
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Date.After(articles[j].Date)
	})

	return articles
}
