package command

import (
	"context"
	"fmt"

	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/operator/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DeploymentListCommand struct {
	Meta
}

// Help implements the cli.Command interface
func (c *DeploymentListCommand) Help() string {
	return `Usage: ensemble deployment list

  List the running deployments.

` + c.Flags().Help()
}

func (c *DeploymentListCommand) Flags() *flagset.Flagset {
	return c.NewFlagSet("deployment list")
}

// Synopsis implements the cli.Command interface
func (c *DeploymentListCommand) Synopsis() string {
	return "List the running deployments"
}

// Run implements the cli.Command interface
func (c *DeploymentListCommand) Run(args []string) int {
	flags := c.Flags()
	if err := flags.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	clt, err := c.Conn()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	resp, err := clt.ListDeployments(context.Background(), &emptypb.Empty{})
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	fmt.Println(formatDeployments(resp.Deployments))
	return 0
}

func formatDeployments(deps []*proto.Deployment) string {
	if len(deps) == 0 {
		return "No deployments found"
	}

	rows := make([]string, len(deps)+1)
	rows[0] = "ID|Name|Backend|Version|Status"
	for i, d := range deps {
		rows[i+1] = fmt.Sprintf("%s|%s|%s|%d|%s",
			d.Id,
			d.Name,
			d.Backend,
			d.Sequence,
			d.Status)
	}
	return formatList(rows)
}
