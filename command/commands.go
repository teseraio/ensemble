package command

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/ryanuber/columnize"
	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/command/server"
	"github.com/teseraio/ensemble/operator/proto"
	"google.golang.org/grpc"
)

// Commands returns the cli commands
func Commands() map[string]cli.CommandFactory {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}
	meta := Meta{
		UI: ui,
	}

	return map[string]cli.CommandFactory{
		"apply": func() (cli.Command, error) {
			return &ApplyCommand{
				Meta: meta,
			}, nil
		},
		"delete": func() (cli.Command, error) {
			return &DeleteCommand{
				Meta: meta,
			}, nil
		},
		"server": func() (cli.Command, error) {
			return &server.Command{
				UI: ui,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				UI: ui,
			}, nil
		},
		"deployment": func() (cli.Command, error) {
			return &DeploymentCommand{}, nil
		},
		"deployment list": func() (cli.Command, error) {
			return &DeploymentListCommand{
				Meta: meta,
			}, nil
		},
		"deployment status": func() (cli.Command, error) {
			return &DeploymentStatusCommand{
				Meta: meta,
			}, nil
		},
		"k8s": func() (cli.Command, error) {
			return &K8sCommand{}, nil
		},
		"k8s init": func() (cli.Command, error) {
			return &K8sInitCommand{
				UI: ui,
			}, nil
		},
		"k8s artifacts": func() (cli.Command, error) {
			return &K8sArtifactsCommand{
				Meta: meta,
			}, nil
		},
	}
}

// Meta is a helper utility for the commands
type Meta struct {
	UI   cli.Ui
	addr string
}

func (m *Meta) NewFlagSet(n string) *flagset.Flagset {
	f := flagset.NewFlagSet(n)

	f.StringFlag(&flagset.StringFlag{
		Name:  "address",
		Value: &m.addr,
		Usage: "Path of the file to apply",
	})

	return f
}

// Conn returns a grpc connection
func (m *Meta) Conn() (proto.EnsembleServiceClient, error) {
	conn, err := grpc.Dial(m.addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}
	clt := proto.NewEnsembleServiceClient(conn)
	return clt, nil
}

func formatList(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = "<none>"
	return columnize.Format(in, columnConf)
}

func formatKV(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = "<none>"
	columnConf.Glue = " = "
	return columnize.Format(in, columnConf)
}
