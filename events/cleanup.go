package events

import (
	"context"
	"log"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

func Cleanup(apiClient *github.Client) error {
	branches, _, err := apiClient.Repositories.ListBranches(context.TODO(), config.Owner(), config.Repo(), nil)
	if err != nil {
		return err
	}

	for _, branch := range branches {
		appId := config.AppID()
		res, _, err := apiClient.Checks.ListCheckRunsForRef(
			context.TODO(),
			config.Owner(),
			config.Repo(),
			branch.GetName(),
			&github.ListCheckRunsOptions{
				AppID: &appId,
			},
		)
		if err != nil {
			return err
		}

		for _, checkRun := range res.CheckRuns {
			if checkRun.GetStatus() == "completed" {
				continue
			}

			title := "z/OS Build & Test"
			summary := "Check Run Cancelled"
			msg := "Stale Check Run - GOZBOT has been terminated since this run has been queued"
			_, _, err := apiClient.Checks.UpdateCheckRun(
				context.TODO(),
				config.Owner(),
				config.Repo(),
				checkRun.GetID(),
				github.UpdateCheckRunOptions{
					Name:        title,
					Status:      &checks.STATUS_COMPLETED,
					Conclusion:  &checks.CONCLUSION_CANCELLED,
					CompletedAt: &github.Timestamp{time.Now()},
					Output:      &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &msg},
				},
			)
			if err != nil {
				return err
			}

			log.Println("Cancelled check run:", checkRun.GetID())
		}
	}

	return nil
}
