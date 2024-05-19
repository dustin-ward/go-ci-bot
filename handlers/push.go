package handlers

import (
	"context"
	"fmt"
	"log"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

func HandlePushEvent(apiClient *github.Client, event *github.PushEvent) error {
	prList, _, err := apiClient.PullRequests.List(context.TODO(), config.Owner(), config.Repo(), nil)
	if err != nil {
		return err
	}

	for _, pr := range prList {
		if "refs/heads/"+pr.GetHead().GetRef() == event.GetRef() {
			if event.GetSize() == 0 {
				// Sanity check
				return fmt.Errorf("no commits in push event %s/%s", event.GetRef(), event.GetHead())
			}

			headCommit := event.Commits[event.GetSize()-1]
			log.Printf("Push event in PR#%d %s - %s (%s)\n",
				pr.GetNumber(),
				pr.GetTitle(),
				headCommit.GetAuthor().GetLogin(),
				headCommit.GetSHA(),
			)
            log.Println("Adding to build queue...")

			checkRun, _, err := apiClient.Checks.CreateCheckRun(context.TODO(), config.Owner(), config.Repo(),
				github.CreateCheckRunOptions{
					Name:      "z/OS Build & Test",
					HeadSHA:   headCommit.GetSHA(),
					Status:    &checks.STATUS_QUEUED,
					StartedAt: &github.Timestamp{time.Now()},
				},
			)
			if err != nil {
				return err
			}

			_ = checkRun
		}
	}

	return nil

	// title := "TEST RUN"
	// summary := "This is the summary"
	// body := "Body text here... The run is now in progress"
	// checkRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
	// 	checkRun.GetID(),
	// 	github.UpdateCheckRunOptions{
	// 		Name:   "z/OS Build",
	// 		Status: &checks.STATUS_IN_PROGRESS,
	// Output: &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatal("Update Check Run 1: ", err)
	// }

	// time.Sleep(15 * time.Second)

	// body += "\n\nRun Completed"
	// checkRun, _, err = apiClient.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
	// 	checkRun.GetID(),
	// 	github.UpdateCheckRunOptions{
	// 		Name:        "z/OS Build",
	// 		Status:      &checks.STATUS_COMPLETED,
	// 		Conclusion:  &checks.CONCLUSION_FAILURE,
	// 		CompletedAt: &github.Timestamp{time.Now()},
	// Output: &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
	// 	},
	// )
	// if err != nil {
	// 	log.Fatal("Update Check Run 2: ", err)
	// }

	// return nil
}
