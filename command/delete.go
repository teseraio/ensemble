package command

import (
	"context"

	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/operator/proto"
)

type DeleteCommand struct {
	Meta

	filename  string
	recursive bool
}

// Synopsis implements the cli.Command interface
func (d *DeleteCommand) Synopsis() string {
	return "Delete a configuration to a resource by filename or stdin"
}

// Synopsis implements the cli.Command interface
func (d *DeleteCommand) Help() string {
	return `Usage: ensemble delete [options]

  Delete a configuration to a resource by filename or stdin.

  Delete a single yaml file:

    $ ensemble delete -f pod.yaml

  Delete multiple files from a directory:

    $ ensemble delete -f ./components

  Delete a configuration from stdin:

    $ cat pod.yaml | ensemble delete -f -

` + d.Flags().Help()
}

func (d *DeleteCommand) Flags() *flagset.Flagset {
	f := d.NewFlagSet("delete")

	f.StringFlag(&flagset.StringFlag{
		Name:  "f",
		Value: &d.filename,
		Usage: "Filename containing the resource to delete",
	})

	f.BoolFlag(&flagset.BoolFlag{
		Name:  "R",
		Value: &d.recursive,
		Usage: "Process the directory used in -f, --filename recursively",
	})

	return f
}

// Run implements the cli.Command interface
func (d *DeleteCommand) Run(args []string) int {
	flags := d.Flags()
	if err := flags.Parse(args); err != nil {
		d.UI.Error(err.Error())
		return 1
	}

	comps, err := readComponents(d.filename, d.recursive)
	if err != nil {
		d.UI.Error(err.Error())
		return 1
	}

	comp, err := parseComponentFromFile([]byte(comps[0]))
	if err != nil {
		d.UI.Error(err.Error())
		return 1
	}
	comp.Action = proto.Component_DELETE

	clt, err := d.Conn()
	if err != nil {
		d.UI.Error(err.Error())
		return 1
	}
	if _, err := clt.Apply(context.Background(), comp); err != nil {
		d.UI.Error(err.Error())
		return 1
	}
	return 0
}
