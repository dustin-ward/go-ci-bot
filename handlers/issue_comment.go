package handlers

import (
	"context"
	"log"
	"regexp"
	"test-org-gozbot/checks"
	"test-org-gozbot/config"
	"time"

	"github.com/google/go-github/v62/github"
)

var (
	reSpinBuild = regexp.MustCompile("(?i)re[- ]*(spin|test|build)")
)

func HandleIssueCommentEvent(apiClient *github.Client, event *github.IssueCommentEvent) error {
	if event.GetIssue().IsPullRequest() && reSpinBuild.MatchString(event.GetComment().GetBody()) {
		pr, _, err := apiClient.PullRequests.Get(context.TODO(), config.Owner(), config.Repo(), event.GetIssue().GetNumber())
		if err != nil {
			return err
		}

		log.Printf("Re-Spinning build for PR#%d (%s/%s) - triggered by %s\n",
			pr.GetNumber(),
			pr.GetHead().GetRef(),
			pr.GetHead().GetSHA()[:6],
			event.GetComment().GetUser().GetLogin(),
		)

		title := "z/OS Build & Test"
		summary := "In Queue"
		msg := "This commit has been added to the build queue. More information will appear here once the build has started"
		checkRun, _, err := apiClient.Checks.CreateCheckRun(context.TODO(), config.Owner(), config.Repo(),
			github.CreateCheckRunOptions{
				Name:      title,
				HeadSHA:   pr.GetHead().GetSHA(),
				Status:    &checks.STATUS_QUEUED,
				StartedAt: &github.Timestamp{time.Now()},
				Output:    &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &msg},
			},
		)
		if err != nil {
			return err
		}

		_ = checkRun
	}

	return nil
}
