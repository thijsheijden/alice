package main

import (
	"fmt"

	"github.com/streadway/amqp"
	"github.com/thijsheijden/alice"
)

type test struct {
	Msg string
}

func main() {
	broker := alice.CreateMockBroker()

	e, _ := alice.CreateExchange("ui-direct-exchange", alice.Direct, true, false, false, false, nil)

	q := alice.CreateQueue(e, "test-queue", false, false, true, false, nil)

	c, _ := broker.CreateConsumer(q, "key", alice.DefaultConsumerErrorHandler)

	go c.ConsumeMessages(nil, true, handleMessage)

	p, _ := broker.CreateProducer(e, alice.DefaultProducerErrorHandler)

	key := "key"
	p.PublishMessage("Hey!", &key, &amqp.Table{})

	select {}
}

func handleMessage(msg amqp.Delivery) {
	fmt.Println(msg.Body)
}
