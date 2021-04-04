package alice

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

// A MockProducer is a mocked producer
type MockProducer struct {
	exchange *Exchange
	broker   *MockBroker
}

// PublishMessage publishes a message
func (p *MockProducer) PublishMessage(msg interface{}, key *string, headers *amqp.Table) {
	// Find the queues this message was meant for
	var queuesToSendTo []*Queue = make([]*Queue, 0, 10)
	for _, q := range p.broker.exchanges[p.exchange] {
		if q.bindingKey == *key {
			queuesToSendTo = append(queuesToSendTo, q)
		}
	}

	d, _ := json.Marshal(msg)

	delivery := amqp.Delivery{
		Headers:         *headers,
		ContentType:     "",
		ContentEncoding: "",
		Body:            d,
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
