package amqp_safe

func (c *Connector) OnConnectionFail(f func() bool) {
	c.mx.Lock()
	c.onConnectionFail = f
	c.mx.Unlock()
}

func (c *Connector) OnChannelFail(f func() bool) {
	c.mx.Lock()
	c.onChannelFail = f
	c.mx.Unlock()
}

func (c *Connector) OnConnect(f func()) {
	c.mx.Lock()
	c.onReConnection = f
	c.mx.Unlock()
}

func (c *Connector) OnReady(f func()) {
	c.mx.Lock()
	c.onReady = f
	c.mx.Unlock()
}

func (c *Connector) OnChannel(f func()) {
	c.mx.Lock()
	c.onReChannel = f
	c.mx.Unlock()
}
