package amqp_safe

import (
	"errors"
	"log"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/streadway/amqp"
)

var ErrServerNAck = errors.New("not ack by server")
var ErrServerReturn = errors.New("returned by server")
var ErrNoConnection = errors.New("not connected")
var ErrNoChannel = errors.New("no channel")

type Logger interface {
	Println(v ...interface{})
}

type Config struct {
	DialTimeout    time.Duration
	HeartbeatEvery time.Duration
	RetryEvery     time.Duration
	Logger         Logger
	Hosts          []string
}

type Connector struct {
	iter   int
	ch     *amqp.Channel
	con    *amqp.Connection
	closed int32

	confirms chan amqp.Confirmation
	returns  chan amqp.Return
	pubmx    sync.Mutex
	mx       sync.RWMutex
	wg       sync.WaitGroup

	onConnectionFail func() bool
	onChannelFail    func() bool
	onReConnection   func()
	onReady          func()
	onReChannel      func()

	once sync.Once

	cfg Config
}

func (c *Connector) Close() error {
	atomic.StoreInt32(&c.closed, 1)

	con := c.con
	if con != nil {
		_ = con.Close()
	}

	c.wg.Wait()

	return nil
}

func (c *Connector) Start() *Connector {
	go func() {
		c.startConnection()
		c.startChannel()
	}()
	return c
}

func (c *Connector) startChannel() {
	c.wg.Add(1)
	defer c.wg.Done()

	for c.closed == 0 {
		err := c.initChannel()
		if err == nil {
			break
		}

		c.cfg.Logger.Println("[channel] failed to start:", err)
		time.Sleep(c.cfg.RetryEvery)
	}
}

func (c *Connector) startConnection() {
	c.wg.Add(1)
	defer c.wg.Done()

	for c.closed == 0 {
		host := c.cfg.Hosts[c.iter%len(c.cfg.Hosts)]
		err := c.initConnection(host)
		if err == nil {
			break
		}

		c.iter++

		u, uerr := url.Parse(host)
		if uerr != nil {
			c.cfg.Logger.Println("[connection] failed to start with (corrupted url) err:", err, c.iter, uerr)
		} else {
			c.cfg.Logger.Println("[connection] failed to start with", u.Host, "err:", err, c.iter)
		}

		time.Sleep(c.cfg.RetryEvery)
	}
}

func (c *Connector) initConnection(url string) error {
	con, err := amqp.DialConfig(url, amqp.Config{
		Heartbeat: c.cfg.HeartbeatEvery,
		Dial:      amqp.DefaultDial(c.cfg.DialTimeout),
	})
	if err != nil {
		return err
	}

	atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&c.con)), unsafe.Pointer(con))

	con.NotifyClose(func() chan *amqp.Error {
		ech := make(chan *amqp.Error)
		go func() {
			e := <-ech

			c.mx.RLock()
			f := c.onConnectionFail
			c.mx.RUnlock()

			if f != nil {
				if !f() {
					_ = c.Close()
					return
				}
			}

			atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&c.con)), unsafe.Pointer(nil))

			if e != nil {
				c.cfg.Logger.Println("[recreate] got connection close with error:", e.Reason)
				for {
					c.iter++
					err := c.initConnection(c.cfg.Hosts[c.iter%len(c.cfg.Hosts)])
					if err != nil {
						c.cfg.Logger.Println("[recreate] connection not recreated, will retry...")
						time.Sleep(c.cfg.RetryEvery)
						continue
					}
					c.cfg.Logger.Println("[recreate] connection successfully recreated!")
					return
				}
			}
			c.cfg.Logger.Println("[done] connection closed by user!")
		}()
		return ech
	}())

	c.mx.RLock()
	f := c.onReConnection
	c.mx.RUnlock()

	if f != nil {
		f()
	}

	return nil
}

func (c *Connector) initChannel() error {
	scon := c.con
	if scon == nil {
		return ErrNoConnection
	}

	ch, err := scon.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}

	atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&c.ch)), unsafe.Pointer(ch))

	c.pubmx.Lock()
	c.confirms = make(chan amqp.Confirmation)
	c.returns = make(chan amqp.Return)

	ch.NotifyPublish(c.confirms)
	ch.NotifyReturn(c.returns)
	c.pubmx.Unlock()

	ch.NotifyClose(func() chan *amqp.Error {
		ech := make(chan *amqp.Error)
		go func() {
			e := <-ech

			c.mx.RLock()
			f := c.onChannelFail
			c.mx.RUnlock()

			if f != nil {
				if !f() {
					_ = c.Close()
					return
				}
			}

			atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&c.ch)), unsafe.Pointer(nil))

			if c.closed != 0 {
				return
			}

			if e != nil {
				c.cfg.Logger.Println("[recreate] got channel close with error:", e.Reason)
			} else {
				c.cfg.Logger.Println("[done] channel closed by user!")
			}

			for {
				err := c.initChannel()
				if err != nil {
					c.cfg.Logger.Println("[recreate] channel not recreated, will retry...")
					time.Sleep(c.cfg.RetryEvery)
					continue
				}
				c.cfg.Logger.Println("[recreate] channel successfully recreated!")
				return
			}
		}()
		return ech
	}())

	c.once.Do(func() {
		c.mx.RLock()
		rf := c.onReady
		c.mx.RUnlock()

		rf()
	})

	c.mx.RLock()
	f := c.onReChannel
	c.mx.RUnlock()

	if f != nil {
		f()
	}

	return nil
}

func NewConnector(cfg Config) *Connector {
	if cfg.Logger == nil {
		cfg.Logger = log.New(os.Stdout, "[amqp-safe]", 0)
	}

	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 3 * time.Second
	}

	if cfg.HeartbeatEvery == 0 {
		cfg.HeartbeatEvery = 3 * time.Second
	}

	if cfg.RetryEvery == 0 {
		cfg.RetryEvery = 1 * time.Second
	}

	if len(cfg.Hosts) == 0 {
		panic("no amqp hosts specified")
	}

	return &Connector{cfg: cfg, onReady: func() {}}
}
