package spdy

import (
	"bytes"
	"crypto/rand"
	"net"
	"net/http"
	"testing"
)

func TestEchoConnection(t *testing.T) {
	conn0, conn1 := net.Pipe()

	server, err := Server(conn0)
	if err != nil {
		t.Fatal(err)
	}

	client, err := Client(conn1)
	if err != nil {
		t.Fatal(err)
	}

	n := 100

	go func() {
		stream := server.(*connection).Accept()

		data := make([]byte, n)
		if _, err := stream.Read(data); err != nil {
			return
		}
		if _, err := stream.Write(data); err != nil {
			return
		}
	}()

	data := make([]byte, n)
	rand.Read(data)

	stream, err := client.CreateStream(http.Header{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stream.Write(data); err != nil {
		t.Fatal(err)
	}

	res := make([]byte, n)
	if _, err := stream.Read(res); err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, res) {
		t.Fatal("bad")
	}
}

func TestMultipleConnections(t *testing.T) {
	conn0, conn1 := net.Pipe()

	server, err := Server(conn0)
	if err != nil {
		t.Fatal(err)
	}

	client, err := Client(conn1)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("example")
	numStreams := 5

	go func() {
		streams := make([]net.Conn, numStreams)
		for i := 0; i < numStreams; i++ {
			if streams[i], err = client.CreateStream(http.Header{}); err != nil {
				return
			}
		}
		for _, stream := range streams {
			if _, err := stream.Write(data); err != nil {
				return
			}
		}
	}()

	streams := make([]net.Conn, numStreams)
	for i := 0; i < numStreams; i++ {
		// first open the connections
		streams[i] = server.(*connection).Accept()
	}
	for _, stream := range streams {
		res := make([]byte, 1024)

		n, err := stream.Read(res)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(res[:n], data) {
			t.Fatal("bad")
		}
	}

}
