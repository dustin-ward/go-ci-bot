package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"test-org-gozbot/auth"
	"test-org-gozbot/buildqueue"
	"test-org-gozbot/config"
	"test-org-gozbot/handlers"

	"github.com/google/go-github/v62/github"
)

func main() {
	config.NewConfig(
		"https://github.ibm.com",
		"test-org-gozbot",
		"goz-workflow-demo",
		"./private-key.pem",
		2206,
		19342,
	)

	apiClient, err := auth.CreateClient()
	if err != nil {
		log.Fatal("Auth: ", err)
	}

	// Create build queue
	buildqueue.Init(5)

	// Poll for events
	events, _, err := apiClient.Activity.ListRepositoryEvents(context.TODO(), config.Owner(), config.Repo(), nil)
	if err != nil {
		log.Fatal("ListRepositoryEvents: ", err)
	}
	if len(events) == 0 {
		fmt.Println("No Events")
		os.Exit(0)
	}

	for _, e := range events {
		payload, err := e.ParsePayload()
		if err != nil {
			fmt.Println("Unable to parse event:", e.GetID())
			continue
		}
		switch t := payload.(type) {
		case *github.PullRequestEvent:
			if err := handlers.HandlePullRequestEvent(apiClient, payload.(*github.PullRequestEvent)); err != nil {
				log.Fatal("HandlePullRequestEvent: ", err)
			}
		case *github.PushEvent:
			if err := handlers.HandlePushEvent(apiClient, payload.(*github.PushEvent)); err != nil {
				log.Fatal("HandlePushEvent: ", err)
			}
		default:
			fmt.Printf("Unhandled event type: %T[%s] - %s (%s)\n",
				t,
				e.GetID(),
				e.GetActor().GetLogin(),
				e.GetCreatedAt(),
			)
		}
	}
}
