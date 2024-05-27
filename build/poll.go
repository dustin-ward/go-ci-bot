package build

import (
	"github.com/google/go-github/v62/github"
)

func Poll(apiClient *github.Client) {
	build, ok := Pop()
	if ok {
		build.Start(apiClient)
	}
}
