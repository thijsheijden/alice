package alice

import (
	"fmt"

	"github.com/streadway/amqp"
)

// LogMessages is a boolean that determines whether Alice logs messages
var LogMessages bool = true

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

func logMessage(msg string) {
	if LogMessages {
		fmt.Println(msg)
	}
}

func failWithError(err error, msg string) {
	fmt.Println(msg)
	panic(err)
}

func logError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %v\n", msg, err)
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
