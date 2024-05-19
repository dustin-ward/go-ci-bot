package buildqueue

import (
	"context"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

type Build struct {
    PR          int
	SHA         string
	SubmittedBy string
}

func Push(apiClient *github.Client, PR int, SHA, SubmittedBy string) (bool, error) {
    title := "z/OS Build & Test"
    summary := "In Queue"
    msg := "This commit has been added to the build queue. More information will appear here once the build has started"
    checkRun, _, err := apiClient.Checks.CreateCheckRun(context.TODO(), config.Owner(), config.Repo(),
        github.CreateCheckRunOptions{
            Name:      title,
            HeadSHA:   SHA,
            Status:    &checks.STATUS_QUEUED,
            StartedAt: &github.Timestamp{time.Now()},
            Output:    &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &msg},
        },
    )
	if err != nil {
		return false, err
	}

	_ = checkRun

	return true, nil
}
