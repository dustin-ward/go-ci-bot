package build

import (
	"os"
	"time"

	"github.com/google/go-github/v62/github"
)

func Poll(apiClient *github.Client, ticker *time.Ticker, stopBuilds chan struct{}, stopMain chan os.Signal) {
	for {
		select {
		case <-ticker.C:
			build, ok := Pop()
			if ok {
				build.Start(apiClient)
			}
		case <-stopMain:
			//TODO: Cancel all in-progress builds
			close(stopBuilds)
			return
		}
	}
}
