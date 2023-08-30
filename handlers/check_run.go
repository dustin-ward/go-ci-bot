package handlers

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/google/go-github/v50/github"
    "test-org-gozbot/auth"
)

func CheckRunEventHandler(deliveryID string, eventName string, event *github.CheckRunEvent) error {
	log.Println("CHECK RUN EVENT:", *event.Action)

	switch event.GetAction() {
	case "created":
		err := doCheckRun(event, "GOZ Build")
		if err != nil {
			return fmt.Errorf("doCheckRun:", err)
		}
	case "rerequested":
		err := createCheckRun(event, "GOZ Build")
		if err != nil {
			return fmt.Errorf("createCheckRun:", err)
		}
	default:
		log.Println("Unhandled action:", event.GetAction())
	}

	return nil
}

func createCheckRun(event any, name string) error {
    var installID int64
    var owner string
    var repo string
    var sha string
    switch t := event.(type) {
        case *github.CheckSuiteEvent:
            se := event.(*github.CheckSuiteEvent)
            installID = se.GetInstallation().GetID()
            owner = se.GetRepo().GetOwner().GetLogin()
            repo = se.GetRepo().GetName()
            sha = se.GetCheckSuite().GetHeadSHA()
        case *github.CheckRunEvent:
            re := event.(*github.CheckRunEvent)
            installID = re.GetInstallation().GetID()
            owner = re.GetRepo().GetOwner().GetLogin()
            repo = re.GetRepo().GetName()
            sha = re.GetCheckRun().GetHeadSHA()
        default:
            return fmt.Errorf("createCheckRun: unknown event type:", t)
    }

	client, err := auth.CreateClient(installID)
	if err != nil {
		return err
	}

	_, _, err = client.Checks.CreateCheckRun(
		context.TODO(),
		owner,
		repo,
		github.CreateCheckRunOptions{
			Name:    name,
			HeadSHA: sha,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func doCheckRun(event *github.CheckRunEvent, name string) error {
    installID := event.GetInstallation().GetID()
    owner := event.GetRepo().GetOwner().GetLogin()
    repo := event.GetRepo().GetName()
    checkRunID := event.GetCheckRun().GetID()

	client, err := auth.CreateClient(installID)
	if err != nil {
		return err
	}

	s := "in_progress"
	s1 := "completed"
	s2 := "success"
	_, _, err = client.Checks.UpdateCheckRun(
		context.TODO(),
		owner,
		repo,
		checkRunID,
		github.UpdateCheckRunOptions{
			Name:   name,
			Status: &s,
		},
	)
	if err != nil {
		return err
	}
	title := "Results"
	text := "No issues here!"

	log.Println("Starting Code Exec...")

	time.Sleep(4*time.Second)
	cmd := exec.Command("./random")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd.Start:", err)
	}
	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
            log.Println("random failed:", err)
			s2 = "failure"
			text = "This test failed on purpose, re-run it :)"
		} else {
			return fmt.Errorf("cmd.Wait:", err)
		}
	}

	summary := "test-summary: " + s2
	output := github.CheckRunOutput{Title: &title, Text: &text, Summary: &summary}
	_, _, err = client.Checks.UpdateCheckRun(
		context.TODO(),
		owner,
		repo,
		checkRunID,
		github.UpdateCheckRunOptions{
			Name:       name,
			Status:     &s1,
			Conclusion: &s2,
			Output:     &output,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
