package consts

import (
	"log"
	"os"
)

var (
	Version = "unknown version"
	Build   = "unknown build time"
	Commit  = "unknown commit"
)

// Event queue
const QueueEventsName = "events_queue_durable"

// Errors
const (
	ErrBody          = "Invalid request body"
	ErrJSON          = "JSON error"
	ErrDatabase      = "Error with database"
	ErrInvalidID     = "ID was invalid"
	ErrAlreadyExists = "'%s' already exists"
	ErrNotFound      = "Not found"
)

// Collections
const (
	CollectionUsers       = "users"
	CollectionEvents      = "events"
	CollectionRepoTraffic = "repo_traffic"
	CollectionRepos       = "repos"
)

// AllCollections should hold anmes of all collections so those can be erased easily
var AllCollections = []string{CollectionUsers, CollectionEvents, CollectionRepoTraffic, CollectionRepos}

// Events
const (
	EventUserCreated         = "user_created"
	EventTrafficJobCompleted = "traffic_completed"
)

// Other
const (
	MaxBodySizeBytes = 64_000
)

// Database
var DatabaseName = "this_will_change"

func init() {
	DatabaseName = "laboratory"
	if os.Getenv("TEST_MODE") == "1" {
		DatabaseName = "test"
		log.Println("TEST MODE")
	}
}
