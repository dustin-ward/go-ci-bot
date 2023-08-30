package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cbrgm/githubevents/githubevents"
    "test-org-gozbot/handlers"
)

func main() {
	webHookSecret := os.Getenv("GOZBOT_WEBHOOK_SECRET")
	if webHookSecret == "" {
		log.Fatal("ERROR: GOZBOT_WEBHOOK_SECRET not set")
	}

	handler := githubevents.New(webHookSecret)

	// Handler functions for each category of event
	// handler.OnIssuesEventAny()
	handler.OnCheckSuiteEventAny(handlers.CheckSuiteEventHandler)
	handler.OnCheckRunEventAny(handlers.CheckRunEventHandler)

	http.HandleFunc("/hook", func(w http.ResponseWriter, r *http.Request) {
		err := handler.HandleEventRequest(r)
		if err != nil {
			if err.Error() != "unknown X-Github-Event in message: security_advisory" {
				log.Println("http error:", err)
			}
		}
	})

	// Start server
	log.Println("Listening on 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
