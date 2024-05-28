package gh

import (
	"context"
	"test-org-gozbot/config"
	"test-org-gozbot/gh/auth"

	"github.com/google/go-github/v62/github"
)

func CreateComment(PR int, body string) (*github.IssueComment, error) {
	client, err := auth.GetClient()
	if err != nil {
		return nil, err
	}

	comment, _, err := client.Issues.CreateComment(
		context.TODO(),
		config.Owner(),
		config.Repo(),
		PR,
		&github.IssueComment{Body: &body},
	)

	return comment, err
}
