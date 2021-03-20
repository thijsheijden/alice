package main

import (
	"fmt"

	"github.com/thijsheijden/alice"

	"github.com/streadway/amqp"
)

func main() {
	a := alice.DefaultConfig

	// Connect to RabbitMQ
	c := alice.Connect(*a)

	exchange, err := alice.CreateDefaultExchange("test-exchange", alice.Direct)
	queue := alice.CreateDefaultQueue(exchange, "test-queue")
	if err != nil {
		fmt.Println(err)
	}

	cons, _ := c.CreateConsumer(queue, "test", alice.DefaultConsumerErrorHandler)
	go cons.ConsumeMessages(nil, true, func(delivery amqp.Delivery) { fmt.Println(delivery.Body) })

	// Create producer on default exchange
	p, _ := c.CreateProducer(exchange, alice.DefaultProducerErrorHandler)
	defer p.Shutdown()

	// Publish message
	p.PublishNMessages(5)

	select {}
}
