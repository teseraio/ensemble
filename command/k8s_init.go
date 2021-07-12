package command

import (
	"github.com/mitchellh/cli"
	"github.com/teseraio/ensemble/command/flagset"
	"github.com/teseraio/ensemble/k8s"
)

// K8sInitCommand is the command to init a kubernetes cluster
type K8sInitCommand struct {
	UI cli.Ui

	name     string
	backend  string
	replicas int
}

// Help implements the cli.Command interface
func (k *K8sInitCommand) Help() string {
	return `Usage: ensemble k8s init

  Display a YAML file for a cluster deployment.

` + k.Flags().Help()
}

func (k *K8sInitCommand) Flags() *flagset.Flagset {
	f := flagset.NewFlagSet("init")

	f.StringFlag(&flagset.StringFlag{
		Name:  "name",
		Value: &k.name,
		Usage: "Name of the cluster",
	})

	f.StringFlag(&flagset.StringFlag{
		Name:  "backend",
		Value: &k.backend,
		Usage: "Backend or database to deploy",
	})

	f.IntFlag(&flagset.IntFlag{
		Name:  "replicas",
		Value: &k.replicas,
		Usage: "Number of replicas",
	})

	return f
}

// Synopsis implements the cli.Command interface
func (k *K8sInitCommand) Synopsis() string {
	return "Display a YAML file for a cluster deployment"
}

// Run implements the cli.Command interface
func (k *K8sInitCommand) Run(args []string) int {
	flags := k.Flags()
	if err := flags.Parse(args); err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	obj := map[string]interface{}{
		"Name":     k.name,
		"Backend":  k.backend,
		"Replicas": k.replicas,
	}
	raw, err := k8s.RunTmpl2("kind-cluster", obj)
	if err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	yamlRes, err := convertJSONToYaml(raw)
	if err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	k.UI.Output(string(yamlRes))
	return 0
}
