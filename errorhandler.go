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
		log.Println(fmt.Sprintf("ALICE: %s", msg))
	}
}

func failWithError(err error, msg string) {
	log.Println(err)
	panic(err)
}

func logError(err error, msg string) {
	if err != nil {
		log.Printf("ALICE: ERROR: %s: %v\n", msg, err)
	}
}
