package handlers

import (
	"fmt"
	"log"

	"github.com/google/go-github/v62/github"
)

func Handle(e *github.Event) error {
	payload, err := e.ParsePayload()
	if err != nil {
		return fmt.Errorf("Unable to parse event: %d %v", e.GetID(), err)
	}

	log.Printf("Event Received: %s[%s] - %s (%s)\n",
		e.GetType(),
		e.GetID(),
		e.GetActor().GetLogin(),
		e.GetCreatedAt(),
	)

	switch p := payload.(type) {
	case *github.PullRequestEvent:
		if err := handlePullRequestEvent(p); err != nil {
			return fmt.Errorf("handlePullRequestEvent: %v", err)
		}
	case *github.PushEvent:
		if err := handlePushEvent(p); err != nil {
			return fmt.Errorf("handlePushEvent: %v", err)
		}
	case *github.IssueCommentEvent:
		if err := handleIssueCommentEvent(p); err != nil {
			return fmt.Errorf("handleIssueCommentEvent: %v", err)
		}
	default:
		log.Printf("Unhandled event type: %T[%s]\n", p, e.GetID())
	}

	return nil
}
