package command

import (
	"context"

	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/k8s"
	"github.com/teseraio/ensemble/operator/proto"
	"gopkg.in/yaml.v2"
)

type ApplyCommand struct {
	filename  string
	recursive bool

	Meta
}

// Help implements the cli.Command interface
func (c *ApplyCommand) Synopsis() string {
	return "Apply a configuration to a resource by filename or stdin"
}

// Synopsis implements the cli.Command interface
func (c *ApplyCommand) Help() string {
	return `Usage: ensemble apply [options]

  Apply a configuration to a resource by filename or stdin.

  Apply a single yaml file:

    $ ensemble apply -f pod.yaml

  Apply multiple files from a directory:

    $ ensemble apply -f ./components

  Apply a configuration from stdin:

    $ cat pod.yaml | ensemble apply -f -

` + c.Flags().Help()
}

func (c *ApplyCommand) Flags() *flagset.Flagset {
	f := c.NewFlagSet("apply")

	f.StringFlag(&flagset.StringFlag{
		Name:  "f",
		Value: &c.filename,
		Usage: "Filename containing the resource to delete",
	})

	f.BoolFlag(&flagset.BoolFlag{
		Name:  "R",
		Value: &c.recursive,
		Usage: "Process the directory used in -f, --filename recursively",
	})

	return f
}

// Run implements the cli.Command interface
func (c *ApplyCommand) Run(args []string) int {
	f := c.Flags()

	err := f.Parse(args)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if c.filename == "" {
		c.UI.Error("-f must be set")
		return 1
	}

	comps, err := readComponents(c.filename, c.recursive)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	comp, err := parseComponentFromFile([]byte(comps[0]))
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	comp.Action = proto.Component_CREATE

	clt, err := c.Conn()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	if _, err := clt.Apply(context.Background(), comp); err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	return 0
}

func parseComponentFromFile(raw []byte) (*proto.Component, error) {
	var item *k8s.Item
	if err := yaml.Unmarshal(raw, &item); err != nil {
		return nil, err
	}
	spec, err := k8s.DecodeItem(item)
	if err != nil {
		return nil, err
	}
	comp := &proto.Component{
		Name:     item.Metadata.Name,
		Spec:     spec,
		Metadata: item.Metadata.Labels,
	}
	return comp, nil
}
