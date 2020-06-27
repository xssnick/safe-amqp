package amqp_safe

import (
	"github.com/streadway/amqp"
)

const (
	ExchangeDirect  = "direct"
	ExchangeFanout  = "fanout"
	ExchangeTopic   = "topic"
	ExchangeHeaders = "headers"
)

type Delivery = amqp.Delivery
type Table = amqp.Table
type Queue = amqp.Queue
type Publishing = amqp.Publishing

func (c *Connector) QueueBind(name, key, exchange string, noWait bool, args Table) error {
	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	return sch.QueueBind(name, key, exchange, noWait, args)
}

func (c *Connector) QueueUnbind(name, key, exchange string, args Table) error {
	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	return sch.QueueUnbind(name, key, exchange, args)
}

func (c *Connector) QueuePurge(name string, noWait bool) (int, error) {
	sch := c.ch
	if sch == nil {
		return 0, ErrNoChannel
	}

	return sch.QueuePurge(name, noWait)
}

func (c *Connector) QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error) {
	sch := c.ch
	if sch == nil {
		return 0, ErrNoChannel
	}

	return sch.QueueDelete(name, ifUnused, ifEmpty, noWait)
}

func (c *Connector) QueueInspect(name string) (Queue, error) {
	sch := c.ch
	if sch == nil {
		return Queue{}, ErrNoChannel
	}

	return sch.QueueInspect(name)
}

func (c *Connector) QueueDeclarePassive(name string, durable, autoDelete, exclusive, noWait bool, args Table) (Queue, error) {
	sch := c.ch
	if sch == nil {
		return Queue{}, ErrNoChannel
	}

	return sch.QueueDeclarePassive(name, durable, autoDelete, exclusive, noWait, args)
}

func (c *Connector) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args Table) (Queue, error) {
	sch := c.ch
	if sch == nil {
		return Queue{}, ErrNoChannel
	}

	return sch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (c *Connector) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args Table) error {
	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	return sch.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

func (c *Connector) ExchangeDeclarePassive(name, kind string, durable, autoDelete, internal, noWait bool, args Table) error {
	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	return sch.ExchangeDeclarePassive(name, kind, durable, autoDelete, internal, noWait, args)
}

func (c *Connector) ExchangeDelete(name string, ifUnused, noWait bool) error {
	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	return sch.ExchangeDelete(name, ifUnused, noWait)
}

func (c *Connector) ExchangeBind(destination, key, source string, noWait bool, args Table) error {
	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	return sch.ExchangeBind(destination, key, source, noWait, args)
}

func (c *Connector) ExchangeUnbind(destination, key, source string, noWait bool, args Table) error {
	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	return sch.ExchangeUnbind(destination, key, source, noWait, args)
}
