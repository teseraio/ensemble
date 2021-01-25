package server

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/k8s"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/state/boltdb"

	"github.com/google/gops/agent"
	"github.com/mitchellh/cli"
)

// Command is the command to run the agent
type Command struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *Command) Help() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (c *Command) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (c *Command) Run(args []string) int {
	var debug bool
	var logLevel, boltdbPath string

	flags := flag.NewFlagSet("operator", flag.ContinueOnError)
	flags.Usage = func() {}

	flags.BoolVar(&debug, "debug", false, "")
	flags.StringVar(&logLevel, "log-level", "", "")
	flags.StringVar(&boltdbPath, "boltdb", "test.db", "")

	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if debug {
		if err := agent.Listen(agent.Options{}); err != nil {
			c.UI.Error(fmt.Sprintf("Failed to start gops: %v", err))
			return 1
		}
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "ensemble",
		Level: hclog.LevelFromString(logLevel),
	})

	// setup resource provider
	k8sProvider, err := k8s.K8sFactory(logger, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create the provider: %v", err))
		return 1
	}
	if err := k8sProvider.Setup(); err != nil {
		c.UI.Error(fmt.Sprintf("Failed to start the provider: %v", err))
		return 1
	}

	// setup state
	state, err := boltdb.Factory(map[string]interface{}{
		"path": boltdbPath,
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to start boltdb state: %v", err))
		return 1
	}

	config := &operator.Config{
		Provider:         k8sProvider,
		State:            state,
		HandlerFactories: BuiltinBackends,
		GRPCAddr:         &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 6001},
	}
	srv, err := operator.NewServer(logger, config)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to start the server: %v", err))
		return 1
	}

	return c.handleSignals(srv.Stop)
}

func (c *Command) handleSignals(closeFn func()) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	var sig os.Signal
	select {
	case sig = <-signalCh:
	}

	c.UI.Output(fmt.Sprintf("Caught signal: %v", sig))
	c.UI.Output("Gracefully shutting down agent...")

	gracefulCh := make(chan struct{})
	go func() {
		if closeFn != nil {
			closeFn()
		}
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-time.After(5 * time.Second):
		return 1
	case <-gracefulCh:
		return 0
	}
}
