package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/logging"
	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/commands"
	"github.com/otiai10/amesh-bot/controllers"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/goapis/google"
	"github.com/otiai10/marmoset"
)

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
			commands.ForecastCommand{SourceURL: "https://www.jma.go.jp/bosai/forecast"},
			commands.GoogleCommand{Search: g},
			commands.LGTMCommand{Service: service.LGTM{}},
		},
		Default:  commands.AmeshCommand{Storage: &service.Cloudstorage{BaseURL: "https://storage.googleapis.com"}},
		NotFound: commands.NotFound{},
		Logger:   lg.Logger("bot"),
	}
	c := controllers.Controller{
		Bot:       b,
		Slack:     &service.SlackOAuthClient{},
		Datastore: service.NewDatastore(os.Getenv("GOOGLE_PROJECT_ID")),
	}
	r.POST("/slack/webhook", c.Webhook)
	r.GET("/slack/oauth", c.OAuth)
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
