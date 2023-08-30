package handlers

import (
	"context"
	"fmt"
	"log"

	"test-org-gozbot/auth"
	"test-org-gozbot/checks"

	"github.com/google/go-github/v50/github"
)

var (
	STATUS_IN_PROGRESS string = "in_progress"
	STATUS_COMPLETED   string = "completed"
)

func CheckRunEventHandler(deliveryID string, eventName string, event *github.CheckRunEvent) error {
	log.Println("CHECK RUN EVENT:", *event.Action)

	switch event.GetAction() {
	case "created":
		err := doCheckRun(event)
		if err != nil {
			return fmt.Errorf("doCheckRun:", err)
		}
	case "rerequested":
		err := createCheckRun(event, event.GetCheckRun().GetName())
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

func doCheckRun(event *github.CheckRunEvent) error {
	installID := event.GetInstallation().GetID()
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	name := event.GetCheckRun().GetName()
	checkRunID := event.GetCheckRun().GetID()

	client, err := auth.CreateClient(installID)
	if err != nil {
		return err
	}

	_, _, err = client.Checks.UpdateCheckRun(
		context.TODO(),
		owner,
		repo,
		checkRunID,
		github.UpdateCheckRunOptions{
			Name:   name,
			Status: &STATUS_IN_PROGRESS,
		},
	)
	if err != nil {
		return err
	}

	var output *github.CheckRunOutput
	var conclusion string
	switch name {
	case "GOZ Build":
		output, conclusion = checks.Build()
	case "GOZ Test":
		output, conclusion = checks.Build()
	case "GOZ PAX'd + Artifactory":
		output, conclusion = checks.Build()
	default:
		output, conclusion = checks.Undef()
		log.Println("WARNING: unknown check name: '%s'", name)
	}

	_, _, err = client.Checks.UpdateCheckRun(
		context.TODO(),
		owner,
		repo,
		checkRunID,
		github.UpdateCheckRunOptions{
			Name:       name,
			Status:     &STATUS_COMPLETED,
			Conclusion: &conclusion,
			Output:     output,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
