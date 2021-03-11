package alice

import "github.com/streadway/amqp"

type Queue struct {
	name       string
	durable    bool
	exclusive  bool
	autoDelete bool
	noWait     bool
	arguments  amqp.Table
}

//Creates a default Queue with given name and durable: true, exclusive: false, autoDelete: false and arguments: nil
func CreateDefaultQueue(name string) *Queue {
	q := &Queue{
		name:       name,
		durable:    true,
		exclusive:  false,
		autoDelete: false,
		noWait:     false,
		arguments:  nil,
	}
	return q
}

//Creates a fully customised queue
func CreateQueue(name string, durable bool, exclusive bool, autoDelete bool, noWait bool, arguments amqp.Table) *Queue {
	q := &Queue{
		name:       name,
		durable:    durable,
		exclusive:  exclusive,
		autoDelete: autoDelete,
		noWait:     noWait,
		arguments:  arguments,
	}
	return q
}
