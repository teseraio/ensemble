package spdy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
)

var emptyData = []byte("{}")

// RemoteCommand executes a remote command on Kubernetes
type RemoteCommand struct {
	URL         string
	TLSConfig   *tls.Config
	BearerToken string
}

// Args is the argument for the execute command
type Args struct {
	Container string
	Command   []string
}

// Execute executes a specific command
func (r *RemoteCommand) Execute(args *Args) ([]byte, error) {
	url := r.URL
	url += "?stderr=true&stdout=true&"
	url += "container=" + args.Container + "&"

	for _, command := range args.Command {
		url += "command=" + command + "&"
	}

	transport := &RoundTripper{
		TLSConfig:   r.TLSConfig,
		BearerToken: r.BearerToken,
	}
	client := &http.Client{
		Transport: transport,
	}

	protocols := []string{
		"v4.channel.k8s.io",
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(emptyData))
	if err != nil {
		return nil, err
	}
	for i := range protocols {
		req.Header.Add("X-Stream-Protocol-Version", protocols[i])
	}
	if r.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.BearerToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	conn, err := transport.NewConnection(resp)
	if err != nil {
		return nil, err
	}

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	streamer := &protocol{
		Stdout: outBuf,
		Stderr: errBuf,
	}
	if err := streamer.Stream(conn); err != nil {
		return nil, err
	}

	if errData := errBuf.Bytes(); len(errData) != 0 {
		return nil, fmt.Errorf(string(errData))
	}

	return outBuf.Bytes(), nil
}

type protocol struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Stream uses a Connection to execute the remote command
func (p *protocol) Stream(conn Connection) error {
	var wg sync.WaitGroup

	errCh := make(chan error, 2)

	createStream := func(value string, dst io.Writer) (net.Conn, error) {
		header := http.Header{}
		header.Set("streamType", value)

		stream, err := conn.CreateStream(header)
		if err != nil {
			return nil, err
		}

		if dst != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()

				_, err := io.Copy(dst, stream)
				errCh <- err
			}()
		}
		return stream, nil
	}

	var errStream, stdin net.Conn
	var err error

	// create streams
	if errStream, err = createStream("error", nil); err != nil {
		return err
	}

	if p.Stdout != nil {
		if _, err = createStream("stdout", p.Stdout); err != nil {
			return err
		}
	}
	if p.Stderr != nil {
		if _, err = createStream("stderr", p.Stderr); err != nil {
			return err
		}
	}

	if p.Stdin != nil {
		if stdin, err = createStream("stdin", nil); err != nil {
			return err
		}
	}

	// copy the input
	if p.Stdin != nil {
		if _, err := io.Copy(stdin, p.Stdin); err != nil {
			return err
		}
	}

	// wait for the outputs to be flushed
	wg.Wait()

	// check the error channels
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}

	// check the error stream
	data, err := ioutil.ReadAll(errStream)
	if err != nil {
		return fmt.Errorf(string(data))
	}

	return nil
}
