package main

import (
	"alice"

	"github.com/streadway/amqp"
)

func main() {
	a := alice.DefaultConfig

	// Connect to RabbitMQ
	c := alice.Connect(*a)

	queue := alice.CreateDefaultQueue("test-queue")
	exchange := alice.CreateDefaultExchange("test-exchange", alice.Direct)

	cons := c.CreateConsumer(*exchange, *queue, "test")
	go cons.ConsumeMessages(nil, func(delivery amqp.Delivery) {})

	// Create producer on default exchange
	p := c.CreateProducer(*exchange)
	defer p.DeferShutdown()

	// Publish message
	p.Produce1000Messages()

	select {}
}
