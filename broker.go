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
func CreateBroker(config *ConnectionConfig) *RabbitBroker {
	return &RabbitBroker{
		config: config,
	}
}

// CreateConsumer creates a consumer consuming from the specified queue using the specified key
func (b *RabbitBroker) CreateConsumer(queue *Queue, bindingKey string, errorHandler func(error)) (Consumer, error) {
	if b.consumerConn == nil {
		b.consumerConn = b.connect()
	}

	return b.consumerConn.CreateConsumer(queue, bindingKey, errorHandler)
}

// CreateProducer creates a producer using this broker
func (b *RabbitBroker) CreateProducer(exchange *Exchange, errorHandler func(ProducerError)) (Producer, error) {
	if b.producerConn == nil {
		b.producerConn = b.connect()
	}

	return b.producerConn.CreateProducer(exchange, errorHandler)
}

func (b *RabbitBroker) connect() *Connection {
	var err error

	config := *b.config

	connection := &Connection{
		conn:         nil,
		errorHandler: config.errorHandler,
		config:       config,
	}

	amqpURI := "amqp://" + config.user + ":" + config.password + "@" + config.host + ":" + fmt.Sprint(config.port)

	// Create a buffered done channel to fill when connection is established
	done := make(chan bool, 1)

	// Go create a connection
	go func() {
		connection.conn, err = amqp.Dial(amqpURI)
		connection.errorHandler(err)

		// If there is no error continue
		if err == nil {
			done <- true
		}
	}()
	<-done

	// Handle the closing notifications
	if config.autoReconnect {
		go connection.reconnect(connection.conn.NotifyClose(make(chan *amqp.Error)))
	} else { // No reconnecting so just send this error into the error channel?

	}

	return connection
}
