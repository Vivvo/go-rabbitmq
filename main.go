package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (r RabbitMQ) handleMessage(d amqp.Delivery) {
	var msg MessageDto
	err := json.Unmarshal(d.Body, &msg)
	failOnError(err, "Unable to marshall body")

	for _, h := range r.Handlers {
		if msg.PublishChangesMessageType == h.Type {
			err := h.HandlerFunc(msg)
			if err != nil {
				log.Printf("HandlerFunc returned an error handling the message, requeueing...")
				err = d.Nack(false, true)
				failOnError(err, "Failed to nack message")
			} else {
				log.Printf("HandlerFunc successfully handled the message")
				err = d.Ack(false)
				failOnError(err, "Failed to ack message")
			}
			return
		}
	}

	log.Printf("No Handlers can handle this message, rejecting...")
	err = d.Reject(false)
	failOnError(err, "Failed to reject message")
}

func (r RabbitMQ) handleConsume(deliveries <-chan amqp.Delivery) {
	forever := make(chan bool)

	go func() {
		for d := range deliveries {
			r.handleMessage(d)
		}
	}()

	<-forever
}

func (r RabbitMQ) Run() {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", r.Config.Username, r.Config.Password, r.Config.Hostname, r.Config.Port))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		r.Config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	deliveries, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	r.handleConsume(deliveries)
}
