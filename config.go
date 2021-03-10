package whiterabbit

import "time"

// RabbitConfig is a config structure to use when setting up a RabbitMQ connection
type RabbitConfig struct {
	amqpUser       string
	amqpPassword   string
	amqpHost       string
	amqpPort       int
	autoReconnect  bool
	reconnectDelay time.Duration
}

// DefaultConfig is the default configuration for RabbitMQ.
// User: guest, password: guest, host: localhost, port: 5672, autoReconnect: true, reconnectDelay: 5 seconds.
var DefaultConfig = &RabbitConfig{
	amqpUser:       "guest",
	amqpPassword:   "guest",
	amqpHost:       "localhost",
	amqpPort:       5672,
	autoReconnect:  true,
	reconnectDelay: time.Second * 5,
}
