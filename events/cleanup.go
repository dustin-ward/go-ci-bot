package events

import (
	"github.ibm.com/open-z/jeff-ci/gh"
	"log"
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
			msg := "Stale Check Run - JeffCI has been terminated since this run has been queued"
			_, err := gh.CompleteCheckRun(checkRun, gh.CHECK_CONCLUSION_CANCELLED, summary, msg)
			if err != nil {
				return err
			}

			log.Println("Cancelled check run:", checkRun.GetID())
		}
	}

	return nil
}
