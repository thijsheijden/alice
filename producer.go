package alice

import (
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
)

// Producer models a RabbitMQ producer
type Producer struct {
	channel  *amqp.Channel
	exchange *Exchange
}

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
		exchange.autodelete,
		exchange.internal,
		exchange.noWait,
		exchange.args,
	)
	logError(err, "Failed to declare producer exchange")

	return p
}

func (p *Producer) ProduceMessage(msg string, key string) {

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

func (p *Producer) Produce1000Messages() {
	for i := 0; i < 1000000; i++ {
		p.ProduceMessage("Honk!", "test")
	}
}

func (p *Producer) DeferShutdown() {
	p.channel.Close()
}
