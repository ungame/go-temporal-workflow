package rmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"go-temporal-workflow/utils"
)

type Client interface {
	ExchangeDeclare(opts *ExchangeOptions) error
	QueueDeclare(opts *QueueOptions) error
	AsConsumer() Consumer
	AsPublisher() Publisher
	Close()
}

type Consumer interface {
	HandleFunc(messageType interface{}, handler func(data []byte) error)
	Listen(opts *ConsumerOptions) error
}

type Publisher interface {
	Send(opts *PublisherOptions) error
}

type client struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	handlers map[string]func(data []byte) error
}

func NewClient(cfg Config) Client {
	var c client
	var err error
	c.conn, err = amqp.Dial(cfg.URL())
	if err != nil {
		log.Panicln(err)
	}
	c.channel, err = c.conn.Channel()
	if err != nil {
		log.Panicln(err)
	}
	c.handlers = make(map[string]func(data []byte) error)
	return &c
}

func (c *client) AsConsumer() Consumer {
	return c
}

func (c *client) AsPublisher() Publisher {
	return c
}

func (c *client) ExchangeDeclare(opts *ExchangeOptions) error {
	return c.channel.ExchangeDeclare(
		opts.Name,
		opts.Kind,
		opts.Durable,
		opts.AutoDelete,
		opts.Internal,
		opts.NoWait,
		opts.Args,
	)
}

func (c *client) QueueDeclare(opts *QueueOptions) error {
	queue, err := c.channel.QueueDeclare(
		opts.Name,
		opts.Durable,
		opts.AutoDelete,
		opts.Exclusive,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		return err
	}

	if opts.BindOptions != nil {

		bindOpts := opts.BindOptions

		err = c.channel.QueueBind(
			queue.Name,
			bindOpts.RoutingKey,
			bindOpts.ExchangeName,
			bindOpts.NoWait,
			bindOpts.Args,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) Listen(opts *ConsumerOptions) error {

	messages, err := c.channel.Consume(
		opts.QueueName,
		opts.Consumer,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
	if err != nil {
		return err
	}

	for message := range messages {

		var msg Message
		err := json.Unmarshal(message.Body, &msg)
		if err != nil {
			return err
		}

		if !opts.AutoAck {
			err := message.Ack(false)
			if err != nil {
				return err
			}
		}

		handler, ok := c.handlers[msg.Type]
		if !ok {
			log.Printf("received message without registered handler: MessageType=%s", msg.Type)
			continue
		}

		err = handler(msg.Data)
		if err != nil {
			log.Printf("error on handler message: err=%s data=%s\n", err, msg.Data)
		}
	}

	return nil
}

func (c *client) HandleFunc(messageType interface{}, handler func(data []byte) error) {
	c.handlers[fmt.Sprintf("%T", messageType)] = handler
}

func (c *client) Send(opts *PublisherOptions) error {

	body, err := json.Marshal(opts.Message)
	if err != nil {
		return err
	}

	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	if opts.Persistent {
		message.DeliveryMode = amqp.Persistent
	}

	return c.channel.Publish(
		opts.ExchangeName,
		opts.RoutingKey,
		opts.Mandatory,
		opts.Immediate,
		message,
	)
}

func (c *client) Close() {
	utils.HandleClose(c.channel)
	utils.HandleClose(c.conn)
}
