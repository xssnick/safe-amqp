package amqp_safe

import (
	"time"
)

func (c *Connector) Consume(queue, consumer string, cb func(Delivery)) {
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

				cb(ev)
			}
		}
		c.wg.Done()
	}()
}
