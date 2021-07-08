package command

import (
	"github.com/mitchellh/cli"
)

// K8sCommand is the command for kubernetes
type K8sCommand struct {
	UI cli.Ui
}

// Help implements the cli.Command interface
func (k *K8sCommand) Help() string {
	return `Usage: ensemble <subcommand>
  This command subgroups actions to interact with Kubernetes.

  $ ensemble k8s artifacts

  $ ensemble k8s init`
}

// Synopsis implements the cli.Command interface
func (k *K8sCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (k *K8sCommand) Run(args []string) int {
	return cli.RunResultHelp
}
