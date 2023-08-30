package handlers

import (
	"log"

	"github.com/google/go-github/v50/github"
)

func CheckSuiteEventHandler(deliveryID string, eventName string, event *github.CheckSuiteEvent) error {
	log.Println("CHECK SUITE EVENT:", *event.Action)

	switch *event.Action {
	case "requested", "rerequested":
		log.Println("Creating CheckRun...")
        log.Printf("%s/%s (%s)\n",
            event.GetRepo().GetOwner().GetLogin(),
            event.GetRepo().GetName(),
            event.GetCheckSuite().GetHeadSHA(),
        )
		err := createCheckRun(event, "GOZ Build")
		if err != nil {
			return err
		}

	default:
		log.Println("Unhandled action:", *event.Action)
	}
	return nil
}
