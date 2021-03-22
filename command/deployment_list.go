package command

import (
	"context"
	"fmt"

	"github.com/teseraio/ensemble/operator/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DeploymentListCommand struct {
	Meta
}

// Help implements the cli.Command interface
func (c *DeploymentListCommand) Help() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (c *DeploymentListCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (c *DeploymentListCommand) Run(args []string) int {
	flags := c.FlagSet("deployment list")
	if err := flags.Parse(args); err != nil {
		panic(err)
	}

	clt, err := c.Conn()
	if err != nil {
		panic(err)
	}

	resp, err := clt.ListDeployments(context.Background(), &emptypb.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(formatDeployments(resp.Deployments))
	return 0
}

func formatDeployments(deps []*proto.Deployment) string {
	if len(deps) == 0 {
		return "No deployments found"
	}

	rows := make([]string, len(deps)+1)
	rows[0] = "Name|Backend|Version|Status"
	for i, d := range deps {
		rows[i+1] = fmt.Sprintf("%s|%s|%d|%s",
			d.Name,
			d.Backend,
			d.Sequence,
			d.Status)
	}
	return formatList(rows)
}
