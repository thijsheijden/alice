package alice

import (
	"github.com/streadway/amqp"
)

type MockBroker struct {
	exchanges map[*Exchange][]*Queue        // The exchanges bound to this broker, with their bound queues
	Messages  map[*Queue]chan amqp.Delivery // The messages sent in a queue
}

func CreateMockBroker() *MockBroker {
	return &MockBroker{
		exchanges: make(map[*Exchange][]*Queue),
		Messages:  make(map[*Queue]chan amqp.Delivery),
	}
}

func (b *MockBroker) CreateConsumer(queue *Queue, bindingKey string, errorHandler func(error)) (Consumer, error) {
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

func (b *MockBroker) CreateProducer(exchange *Exchange, errorHandler func(ProducerError)) (Producer, error) {
	p := &MockProducer{
		exchange: exchange,
		broker:   b,
	}

	return p, nil
}
