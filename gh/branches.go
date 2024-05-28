package gh

import (
	"context"
	"test-org-gozbot/config"
    "test-org-gozbot/gh/auth"

	"github.com/google/go-github/v62/github"
)

func GetBranches() ([]*github.Branch, error) {
	client, err := auth.GetClient()
	if err != nil {
		return []*github.Branch{}, err
	}

	branches, _, err := client.Repositories.ListBranches(
		context.TODO(),
		config.Owner(),
		config.Repo(),
		nil,
	)
	if err != nil {
		return []*github.Branch{}, err
	}

	return branches, nil
}
