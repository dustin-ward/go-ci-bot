package gh

import (
	"context"
	"github.ibm.com/open-z/jeff-ci/config"
	"github.ibm.com/open-z/jeff-ci/gh/auth"

	"github.com/google/go-github/v62/github"
)

func GetRepositoryEvents() ([]*github.Event, error) {
	client, err := auth.GetClient()
	if err != nil {
		return []*github.Event{}, err
	}

	events, _, err := client.Activity.ListRepositoryEvents(
		context.TODO(),
		config.Owner(),
		config.Repo(),
		nil,
	)

	return events, err
}
