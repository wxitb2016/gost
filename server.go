package gost

import (
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-log/log"
	"github.com/phuslu/glog"
)

// Server is a proxy server.
type Server struct {
	Listener Listener
	options  *ServerOptions
}

// Init intializes server with given options.
func (s *Server) Init(opts ...ServerOption) {
	if s.options == nil {
		s.options = &ServerOptions{}
	}
	for _, opt := range opts {
		opt(s.options)
	}
}

// Addr returns the address of the server
func (s *Server) Addr() net.Addr {
	return s.Listener.Addr()
}

// Close closes the server
func (s *Server) Close() error {
	return s.Listener.Close()
}

// Serve serves as a proxy server.
func (s *Server) Serve(h Handler, opts ...ServerOption) error {
	s.Init(opts...)

	if s.Listener == nil {
		ln, err := TCPListener("")
		if err != nil {
			return err
		}
		s.Listener = ln
	}
	if h == nil {
		h = HTTPHandler()
	}

	l := s.Listener
	var tempDelay time.Duration
	for {
		conn, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Logf("server: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0

		if s.options.Bypass.Contains(conn.RemoteAddr().String()) {
			log.Log("[bypass]", conn.RemoteAddr())
			conn.Close()
			continue
		}

		go h.Handle(conn)
	}
}

// ServerOptions holds the options for Server.
type ServerOptions struct {
	Bypass *Bypass
}

// ServerOption allows a common way to set server options.
type ServerOption func(opts *ServerOptions)

// BypassServerOption sets the bypass option of ServerOptions.
func BypassServerOption(bypass *Bypass) ServerOption {
	return func(opts *ServerOptions) {
		opts.Bypass = bypass
	}
}

// Listener is a proxy server listener, just like a net.Listener.
type Listener interface {
	net.Listener
}

type tcpListener struct {
	net.Listener
}

// TCPListener creates a Listener for TCP proxy server.
func TCPListener(addr string) (Listener, error) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	// Set QUIC_CONNID_PORT if not set
	if _, ok := os.LookupEnv("QUIC_CONNID_PORT"); !ok {
		err = os.Setenv("QUIC_CONNID_PORT", strconv.Itoa(laddr.Port))
		glog.Infof("Setting QUIC_CONNID_PORT to %v", laddr.Port)
		if err != nil {
			return nil, err
		}
	}
	ln, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return &tcpListener{Listener: tcpKeepAliveListener{ln}}, nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(KeepAliveTime)
	return tc, nil
}

var (
	trPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32*1024)
		},
	}
)

func transport(rw1, rw2 io.ReadWriter) error {
	errc := make(chan error, 1)
	go func() {
		buf := trPool.Get().([]byte)
		defer trPool.Put(buf)

		_, err := io.CopyBuffer(rw1, rw2, buf)
		errc <- err
	}()

	go func() {
		buf := trPool.Get().([]byte)
		defer trPool.Put(buf)

		_, err := io.CopyBuffer(rw2, rw1, buf)
		errc <- err
	}()

	err := <-errc
	if err != nil && err == io.EOF {
		err = nil
	}
	return err
}
