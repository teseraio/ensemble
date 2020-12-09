package spdy

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/docker/spdystream"
)

// Connection represents an upgraded HTTP connection with SDPY
type Connection interface {
	// CreateStream creates a new Stream with the supplied headers.
	CreateStream(headers http.Header) (net.Conn, error)

	// Accept accepts a new stream
	Accept() net.Conn

	// Close resets all streams and closes the connection.
	Close() error
}

type connection struct {
	conn *spdystream.Connection

	streams    []*spdystream.Stream
	streamLock sync.Mutex

	streamCh chan *spdystream.Stream
}

// Client starts a new SPDY client connection on the net.Conn
func Client(conn net.Conn) (Connection, error) {
	return newConnection(conn, true)
}

// Server starts a new SPDY server connection on the net.Conn
func Server(conn net.Conn) (Connection, error) {
	return newConnection(conn, false)
}

func newConnection(conn net.Conn, client bool) (Connection, error) {
	spdyConn, err := spdystream.NewConnection(conn, true)
	if err != nil {
		defer conn.Close()
		return nil, err
	}

	c := &connection{
		conn:     spdyConn,
		streamCh: make(chan *spdystream.Stream),
	}
	go c.conn.Serve(c.handleStream)
	if client {
		go c.sendPings()
	}
	return c, nil
}

const createStreamResponseTimeout = 30 * time.Second

func (c *connection) Accept() net.Conn {
	for {
		select {
		case stream := <-c.streamCh:
			stream.SendReply(http.Header{}, true)
			return stream
		}
	}
}

// Close closes the connection and all the streams
func (c *connection) Close() error {
	c.streamLock.Lock()
	for _, s := range c.streams {
		s.Reset()
	}
	c.streamLock.Unlock()
	return c.conn.Close()
}

// CreateStream creates a new stream in the connection
func (c *connection) CreateStream(headers http.Header) (net.Conn, error) {
	stream, err := c.conn.CreateStream(headers, nil, false)
	if err != nil {
		return nil, err
	}
	if err = stream.WaitTimeout(createStreamResponseTimeout); err != nil {
		return nil, err
	}

	c.registerStream(stream)
	return stream, nil
}

func (c *connection) registerStream(s *spdystream.Stream) {
	c.streamLock.Lock()
	if len(c.streams) == 0 {
		c.streams = []*spdystream.Stream{}
	}
	c.streams = append(c.streams, s)
	c.streamLock.Unlock()
}

func (c *connection) handleStream(stream *spdystream.Stream) {
	c.streamCh <- stream

	c.registerStream(stream)
	stream.SendReply(http.Header{}, false)
}

func (c *connection) sendPings() {
	for {
		select {
		case <-c.conn.CloseChan():
			return
		case <-time.After(3 * time.Second):
			if _, err := c.conn.Ping(); err != nil {
				// TODO: log
			}
		}
	}
}
