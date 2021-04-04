package alice

import (
	"github.com/streadway/amqp"
)

type MockConsumer struct {
	queue            *Queue
	broker           *MockBroker
	ReceivedMessages []amqp.Delivery
}

// ConsumeMessages consumes messages sent to the consumer
func (c *MockConsumer) ConsumeMessages(args amqp.Table, autoAck bool, messageHandler func(amqp.Delivery)) {
	for msg := range c.broker.Messages[c.queue] {
		c.ReceivedMessages = append(c.ReceivedMessages, msg)
		go messageHandler(msg)
	}
}

// Shutdown shuts down the consumer
func (c *MockConsumer) Shutdown() error {
	return nil
}
