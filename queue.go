package alice

import "github.com/streadway/amqp"

// Queue models a RabbitMQ queue
type Queue struct {
	exchange   *Exchange  // The exchange this queue will be bound to
	name       string     // Name of the queue
	durable    bool       // Does this queue persist during broker restarts?
	exclusive  bool       // Is this queue exclusive to one consumer? This also sets autoDelete to true.
	autoDelete bool       // Does this queue get deleted when no-one is consuming from it?
	noWait     bool       // Should we skip waiting for an acknowledgement from the broker?
	bindingKey string     // The key with which this queue is bound to the exchange
	args       amqp.Table // Additional amqp arguments to configure the exchange
}

// CreateDefaultQueue creates and returns a queue with the following parameters:
//	durable: true, exclusive, false, autoDelete: false, noWait: false, args: nil
func CreateDefaultQueue(exchange *Exchange, name string) *Queue {
	return CreateQueue(exchange, name, true, false, false, false, nil)
}

// CreateQueue returns a queue created with the accompanied specifications
func CreateQueue(exchange *Exchange, name string, durable bool, exclusive bool, autoDelete bool, noWait bool, arguments amqp.Table) *Queue {
	q := &Queue{
		exchange:   exchange,
		name:       name,
		durable:    durable,
		exclusive:  exclusive,
		autoDelete: autoDelete,
		noWait:     noWait,
		args:       arguments,
	}
	return q
}

// SetName sets the queue name
func (q *Queue) SetName(name string) {
	q.name = name
}

// SetDurable set exchange durability
func (q *Queue) SetDurable(durable bool) {
	q.durable = durable
}

// SetExclusive sets queue exclusivity
func (q *Queue) SetExclusive(exclusive bool) {
	q.exclusive = exclusive
}

// SetAutoDelete sets queue auto deletion
func (q *Queue) SetAutoDelete(autoDelete bool) {
	q.autoDelete = autoDelete
}

// SetNoWait sets the noWait flag
func (q *Queue) SetNoWait(noWait bool) {
	q.noWait = noWait
}

// SetArgs sets additional queue arguments
func (q *Queue) SetArgs(args amqp.Table) {
	q.args = args
}
