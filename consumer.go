package alice

import (
	"fmt"

	"github.com/streadway/amqp"
)

// RabbitConsumer models a RabbitMQ consumer
type RabbitConsumer struct {
	channel      *amqp.Channel // Channel this consumer uses to communicate with broker
	queue        *Queue        // The queue this consumer consumes from
	errorHandler func(error)   // Error Handler for this consumer
	conn         *Connection   // Pointer to broker connection
}

// ConsumeMessages consumes messages sent to the consumer
// Keep the consumerName empty if you wish to get a randomly generated consumer name
func (c *RabbitConsumer) ConsumeMessages(args amqp.Table, consumerName string, autoAck bool, messageHandler func(amqp.Delivery)) {
	messages, err := c.channel.Consume(
		c.queue.name,
		consumerName,
		autoAck,
		false,
		false,
		false,
		args,
	)
	logError(err, "Failed to consume messages")

	for message := range messages {
		logMessage(fmt.Sprintf("Received message of '%d' bytes from exchange '%v'", len(message.Body), message.Exchange))
		go messageHandler(message)
	}
}

// CreateConsumer creates a new Consumer
func (c *Connection) CreateConsumer(queue *Queue, routingKey string, errorHandler func(error)) (*RabbitConsumer, error) {

	consumer := &RabbitConsumer{
		channel:      nil,
		queue:        queue,
		errorHandler: errorHandler,
		conn:         c,
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
	err = consumer.bindQueue(queue, routingKey)
	if err != nil {
		return nil, err
	}

	//Prints the specifications
	logMessage(fmt.Sprintf("Declared queue %s with currently %d consumers, binding to exchange %q",
		queue.name, q.Consumers, queue.exchange.name))
	logMessage(fmt.Sprintf("Created consumer on queue '%s' with routing key '%s'", queue.name, routingKey))

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

func (c *RabbitConsumer) listenForClose() {
	closeChan := c.channel.NotifyClose(make(chan *amqp.Error))
	go func() {
		err := <-closeChan
		logMessage(err.Reason)
		if err.Recover {
			logMessage("Channel close is recoverable")
			c.ReconnectChannel()
		}
	}()
}

// ReconnectChannel tries to re-open this consumers channel
func (c *RabbitConsumer) ReconnectChannel() error {
	logMessage("Attempting to re-open consumer channel")
	var err error
	c.channel, err = c.conn.conn.Channel()
	return err
}

// Shutdown shuts down the consumer
func (c *RabbitConsumer) Shutdown() error {
	logMessage("Shutting down consumer")
	return c.channel.Close()
}

// DefaultConsumerErrorHandler handles the errors of this consumer
func DefaultConsumerErrorHandler(err error) {
	logMessage(err.Error())
}
