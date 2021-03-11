package alice

import (
	"log"

	"github.com/streadway/amqp"
)

// Consumer models a RabbitMQ consumer
type Consumer struct {
	channel *amqp.Channel
	queue   *Queue
}

// ConsumeMessages consumes messages sent to the consumer
func (c *Consumer) ConsumeMessages(args amqp.Table, messageHandler func(amqp.Delivery)) {
	messages, err := c.channel.Consume(
		c.queue.name,
		"",
		false,
		false,
		false,
		false,
		args,
	)
	logError(err, "Failed to consume messages")

	for message := range messages {
		go messageHandler(message)
		message.Ack(true)
	}
}

// CreateConsumer creates a new Consumer
func (c *Connection) CreateConsumer(exchange Exchange, q Queue, bindingKey string) *Consumer {
	cons := &Consumer{
		channel: nil,
		queue:   &q,
	}

	var err error

	cons.channel, err = c.conn.Channel()
	logError(err, "Failed to open channel")

	//Connect to this specific exchange, default taken from exchange.go
	err = cons.channel.ExchangeDeclare(
		exchange.name,
		string(exchange.exchangeType),
		exchange.durable,
		exchange.autodelete,
		exchange.internal,
		exchange.noWait,
		exchange.args,
	)
	logError(err, "Failed to declare exchange")

	//Create a queue with the same name as Kubernetes pod
	queue, err := cons.channel.QueueDeclare(
		q.name,
		q.durable,
		q.autoDelete,
		q.exclusive,
		q.noWait,
		q.arguments,
	)
	logError(err, "Failed to declare queue")

	log.Printf("Declared queue (%q %d messages, %d consumer), binding to exchange %q",
		q.name, queue.Messages, queue.Consumers, exchange.name)

	//Bind the queue to the exchange
	err = cons.channel.QueueBind(
		q.name,
		bindingKey,
		exchange.name,
		false,
		nil,
	)
	logError(err, "Failed to bind queue to exchange")

	return cons
}

// Shutdown shuts down the consumer
func (cons *Consumer) Shutdown() error {
	return cons.channel.Close()
}
