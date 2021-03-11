package alice

import (
	"fmt"

	"github.com/streadway/amqp"
)

// LogMessages is a boolean that determines whether whiterabbit logs messages
// Primarily useful during debugging.
var LogMessages bool = true

// DefaultErrorHandler is the default whiterabbit error handler
func DefaultErrorHandler(errs <-chan error) {
	for err := range errs {
		if err == amqp.ErrCredentials {
			failWithError(err)
		}
	}
}

func logMessage(msg string) {
	if LogMessages {
		fmt.Println(msg)
	}
}

func failWithError(err error) {
	panic(err)
}

func logError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %v\n", msg, err)
	}
}
