package alice

import "github.com/streadway/amqp"

// A Broker models a broker
type Broker interface {
	CreateConsumer(queue *Queue, bindingKey string, consumerTag string, errorHandler func(error)) (Consumer, error)
	CreateProducer(exchange *Exchange, errorHandler func(ProducerError)) (Producer, error)
}

// A Consumer models a broker consumer
type Consumer interface {
	ConsumeMessages(args amqp.Table, messageHandler func(amqp.Delivery))
	Shutdown() error
}

// A Producer models a broker producer
type Producer interface {
	PublishMessage(msg []byte, key *string, headers *amqp.Table)
	Shutdown() error
}
