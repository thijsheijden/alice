package main

import (
	"encoding/json"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/thijsheijden/alice"
)

var done chan bool

func TestBroker(t *testing.T) {
	broker := alice.CreateMockBroker()

	done = make(chan bool)

	e, _ := alice.CreateExchange("ui-direct-exchange", alice.Direct, true, false, false, false, nil)

	q := alice.CreateQueue(e, "test-queue", false, false, true, false, nil)

	c, _ := broker.CreateConsumer(q, "key", alice.DefaultConsumerErrorHandler)

	go c.ConsumeMessages(nil, "", true, testHandleMessage)

	p, _ := broker.CreateProducer(e, alice.DefaultProducerErrorHandler)

	key := "key"
	p.PublishMessage([]byte("Hey!"), &key, &amqp.Table{})

	<-done

	ce := c.(*alice.MockConsumer)

	d, _ := json.Marshal("Hey!")

	assert.Equal(t, d, ce.ReceivedMessages[0].Body)
}

func testHandleMessage(msg amqp.Delivery) {
	done <- true
}
