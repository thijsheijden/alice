package alice

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// RabbitConsumer models a RabbitMQ consumer
type RabbitConsumer struct {
	channel      *amqp.Channel // Channel this consumer uses to communicate with broker
	queue        *Queue        // The queue this consumer consumes from
	errorHandler func(error)   // Error Handler for this consumer
}

// ConsumeMessages consumes messages sent to the consumer
func (c *RabbitConsumer) ConsumeMessages(args amqp.Table, autoAck bool, messageHandler func(amqp.Delivery)) {
	messages, err := c.channel.Consume(
		c.queue.name,
		"",
		autoAck,
		false,
		false,
		false,
		args,
	)
	logError(err, "Failed to consume messages")

	for message := range messages {
		go messageHandler(message)
	}
}

// CreateConsumer creates a new Consumer
func (c *Connection) CreateConsumer(queue *Queue, bindingKey string, errorHandler func(error)) (*RabbitConsumer, error) {

	consumer := &RabbitConsumer{
		channel:      nil,
		queue:        queue,
		errorHandler: errorHandler,
	}

	var err error

	//Connects to the channel
	consumer.channel, err = c.conn.Channel()
	if err != nil {
		return nil, err
	}

	//Connects to exchange
	err = consumer.declareExchange(queue.exchange)
	if err != nil {
		return nil, err
	}

	//Creates the queue
	q, err := consumer.declareQueue(queue)
	if err != nil {
		return nil, err
	}

	//Binds the queue to the exchange
	err = consumer.bindQueue(queue, bindingKey)
	if err != nil {
		return nil, err
	}

	//Prints the specifications
	log.Printf("Declared queue (%q %d messages, %d consumer), binding to exchange %q",
		queue.name, q.Messages, q.Consumers, queue.exchange.name)

	return consumer, nil
}

func (c *RabbitConsumer) declareExchange(exchange *Exchange) error {
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

func (c *RabbitConsumer) declareQueue(queue *Queue) (amqp.Queue, error) {
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

func (c *RabbitConsumer) bindQueue(queue *Queue, bindingKey string) error {
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
func (c *RabbitConsumer) Shutdown() error {
	return c.channel.Close()
}

// DefaultConsumerErrorHandler handles the errors of this consumer
func DefaultConsumerErrorHandler(err error) {
	fmt.Println(err)
}
