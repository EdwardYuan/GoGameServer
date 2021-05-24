package lib

import (
	"github.com/streadway/amqp"
)

// RabbitClient hold the connection, exchanges and queues of RabbitMQ
type RabbitClient struct {
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	Exchanges map[string]string
	Queues    map[string]*amqp.Queue
}

func NewRabbitClient() *RabbitClient {
	return &RabbitClient{}
}

//Start makes a connection of RabbitMQ and a channel, then declare an exchange and a queue
//it receives four arguments: url represents RabbitMQ host, xName represents the exchange
//name, qName means the queue name and the xType defines the type of exchange.
func (r *RabbitClient) Start(url string, xName string, qName string, xType string) error {
	var err error
	r.Conn, err = amqp.Dial(url)
	SysLoggerFatal(err, "Connect to RabbitMQ Error")

	r.Channel, err = r.Conn.Channel()
	SysLoggerFatal(err, "RabbitMQ New Channel Error")
	err = r.Channel.ExchangeDeclare(xName, xType,
		false,
		false,
		false,
		false,
		nil,
	)
	SysLoggerFatal(err, "Rabbit Declare Exchange Error")
	r.Exchanges = make(map[string]string)
	r.Exchanges[xName] = xName
	q, err := r.Channel.QueueDeclare(qName,
		false,
		false,
		false,
		false,
		nil,
	)
	SysLoggerFatal(err, "RabbitMQ Declare Queue Error")
	r.Queues = make(map[string]*amqp.Queue)
	r.Queues[qName] = &q
	return err
}

func (r *RabbitClient) Stop() {
	defer r.Channel.Close()
	defer r.Conn.Close()
}

//NewExchange declare a new exchange and add to map
func (r *RabbitClient) NewExchange(xName string, kind string, durable bool,
	autoDel bool, internal bool, noWait bool, args amqp.Table) error {
	err := r.Channel.ExchangeDeclare(
		xName,
		kind,
		durable,
		autoDel,
		internal,
		noWait,
		args,
	)
	SysLoggerFatal(err, "RabbitMQ Declare exchange error")
	if err == nil {
		r.Exchanges[xName] = xName
	}
	return err
}

// NewQueue Declares a new queue and add to map
func (r *RabbitClient) NewQueue(qName string, durable bool, autoDel bool,
	exclusive bool, noWait bool, args amqp.Table) error {
	q, err := r.Channel.QueueDeclare(
		qName,
		durable,
		autoDel,
		exclusive,
		noWait,
		args,
	)
	SysLoggerFatal(err, "RabbitMQ Declare Queue Error")
	if err == nil {
		r.Queues[qName] = &q
	}
	return err
}

//ExchangeBind wrap amqp.ExchangeBind method
func (r *RabbitClient) ExchangeBind(dest string, src string, routingKey string,
	noWait bool, args amqp.Table) error {
	return r.Channel.ExchangeBind(
		dest,
		routingKey,
		src,
		noWait,
		args,
	)
}

func (r *RabbitClient) DefaultBind(dest string, src string, key string) error {
	return r.Channel.ExchangeBind(
		dest,
		key,
		src,
		false,
		nil,
	)
}

//QueueBind wrap amqp.QueueBind method
func (r *RabbitClient) QueueBind(name string, bindingKey string, xName string,
	noWait bool, args amqp.Table) error {
	return r.Channel.QueueBind(
		name,
		bindingKey,
		xName,
		noWait,
		args,
	)
}

func (r *RabbitClient) Publish(xName string, key string, mandatory bool,
	immediate bool, contentType string, encoding string, data []byte) error {
	p := amqp.Publishing{
		ContentType:     contentType,
		ContentEncoding: encoding,
		Body:            data,
	}
	return r.Channel.Publish(xName, key, mandatory, immediate, p)
}

func (r *RabbitClient) SimplePublish(xName string, key string, data []byte) error {
	p := amqp.Publishing{
		ContentType: "plain/text",
		Body:        data,
	}
	return r.Channel.Publish(xName, key, false, false, p)
}

func (r *RabbitClient) Consume(qName string, consumer string, autoAck bool,
	exclusive bool, noLocal bool, noWait bool, args amqp.Table) error {
	_, err := r.Channel.Consume(
		qName,
		consumer,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)
	return err
}

func (r *RabbitClient) SimpleConsume(qName string, consumer string) error {
	_, err := r.Channel.Consume(
		qName,
		consumer,
		false,
		false,
		false,
		false,
		nil,
	)
	return err
}
