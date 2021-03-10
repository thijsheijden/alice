package whiterabbit

import "github.com/streadway/amqp"

// LogMessages is a boolean that determines whether whiterabbit logs messages
// Primarily useful during debugging.
var LogMessages bool = true

// DefaultErrorHandler is the default whiterabbit error handler
func DefaultErrorHandler(err *amqp.Error) {

}
