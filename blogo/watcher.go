package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func InitWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					if strings.HasSuffix(event.Name, ".md") {
						log.Printf("Reloading article: %v", event.Name)
						article, _ := GetArticleFromFile(event.Name)
						if article.Slug != "" {
							LoadArticle(article)
							GenerateArticleStatic(article)
						}
						UpdateFeed()
					}
				}

				// On article delete, remove it from the map
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if strings.HasSuffix(event.Name, ".md") {
						log.Printf("Removing article: %v", event.Name)
						RemoveArticle(event.Name)
						RemoveArticleStatic(event.Name)
						UpdateFeed()
					}
				}

				// If renamed or moved, remove the old article from the map
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					if strings.HasSuffix(event.Name, ".md") {
						log.Printf("Replacing article: %v", event.Name)
						RemoveArticle(event.Name)
						RemoveArticleStatic(event.Name)
						UpdateFeed()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(fmt.Sprintf("%v/articles", os.Getenv("CONTENT_PATH")))
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
