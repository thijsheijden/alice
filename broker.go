package alice

import (
	"fmt"

	"github.com/streadway/amqp"
)

// A RabbitBroker is a rabbitmq broker
type RabbitBroker struct {
	config       *ConnectionConfig
	consumerConn *Connection
	producerConn *Connection
}

// CreateBroker creates a RabbitMQ broker
func CreateBroker(config *ConnectionConfig) (*RabbitBroker, error) {
	broker := RabbitBroker{
		config: config,
	}

	// Test connection
	conn, err := broker.connect()
	if err != nil {
		return nil, err
	}

	// Close connection as it is not needed right now
	conn.conn.Close()
	return &broker, nil
}

// CreateConsumer creates a consumer consuming from the specified queue using the specified key
func (b *RabbitBroker) CreateConsumer(queue *Queue, bindingKey string, errorHandler func(error)) (Consumer, error) {
	if b.consumerConn == nil {
		b.consumerConn, _ = b.connect()
	}

	return b.consumerConn.CreateConsumer(queue, bindingKey, errorHandler)
}

// CreateProducer creates a producer using this broker
func (b *RabbitBroker) CreateProducer(exchange *Exchange, errorHandler func(ProducerError)) (Producer, error) {
	if b.producerConn == nil {
		b.producerConn, _ = b.connect()
	}

	return b.producerConn.CreateProducer(exchange, errorHandler)
}

func (b *RabbitBroker) connect() (*Connection, error) {
	var err error

	config := *b.config

	connection := &Connection{
		conn:         nil,
		errorHandler: config.errorHandler,
		config:       config,
	}

	amqpURI := "amqp://" + config.user + ":" + config.password + "@" + config.host + ":" + fmt.Sprint(config.port)

	// Create a buffered done channel to fill when connection is established
	done := make(chan error, 1)

	// Go create a connection
	go func() {
		connection.conn, err = amqp.Dial(amqpURI)
		done <- err
	}()
	err = <-done
	if err != nil {
		return nil, err
	}

	// Handle the closing notifications
	if config.autoReconnect {
		go connection.reconnect(connection.conn.NotifyClose(make(chan *amqp.Error)))
	} else { // No reconnecting so just send this error into the error channel?

	}

	return connection, nil
}
