package alice

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// Consumer models a RabbitMQ consumer
type Consumer struct {
	channel      *amqp.Channel // Channel this consumer uses to communicate with broker
	queue        *Queue        // The queue this consumer consumes from
	errorHandler func(error)   // Error Handler for this consumer
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
func (c *Connection) CreateConsumer(queue *Queue, bindingKey string, errorHandler func(error)) *Consumer {

	consumer := &Consumer{
		channel:      nil,
		queue:        queue,
		errorHandler: errorHandler,
	}

	var err error

	//Connects to the channel
	consumer.channel, err = c.conn.Channel()
	logError(err, "Failed to open channel")

	//Connects to exchange
	err = consumer.ConnectToExchange(queue.exchange)
	logError(err, "Failed to declare exchange")

	//Creates the queue
	q, err := consumer.CreateQueue(queue)
	logError(err, "Failed to declare queue")

	//Binds the queue to the exchange
	err = consumer.BindQueue(queue, bindingKey)
	logError(err, "Failed to bind queue to exchange")

	//Prints the specifications
	log.Printf("Declared queue (%q %d messages, %d consumer), binding to exchange %q",
		queue.name, q.Messages, q.Consumers, queue.exchange.name)

	return consumer
}

// ConnectToExchange connects to a specific exchange
func (c *Consumer) ConnectToExchange(exchange *Exchange) error {
	e := c.channel.ExchangeDeclare(
		exchange.name,
		string(exchange.exchangeType),
		exchange.durable,
		exchange.autoDelete,
		exchange.internal,
		exchange.noWait,
		exchange.args,
	)
	return e
}

// CreateQueue creates a queue with the same name as Kubernetes pod
func (c *Consumer) CreateQueue(queue *Queue) (amqp.Queue, error) {
	q, e := c.channel.QueueDeclare(
		queue.name,
		queue.durable,
		queue.autoDelete,
		queue.exclusive,
		queue.noWait,
		queue.args,
	)
	return q, e
}

// BindQueue binds the queue to the exchange
func (c *Consumer) BindQueue(queue *Queue, bindingKey string) error {
	e := c.channel.QueueBind(
		queue.name,
		bindingKey,
		queue.exchange.name,
		false,
		nil,
	)
	return e
}

// Shutdown shuts down the consumer
func (c *Consumer) Shutdown() error {
	return c.channel.Close()
}

// DefaultConsumerErrorHandler handles the errors of this consumer
func DefaultConsumerErrorHandler(err error) {
	fmt.Println(err)
}
