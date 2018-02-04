package client

import (
	"bufio"
	"errors"
	"net"
	"time"

	"github.com/aliaksandrb/cachy/proto"

	log "github.com/aliaksandrb/cachy/logger"
)

type Client interface {
	Get(key string) (val interface{}, err error)
	Set(key string, val interface{}, ttl time.Duration) error
	Update(key string, val interface{}, ttl time.Duration) error
	Remove(key string) error
	Keys() ([]string, error)
	Close() error
}

func New(addr string, poolSize int) (Client, error) {
	c := &client{
		addr:     addr,
		poolSize: poolSize,
		connPool: make(chan net.Conn, poolSize),
		closing:  make(chan struct{}),
	}

	for i := 0; i < poolSize; i++ {
		conn, err := makeConn(addr)
		if err != nil {
			return nil, err
		}

		c.connPool <- conn
	}

	return c, nil
}

func makeConn(addr string) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}

	return net.DialTCP("tcp", nil, tcpAddr)
}

type client struct {
	addr     string
	poolSize int
	connPool chan net.Conn
	closing  chan struct{}
}

var ErrTerminated = errors.New("terminated")

func (c *client) acquireConn() (net.Conn, error) {
	select {
	case conn := <-c.connPool:
		return conn, nil
	case <-c.closing:
		return nil, ErrTerminated
	}

	return nil, ErrTerminated
}

func (c *client) releaseConn(conn net.Conn) {
	go func() {
		c.connPool <- conn
	}()
}

func (c *client) send(b []byte) (s *bufio.Scanner, err error) {
	conn, err := c.acquireConn()
	if err != nil {
		return nil, err
	}
	defer c.releaseConn(conn)

	if _, err = conn.Write(b); err != nil {
		return
	}

	return proto.NewResponseScanner(conn)
}

func (c *client) processMessage(b []byte) (val interface{}, err error) {
	response, err := c.send(b)
	if err != nil {
		return nil, err
	}

	val, err = proto.Decode(response)
	if err != nil {
		return
	}

	err, _ = val.(error)

	return
}

func (c *client) Get(key string) (val interface{}, err error) {
	msg, err := proto.NewMessage(proto.CmdGet, key, nil, 0)
	if err != nil {
		return
	}

	return c.processMessage(msg)
}

func (c *client) Set(key string, val interface{}, ttl time.Duration) (err error) {
	msg, err := proto.NewMessage(proto.CmdSet, key, val, ttl)
	if err != nil {
		return
	}

	_, err = c.processMessage(msg)
	return err
}

func (c *client) Update(key string, val interface{}, ttl time.Duration) (err error) {
	msg, err := proto.NewMessage(proto.CmdUpdate, key, val, ttl)
	if err != nil {
		return
	}

	_, err = c.processMessage(msg)
	return err
}

func (c *client) Remove(key string) error {
	msg, err := proto.NewMessage(proto.CmdRemove, key, nil, 0)
	if err != nil {
		return err
	}

	_, err = c.processMessage(msg)
	return err
}

func (c *client) Keys() (keys []string, err error) {
	msg, err := proto.NewMessage(proto.CmdKeys, "", nil, 0)
	if err != nil {
		return
	}

	response, err := c.processMessage(msg)
	if err != nil {
		return
	}

	vals, ok := response.([]interface{})
	if !ok {
		log.Err("keys should return slice, got %T - % q", response, response)
		return nil, proto.ErrUnknown
	}

	keys = make([]string, len(vals))
	for i, key := range vals {
		k, ok := key.(string)
		if !ok {
			log.Err("keys should strings, got %T - % q", key, key)
			continue
		}
		keys[i] = k
	}

	return keys, err
}

func (c *client) Close() (err error) {
	log.Info("closing client...")
	close(c.closing)

	for i := 0; i < c.poolSize; i++ {
		conn := <-c.connPool
		log.Info("[%d] closing connection: %s", i, conn.RemoteAddr())
		if err = conn.Close(); err != nil {
			return
		}
	}

	log.Info("closing client, done.")
	return nil
}
