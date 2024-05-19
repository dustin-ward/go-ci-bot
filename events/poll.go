package events

import (
	"context"
	"fmt"
	"log"
	"os"
	"test-org-gozbot/config"
	"test-org-gozbot/handlers"
	"time"

	"github.com/google/go-github/v62/github"
)

func Poll(apiClient *github.Client, ticker *time.Ticker, lastPollTime time.Time, stopPoll chan struct{}, stopMain chan os.Signal) {
	for {
		select {
		case <-ticker.C:
			log.Println("tick")
			events, _, err := apiClient.Activity.ListRepositoryEvents(context.TODO(), config.Owner(), config.Repo(), nil)
			pollTime := time.Now()
			if err != nil {
				log.Fatal("ListRepositoryEvents: ", err)
			}

			for _, e := range events {
				createdAt := e.GetCreatedAt()
				if createdAt.IsZero() {
					// Skip events with no timestamp (?)
					continue
				}
				if createdAt.Before(lastPollTime) {
					// Dont consider any events older than the last poll
					break
				}

				payload, err := e.ParsePayload()
				if err != nil {
					fmt.Println("Unable to parse event:", e.GetID())
					continue
				}
				log.Printf("Event Received: %s[%s] - %s (%s)\n",
					e.GetType(),
					e.GetID(),
					e.GetActor().GetLogin(),
					e.GetCreatedAt(),
				)

				switch p := payload.(type) {
				case *github.PushEvent:
					if err := handlers.HandlePushEvent(apiClient, p); err != nil {
						log.Fatal("HandlePushEvent: ", err)
					}
				case *github.IssueCommentEvent:
					if err := handlers.HandleIssueCommentEvent(apiClient, p); err != nil {
						log.Fatal("HandleIssueCommentEvent: ", err)
					}
				default:
					log.Printf("Unhandled event type: %T[%s]\n",
						p,
						e.GetID(),
					)
				}
			}
			lastPollTime = pollTime
		case <-stopPoll:
			// End Program
			close(stopMain)
			return
		}
	}
}
