package main

import (
	"context"
	"log"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"miikka.xyz/devops-app/cache"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/events"
	"miikka.xyz/devops-app/jobs/github_traffic"
	"miikka.xyz/devops-app/server"
	"miikka.xyz/devops-app/store"
	"miikka.xyz/devops-app/utils"
)

func main() {
	// Init RabbitMQ
	rabbitConn, rabbitCh := events.CreateQueue(consts.QueueEventsName)
	defer rabbitCh.Close()
	defer rabbitConn.Close()

	// TODO: Now store has init() function, which will be called automatically. Refactor that
	defer store.Close()

	cacheClient, _ := cache.New(false)

	// Create server and pass event queue for it
	s := server.New("8080", rabbitCh, cacheClient)

	if utils.GetEnv("RUN_JOBS_ON_STARTUP", "false") == "true" {
		log.Println("run jobs on startup is on")
		go func() {
			runJobs(rabbitCh, cacheClient)
		}()
	} else {
		log.Println("run jobs on startup is off")
	}

	log.Println("Server listening 8080...")
	if err := s.HTTP.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func runJobs(rabbitCh *amqp.Channel, cacheClient *cache.Cache) {
	if err := github_traffic.DoGithubTrafficStats(); err != nil {
		log.Println("job failed", err)
		return
	}
	event := events.Event{
		CreatedAt: time.Now(),
		ObjectID:  primitive.NilObjectID,
		Type:      consts.EventTrafficJobCompleted,
	}
	if err := events.Publish(rabbitCh, &event); err != nil {
		log.Println("publishing event failed", err)
		return
	}
	log.Println("traffic data saved to database")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Update cache
	err := cacheClient.UpdateTrafficCache(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}
}
