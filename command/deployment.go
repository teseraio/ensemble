package command

import (
	"github.com/mitchellh/cli"
)

type DeploymentCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *DeploymentCommand) Help() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (c *DeploymentCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (c *DeploymentCommand) Run(args []string) int {
	return 0
}
