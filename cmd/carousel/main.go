package main

import (
	"context"
	"log"
	"net/http"

	"github.com/charconstpointer/carousel/slack"
)

func main() {
	tracker := slack.NewTracker()
	slackHandler := slack.NewHandler("xoxb-3713913975109-3702272114647-UD9vftiBV0xrw9u5gCqxNA6Y")
	slackServer := slack.NewServer(tracker)
	http.HandleFunc("/", slackServer.HandleReadinessResponse())
	go func() {
		cid, err := slackHandler.RequestUserReadiness(context.Background(), "U03LN6Z60PR")
		if err != nil {
			log.Fatalf("could not request user readiness: %v", err)
		}
		res, err := tracker.Track(cid.CallbackID)
		if err != nil {
			log.Fatalf("could not track request: %v", err)
		}

		log.Println("user responded", <-res)
	}()
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start http server: %v", err)
	}
}
