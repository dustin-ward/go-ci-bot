package auth

import (
	"context"
	"log"
	"net/http"
	"test-org-gozbot/config"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

func CreateClient() (*github.Client, error) {
	itr, err := ghinstallation.NewAppsTransportKeyFromFile(
		http.DefaultTransport,
		config.AppID(),
		config.CertPath(),
	)
	if err != nil {
		log.Fatalf("faild to create app transport: %v\n", err)
	}
	itr.BaseURL = config.GHEHost()

	//create git client with app transport
	client, err := github.NewClient(
		&http.Client{
			Transport: itr,
			Timeout:   time.Second * 30,
		},
	).WithEnterpriseURLs(config.GHEHost(), config.GHEHost())
	if err != nil {
		log.Fatalf("faild to create git client for app: %v\n", err)
	}

	token, _, err := client.Apps.CreateInstallationToken(
		context.Background(),
		config.InstallID(),
		nil,
	)
	if err != nil {
		log.Fatalf("failed to create installation token: %v\n", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.GetToken()},
	)
	oAuthClient := oauth2.NewClient(context.Background(), ts)

	apiClient, err := github.NewClient(oAuthClient).WithEnterpriseURLs(config.GHEHost(), config.GHEHost())
	return apiClient, err
}
