package whiterabbit

import "github.com/streadway/amqp"

// Connection models a RabbitMQ connection
type Connection struct {
	conn         *amqp.Connection
	notifyClose  chan *amqp.Error
	errorHandler func(chan *amqp.Error)
}
