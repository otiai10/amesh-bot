package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/logging"
	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/commands"
	"github.com/otiai10/amesh-bot/controllers"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/goapis/google"
	"github.com/otiai10/marmoset"
)

var (
	timezone *time.Location
)

func init() {
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}
	timezone = tokyo
}

func main() {

	r := marmoset.NewRouter()
	g := &google.Client{
		APIKey:               os.Getenv("GOOGLE_CUSTOMSEARCH_API_KEY"),
		CustomSearchEngineID: os.Getenv("GOOGLE_CUSTOMSEARCH_ENGINE_ID"),
	}

	lg, err := logging.NewClient(context.Background(), os.Getenv("GOOGLE_PROJECT_ID"))
	if err != nil {
		panic(err)
	}
	defer lg.Close()

	b := &bot.Bot{
		Commands: []bot.Command{
			commands.ImageCommand{Search: g},
			commands.ForecastCommand{
				SourceURL: "https://www.jma.go.jp/bosai/forecast", Timezone: timezone,
			},
			commands.TyphoonCommand{},
			commands.GoogleCommand{Search: g},
			commands.LGTMCommand{Service: service.LGTM{}},
			commands.EchoCommand{},
		},
		Default: commands.AmeshCommand{
			Storage:  &service.Cloudstorage{BaseURL: "https://storage.googleapis.com"},
			Timezone: timezone,
		},
		NotFound: commands.NotFound{},
		Logger:   lg.Logger("bot"),
	}
	c := controllers.Controller{
		Bot:       b,
		Slack:     &service.SlackOAuthClient{},
		Datastore: service.NewDatastore(os.Getenv("GOOGLE_PROJECT_ID")),
		Storage:   &service.Cloudstorage{BaseURL: "https://storage.googleapis.com"},
	}
	r.POST("/slack/webhook", c.Webhook)
	r.GET("/slack/oauth", c.OAuth)

	// 画像フィルタリング
	r.GET("/image", c.Image)

	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
