package alice

import (
	"fmt"
	"time"

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

	// Check if reconnect is turned on
	if config.autoReconnect {
		// Create a ticker with the reconnect delay
		ticker := time.NewTicker(config.reconnectDelay)

		// Create done channel (for when the connection is successfully established)
		done := make(chan bool, 1)

		for {
			select {
			case <-ticker.C:
				logMessage("Attempting RabbitMQ connection")

				// Attempt to connect to the broker
				conn, err := broker.connect()

				// If there is no error
				if err == nil {
					// Close the connection for now
					conn.conn.Close()

					// Stop the ticker
					ticker.Stop()

					// Signal that we are done
					done <- true
				}
			case <-done:
				logMessage("Succesfully connected to RabbitMQ broker")
				return &broker, nil
			}
		}
	} else {
		// Test connection
		conn, err := broker.connect()
		if err != nil {
			return nil, err
		}

		// Close connection as it is not needed right now
		conn.conn.Close()
		return &broker, nil
	}
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
	}

	return connection, nil
}
