package github_traffic

import (
	"testing"

	"miikka.xyz/devops-app/store"
)

func TestJob(t *testing.T) {
	teardown := store.SetupTest(t)
	defer teardown()

	err := DoGithubTrafficStats()
	if err != nil {
		t.Fatal(err)
	}
}
