package command

import (
	"context"
	"io/ioutil"

	"github.com/teseraio/ensemble/k8s"
	"github.com/teseraio/ensemble/operator/proto"
	"gopkg.in/yaml.v2"
)

type ApplyCommand struct {
	Meta
}

// Help implements the cli.Command interface
func (c *ApplyCommand) Help() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (c *ApplyCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (c *ApplyCommand) Run(args []string) int {
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
