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

	comps := []*proto.Component{}
	for _, arg := range args {
		raw, err := ioutil.ReadFile(arg)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		var item *k8s.Item
		if err := yaml.Unmarshal(raw, &item); err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		spec, err := k8s.DecodeItem(item)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		comps = append(comps, &proto.Component{
			Name:     item.Metadata.Name,
			Spec:     spec,
			Metadata: item.Metadata.Labels,
			Action:   proto.Component_CREATE,
		})
	}

	clt, err := c.Conn()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	for _, comp := range comps {
		if _, err := clt.Apply(context.Background(), comp); err != nil {
			c.UI.Error(err.Error())
			return 1
		}
	}
	return 0
}
