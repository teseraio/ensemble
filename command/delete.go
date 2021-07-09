package command

import (
	"fmt"

	"github.com/teseraio/ensemble/command/flagset"
)

type DeleteCommand struct {
	Meta

	filename  string
	recursive bool
}

// Synopsis implements the cli.Command interface
func (d *DeleteCommand) Synopsis() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (d *DeleteCommand) Help() string {
	return `Usage: ensemble delete [options]

  $ ensemble delete -f pod.yaml

  $ ensemble delete -f ./components

` + d.Flags().Help()
}

func (d *DeleteCommand) Flags() *flagset.Flagset {
	f := d.NewFlagSet("apply")

	f.StringFlag(&flagset.StringFlag{
		Name:  "f",
		Value: &d.filename,
		Usage: "Path of the file to apply",
	})

	f.BoolFlag(&flagset.BoolFlag{
		Name:  "R",
		Value: &d.recursive,
		Usage: "Follow the directory in -f recursively",
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
	fmt.Println(comps)

	panic("TODO")

	/*
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
	*/

	return 0
}
