package alice

import (
	"fmt"

	"github.com/streadway/amqp"
)

// Connection models a RabbitMQ connection
type Connection struct {
	conn         *amqp.Connection
	errorChan    chan *amqp.Error
	errorHandler func(chan *amqp.Error)
}

// Connect connects to the RabbitMQ broker using the supplied config
func Connect(config ConnectionConfig) *Connection {

	var err error

	connection := &Connection{
		conn:         nil,
		errorChan:    make(chan *amqp.Error),
		errorHandler: nil,
	}

	amqpURI := "amqp://" + config.amqpUser + ":" + config.amqpPassword + "@" + config.amqpHost + ":" + fmt.Sprint(config.amqpPort)

	connection.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		fmt.Println(err)
	}

	connection.errorHandler = config.errorHandler

	// Handle the closing notifications
	go func() {
		connection.conn.NotifyClose(connection.errorChan)
	}()

	return connection
}
