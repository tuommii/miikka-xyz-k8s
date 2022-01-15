package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"miikka.xyz/devops-app/cache"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/events"
	"miikka.xyz/devops-app/jobs/github_traffic"
	"miikka.xyz/devops-app/store"
)

func main() {
	// Init RabbitMQ
	rabbitConn, rabbitCh := events.CreateQueue(consts.QueueEventsName)
	defer rabbitCh.Close()
	defer rabbitConn.Close()
	// TODO: Refactor init() in store
	defer store.Close()

	// Run job
	log.Println("Starting to run a github repository traffic job")
	err := github_traffic.DoGithubTrafficStats()
	if err != nil {
		log.Println("job failed", err)
	}
	log.Println("job completed successfully")

	// Publish event
	event := events.Event{
		CreatedAt: time.Now(),
		ObjectID:  primitive.NilObjectID,
		Type:      consts.EventTrafficJobCompleted,
	}
	if err := events.Publish(rabbitCh, &event); err != nil {
		log.Println("publishing event failed", err)
		return
	}
	log.Println("traffic job completed event published")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Update cache
	cacheClient, _ := cache.New(false)
	cacheClient.UpdateTrafficCache(ctx, 0)
	log.Println("cache updated")
	log.Println("exit...")
}
