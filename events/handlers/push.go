package handlers

import (
	"fmt"
	"log"

	"github.ibm.com/open-z/jeff-ci/gh"
	"github.ibm.com/open-z/jeff-ci/tasks"

	"github.com/google/go-github/v62/github"
)

func handlePushEvent(event *github.PushEvent) error {
	prList, err := gh.GetPullRequests()
	if err != nil {
		return err
	}

	// I dont exactly remeber why I need to match the event to its PR struct...
	// but I think its because we need the PR#
	for _, pr := range prList {
		if "refs/heads/"+pr.GetHead().GetRef() == event.GetRef() {
			return triggerNewBuild(event, pr)
		}
	}

	return nil
}

func triggerNewBuild(event *github.PushEvent, pr *github.PullRequest) error {
	if event.GetSize() == 0 {
		// Sanity check
		return fmt.Errorf("no commits in push event %s/%s", event.GetRef(), event.GetHead())
	}

	headCommit := event.Commits[event.GetSize()-1]
	log.Printf("Push event in PR#%d %s - %s (%s)\n",
		pr.GetNumber(),
		pr.GetTitle(),
		headCommit.GetAuthor().GetLogin(),
		headCommit.GetSHA()[:6],
	)

	//TODO: Redundant code here and in pull_request.go?
	tasks.Build{
		PR:          pr.GetNumber(),
		SHA:         headCommit.GetSHA(),
		BaseBranch:  pr.GetBase().GetRef(),
		HeadBranch:  pr.GetHead().GetRef(),
		SubmittedBy: headCommit.GetAuthor().GetLogin(),
	}.Enqueue()

	return nil
}
