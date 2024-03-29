package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/k8s"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/state/boltdb"

	"github.com/google/gops/agent"
	"github.com/mitchellh/cli"
)

// Command is the command to run the agent
type Command struct {
	UI cli.Ui

	debug      bool
	logLevel   string
	boltdbPath string
	bind       string
}

// Help implements the cli.Command interface
func (c *Command) Help() string {
	return `Usage: ensemble server [options]
  
  Run the Ensemble operator server.

` + c.Flags().Help()
}

func (c *Command) Flags() *flagset.Flagset {
	f := flagset.NewFlagSet("server")

	f.BoolFlag(&flagset.BoolFlag{
		Name:  "debug",
		Value: &c.debug,
		Usage: "Path of the file to apply",
	})

	f.StringFlag(&flagset.StringFlag{
		Name:  "log-level",
		Value: &c.logLevel,
		Usage: "Follow the directory in -f recursively",
	})

	f.StringFlag(&flagset.StringFlag{
		Name:    "boltdb",
		Value:   &c.boltdbPath,
		Usage:   "Follow the directory in -f recursively",
		Default: "test.db",
	})

	f.StringFlag(&flagset.StringFlag{
		Name:    "bind",
		Value:   &c.bind,
		Usage:   "Bind IP address for the GRPC server",
		Default: "127.0.0.1",
	})

	return f
}

// Synopsis implements the cli.Command interface
func (c *Command) Synopsis() string {
	return "Run the Ensemble operator server"
}

// Run implements the cli.Command interface
func (c *Command) Run(args []string) int {
	flags := c.Flags()
	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if c.debug {
		if err := agent.Listen(agent.Options{}); err != nil {
			c.UI.Error(fmt.Sprintf("Failed to start gops: %v", err))
			return 1
		}
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "ensemble",
		Level: hclog.LevelFromString(c.logLevel),
	})

	// setup resource provider
	k8sProvider, err := k8s.K8sFactory(logger, nil)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to create the provider: %v", err))
		return 1
	}

	// setup state
	state, err := boltdb.Factory(map[string]interface{}{
		"path": c.boltdbPath,
	})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to start boltdb state: %v", err))
		return 1
	}

	config := &operator.Config{
		Provider:         k8sProvider,
		State:            state,
		HandlerFactories: BuiltinBackends,
		GRPCAddr:         &net.TCPAddr{IP: net.ParseIP(c.bind), Port: 6001},
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
