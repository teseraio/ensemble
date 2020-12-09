package spdy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	headerConnection               = "Connection"
	headerUpgrade                  = "Upgrade"
	headerProtocolVersion          = "X-Stream-Protocol-Version"
	headerAcceptedProtocolVersions = "X-Accepted-Stream-Protocol-Versions"
	headerSpdy31                   = "SPDY/3.1"
)

// RoundTripper upgrades an HTTP connection to enable streams
// with the SDPY protocol
type RoundTripper struct {
	//tlsConfig holds the TLS configuration for the connection.
	TLSConfig *tls.Config

	BearerToken string

	// conn is the underlying network connection.
	conn net.Conn
}

func (r *RoundTripper) dialWithoutProxy(ctx context.Context, url *url.URL) (net.Conn, error) {
	dialAddr := url.Host

	if url.Scheme == "http" {
		return net.Dial("tcp", dialAddr)
	}

	conn, err := tls.Dial("tcp", dialAddr, r.TLSConfig)
	if err != nil {
		return nil, err
	}
	if r.TLSConfig != nil && r.TLSConfig.InsecureSkipVerify {
		return conn, nil
	}

	host, _, err := net.SplitHostPort(dialAddr)
	if err != nil {
		return nil, err
	}
	if r.TLSConfig != nil && len(r.TLSConfig.ServerName) > 0 {
		host = r.TLSConfig.ServerName
	}

	if err = conn.VerifyHostname(host); err != nil {
		return nil, err
	}
	return conn, nil
}

func cloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)

	// shallow clone
	*r = *req

	// deep copy headers
	r.Header = make(http.Header, len(req.Header))
	for key, values := range req.Header {
		r.Header[key] = make([]string, len(values))
		copy(r.Header[key], values)
	}

	return r
}

// RoundTrip executes the Request and upgrades it. After a successful upgrade,
// clients may call SpdyRoundTripper.Connection() to retrieve the upgraded
// connection.
func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)
	req.Header.Add(headerConnection, headerUpgrade)
	req.Header.Add(headerUpgrade, headerSpdy31)

	conn, err := r.dialWithoutProxy(req.Context(), req.URL)
	if err != nil {
		return nil, err
	}
	if err := req.Write(conn); err != nil {
		conn.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	r.conn = conn
	return resp, nil
}

func containsLower(target, s string) bool {
	return strings.Contains(strings.ToLower(target), strings.ToLower(s))
}

func isValidUpgradeHeader(headers http.Header) bool {
	if !containsLower(headers.Get(headerConnection), headerUpgrade) {
		return false
	}
	if !containsLower(headers.Get(headerUpgrade), headerSpdy31) {
		return false
	}
	return true
}

// NewConnection validates the upgrade response.
func (r *RoundTripper) NewConnection(resp *http.Response) (Connection, error) {
	if resp.StatusCode != http.StatusSwitchingProtocols || !isValidUpgradeHeader(resp.Header) {
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error: %v", err)
		}
		return nil, fmt.Errorf("unable to upgrade connection: %s", string(data))
	}
	return Client(r.conn)
}
