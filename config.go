package alice

import (
	"time"
)

// ConnectionConfig is a config structure to use when setting up a RabbitMQ connection
type ConnectionConfig struct {
	AmqpUser       string
	AmqpPassword   string
	AmqpHost       string
	AmqpPort       int
	AutoReconnect  bool
	ReconnectDelay time.Duration
	ErrorHandler   func(<-chan error)
}

// DefaultConfig is the default configuration for RabbitMQ.
// User: guest, password: guest, host: localhost, port: 5672, autoReconnect: true, reconnectDelay: 5 seconds.
var DefaultConfig = &ConnectionConfig{
	AmqpUser:       "guest",
	AmqpPassword:   "guest",
	AmqpHost:       "localhost",
	AmqpPort:       5672,
	AutoReconnect:  true,
	ReconnectDelay: time.Second * 1,
	ErrorHandler:   DefaultErrorHandler,
}
