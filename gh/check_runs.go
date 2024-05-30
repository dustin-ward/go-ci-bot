package gh

import (
	"context"
	"log"
	"test-org-gozbot/config"
	"test-org-gozbot/gh/auth"
	"time"

	"github.com/google/go-github/v62/github"
)

func CreateCheckRun(sha, title, summary, body string) (*github.CheckRun, error) {
	client, err := auth.GetClient()
	if err != nil {
		return nil, err
	}

	checkRun, _, err := client.Checks.CreateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		github.CreateCheckRunOptions{
			Name:      title,
			HeadSHA:   sha,
			Status:    &CHECK_STATUS_QUEUED,
			StartedAt: &github.Timestamp{time.Now()},
			Output:    &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
		},
	)

	return checkRun, err
}

func UpdateCheckRun(checkRun *github.CheckRun, summary, body string) (*github.CheckRun, error) {
	client, err := auth.GetClient()
	if err != nil {
		return nil, err
	}

	title := checkRun.GetName()
	newCheckRun, _, err := client.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		checkRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:   title,
			Status: &CHECK_STATUS_IN_PROGRESS,
			Output: &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
		},
	)

	return newCheckRun, err
}

func CompleteCheckRun(checkRun *github.CheckRun, conclusion, summary, body string) (*github.CheckRun, error) {
	log.Println("CompleteCheckRun:", conclusion, "body len:", len(body), body[max(0, len(body)-100):])
	log.Println("CheckRun:", checkRun.GetName(), checkRun.GetID())
	client, err := auth.GetClient()
	if err != nil {
		return nil, err
	}

	title := checkRun.GetName()
	newCheckRun, _, err := client.Checks.UpdateCheckRun(context.TODO(), config.Owner(), config.Repo(),
		checkRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:        title,
			Status:      &CHECK_STATUS_COMPLETED,
			Conclusion:  &conclusion,
			CompletedAt: &github.Timestamp{time.Now()},
			Output:      &github.CheckRunOutput{Title: &title, Summary: &summary, Text: &body},
		},
	)

	return newCheckRun, err
}

func GetCheckRunsForRef(ref string) ([]*github.CheckRun, error) {
	client, err := auth.GetClient()
	if err != nil {
		return []*github.CheckRun{}, err
	}

	appId := config.AppID()
	res, _, err := client.Checks.ListCheckRunsForRef(
		context.TODO(),
		config.Owner(),
		config.Repo(),
		ref,
		&github.ListCheckRunsOptions{
			AppID: &appId,
		},
	)

	if err != nil {
		return []*github.CheckRun{}, err
	}

	return res.CheckRuns, nil
}
