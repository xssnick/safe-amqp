package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	amqp "github.com/xssnick/amqp-safe"
)

func main() {
	c := amqp.NewConnector(amqp.Config{
		Hosts: []string{"amqp://admin:password@127.0.0.1"},
	}).Start()

	c.OnReady(func() {
		exchangeName := "test-exchange"
		queueName := "test-queue"

		err := c.ExchangeDeclare(exchangeName, amqp.ExchangeDirect, true, false, false, false, nil)
		if err != nil {
			log.Panic(err)
		}

		_, err = c.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			log.Panic(err)
		}

		if err := c.QueueBind(queueName, "", exchangeName, false, nil); err != nil {
			log.Panic(err)
		}

		go func() {
			i := 0

			for {
				err := c.Publish("test-exchange", "", amqp.Publishing{
					Body: []byte("hey" + strconv.Itoa(i)),
				})

				time.Sleep(1 * time.Second)

				if err != nil {
					log.Println("puberr:", err)
					continue
				}

				i++
			}
		}()

		c.Consume("test-queue", "", func(bytes []byte) amqp.Result {
			log.Println("ev:", string(bytes))
			return amqp.ResultOK
		})
	})

	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT)

	<-sg

	log.Println("closing...")
	c.Close()
	log.Println("done!")
}
