package handlers

import (
	"context"
	"log"
	"regexp"
	"test-org-gozbot/buildqueue"
	"test-org-gozbot/config"

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

		ok, err := buildqueue.Push(apiClient, pr.GetNumber(), pr.GetHead().GetSHA(), event.GetComment().GetUser().GetLogin())
		if err != nil {
			return err
		}
		if ok {
			log.Println("Added to build queue")
		} else {
			log.Println("Not added to build queue")
		}
	}

	return nil
}
