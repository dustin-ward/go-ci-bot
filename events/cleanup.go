package events

import (
	"context"
	"fmt"
	"log"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

func Cleanup(apiClient *github.Client) error {
	events, _, err := apiClient.Activity.ListRepositoryEvents(context.TODO(), config.Owner(), config.Repo(), nil)
	if err != nil {
		log.Fatal("ListRepositoryEvents: ", err)
	}

	for _, e := range events {
		payload, err := e.ParsePayload()
		if err != nil {
			fmt.Println("Unable to parse event:", e.GetID())
			continue
		}

		switch p := payload.(type) {
		case *github.PullRequestEvent:
			pr := p.GetPullRequest()
			log.Printf("Pull Request Event: %s [#%d] - %s (%s) %s\n",
				p.GetAction(),
				pr.GetNumber(),
				pr.GetTitle(),
				pr.GetUser().GetLogin(),
				pr.GetState(),
			)

			appId := config.AppID()
			res, _, err := apiClient.Checks.ListCheckRunsForRef(
				context.TODO(),
				config.Owner(),
				config.Repo(),
				pr.GetHead().GetRef(),
				&github.ListCheckRunsOptions{
					AppID: &appId,
				},
			)
			if err != nil {
				return err
			}

			for _, checkRun := range res.CheckRuns {
				log.Printf("Found unfinished check run: %s - %s[%d] %s %s\n",
					checkRun.GetName(),
					checkRun.GetHeadSHA(),
					checkRun.GetID(),
					checkRun.GetStatus(),
					checkRun.GetStartedAt(),
				)

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
						Name:        checkRun.GetName(),
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
		case *github.PushEvent:
		default:
			log.Printf("Unhandled stale event: %s[%s] - %s (%s)\n",
				e.GetType(),
				e.GetID(),
				e.GetActor().GetLogin(),
				e.GetCreatedAt(),
			)
		}
	}

	return nil
}
