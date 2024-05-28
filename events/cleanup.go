package events

import (
	"log"
	"test-org-gozbot/checks"
	"test-org-gozbot/gh"
)

func Cleanup() error {
	branches, err := gh.GetBranches()
	if err != nil {
		return err
	}

	for _, branch := range branches {
		checkRuns, err := gh.GetCheckRunsForRef(branch.GetName())
		if err != nil {
			return err
		}

		for _, checkRun := range checkRuns {
			if checkRun.GetStatus() == "completed" {
				continue
			}

			summary := "Check Run Cancelled"
			msg := "Stale Check Run - GOZBOT has been terminated since this run has been queued"
			_, err := gh.CompleteCheckRun(checkRun, checks.CONCLUSION_CANCELLED, summary, msg)
			if err != nil {
				return err
			}

			log.Println("Cancelled check run:", checkRun.GetID())
		}
	}

	return nil
}
