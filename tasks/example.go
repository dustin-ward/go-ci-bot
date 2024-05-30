package tasks

import (
	"fmt"
	"log"
	"os/exec"
	"test-org-gozbot/gh"
	"time"

	"github.com/google/go-github/v62/github"
)

type ExampleTask struct {
	SHA         string
	SubmittedBy string
	CheckRun    *github.CheckRun
}

var (
	exampleTitle             = "Example Task"
	exampleSummaryInQueue    = "Example - In Queue"
	exampleSummaryInProgress = "Example - In Progress"
	exampleSummaryCompleted  = "Example - Completed"
)

func (e ExampleTask) Do() error {
	body := `This is an example task created by GOZBOT. See 'tasks/example.go' in the source for more details
When the 'random-exit-code' task has run, we will see the output here:`
	checkRun, err := gh.UpdateCheckRun(e.CheckRun, exampleSummaryInProgress, body)
	if err != nil {
		return err
	}

	output, ok := e.exampleWork()
	body += "\n```" + output + "\n```"

	var conclusion string
	if ok {
		conclusion = gh.CHECK_CONCLUSION_SUCCESS
	} else {
		conclusion = gh.CHECK_CONCLUSION_NEUTRAL
	}

	_, err = gh.CompleteCheckRun(checkRun, conclusion, exampleSummaryCompleted, body)
	if err != nil {
		return err
	}

	return nil
}

func (e *ExampleTask) exampleWork() (output string, ok bool) {
	ok = true

	time.Sleep(time.Second * 5)
	cmd := exec.Command("./random-exit-code")

	out, err := cmd.CombinedOutput()
	output = string(out)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			output += fmt.Sprintf("\nExit Code: %d", exitError.ExitCode())
		} else {
			output += fmt.Sprintf("\nUndefined Error: %v", err)
			log.Println("Build Error:", err)
		}
		ok = false
	}

	return
}

func PushExampleTask(SHA, SubmittedBy string) (ok bool, err error) {
	ok = false

	// See if any exampleTasks for this SHA are already queued
	for _, t := range task_queue {
		if e, ok := t.(ExampleTask); ok && e.SHA == SHA {
			return false, nil
		}
	}

	msg := "This commit has been added to the task queue. More information will appear here once the task has started"
	checkRun, err := gh.CreateCheckRun(SHA, exampleTitle, exampleSummaryInQueue, msg)
	if err != nil {
		return
	}

	Push(ExampleTask{
		SHA:         SHA,
		SubmittedBy: SubmittedBy,
		CheckRun:    checkRun,
	})
	if err != nil {
		return
	}

	ok = true
	return
}
