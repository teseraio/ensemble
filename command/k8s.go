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

  This command groups actions to display Kubernetes YAML files to interact with Ensemble.

  Display the YAML artifacts to deploy Ensemble:

    $ ensemble k8s artifacts

  Display a YAML file for a cluster deployment:

    $ ensemble k8s init`
}

// Synopsis implements the cli.Command interface
func (k *K8sCommand) Synopsis() string {
	return "Create Kubernets YAML artifacts"
}

// Run implements the cli.Command interface
func (k *K8sCommand) Run(args []string) int {
	return cli.RunResultHelp
}
