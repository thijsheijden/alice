package alice

import (
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
)

// Producer models a RabbitMQ producer
type Producer struct {
	channel  *amqp.Channel // The channel this producer uses to communicate with the broker
	exchange *Exchange     // The exchange this producer produces to
}

// CreateProducer creates and returns a producer attached to the given exchange
func (c *Connection) CreateProducer(exchange Exchange) *Producer {
	var err error

	p := &Producer{
		channel:  nil,
		exchange: &exchange,
	}

	p.channel, err = c.conn.Channel()
	logError(err, "Could not open producer channel")

	err = p.channel.ExchangeDeclare(
		exchange.name,
		exchange.exchangeType.String(),
		exchange.durable,
		exchange.autoDelete,
		exchange.internal,
		exchange.noWait,
		exchange.args,
	)
	logError(err, "Failed to declare producer exchange")

	return p
}

// PublishMessage publishes a message with the given routing key
func (p *Producer) PublishMessage(msg interface{}, key string) {

	m, _ := json.Marshal(msg)

	err := p.channel.Publish(
		p.exchange.name,
		key,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "plaintext",
			Body:         m,
			Timestamp:    time.Now(),
			Headers:      nil,
		},
	)
	logError(err, "Failed to publish message")
}

// PublishNMessages publishes n messages
func (p *Producer) PublishNMessages(n int) {
	for i := 0; i < n; i++ {
		p.PublishMessage("Honk!", "test")
	}
}

// Shutdown closes this producer's channel
func (p *Producer) Shutdown() {
	p.channel.Close()
}
