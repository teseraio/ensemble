package spdy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpgradeResponse(t *testing.T) {
	cases := []struct {
		connectionHeader string
		upgradeHeader    string
		err              bool
	}{
		{
			connectionHeader: "",
			upgradeHeader:    "",
			err:              true,
		},
		{
			connectionHeader: "Upgrade",
			upgradeHeader:    "",
			err:              true,
		},
		{
			connectionHeader: "",
			upgradeHeader:    "SPDY/3.1",
			err:              true,
		},
		{
			connectionHeader: "Upgrade",
			upgradeHeader:    "SPDY/3.1",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				_, err := UpgradeResponse(w, req)
				if err != nil && !c.err {
					t.Fatal(err)
				}
				if err == nil && c.err {
					t.Fatal("bad")
				}
			}))
			defer server.Close()

			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Connection", c.connectionHeader)
			req.Header.Set("Upgrade", c.upgradeHeader)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if !c.err {
				if resp.StatusCode != http.StatusSwitchingProtocols {
					t.Fatal("bad")
				}
			}
		})
	}
}

type mockResponseWriter struct {
	header     http.Header
	statusCode *int
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) WriteHeader(code int) {
	m.statusCode = &code
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func TestHandshake(t *testing.T) {
	cases := map[string]struct {
		client   []string
		server   []string
		expected string
		err      bool
	}{
		"no common protocol": {
			client:   []string{"c"},
			server:   []string{"a", "b"},
			expected: "",
			err:      true,
		},
		"no common protocol with comma separated list": {
			client:   []string{"c, d"},
			server:   []string{"a", "b"},
			expected: "",
			err:      true,
		},
		"common protocol": {
			client:   []string{"b"},
			server:   []string{"a", "b"},
			expected: "b",
		},
		"common protocol with comma separated list": {
			client:   []string{"b, c"},
			server:   []string{"a", "b"},
			expected: "b",
		},
	}

	for _, c := range cases {
		req, err := http.NewRequest("GET", "http://www.example.com/", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range c.client {
			req.Header.Add(headerProtocolVersion, p)
		}

		w := &mockResponseWriter{
			header: make(http.Header),
		}
		negotiated, err := Handshake(req, w, c.server)

		// verify negotiated protocol
		if c.expected != negotiated {
			t.Fatal("bad")
		}

		if err != nil && !c.err {
			t.Fatal(err)
		}
		if err == nil && c.err {
			t.Fatal("expected error")
		}
	}
}
