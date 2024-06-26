package handlers

import (
	"log"
	"regexp"

	"github.ibm.com/open-z/jeff-ci/gh"
	"github.ibm.com/open-z/jeff-ci/tasks"

	"github.com/google/go-github/v62/github"
)

var (
	//TODO: Make a more modular system to allow for renaming the bot
	reSpinBuild = regexp.MustCompile("(?i)@jeffci re[- ]*(spin|test|build|try)")
	exampleTask = regexp.MustCompile("(?i)@jeffci example")
)

func handleIssueCommentEvent(event *github.IssueCommentEvent) error {
	if event.GetIssue().IsPullRequest() {
		body := event.GetComment().GetBody()
		if reSpinBuild.MatchString(body) {
			return respinBuild(event)
		}
		if exampleTask.MatchString(body) {
			return startExampleTask(event)
		}
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

	//TODO: Redundant code here and in pull_request.go?
	tasks.Build{
		PR:          pr.GetNumber(),
		SHA:         pr.GetHead().GetSHA(),
		BaseBranch:  pr.GetBase().GetRef(),
		HeadBranch:  pr.GetHead().GetRef(),
		SubmittedBy: event.GetComment().GetUser().GetLogin(),
	}.Enqueue()

	return nil
}

func startExampleTask(event *github.IssueCommentEvent) error {
	pr, err := gh.GetPullRequest(event.GetIssue().GetNumber())
	if err != nil {
		return err
	}

	log.Printf("Starting Example Task for PR#%d (%s/%s) - triggered by %s\n",
		pr.GetNumber(),
		pr.GetHead().GetRef(),
		pr.GetHead().GetSHA()[:6],
		event.GetComment().GetUser().GetLogin(),
	)

	tasks.ExampleTask{
		SHA:         pr.GetHead().GetSHA(),
		SubmittedBy: event.GetComment().GetUser().GetLogin(),
	}.Enqueue()

	return nil
}
