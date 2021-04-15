package alice

import (
	"errors"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// RabbitProducer models a RabbitMQ producer
type RabbitProducer struct {
	channel      *amqp.Channel       // The channel this producer uses to communicate with the broker
	exchange     *Exchange           // The exchange this producer produces to
	errorHandler func(ProducerError) // Error handler for this producer
	conn         *Connection         // Pointer to broker connection
}

// CreateProducer creates and returns a producer attached to the given exchange.
// The errorHandler can be the DefaultProducerErrorHandler or a custom handler.
func (c *Connection) CreateProducer(exchange *Exchange, errorHandler func(ProducerError)) (*RabbitProducer, error) {

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

	logMessage(fmt.Sprintf("Created producer on exchange '%s'", exchange.name))

	return p, nil
}

// Open channel to broker
func (p *RabbitProducer) openChannel(c *Connection) error {
	var err error
	p.channel, err = c.conn.Channel()
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
		closeError := <-closeChan
		err := ProducerError{
			producer:    p,
			err:         closeError,
			status:      320,
			recoverable: closeError.Recover,
		}
		p.errorHandler(err)
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
		}
	}()
}

// PublishMessage publishes a message with the given routing key
func (p *RabbitProducer) PublishMessage(msg []byte, key *string, headers *amqp.Table) {

	logMessage(fmt.Sprintf("Publishing message of size '%d' bytes to exchange '%s' with routing key '%s'", len(msg), p.exchange.name, *key))

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
	if err != nil {
		p.errorHandler(ProducerError{producer: p, err: err, status: 504}) // Error can only be channel or connection not being open
	}
}

// ReconnectChannel tries to re-open this producer's channel
func (p *RabbitProducer) ReconnectChannel() {
	p.openChannel(p.conn)
}

// Shutdown closes this producer's channel
func (p *RabbitProducer) Shutdown() error {
	return p.channel.Close()
}
