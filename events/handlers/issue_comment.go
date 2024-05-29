package handlers

import (
	"log"
	"regexp"
	"test-org-gozbot/gh"
	"test-org-gozbot/tasks"

	"github.com/google/go-github/v62/github"
)

var (
	reSpinBuild = regexp.MustCompile("(?i)re[- ]*(spin|test|build|try)")
)

func handleIssueCommentEvent(event *github.IssueCommentEvent) error {
	if event.GetIssue().IsPullRequest() && reSpinBuild.MatchString(event.GetComment().GetBody()) {
		return respinBuild(event)
	}

	return nil
}

func respinBuild(event *github.IssueCommentEvent) error {
	pr, err := gh.GetPullRequest(event.GetIssue().GetNumber())
	if err != nil {
		return err
	}

	log.Printf("Re-Spinning build for PR#%d (%s/%s) - triggered by %s\n",
		pr.GetNumber(),
		pr.GetHead().GetRef(),
		pr.GetHead().GetSHA()[:6],
		event.GetComment().GetUser().GetLogin(),
	)

	ok, err := tasks.PushBuild(
		pr.GetNumber(),
		pr.GetHead().GetSHA(),
		event.GetComment().GetUser().GetLogin(),
	)
	if err != nil {
		return err
	}

	if ok {
		log.Println("Added to build queue")
	} else {
		log.Println("Not added to build queue")
	}

	return nil
}
