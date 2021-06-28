package command

import (
	"context"

	"github.com/teseraio/ensemble/operator/proto"
)

type DeleteCommand struct {
	Meta
}

// Help implements the cli.Command interface
func (d *DeleteCommand) Help() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (d *DeleteCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (d *DeleteCommand) Run(args []string) int {
	flags := d.FlagSet("delete")
	if err := flags.Parse(args); err != nil {
		d.UI.Error(err.Error())
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		d.UI.Error("at least one file expected")
		return 1
	}

	comp, err := readComponentFromFile(args[0])
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
