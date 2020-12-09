package command

import (
	"fmt"

	"github.com/teseraio/ensemble/k8s"
)

// K8sInitCommand is the command to init a kubernetes cluster
type K8sInitCommand struct {
	Meta
}

// Help implements the cli.Command interface
func (k *K8sInitCommand) Help() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (k *K8sInitCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (k *K8sInitCommand) Run(args []string) int {
	var name, backend string
	var replicas int

	flags := k.Meta.FlagSet("k8s init")
	flags.Usage = func() {}

	flags.StringVar(&name, "name", "", "")
	flags.StringVar(&backend, "backend", "", "")
	flags.IntVar(&replicas, "replicas", 1, "")

	if err := flags.Parse(args); err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	obj := map[string]interface{}{
		"Name":     name,
		"Backend":  backend,
		"Replicas": replicas,
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

	fmt.Println(string(yamlRes))
	return 0
}
