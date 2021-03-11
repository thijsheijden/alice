package alice

import (
	"fmt"
	"runtime"
	"time"

	"github.com/streadway/amqp"
)

// Connection models a RabbitMQ connection
type Connection struct {
	conn         *amqp.Connection
	errorChan    chan error
	errorHandler func(<-chan error)
	config       ConnectionConfig
}

// Connect connects to the RabbitMQ broker using the supplied config
func Connect(config ConnectionConfig) *Connection {

	var err error

	connection := &Connection{
		conn:         nil,
		errorChan:    make(chan error),
		errorHandler: config.ErrorHandler,
		config:       config,
	}

	// Go handle errors
	go connection.errorHandler(connection.errorChan)

	amqpURI := "amqp://" + config.AmqpUser + ":" + config.AmqpPassword + "@" + config.AmqpHost + ":" + fmt.Sprint(config.AmqpPort)

	connection.conn, err = amqp.Dial(amqpURI)
	connection.errorChan <- err

	// Handle the closing notifications
	if config.AutoReconnect {
		go connection.reconnect(connection.conn.NotifyClose(make(chan *amqp.Error)))
	} else { // No reconnecting so just send this error into the error channel?

	}

	return connection
}

// Shutdown shuts down the connection to rabbitmq
func (c *Connection) Shutdown() {
	c.conn.Close()
}

// Handle automatic restarting on connection closed
func (c *Connection) reconnect(ch chan *amqp.Error) {
	<-ch // Connection was closed for some reason

	// Create new ticker with the desired connection delay time
	ticker := time.NewTicker(c.config.ReconnectDelay)

	for {
		<-ticker.C // New tick

		var err error

		logMessage("Attempting to reconnect to RabbitMQ")
		c.conn, err = amqp.Dial("amqp://" + c.config.AmqpUser + ":" + c.config.AmqpPassword + "@" + c.config.AmqpHost + ":" + fmt.Sprint(c.config.AmqpPort))

		logError(err, "Failed to reconnect")

		if !c.conn.IsClosed() {
			go c.reconnect(c.conn.NotifyClose(make(chan *amqp.Error)))
			ticker.Stop()
			break
		}
	}

	logMessage("Reconnected")
	logMessage(fmt.Sprint(runtime.NumGoroutine()))
}
