package gh

import (
	"context"
	"github.ibm.com/open-z/jeff-ci/config"
	"github.ibm.com/open-z/jeff-ci/gh/auth"

	"github.com/google/go-github/v62/github"
)

func GetPullRequest(number int) (*github.PullRequest, error) {
	client, err := auth.GetClient()
	if err != nil {
		return nil, err
	}

	pr, _, err := client.PullRequests.Get(
		context.TODO(),
		config.Owner(),
		config.Repo(),
		number,
	)

	return pr, err
}

func GetPullRequests() ([]*github.PullRequest, error) {
	client, err := auth.GetClient()
	if err != nil {
		return []*github.PullRequest{}, err
	}

	prList, _, err := client.PullRequests.List(
		context.TODO(),
		config.Owner(),
		config.Repo(),
		nil,
	)

	return prList, err
}
