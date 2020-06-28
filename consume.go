package amqp_safe

import (
	"time"
)

type Result int

const ResultOK Result = 1
const ResultError Result = 2
const ResultReject Result = 3

func (c *Connector) Consume(queue, consumer string, cb func([]byte) Result) {
	c.wg.Add(1)
	go func() {
		for c.closed == 0 {
			sch := c.ch
			if sch == nil {
				time.Sleep(c.cfg.RetryEvery)
				continue
			}

			d, err := sch.Consume(queue, consumer, false, false, false, false, nil)
			if err != nil {
				time.Sleep(c.cfg.RetryEvery)
				continue
			}

			for {
				ev, ok := <-d
				if !ok {
					break
				}

				res := cb(ev.Body)

				var err error
				switch res {
				case ResultOK:
					err = ev.Ack(false)
				case ResultError:
					err = ev.Nack(false,true)
				case ResultReject:
					err = ev.Nack(false,false)
				}

				if err != nil {
					c.cfg.Logger.Println("[ack/nack] failed, error:", err)
				}
			}
		}
		c.wg.Done()
	}()
}
