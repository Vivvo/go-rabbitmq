package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
	"log"
	"reflect"
	"sort"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func transform(t reflect.Type, input interface{}) interface{} {
	v := reflect.New(t).Interface()
	err := mapstructure.Decode(input, v)
	failOnError(err, fmt.Sprintf("Unable to transform message to type %s", t.String()))
	return v
}

func (r RabbitMQ) handleMessage(d amqp.Delivery) {
	var msg MessageDto
	err := json.Unmarshal(d.Body, &msg)
	failOnError(err, "Unable to marshall body")

	for _, h := range r.Handlers {
		if h.Type == "*" || h.Type == msg.PublishChangesMessageType {
			if h.HandlerFunc == nil {
				log.Printf("HandlerFunc not implemented")
				return
			}

			var v interface{}
			if h.ExpectedClass != nil {
				v = transform(reflect.TypeOf(h.ExpectedClass), msg.Payload)
			} else {
				v = msg.Payload
			}

			err := h.HandlerFunc(v)
			if err != nil {
				log.Printf("HandlerFunc returned an error handling the message")
				err = d.Nack(false, false)
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

	// Always use wildcard handler last
	sort.Slice(r.Handlers[:], func(i, j int) bool {
		return r.Handlers[i].Type != r.Handlers[j].Type
	})

	r.handleConsume(deliveries)
}
