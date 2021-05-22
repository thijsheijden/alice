package alice

import (
	"github.com/streadway/amqp"
)

// MockBroker implements the Broker interface (mock)
type MockBroker struct {
	exchanges map[*Exchange][]*Queue        // The exchanges bound to this broker, with their bound queues
	Messages  map[*Queue]chan amqp.Delivery // The messages sent in a queue
}

// CreateMockBroker creates a new MockBroker (mock)
func CreateMockBroker() Broker {
	return &MockBroker{
		exchanges: make(map[*Exchange][]*Queue),
		Messages:  make(map[*Queue]chan amqp.Delivery),
	}
}

// CreateConsumer creates a new consumer (mock)
func (b *MockBroker) CreateConsumer(queue *Queue, bindingKey string, consumerTag string, errorHandler func(error)) (Consumer, error) {
	queue.bindingKey = bindingKey
	c := &MockConsumer{
		queue:            queue,
		broker:           b,
		ReceivedMessages: make([]amqp.Delivery, 0),
	}

	// Add this queue to this exchange
	b.exchanges[queue.exchange] = append(b.exchanges[queue.exchange], queue)
	b.Messages[queue] = make(chan amqp.Delivery, 0)

	return c, nil
}

// CreateProducer creates a new producer (mock)
func (b *MockBroker) CreateProducer(exchange *Exchange, errorHandler func(ProducerError)) (Producer, error) {
	p := &MockProducer{
		exchange: exchange,
		broker:   b,
	}

	return p, nil
}
