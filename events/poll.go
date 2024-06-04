package events

import (
	"github.ibm.com/open-z/jeff-ci/events/handlers"
	"github.ibm.com/open-z/jeff-ci/gh"
	"log"
	"time"
)

func Poll(earliestTime, now time.Time) time.Time {
	events, err := gh.GetRepositoryEvents()
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

	return now
}
