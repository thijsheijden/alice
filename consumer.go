package alice

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// RabbitConsumer models a RabbitMQ consumer
type RabbitConsumer struct {
	channel        *amqp.Channel       // Channel this consumer uses to communicate with broker
	queue          *Queue              // The queue this consumer consumes from
	errorHandler   func(error)         // Error Handler for this consumer
	conn           *connection         // Pointer to broker connection
	autoAck        bool                // Whether this consumer want autoAck
	tag            string              // Consumer tag
	args           amqp.Table          // Additional arguments when consuming messages
	messageHandler func(amqp.Delivery) // Message handler to call if this consumer receives a message
	routingKey     string              // Routing key this consumer listens to
}

/*
ConsumeMessages starts the consumption of messages from the queue the consumer is bound to
	args: amqp.Table, additional arguments for this consumer
	messageHandler: func(amqp.Delivery), a handler for incoming messages. Every message the handler is called in a new goroutine
*/
func (c *RabbitConsumer) ConsumeMessages(args amqp.Table, messageHandler func(amqp.Delivery)) {
	messages, err := c.channel.Consume(
		c.queue.name,
		c.tag,
		c.autoAck,
		false,
		false,
		false,
		args,
	)
	logError(err, "Failed to consume messages")

	// Set some more consumer attributes
	c.args = args
	c.messageHandler = messageHandler

	// Listen for incoming messages and pass them to the message handler
	for message := range messages {
		logMessage(fmt.Sprintf("Received message of '%d' bytes from exchange '%v'", len(message.Body), message.Exchange))
		go messageHandler(message)
	}
}

// createConsumer creates a new Consumer on this connection
func (c *connection) createConsumer(queue *Queue, routingKey string, consumerTag string, autoAck bool, errorHandler func(error)) (*RabbitConsumer, error) {

	consumer := &RabbitConsumer{
		channel:      nil,
		queue:        queue,
		errorHandler: errorHandler,
		conn:         c,
		tag:          consumerTag,
		autoAck:      autoAck,
		routingKey:   routingKey,
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

	consumer.listenForClose()

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
		closeErr := <-closeChan
		logError(closeErr, closeErr.Reason)
		c.reconnect()
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

func (c *RabbitConsumer) reconnect() error {
	// Wait for the connection to be open again
	// Create new ticker with the desired connection delay time
	ticker := time.NewTicker(time.Second * 10)

	for {
		<-ticker.C // New tick

		// Check if connection is open yet
		if !c.conn.conn.IsClosed() {
			// Attempt to re-connect
			var err error

			//Connects to the channel
			c.channel, err = c.conn.conn.Channel()
			if err != nil {
				return err
			}

			//Connects to exchange
			err = c.declareExchange(c.queue.exchange)
			if err != nil {
				return err
			}

			//Creates the queue
			_, err = c.declareQueue(c.queue)
			if err != nil {
				return err
			}

			//Binds the queue to the exchange
			err = c.bindQueue(c.queue, c.routingKey)
			if err != nil {
				return err
			}

			c.listenForClose()

			logMessage("consumer reconnected")

			go c.ConsumeMessages(c.args, c.messageHandler)

			return nil
		}
	}
}

// DefaultConsumerErrorHandler handles the errors of this consumer
func DefaultConsumerErrorHandler(err error) {
	logMessage(err.Error())
}

// MARK: Consumer errors
