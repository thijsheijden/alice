package alice

import (
	"fmt"
	"runtime"
	"time"

	"github.com/streadway/amqp"
)

// Connection models a RabbitMQ connection
type Connection struct {
	conn         *amqp.Connection // The connection to the RabbitMQ broker
	errorHandler func(error)      // The error handler for this connection
	config       ConnectionConfig // Configuration for connection
}

// Connect connects to the RabbitMQ broker using the supplied configuration
func Connect(config ConnectionConfig) *Connection {

	var err error

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

// Shutdown shuts down the connection to rabbitmq
func (c *Connection) Shutdown() {
	c.conn.Close()
}

// Handle automatic restarting on connection closed
func (c *Connection) reconnect(ch chan *amqp.Error) {
	<-ch // Connection was closed for some reason

	// Create new ticker with the desired connection delay time
	ticker := time.NewTicker(c.config.reconnectDelay)

	for {
		<-ticker.C // New tick

		var err error

		logMessage("Attempting to reconnect to RabbitMQ")
		c.conn, err = amqp.Dial("amqp://" + c.config.user + ":" + c.config.password + "@" + c.config.host + ":" + fmt.Sprint(c.config.port))

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
