package handlers

import (
	"log"
	"regexp"

	"github.ibm.com/open-z/jeff-ci/tasks"

	"github.com/google/go-github/v62/github"
)

func handlePullRequestEvent(event *github.PullRequestEvent) error {
	action := event.GetAction()
	if action == "opened" || action == "reopened" {
		return registerNewPR(event)
	}

	if action == "closed" {
		return triggerPackageRelease(event)
	}

	return nil
}

func registerNewPR(event *github.PullRequestEvent) error {
	pr := event.GetPullRequest()
	log.Printf("Pull request opened: #%d %s - %s (%s)\n",
		pr.GetNumber(),
		pr.GetTitle(),
		pr.GetUser().GetLogin(),
		pr.GetHead().GetSHA()[:6],
	)

	//TODO: Redundant code here and in pull_request.go?
	tasks.Build{
		PR:          pr.GetNumber(),
		SHA:         pr.GetHead().GetSHA(),
		BaseBranch:  pr.GetBase().GetRef(),
		HeadBranch:  pr.GetHead().GetRef(),
		SubmittedBy: pr.GetHead().GetUser().GetLogin(),
	}.Enqueue()

	return nil
}

func triggerPackageRelease(event *github.PullRequestEvent) error {
	pr := event.GetPullRequest()

	// See if the PR was actually merged or just closed
	if pr == nil || !pr.GetMerged() {
		return nil
	}

	re := regexp.MustCompile(`^release-branch\.go\d+\.\d+-zos$`)
	baseRefName := pr.GetBase().GetRef()
	if re.MatchString(baseRefName) {
		// Create build with base == head == release-branch.go1.xx.x-zos
		tasks.Build{
			PR:          pr.GetNumber(),
			SHA:         pr.GetBase().GetSHA(),
			BaseBranch:  pr.GetBase().GetRef(),
			HeadBranch:  pr.GetBase().GetRef(),
			SubmittedBy: pr.GetMergedBy().GetLogin(),
		}.Enqueue()
	}

	return nil
}
