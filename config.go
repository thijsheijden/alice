package alice

import (
	"time"
)

// ConnectionConfig is a config structure to use when setting up a RabbitMQ connection
type ConnectionConfig struct {
	AmqpUser       string             // Username for connection
	AmqpPassword   string             // Password for connection
	AmqpHost       string             // URI for RabbitMQ broker
	AmqpPort       int                // RabbitMQ broker port
	AutoReconnect  bool               // Whether to try to reconnect after a unexpected disconnect
	ReconnectDelay time.Duration      // The delay between reconnection attempts
	ErrorHandler   func(<-chan error) // The error handler for the connection
}

// DefaultConfig is the default configuration for RabbitMQ.
//	User: "guest", password: "guest", host: "localhost", port: 5672, autoReconnect: true, reconnectDelay: time.Second * 10
var DefaultConfig = &ConnectionConfig{
	AmqpUser:       "guest",
	AmqpPassword:   "guest",
	AmqpHost:       "localhost",
	AmqpPort:       5672,
	AutoReconnect:  true,
	ReconnectDelay: time.Second * 10,
	ErrorHandler:   DefaultErrorHandler,
}
