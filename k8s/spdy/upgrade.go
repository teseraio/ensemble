package spdy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
)

type connWrapper struct {
	net.Conn
	closed    int32
	bufReader *bufio.Reader
}

func (w *connWrapper) Read(b []byte) (n int, err error) {
	if atomic.LoadInt32(&w.closed) == 1 {
		return 0, io.EOF
	}
	return w.bufReader.Read(b)
}

func (w *connWrapper) Close() error {
	err := w.Conn.Close()
	atomic.StoreInt32(&w.closed, 1)
	return err
}

// UpgradeResponse upgrades an HTTP response to a SPDY multi-stream connection
func UpgradeResponse(w http.ResponseWriter, req *http.Request) (Connection, error) {
	if !isValidUpgradeHeader(req.Header) {
		return nil, fmt.Errorf("missing upgrade headers in request: %#v", req.Header)
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("unable to hijack response")
	}

	w.Header().Add(headerConnection, headerUpgrade)
	w.Header().Add(headerUpgrade, headerSpdy31)
	w.WriteHeader(http.StatusSwitchingProtocols)

	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("error hijacking response: %v", err)
	}

	connWithBuf := &connWrapper{
		Conn:      conn,
		bufReader: bufrw.Reader,
	}
	spdyConn, err := Server(connWithBuf)
	if err != nil {
		return nil, fmt.Errorf("error creating SPDY server connection: %v", err)
	}

	return spdyConn, nil
}

func negotiateProtocol(clientProtocols, serverProtocols []string) string {
	for i := range clientProtocols {
		for j := range serverProtocols {
			if clientProtocols[i] == serverProtocols[j] {
				return clientProtocols[i]
			}
		}
	}
	return ""
}

func commaSeparatedHeaderValues(header []string) []string {
	var parsedClientProtocols []string
	for i := range header {
		for _, clientProtocol := range strings.Split(header[i], ",") {
			if proto := strings.Trim(clientProtocol, " "); len(proto) > 0 {
				parsedClientProtocols = append(parsedClientProtocols, proto)
			}
		}
	}
	return parsedClientProtocols
}

// Handshake performs a subprotocol negotiation.
func Handshake(req *http.Request, w http.ResponseWriter, serverProtocols []string) (string, error) {
	clientProtocols := commaSeparatedHeaderValues(req.Header[http.CanonicalHeaderKey(headerProtocolVersion)])
	if len(clientProtocols) == 0 {
		return "", fmt.Errorf("unable to upgrade: %s is required", headerProtocolVersion)
	}

	if len(serverProtocols) == 0 {
		panic(fmt.Errorf("unable to upgrade: serverProtocols is required"))
	}

	negotiatedProtocol := negotiateProtocol(clientProtocols, serverProtocols)
	if len(negotiatedProtocol) == 0 {
		for i := range serverProtocols {
			w.Header().Add(headerAcceptedProtocolVersions, serverProtocols[i])
		}
		err := fmt.Errorf("unable to upgrade: unable to negotiate protocol: client supports %v, server accepts %v", clientProtocols, serverProtocols)
		http.Error(w, err.Error(), http.StatusForbidden)
		return "", err
	}

	w.Header().Add(headerProtocolVersion, negotiatedProtocol)
	return negotiatedProtocol, nil
}
