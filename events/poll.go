package events

import (
	"log"
	"test-org-gozbot/events/handlers"
	"test-org-gozbot/gh"
	"time"
)

func Poll(earliestTime time.Time) time.Time {
	events, err := gh.GetRepositoryEvents()
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
		if createdAt.Before(earliestTime) {
			// Dont consider any events older than the last poll
			break
		}

		err = handlers.Handle(e)
		if err != nil {
			log.Fatal(err)
		}
	}

	return pollTime
}
