package alice

import (
	"log"

	"github.com/streadway/amqp"
)

// Consumer models a RabbitMQ consumer
type Consumer struct {
	channel *amqp.Channel // Channel this consumer uses to communicate with broker
	queue   *Queue        // The queue this consumer consumes from
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
func (c *Connection) CreateConsumer(exchange Exchange, queue Queue, bindingKey string) *Consumer {
	consumer := &Consumer{
		channel: nil,
		queue:   &queue,
	}

	var err error

	consumer.channel, err = c.conn.Channel()
	logError(err, "Failed to open channel")

	//Connect to this specific exchange, default taken from exchange.go
	err = consumer.channel.ExchangeDeclare(
		exchange.name,
		string(exchange.exchangeType),
		exchange.durable,
		exchange.autoDelete,
		exchange.internal,
		exchange.noWait,
		exchange.args,
	)
	logError(err, "Failed to declare exchange")

	//Create a queue with the same name as Kubernetes pod
	q, err := consumer.channel.QueueDeclare(
		queue.name,
		queue.durable,
		queue.autoDelete,
		queue.exclusive,
		queue.noWait,
		queue.args,
	)
	logError(err, "Failed to declare queue")

	log.Printf("Declared queue (%q %d messages, %d consumer), binding to exchange %q",
		queue.name, q.Messages, q.Consumers, exchange.name)

	//Bind the queue to the exchange
	err = consumer.channel.QueueBind(
		queue.name,
		bindingKey,
		exchange.name,
		false,
		nil,
	)
	logError(err, "Failed to bind queue to exchange")

	return consumer
}

// Shutdown shuts down the consumer
func (c *Consumer) Shutdown() error {
	return c.channel.Close()
}
