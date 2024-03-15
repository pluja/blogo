package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
)

var IndexTmpl *template.Template
var TagTmpl *template.Template
var PostTmpl *template.Template
var AboutTmpl *template.Template

func InitRoutes() *chi.Mux {
	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	if os.Getenv("DEV") == "true" {
		// Parse templates on every request
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				InitTemplates()
				next.ServeHTTP(w, r)
			})
		})
	}

	// Setup CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	fileServer := http.FileServer(http.Dir(fmt.Sprintf("%v/static", os.Getenv("CONTENT_PATH"))))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", GetIndex)
	r.Get("/p/{slug}", ServeBlogPost)
	r.Get("/p/{slug}/raw", GetRawMarkdown)
	r.Get("/t/{tag}", GetTagPosts)
	r.Get("/about", GetAbout)

	r.Get("/rss", HandleRssFeed)
	r.Get("/atom", HandleAtomFeed)
	r.Get("/json", HandleJsonFeed)

	return r
}

func createTemplate(files []string) *template.Template {
	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
		"truncate": func(s string) string {
			if len(s) > 250 {
				return s[:250]
			} else {
				return s
			}
		},
		"slice": func(args ...interface{}) []interface{} {
			return args
		},
		"add": func(a, b int) int {
			return a + b
		},
		"html": func(value string) template.HTML { return template.HTML(value) },
		"safeUrl": func(s string) template.URL {
			return template.URL(s)
		},
		"getObject": func(d datatypes.JSON) []string {
			var obj []string
			if err := json.Unmarshal(d, &obj); err != nil {
				log.Error().Err(err).Msg("Error unmarshalling JSON:")
			}
			return obj
		},
		"baseUrl": func(u string) string {
			parsedUrl, err := url.Parse(u)
			if err != nil {
				fmt.Println(err)
			}

			return (parsedUrl.Scheme + "://" + parsedUrl.Host)
		},
		"humanizeTime": func(t time.Time) string {
			return humanize.Time(t)
		},
		"readTime": func(s string) int {
			words := strings.Fields(s)
			wordCount := len(words)

			readSpeed := 225

			// Calculate time in minutes, use math.Ceil to round up to nearest whole number.
			readTime := math.Ceil(float64(wordCount) / float64(readSpeed))
			return int(readTime)
		},
		"dateString": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}

	tmpl := template.New("").Funcs(funcMap)

	_, err := tmpl.ParseFiles(files...)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing template:")
	}

	return tmpl
}

func InitTemplates() {
	IndexTmpl = createTemplate([]string{
		fmt.Sprintf("%v/templates/base.html", os.Getenv("CONTENT_PATH")),
		fmt.Sprintf("%v/templates/index.html", os.Getenv("CONTENT_PATH")),
	})
	TagTmpl = createTemplate([]string{
		fmt.Sprintf("%v/templates/base.html", os.Getenv("CONTENT_PATH")),
		fmt.Sprintf("%v/templates/tag.html", os.Getenv("CONTENT_PATH")),
	})
	PostTmpl = createTemplate([]string{
		fmt.Sprintf("%v/templates/base.html", os.Getenv("CONTENT_PATH")),
		fmt.Sprintf("%v/templates/post.html", os.Getenv("CONTENT_PATH")),
	})
	AboutTmpl = createTemplate([]string{
		fmt.Sprintf("%v/templates/base.html", os.Getenv("CONTENT_PATH")),
		fmt.Sprintf("%v/templates/about.html", os.Getenv("CONTENT_PATH")),
	})
}
