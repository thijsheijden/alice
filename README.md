<p align="center">
  <a href="https://it_me-ian.artstation.com/"></a><img width="40%" src="images/down_the_rabbit_hole.png" alt="Follow me down the rabbit hole...">
</p>
<p align="center">
  “I'm late! I'm late! For a very important date!" - White Rabbit.
</p>

# Alice
[![GoDoc](https://pkg.go.dev/badge/github.com/thijsheijden/alice?utm_source=godoc)](https://pkg.go.dev/github.com/thijsheijden/alice#section-documentation)

Alice is a wrapper around the Streadway <a href="">amqp</a> package, designed to be easier to use and offer automatic handling of many errors out of the box.

<hr>
Credit for the cute Gopher goes to <a href="https://it_me-ian.artstation.com/">Ian Derksen</a>

## Features
- Automatic broker reconnect (attempted at a user-defined interval)
- Automatic producer and consumer reconnect upon channel error
- Every message handled in a new routine
- Separate TCP connections for producers and consumers
- Queues and exchanges are objects, which can be reused for multiple consumers and/or producers

## Installation
```shell
go get github.com/thijsheijden/alice
```

## Quickstart
```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
	"github.com/thijsheijden/alice"
)

func main() {
	// Turn on logging
	alice.SetLogging()

	// Create a connection configuration
	connectionConfig := alice.CreateConfig(
		"guest",
		"guest",
		"localhost:5672",
		5672,
		true,
		time.Second*10,
		alice.DefaultErrorHandler,
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
	c, err := broker.CreateConsumer(q, "key", "consumer-tag", alice.DefaultConsumerErrorHandler)

	// Start consuming messages
	// Every received message is passed to the handleMessage function
	go c.ConsumeMessages(nil, false, handleMessage)

	select {}
}

func handleMessage(msg amqp.Delivery) {
	fmt.Println(msg.Body)
	msg.Ack(true)
}
```
