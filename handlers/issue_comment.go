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
	// fmt.Printf("Issue Comment Event: %s\n", event.GetAction())
	// fmt.Println("  #:", event.GetIssue().GetNumber())
	// fmt.Println("  State:", event.GetIssue().GetState())
	// fmt.Println("  Title:", event.GetIssue().GetTitle())
	// fmt.Println("  IsPR:", event.GetIssue().IsPullRequest())
	// fmt.Println("  Author:", event.GetComment().GetUser().GetLogin())
	// fmt.Println("  Body:", event.GetComment().GetBody())

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

		checkRun, _, err := apiClient.Checks.CreateCheckRun(context.TODO(), config.Owner(), config.Repo(),
			github.CreateCheckRunOptions{
				Name:      "z/OS Build & Test",
				HeadSHA:   pr.GetHead().GetSHA(),
				Status:    &checks.STATUS_QUEUED,
				StartedAt: &github.Timestamp{time.Now()},
			},
		)
		if err != nil {
			return err
		}

		_ = checkRun
	}

	return nil
}
