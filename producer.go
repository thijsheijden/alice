package alice

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

// RabbitProducer models a RabbitMQ producer
type RabbitProducer struct {
	channel      *amqp.Channel       // The channel this producer uses to communicate with the broker
	exchange     *Exchange           // The exchange this producer produces to
	errorHandler func(ProducerError) // Error handler for this producer
	conn         *connection         // Pointer to broker connection
}

// CreateProducer creates and returns a producer attached to the given exchange.
// The errorHandler can be the DefaultProducerErrorHandler or a custom handler.
func (c *connection) createProducer(exchange *Exchange, errorHandler func(ProducerError)) (*RabbitProducer, error) {

	var err error

	// Create producer object
	p := &RabbitProducer{
		channel:      nil,
		exchange:     exchange,
		errorHandler: errorHandler,
		conn:         c,
	}

	// Open channel to broker
	err = p.openChannel(c)
	if err != nil {
		// Throw error that channel could not be opened
		return nil, err
	}

	// Declare the exchange
	err = p.declareExchange(exchange)
	if err != nil {
		// Throw error that the exchange could not be declared
		return nil, err
	}

	// Listen for channel or connection close message
	p.listenForClose()

	// Listen for overflow messages from broker
	p.listenForFlow()

	// Listen for returned messages from the broker
	p.listenForReturnedMessages()

	log.Info().Str("exchange", exchange.name).Msg("created producer")

	return p, nil
}

// Open channel to broker
func (p *RabbitProducer) openChannel(c *connection) error {
	var err error
	log.Info().Str("type", "producer").Str("exchange", p.exchange.name).Msg("attempting to open channel")
	p.channel, err = c.conn.Channel()
	log.Err(err).Str("type", "producer").Str("exchange", p.exchange.name).Msg("failed to open channel")
	return err
}

// Declare exchange this producer will produce to
func (p *RabbitProducer) declareExchange(exchange *Exchange) error {
	err := p.channel.ExchangeDeclare(
		exchange.name,
		exchange.exchangeType.String(),
		exchange.durable,
		exchange.autoDelete,
		exchange.internal,
		exchange.noWait,
		exchange.args,
	)
	return err
}

// Subscribe to channel close events and make sure to respond to them
func (p *RabbitProducer) listenForClose() {
	closeChan := p.channel.NotifyClose(make(chan *amqp.Error))
	go func() {
		closeErr := <-closeChan
		log.Error().Str("type", "producer").AnErr("err", closeErr).Str("exchange", p.exchange.name).Msg("connection was closed")
	}()
}

// Listen for flow messages from the broker
// If a flow message comes in we need to do some rate limiting
func (p *RabbitProducer) listenForFlow() {
	flowChan := p.channel.NotifyFlow(make(chan bool))
	go func() {
		for {
			<-flowChan
			err := ProducerError{
				producer:    p,
				err:         errors.New("Too many messages being produced"),
				status:      505,
				recoverable: true,
			}
			p.errorHandler(err)
		}
		log.Error().Str("type", "producer").Str("exchange", p.exchange.name).Msg("too many messages being produced")
	}()
}

// Listen for a returned message from the broker
// Will only be called if the mandatory flag is set and there is no queue bound, or the immediate flag is set and there is no free consumer
func (p *RabbitProducer) listenForReturnedMessages() {
	returnedMessageChan := p.channel.NotifyReturn(make(chan amqp.Return))
	go func() {
		for {
			returnedMsg := <-returnedMessageChan
			err := ProducerError{
				producer: p,
				err:      errors.New("Message was returned from the exchange"),
				status:   100,
				returned: returnedMsg,
			}
			p.errorHandler(err)
			log.Error().Str("type", "producer").Str("exchange", p.exchange.name).Interface("msg", returnedMsg).Msg("message was returned")
		}
	}()
}

// PublishMessage publishes a message with the given routing key
func (p *RabbitProducer) PublishMessage(msg []byte, key *string, headers *amqp.Table) {

	log.Trace().Str("type", "producer").Str("routingKey", *key).Str("exchange", p.exchange.name).Int("msgSize", len(msg)).Msg("producing message")

	err := p.channel.Publish(
		p.exchange.name,
		*key,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "plaintext",
			Body:         msg,
			Timestamp:    time.Now(),
			Headers:      *headers,
		},
	)
	log.Err(err).Str("type", "producer").Str("routingKey", *key).Str("exchange", p.exchange.name).Msg("error during message production")
}

// ReconnectChannel tries to re-open this producer's channel
func (p *RabbitProducer) ReconnectChannel() {
	p.openChannel(p.conn)
}

// Shutdown closes this producer's channel
func (p *RabbitProducer) Shutdown() error {
	log.Info().Str("type", "producer").Str("exchange", p.exchange.name).Msg("shutting down")
	return p.channel.Close()
}

// ProducerError are errors the producer can throw and which need to be handled by the producer error handler
type ProducerError struct {
	producer    *RabbitProducer // The producer that had this error
	err         error           // The actual error that occurred
	status      int             // Status code belonging to this error
	returned    amqp.Return     // Message that this error is about
	recoverable bool            // Whether this error is recoverable by retrying at a later moment
}

// MARK: Producer errors

func (pe *ProducerError) Error() string {
	return fmt.Sprintf("producer error: %v", pe.err)
}

// DefaultProducerErrorHandler is the default producer error handler
func DefaultProducerErrorHandler(err ProducerError) {
	switch err.status {
	case 504: // Error while publishing message
		// Publishing channel or RabbitMQ connection not open
		logMessage(err.Error())
	case 320: // Error thrown on NotifyClose
		// Channel or connection was closed
		logMessage(err.Error())

		// If this is recoverable try and recover
		if err.recoverable {
			err.producer.ReconnectChannel()
		}
	case 300: // Error opening channel
		logMessage(err.Error())
	case 202: // Error declaring exchange
		logMessage(err.Error())
	case 505: // Too many messages being published, rate limit messages
		// TODO: Rate limit messages for some time
		logMessage(err.Error())
	case 100: // Message was returned from the broker due to being undeliverable
		// TODO: Resend message?
		logMessage(err.Error())
	default:
		logMessage(err.Error())
	}
}
