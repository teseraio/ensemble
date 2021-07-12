package command

import (
	"fmt"
	"io/ioutil"

	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/k8s"
	"github.com/teseraio/ensemble/operator/proto"
	"gopkg.in/yaml.v2"
)

type ApplyCommand struct {
	filename  string
	directory string

	Meta
}

// Help implements the cli.Command interface
func (c *ApplyCommand) Synopsis() string {
	return "Apply a configuration to a resource by filename or stdin"
}

// Synopsis implements the cli.Command interface
func (c *ApplyCommand) Help() string {
	return `Usage: ensemble apply [options]

  $ ensemble apply -f pod.yaml

  $ ensemble apply -d ./components

` + c.Flags().Help()
}

func (c *ApplyCommand) Flags() *flagset.Flagset {
	f := c.NewFlagSet("apply")

	f.StringFlag(&flagset.StringFlag{
		Name:  "f",
		Value: &c.filename,
		Usage: "Path of the file to apply",
	})

	f.StringFlag(&flagset.StringFlag{
		Name:  "d",
		Value: &c.directory,
		Usage: "Path of the directory to apply",
	})

	return f
}

// Run implements the cli.Command interface
func (c *ApplyCommand) Run(args []string) int {

	f := c.Flags()
	if err := f.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	fmt.Println(c.filename)
	fmt.Println(c.directory)

	/*
		flags := c.FlagSet("apply")
		if err := flags.Parse(args); err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		args = flags.Args()
		if len(args) != 1 {
			c.UI.Error("at least one file expected")
			return 1
		}

		comp, err := readComponentFromFile(args[0])
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
	*/
	return 0
}

func readComponentFromFile(path string) (*proto.Component, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
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
