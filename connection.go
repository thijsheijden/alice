package alice

import (
	"fmt"
	"runtime"
	"time"

	"github.com/streadway/amqp"
)

// connection models a RabbitMQ connection
type connection struct {
	conn         *amqp.Connection // The connection to the RabbitMQ broker
	errorHandler func(error)      // The error handler for this connection
	config       ConnectionConfig // Configuration for connection
}

// shutdown shuts down the connection to rabbitmq
func (c *connection) shutdown() {
	c.conn.Close()
}

// Handle automatic restarting on connection closed
func (c *connection) reconnect(ch chan *amqp.Error) {
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
