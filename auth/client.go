package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

var (
	gitHost  = "https://github.ibm.com/api/v3"
	appID    = int64(2206)
	certPath = "./private-key.pem"
)

func CreateClient(installID int64) (*github.Client, error) {
	privatePem, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatalf("failed to read pem: %v", err)
	}

	itr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, privatePem)
	if err != nil {
		log.Fatalf("faild to create app transport: %v\n", err)
	}
	itr.BaseURL = gitHost

	//create git client with app transport
	client, err := github.NewEnterpriseClient(
		gitHost,
		gitHost,
		&http.Client{
			Transport: itr,
			Timeout:   time.Second * 30,
		})
	if err != nil {
		log.Fatalf("faild to create git client for app: %v\n", err)
	}

	token, _, err := client.Apps.CreateInstallationToken(
		context.Background(),
		installID,
		&github.InstallationTokenOptions{})
	if err != nil {
		log.Fatalf("failed to create installation token: %v\n", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.GetToken()},
	)
	oAuthClient := oauth2.NewClient(context.Background(), ts)

	apiClient, err := github.NewEnterpriseClient(gitHost, gitHost, oAuthClient)
	return apiClient, err
}
