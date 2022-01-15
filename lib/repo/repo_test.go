package repo

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestQueryReposWithMeta(t *testing.T) {
	ts := time.Now().AddDate(0, 0, -365)
	repos, err := StoreGetRepositoryTrafficWithMeta(context.Background(), ts)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("repos: %+v\n", repos[0])
}
