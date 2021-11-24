package alice

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestConsumer(t *testing.T) {
	key := "key"
	exchange, _ := CreateExchange("test-exchange", Direct, false, false, false, false, nil)
	queue := CreateQueue(exchange, "test-queue", false, false, false, false, nil)
	c, _ := broker.CreateConsumer(queue, key, "", DefaultConsumerErrorHandler)

	messageHandler := func(msg amqp.Delivery) {
		log.Println(string(msg.Body))
		s, _ := strconv.Atoi(string(msg.Body))
		divide(s, 0) // Divide by 0
	}

	go c.ConsumeMessages(nil, true, messageHandler)

	p, _ := broker.CreateProducer(exchange, DefaultProducerErrorHandler)
	p.PublishMessage([]byte("10"), &key, &amqp.Table{})

	time.Sleep(time.Second * 2)
}

func divide(a, b int) int {
	return a / b
}
