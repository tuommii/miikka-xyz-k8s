package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/events"
)

func main() {
	// The messagesChannel is the event queue
	rabbitConn, rabbitCh, messagesChannel := events.CreateEventQueue(consts.QueueEventsName)
	defer rabbitConn.Close()
	defer rabbitCh.Close()

	// Make a blocking channel
	foreverCh := make(chan bool)

	// Loop event queue in background
	go func() {
		// Each dot in message increases sleep time
		for msg := range messagesChannel {
			// Process each message from queue
			processMessage(msg)
		}
	}()

	log.Println("Waiting messages...")
	<-foreverCh
}

func processMessage(msg amqp.Delivery) {
	log.Printf("Received message: \n%s\n", string(msg.Body))

	event := events.Event{}
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		log.Println(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*45)
	defer cancel()

	switch event.Type {
	case consts.EventTrafficJobCompleted:
		log.Println("received traffic job completed event")
		id, err := events.StoreCreateEvent(ctx, &events.Event{
			CreatedAt: time.Now(),
			Type:      consts.EventTrafficJobCompleted,
		})
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("event stored to database with id:", id.Hex())
	}
	log.Println("Done")
	msg.Ack(false)
}
