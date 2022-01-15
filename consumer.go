package alice

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

// RabbitConsumer models a RabbitMQ consumer
type RabbitConsumer struct {
	channel        *amqp.Channel       // Channel this consumer uses to communicate with broker
	queue          *Queue              // The queue this consumer consumes from
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
	autoAck: bool, whether to automatically acknowledge messages
	messageHandler: func(amqp.Delivery), a handler for incoming messages. Every message the handler is called in a new goroutine
*/
func (c *RabbitConsumer) ConsumeMessages(args amqp.Table, autoAck bool, messageHandler func(amqp.Delivery)) {
	messages, err := c.channel.Consume(
		c.queue.name,
		c.tag,
		c.autoAck,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		log.Error().AnErr("err", err).Str("type", "consumer").Str("consumerTag", c.tag).Str("routingKey", c.routingKey).Msg("failed to consume messages")
	}

	// Set some more consumer attributes
	c.autoAck = autoAck
	c.args = args
	c.messageHandler = messageHandler

	// Listen for incoming messages and pass them to the message handler
	log.Info().Str("type", "consumer").Str("consumerTag", c.tag).Str("routingKey", c.routingKey).Msg("starting message consumption")
	for message := range messages {
		log.Trace().Str("type", "consumer").Str("consumerTag", c.tag).Str("routingKey", c.routingKey).Str("exchange", message.Exchange).Int("msgSize", len(message.Body)).Msg("received message")
		go func(message amqp.Delivery) {
			// Intercept any errors propagating up the stack
			defer func() {
				if err := recover(); err != nil {
					log.Error().Str("type", "consumer").AnErr("err", err.(error)).Msg("error occurred in message handler")
				}

				if autoAck {
					message.Ack(true)
					log.Trace().Str("type", "consumer").Str("consumerTag", c.tag).Str("routingKey", c.routingKey).Str("msgID", message.MessageId).Msg("automatically acked message")
				}
			}()

			// Call the message handler
			messageHandler(message)
		}(message)
	}
}

// createConsumer creates a new Consumer on this connection
func (c *connection) createConsumer(queue *Queue, routingKey string, consumerTag string) (Consumer, error) {
	consumer := &RabbitConsumer{
		channel:    nil,
		queue:      queue,
		conn:       c,
		tag:        consumerTag,
		routingKey: routingKey,
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
	_, err = consumer.declareQueue(queue)
	if err != nil {
		return nil, err
	}

	//Binds the queue to the exchange
	err = consumer.bindQueue(queue, routingKey)
	if err != nil {
		return nil, err
	}

	consumer.listenForClose()

	log.Info().Str("type", "consumer").Str("queue", queue.name).Str("routingKey", routingKey).Str("consumerTag", consumerTag).Msg("created consumer")

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
		log.Error().Str("type", "consumer").AnErr("err", closeErr).Str("routingKey", c.routingKey).Str("consumerTag", c.tag).Msg("connection was closed")
		c.reconnect()
	}()
}

// ReconnectChannel tries to re-open this consumers channel
func (c *RabbitConsumer) ReconnectChannel() error {
	log.Info().Str("type", "consumer").Str("routingKey", c.routingKey).Str("consumerTag", c.tag).Msg("attempting to re-open channel")
	var err error
	c.channel, err = c.conn.conn.Channel()
	if err != nil {
		log.Error().AnErr("err", err).Str("type", "consumer").Str("routingKey", c.routingKey).Str("consumerTag", c.tag).Msg("failed to re-open channel")
	}
	return err
}

// Shutdown shuts down the consumer
func (c *RabbitConsumer) Shutdown() error {
	log.Info().Str("type", "consumer").Str("routingKey", c.routingKey).Str("consumerTag", c.tag).Msg("shutting down consumer")
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

			log.Info().Str("type", "consumer").Str("routingKey", c.routingKey).Str("consumerTag", c.tag).Msg("reconnected")

			go c.ConsumeMessages(c.args, c.autoAck, c.messageHandler)

			return nil
		}
	}
}
