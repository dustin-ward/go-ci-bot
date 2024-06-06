package gh

import (
	"context"
	"github.ibm.com/open-z/jeff-ci/config"
	"github.ibm.com/open-z/jeff-ci/gh/auth"

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

	return branches, err
}

func GetBranch(branchName string) (*github.Branch, error) {
	client, err := auth.GetClient()
	if err != nil {
		return nil, err
	}

	branch, _, err := client.Repositories.GetBranch(
		context.TODO(),
		config.Owner(),
		config.Repo(),
		branchName,
		0,
	)

	return branch, err
}
