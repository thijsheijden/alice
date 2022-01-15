package alice

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
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
// t is either "consumer" or "producer"
func (c *connection) reconnect(t string, ch chan *amqp.Error) {
	err := <-ch // Connection was closed for some reason

	log.Err(err).Str("connType", t).Msg("connection was closed")

	// Create new ticker with the desired connection delay time
	ticker := time.NewTicker(c.config.reconnectDelay)

	for {
		<-ticker.C // New tick

		var err error

		log.Info().Str("connType", t).Msg("attempting to reconnect")
		c.conn, err = amqp.Dial("amqp://" + c.config.user + ":" + c.config.password + "@" + c.config.host + ":" + fmt.Sprint(c.config.port))

		if err != nil {
			log.Error().AnErr("err", err).Str("connType", t).Msg("failed to reconnect")
		}

		if !c.conn.IsClosed() {
			go c.reconnect(t, c.conn.NotifyClose(make(chan *amqp.Error)))
			ticker.Stop()
			break
		}
	}

	log.Info().Str("connType", t).Msg("successfully reconnected")
}
