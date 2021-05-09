package alice

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// LogMessages is a boolean that determines whether Alice logs messages
var logMessages bool = false

// DefaultErrorHandler is the default Connection error handler
func DefaultErrorHandler(err error) {
	switch err {
	case amqp.ErrCredentials:
		failWithError(err, "Invalid credentials")
	case nil:

	default:
		failWithError(err, "Error occurred")
	}
}

// SetLogging turns logging on
func SetLogging() {
	logMessages = true
}

func logMessage(msg string) {
	if logMessages {
		log.Println(msg)
	}
}

func failWithError(err error, msg string) {
	log.Println(err)
	panic(err)
}

func logError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %v\n", msg, err)
	}
}

// Error is an error
type Error struct {
	Reason      string
	Recoverable string
}

var (
	// ErrConnectionRefused connection was refused
	ErrConnectionRefused = Error{Reason: ""}
)

// MARK: Producer errors

// ProducerError are errors the producer can throw and which need to be handled by the producer error handler
type ProducerError struct {
	producer    *RabbitProducer // The producer that had this error
	err         error           // The actual error that occurred
	status      int             // Status code belonging to this error
	returned    amqp.Return     // Message that this error is about
	recoverable bool            // Whether this error is recoverable by retrying at a later moment
}

func (pe *ProducerError) Error() string {
	return fmt.Sprintf("producer error: %v", pe.err)
}

// DefaultProducerErrorHandler is the default producer error handler
func DefaultProducerErrorHandler(err ProducerError) {
	switch err.status {
	case 504: // Error while publishing message
		// Publishing channel or RabbitMQ connection not open
		logMessage(err.Error())
	case 320: // Error thrown on NotifyClose
		// Channel or connection was closed
		logMessage(err.Error())

		// If this is recoverable try and recover
		if err.recoverable {
			err.producer.ReconnectChannel()
		}
	case 300: // Error opening channel
		logMessage(err.Error())
	case 202: // Error declaring exchange
		logMessage(err.Error())
	case 505: // Too many messages being published, rate limit messages
		// TODO: Rate limit messages for some time
		logMessage(err.Error())
	case 100: // Message was returned from the broker due to being undeliverable
		// TODO: Resend message?
		logMessage(err.Error())
	default:
		logMessage(err.Error())
	}
}
