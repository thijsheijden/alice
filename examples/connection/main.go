package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/streadway/amqp"
	"github.com/thijsheijden/alice"
)

type test struct {
	Msg string
}

func main() {
	alice.SetLogging()
	broker, err := alice.CreateBroker(alice.DefaultConfig)
	if err != nil {
		log.Println(err)
	}

	exchange, err := alice.CreateExchange("test-exchange", alice.Direct, false, true, false, false, nil)
	if err != nil {
		log.Println(err)
	}

	q := alice.CreateQueue(exchange, "test-queue", false, false, true, false, nil)

	c, err := broker.CreateConsumer(q, "key", "", false, alice.DefaultConsumerErrorHandler)

	go c.ConsumeMessages(nil, handleMessage)

	select {}
}

func handleMessage(msg amqp.Delivery) {
	fmt.Println(msg.Body)
	msg.Ack(true)
	log.Println(fmt.Sprintf("%v", runtime.NumGoroutine()))
}
