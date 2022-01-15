package alice

import (
	"time"
)

// ConnectionConfig is a config structure to use when setting up a RabbitMQ connection
type ConnectionConfig struct {
	user           string        // Username for connection
	password       string        // Password for connection
	host           string        // URI for RabbitMQ broker
	port           int           // RabbitMQ broker port
	autoReconnect  bool          // Whether to try to reconnect after a unexpected disconnect
	reconnectDelay time.Duration // The delay between reconnection attempts
}

// DefaultConfig is the default configuration for RabbitMQ.
//	User: "guest", password: "guest", host: "localhost", port: 5672, autoReconnect: true, reconnectDelay: time.Second * 10
var DefaultConfig = &ConnectionConfig{
	user:           "guest",
	password:       "guest",
	host:           "localhost",
	port:           5672,
	autoReconnect:  true,
	reconnectDelay: time.Second * 10,
}

// CreateConfig creates a connection configuration with the supplied parameters
func CreateConfig(user string, password string, host string, port int, autoReconnect bool, reconnectDelay time.Duration) *ConnectionConfig {
	config := &ConnectionConfig{
		user:           user,
		password:       password,
		host:           host,
		port:           port,
		autoReconnect:  autoReconnect,
		reconnectDelay: reconnectDelay,
	}
	return config
}

// SetUser sets the user to use for the connection to the broker
func (config *ConnectionConfig) SetUser(user string) {
	config.user = user
}

// SetPassword sets the password to use for the connection to the broker
func (config *ConnectionConfig) SetPassword(password string) {
	config.password = password
}

// SetHost sets the broker connection host URI
func (config *ConnectionConfig) SetHost(host string) {
	config.host = host
}

// SetPort sets the broker connection port
func (config *ConnectionConfig) SetPort(port int) {
	config.port = port
}

// SetAutoReconnect sets whether the connection should try reconnecting
func (config *ConnectionConfig) SetAutoReconnect(autoReconnect bool) {
	config.autoReconnect = autoReconnect
}

// SetReconnectDelay sets the delay between reconnection attempts
func (config *ConnectionConfig) SetReconnectDelay(reconnectDelay time.Duration) {
	config.reconnectDelay = reconnectDelay
}
