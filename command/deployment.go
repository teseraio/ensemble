package command

import (
	"github.com/mitchellh/cli"
)

type DeploymentCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (c *DeploymentCommand) Help() string {
	return `Usage: ensemble deployment <subcommand>

  This command groups actions to interact with deployments.
  
  List the running deployments:

    $ ensemble deployment list
  
  Check the status of a specific deployment:

    $ ensemble deployment status <deployment_id>`
}

// Synopsis implements the cli.Command interface
func (c *DeploymentCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (c *DeploymentCommand) Run(args []string) int {
	return cli.RunResultHelp
}
