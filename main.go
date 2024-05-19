package main

import (
	"log"
	"os"
	"os/signal"
	"test-org-gozbot/auth"
	"test-org-gozbot/config"
	"test-org-gozbot/events"
	"time"
)

const (
	EventPollInterval = time.Second * 10
)

func main() {
	config.NewConfig(
		"https://github.ibm.com",
		"test-org-gozbot",
		"goz-workflow-demo",
		"./private-key.pem",
		2206,
		19342,
	)

	apiClient, err := auth.CreateClient()
	if err != nil {
		log.Fatal("Auth: ", err)
	}
	log.Println("GOZBOT has started")

	// Wait for interrupt to end program
	stopMain := make(chan os.Signal, 1)
	signal.Notify(stopMain, os.Interrupt)
	log.Println("Press Ctrl+C to exit")

	// Cancel any checks that were previously queued
	err = events.Cleanup(apiClient)
	if err != nil {
		log.Fatal(err)
	}

	// Poll for events
	eventTicker := time.NewTicker(EventPollInterval)
	lastPollTime := time.Now()
	stopPoll := make(chan struct{})
	go events.Poll(apiClient, eventTicker, lastPollTime, stopPoll, stopMain)

	//TODO:Perform any takedown operations needed
	<-stopMain
	log.Println("Shutting Down...")
}
