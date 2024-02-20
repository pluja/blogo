package main

import (
	"fmt"
	"os"
	"path"
)

// Loads an article from a markdown file and stores it in Redis
func GenerateArticleStatic(article ArticleData) (err error) {
	varmap := map[string]interface{}{
		"Article": article,
		"Blogo":   Blogo,
	}
	switch article.Slug {
	case "about":
		About.Slug = "about"
		About.Data = article
	default:
		blogPath := fmt.Sprintf("%v/content", os.Getenv("CONTENT_PATH"))
		_, err := os.Stat(blogPath)

		if os.IsNotExist(err) {
			if err := os.Mkdir(blogPath, os.ModePerm); err != nil {
				return fmt.Errorf("error creating blog directory: %v", err)
			}
		} else if err != nil {
			return fmt.Errorf("error checking blog directory: %v", err)
		}

		filePath := fmt.Sprintf("%v/%v.html", blogPath, article.Slug)
		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("error creating static HTML file: %v", err)
		}

		err = PostTmpl.ExecuteTemplate(file, "base", varmap)
		if err != nil {
			return fmt.Errorf("error writing static HTML: %v", err)
		}

	}
	return nil
}

func RemoveArticleStatic(filepath string) (err error) {
	slug, _ := ParseFilePath(filepath)
	filename := fmt.Sprintf("%v.html", slug)

	blogPath := fmt.Sprintf("%v/content", os.Getenv("CONTENT_PATH"))

	return os.Remove(path.Join(blogPath, filename))
}
