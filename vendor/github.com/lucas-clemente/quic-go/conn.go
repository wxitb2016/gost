package quic

import (
	"net"
	"sync"
)

type connection interface {
	Write([]byte) error
	Read([]byte) (int, net.Addr, error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetCurrentRemoteAddr(net.Addr) net.Addr
}

type conn struct {
	mutex sync.RWMutex

	pconn       net.PacketConn
	currentAddr net.Addr
}

var _ connection = &conn{}

func (c *conn) Write(p []byte) error {
	_, err := c.pconn.WriteTo(p, c.currentAddr)
	return err
}

func (c *conn) Read(p []byte) (int, net.Addr, error) {
	return c.pconn.ReadFrom(p)
}

func (c *conn) SetCurrentRemoteAddr(addr net.Addr) net.Addr {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.currentAddr.String() == addr.String() {
		return nil
	}
	oldAddr := c.currentAddr
	c.currentAddr = addr
	return oldAddr
}

func (c *conn) LocalAddr() net.Addr {
	return c.pconn.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	c.mutex.RLock()
	addr := c.currentAddr
	c.mutex.RUnlock()
	return addr
}

func (c *conn) Close() error {
	return c.pconn.Close()
}
