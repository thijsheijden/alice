package main

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"github.com/thijsheijden/alice"
)

func main() {
	alice.SetLogLevel(-1)

	// Create a connection configuration
	connectionConfig := alice.CreateConfig(
		"guest",
		"guest",
		"localhost:5672",
		5672,
		true,
		time.Second*10,
	)

	// Create a broker using the connection config
	broker, err := alice.CreateBroker(connectionConfig)
	if err != nil {
		log.Println(err)
	}

	// Create an exchange called 'test-exchange' using direct routing
	exchange, err := alice.CreateExchange("test-exchange", alice.Direct, false, true, false, false, nil)
	if err != nil {
		log.Println(err)
	}

	// Create a queue called 'test-queue'
	q := alice.CreateQueue(exchange, "test-queue", false, false, true, false, nil)

	// Create a consumer bound to this queue, listening for messages with routing key 'key'
	c, err := broker.CreateConsumer(q, "key", "consumer-tag")

	// Start consuming messages
	// Every received message is passed to the handleMessage function
	go c.ConsumeMessages(nil, false, handleMessage)

	select {}
}

func handleMessage(msg amqp.Delivery) {
	fmt.Println(msg.Body)
	msg.Ack(true)
}
