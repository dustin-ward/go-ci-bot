package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.ibm.com/open-z/jeff-ci/config"
	"github.ibm.com/open-z/jeff-ci/events"
	"github.ibm.com/open-z/jeff-ci/tasks"
)

const (
	EventPollInterval = time.Second * 10
	TaskPollInterval  = time.Second * 10
)

const (
	NumWorkers = 3
)

func main() {
	config.NewConfig(
	)

	// Cancel any checks that were previously queued
	log.Println("Cleaning Stale Events...")
	err := events.Cleanup()
	if err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt to end program
	stopMain := make(chan os.Signal, 1)
	signal.Notify(stopMain, os.Interrupt)

	// Goroutine communication
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start worker goroutines
	for _ = range NumWorkers {
		worker := tasks.NewWorker()
		log.Printf("Starting Worker #%04d\n", worker.ID)

		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(TaskPollInterval)

			for {
				worker.Poll()

				// Wait for next poll or end.
				select {
				case <-ticker.C:
					continue
				case <-ctx.Done():
					// Stop worker
					return
				}
			}
		}()
	}

	// Poll for events
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(EventPollInterval)
		lastEventID := events.GetMostRecentEventID()
		log.Println("Polling for events...")

		for {
			// Events poll returns the ID of the last event that was handled
			lastEventID = events.Poll(lastEventID)

			// Wait for next poll or end.
			select {
			case <-ticker.C:
				// Do the poll
			case <-ctx.Done():
				// End Program
				return
			}
		}
	}()

	log.Println("JeffCI has started")
	log.Println("Press Ctrl+C to exit")

	//TODO:Perform any takedown operations needed
	<-stopMain
	cancel()
	wg.Wait()
	log.Println("Shutting Down...")
}
