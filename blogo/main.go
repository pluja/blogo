package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	dev := flag.Bool("dev", false, "sets dev mode")
	new := flag.String("new", "", "Creates a new post file in articles/ with the name specified after the flag. Example: -new my-post")
	path := flag.String("path", "", "Sets the path to the content folder. Example: -path /home/user/my-blog/articles")
	nkeys := flag.Bool("nkeys", false, "Generates a new nostr key set.")
	port := flag.Int("port", 3000, "Sets the port to run the server on. Example: -port 3000")
	flag.Parse()

	if *nkeys {
		_, _, nsec, npub, err := GetNewKeySet()
		if err != nil {
			log.Error().Err(err).Msg("Error generating new key set:")
			os.Exit(1)
		}
		log.Info().Msgf("Generated new Nostr key set:\n\nnsec: %v\nnpub: %v\n\n", nsec, npub)
		os.Exit(0)
	}

	if *dev {
		os.Setenv("DEV", "true")
		//os.Setenv("CONTENT_PATH", "..")
		os.Setenv("DEV", "true")
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: "15:04:05",
			},
		).With().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if *path != "" {
		log.Info().Msgf("Using path: %v", *path)
		os.Setenv("CONTENT_PATH", *path)
	} else {
		if os.Getenv("CONTENT_PATH") == "" {
			log.Warn().Msg("No path specified, using default path: ./")
			os.Setenv("CONTENT_PATH", ".")
		}
	}

	if strings.HasSuffix(os.Getenv("CONTENT_PATH"), "/") {
		os.Setenv("CONTENT_PATH", os.Getenv("CONTENT_PATH")[:len(*path)-1])
	}

	if *new != "" {
		CreateArticleTemplate(*new)
		log.Info().Msgf("New template created: articles/%v.md", *new)
		os.Exit(0)
	}

	// Load .env file
	err := godotenv.Load(fmt.Sprintf("%v/.env", os.Getenv("CONTENT_PATH")))
	if err != nil {
		log.Warn().Msg("No .env file found, using default settings or environment variables.")
	}

	InitSettings()
	InitBadger()
	//InitRedis()
	InitTemplates()
	r := InitRoutes()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch},
	})

	handler := c.Handler(r)

	if os.Getenv("PUBLISH_TO_NOSTR") == "true" {
		err = InitNostr()
		if err != nil {
			log.Error().Err(err).Msg("Error initializing nostr:")
		}
	}

	err = LoadArticles()
	if err != nil {
		log.Error().Err(err).Msg("Error loading articles metadata:")
	}

	go InitWatcher()

	log.Info().Msgf("Starting server on port %v...", *port)
	http.ListenAndServe(fmt.Sprintf(":%v", *port), handler)
}

func InitSettings() {
	if os.Getenv("BLOGO_DESCRIPTION") != "false" {
		if os.Getenv("BLOGO_DESCRIPTION") == "" {
			Blogo.Description = "Welcome to my Blogo 🎈"
		} else {
			Blogo.Description = os.Getenv("BLOGO_DESCRIPTION")
		}
	}

	if os.Getenv("BLOGO_TITLE") != "" {
		Blogo.Title = os.Getenv("BLOGO_TITLE")
	} else {
		Blogo.Title = "Blogo"
	}

	if os.Getenv("BLOGO_URL") != "" {
		Blogo.Url = os.Getenv("BLOGO_URL")
		Blogo.Url = strings.TrimSuffix(Blogo.Url, "/")
	} else {
		Blogo.Url = "http://localhost:3000"
	}

	if os.Getenv("BLOGO_KEYWORDS") != "" {
		Blogo.Keywords = os.Getenv("BLOGO_KEYWORDS")
	} else {
		Blogo.Keywords = "blog, blogo"
	}

	if os.Getenv("BLOGO_ANALYTICS") != "" {
		Blogo.Analytics = os.Getenv("BLOGO_ANALYTICS")
	}

	LogSettings()
}

func LogSettings() {
	log.Info().Msg("Loaded settings:")
	log.Info().Msgf("\t~ Title: %v", Blogo.Title)
	log.Info().Msgf("\t~ Description: %v", Blogo.Description)
	log.Info().Msgf("\t~ Url: %v", Blogo.Url)
	log.Info().Msgf("\t~ Keywords: %v", Blogo.Keywords)
	if Blogo.Analytics != "" {
		log.Info().Msgf("\t~ Analytics: yes\n")
	}
}
