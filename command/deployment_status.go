package command

import (
	"context"
	"fmt"

	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/operator/proto"
)

type DeploymentStatusCommand struct {
	Meta
}

// Help implements the cli.Command interface
func (c *DeploymentStatusCommand) Help() string {
	return `Usage: ensemble deployment status <id>

  Display the status of a specific deployment.

` + c.Flags().Help()
}

func (c *DeploymentStatusCommand) Flags() *flagset.Flagset {
	return c.NewFlagSet("deployment status")
}

// Synopsis implements the cli.Command interface
func (c *DeploymentStatusCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (c *DeploymentStatusCommand) Run(args []string) int {
	flags := c.Flags()
	if err := flags.Parse(args); err != nil {
		panic(err)
	}

	args = flags.Args()
	if len(args) != 1 {
		panic("bad")
	}
	depID := args[0]

	clt, err := c.Conn()
	if err != nil {
		panic(err)
	}

	dep, err := clt.GetDeployment(context.Background(), &proto.GetDeploymentReq{Cluster: depID})
	if err != nil {
		panic(err)
	}

	fmt.Println(formatDeployment(dep))
	return 0
}

func formatDeployment(dep *proto.Deployment) string {

	base := formatKV([]string{
		fmt.Sprintf("Name|%s", dep.Name),
		fmt.Sprintf("Backend|%s", dep.Backend),
		fmt.Sprintf("Version|%d", dep.Sequence),
		fmt.Sprintf("Status|%s", dep.Status),
	})

	if len(dep.Instances) != 0 {
		rows := make([]string, len(dep.Instances)+1)
		rows[0] = "Name|Healthy|Status"
		for i, d := range dep.Instances {
			rows[i+1] = fmt.Sprintf("%s|%v|%s",
				d.Name,
				d.Healthy,
				d.Status,
			)
		}
		base += "\n\n" + formatList(rows)
	}
	return base
}
