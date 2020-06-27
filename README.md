# amqp-safe
Golang AMQP with reconnect, clustering and delivery guarantee.

```go
import (
	amqp "github.com/xssnick/amqp-safe"
)

// Start connection and open channel, async
c := amqp.NewConnector(amqp.Config{
    Hosts: []string{"amqp://admin:password@127.0.0.1"},
}).Start()

// Callback on channel ready
c.OnReady(func() {
    err := c.ExchangeDeclare("test-exchange", amqp.ExchangeDirect, true, false, false, false, nil)
    if err != nil {
        log.Panic(err)
    }

    _, err = c.QueueDeclare("test-queue", true, false, false, false, nil)
    if err != nil {
        log.Panic(err)
    }

    if err := c.QueueBind("test-queue", "", "test-exchange", false, nil); err != nil {
        log.Panic(err)
    }

    err := c.Publish("test-exchange", "", amqp.Publishing{
        Body: []byte("hey),
    })

    c.Consume("test-queue", "", func(ev amqp.Delivery) {
        log.Println("got:", string(ev.Body))
        ev.Ack(false)
    })
})
```