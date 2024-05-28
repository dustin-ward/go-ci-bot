package gh

import (
	"context"
	"test-org-gozbot/config"
    "test-org-gozbot/gh/auth"

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
	if err != nil {
		return []*github.Event{}, err
	}

	return events, nil
}
