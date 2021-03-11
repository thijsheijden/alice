package alice

import "github.com/streadway/amqp"

// Exchange models a RabbitMQ exchange
type Exchange struct {
	name         string       // Name of the exchange
	exchangeType ExchangeType // Type of the exchange
	durable      bool
	autodelete   bool
	internal     bool
	noWait       bool
	args         amqp.Table
}

// ExchangeType denotes the types of exchanges RabbitMQ has
type ExchangeType string

const (
	// Direct delivers messages to queues based on the message routing key
	Direct ExchangeType = "direct"

	// Fanout delivers messages to all connected queues
	Fanout ExchangeType = "fanout"

	// Topic delivers messages based on the matching between the message routing key and the pattern used to bind queues to the exchange
	Topic ExchangeType = "topic"

	// Headers delivers messages based on header values, similar to direct routing
	Headers ExchangeType = "headers"
)

func (t ExchangeType) String() string {
	return string(t)
}

// CreateDefaultExchange returns an exchange with the following specifications:
//	durable: true, autodelete: false, internal: false, noWait: false, args: nil
func CreateDefaultExchange(name string, exchangeType ExchangeType) *Exchange {
	e := &Exchange{
		name:         name,
		exchangeType: exchangeType,
		durable:      true,
		autodelete:   false,
		internal:     false,
		noWait:       false,
		args:         nil,
	}

	return e
}

// CreateExchange creates an exchange according to the specified arguments
func CreateExchange(name string, exchangeType ExchangeType, durable bool, autodelete bool, internal bool, noWait bool, args amqp.Table) *Exchange {
	e := &Exchange{
		name:         name,
		exchangeType: exchangeType,
		durable:      durable,
		autodelete:   autodelete,
		internal:     internal,
		noWait:       noWait,
		args:         args,
	}

	return e
}
