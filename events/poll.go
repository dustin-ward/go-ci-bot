package events

import (
	"log"

	"github.ibm.com/open-z/jeff-ci/events/handlers"
	"github.ibm.com/open-z/jeff-ci/gh"
)

func GetMostRecentEventID() (id string) {
	events, err := gh.GetRepositoryEvents()
	if err != nil {
		log.Fatal("GetMostRecentEventID: ", err)
	}

	if len(events) != 0 {
		id = events[0].GetID()
	}

	return
}

func Poll(lastEventID string) (newestEventID string) {
	events, err := gh.GetRepositoryEvents()
	if err != nil {
		log.Fatal("ListRepositoryEvents: ", err)
	}

	for _, e := range events {
		if newestEventID == "" {
			newestEventID = e.GetID()
		}

		if e.GetID() == lastEventID {
			// Dont consider any events older than the last poll
			break
		}

		err = handlers.Handle(e)
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}
