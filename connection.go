package alice

import (
	"fmt"
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
func (c *connection) reconnect(t string, ch chan *amqp.Error) {
	err := <-ch // Connection was closed for some reason

	logError(err, fmt.Sprintf("%s connection was closed", t))

	// Create new ticker with the desired connection delay time
	ticker := time.NewTicker(c.config.reconnectDelay)

	for {
		<-ticker.C // New tick

		var err error

		logMessage(fmt.Sprintf("Attempting to re-open %s connection", t))
		c.conn, err = amqp.Dial("amqp://" + c.config.user + ":" + c.config.password + "@" + c.config.host + ":" + fmt.Sprint(c.config.port))

		logError(err, fmt.Sprintf("Failed to re-open %s connection", t))

		if !c.conn.IsClosed() {
			go c.reconnect(t, c.conn.NotifyClose(make(chan *amqp.Error)))
			ticker.Stop()
			break
		}
	}

	logMessage(fmt.Sprintf("Successfully re-opened %s connection", t))
}
