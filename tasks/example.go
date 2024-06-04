package tasks

import (
	"fmt"
	"github.ibm.com/open-z/jeff-ci/gh"
	"log"
	"os/exec"
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

func (e ExampleTask) Enqueue() {
	// See if any workers are already doing this task
	for _, worker := range workerPool {
		if worker.CurTask != nil {
			if eTask, ok := (*worker.CurTask).(ExampleTask); ok && eTask.SHA == e.SHA {
				log.Printf("ExampleTask not queued: A worker is already proccessing this task\n")
				return
			}
		}
	}

	// See if any exampleTasks for this SHA are already queued
	for _, t := range task_queue {
		if eTask, ok := t.(ExampleTask); ok && eTask.SHA == e.SHA {
			log.Printf("ExampleTask not queued: This task is already in the queue\n")
			return
		}
	}

	msg := "This commit has been added to the task queue. More information will appear here once the task has started"
	var err error
	e.CheckRun, err = gh.CreateCheckRun(e.SHA, exampleTitle, exampleSummaryInQueue, msg)
	if err != nil {
		log.Printf("ExampleTask not queued: error creating check run: %v\n", err)
		return
	}

	pushTask(e)

	log.Printf("ExampleTask added to queue")
	return
}

func (e ExampleTask) Provision() (string, bool) {
	// This task will run locally
	return "local", true
}

func (e ExampleTask) Do(host string) error {
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

	time.Sleep(time.Second * 30)
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
