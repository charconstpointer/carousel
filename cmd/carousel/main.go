package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/charconstpointer/carousel/slack"
)

var (
	slackToken = flag.String("slack-token", "", "slack token")
	testUserID = flag.String("test-user-id", "U03LN6Z60PR", "test user id")
)

func main() {
	flag.Parse()
	tracker := slack.NewTracker()
	slackHandler := slack.NewHandler(*slackToken)
	slackServer := slack.NewServer(tracker)
	http.HandleFunc("/", slackServer.HandleReadinessResponse())
	go func() {
		cid, err := slackHandler.RequestUserReadiness(context.Background(), *testUserID)
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
