package amqp_safe

func (c *Connector) Publish(exchange, key string, publishing Publishing) error {
	c.wg.Add(1)
	defer c.wg.Done()

	c.pubmx.Lock()
	defer c.pubmx.Unlock()

	sch := c.ch
	if sch == nil {
		return ErrNoChannel
	}

	err := sch.Publish(exchange, key, true, false, publishing)
	if err != nil {
		return err
	}

	select {
	case ret := <-c.returns:
		c.cfg.Logger.Println("[publish] message returned with err:", ret.ReplyText)

		ack := <-c.confirms
		if !ack.Ack {
			c.cfg.Logger.Println("[publish] message confirm nack after return")
			return ErrServerNAck
		}

		return ErrServerReturn
	case cnf := <-c.confirms:
		if !cnf.Ack {
			c.cfg.Logger.Println("[publish] message confirm nack")

			return ErrServerNAck
		}
	}

	return nil
}
