package events

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	"miikka.xyz/devops-app/consts"
	"miikka.xyz/devops-app/utils"
)

func newConn() *amqp.Connection {
	url := utils.GetEnv("AMQP_SERVER_URL", "amqp://guest:guest@localhost:5672/")
	log.Println("rabbitmq connection url", url)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(url, err)
	}
	return conn
}

func CreateQueue(name string) (*amqp.Connection, *amqp.Channel) {
	rabbit := newConn()
	ch, err := rabbit.Channel()
	if err != nil {
		log.Fatal(err)
	}
	_, err = ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatal(err)
	}
	return rabbit, ch
}

func CreateEventQueue(name string) (*amqp.Connection, *amqp.Channel, <-chan amqp.Delivery) {
	conn, ch := CreateQueue(name)
	msgs, err := ch.Consume(
		name,  // queue name
		"",    // consumer
		false, // auto ack(nowledge), now acknowledging manually
		false, // exclusive
		false, // no local
		false, // no wait
		nil,   // args
	)
	if err != nil {
		log.Fatal(err)
	}
	return conn, ch, msgs
}

func Publish(ch *amqp.Channel, event *Event) error {
	bytes, err := json.Marshal(*event)
	if err != nil {
		return err
	}
	err = ch.Publish(
		"",
		consts.QueueEventsName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         bytes,
		})
	return err
}
