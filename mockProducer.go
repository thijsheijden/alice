package alice

import (
	"github.com/streadway/amqp"
)

// A MockProducer implements the Producer interface
type MockProducer struct {
	exchange *Exchange
	broker   *MockBroker
}

// PublishMessage publishes a message
func (p *MockProducer) PublishMessage(msg []byte, key *string, headers *amqp.Table) {
	// Find the queues this message was meant for
	var queuesToSendTo []*Queue = make([]*Queue, 0, 10)
	for _, q := range p.broker.exchanges[p.exchange] {
		if q.bindingKey == *key {
			queuesToSendTo = append(queuesToSendTo, q)
		}
	}

	delivery := amqp.Delivery{
		Headers:         *headers,
		ContentType:     "",
		ContentEncoding: "",
		Body:            msg,
	}

	// Send message to the queues
	for _, q := range queuesToSendTo {
		p.broker.Messages[q] <- delivery
	}
}

// Shutdown shuts this producer down
func (p *MockProducer) Shutdown() error {
	return nil
}
