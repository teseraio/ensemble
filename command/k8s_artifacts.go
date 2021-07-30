package command

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/teseraio/ensemble/command/artifacts"
	"github.com/teseraio/ensemble/command/flagset"
	engine "github.com/teseraio/helm-template-engine"
	"github.com/teseraio/helm-template-engine/chart"
	"github.com/teseraio/helm-template-engine/loader"
	"gopkg.in/yaml.v2"
)

//go:generate go-bindata -pkg artifacts -o ./artifacts/artifacts.go ../charts/operator/...

// K8sArtifactsCommand is the command for kubernetes
type K8sArtifactsCommand struct {
	Meta

	service     bool
	crd         bool
	dev         bool
	serviceAcct string
}

// Help implements the cli.Command interface
func (k *K8sArtifactsCommand) Help() string {
	return `Usage: ensemble k8s artifacts <options>

  Display the YAML artifacts to deploy Ensemble.

  Print only the CRD files:

    $ ensemble k8s artifacts --crd

  Print the service object to deploy the operator on Kubernetes:

    $ ensemble k8s artifacts --service

` + k.Flags().Help()
}

func (c *K8sArtifactsCommand) Flags() *flagset.Flagset {
	f := flagset.NewFlagSet("k8s artifacts")

	f.BoolFlag(&flagset.BoolFlag{
		Name:  "dev",
		Value: &c.dev,
		Usage: "Use local development image in service artifacts",
	})

	f.BoolFlag(&flagset.BoolFlag{
		Name:  "service",
		Value: &c.service,
		Usage: "Filter by service artifacts",
	})

	f.BoolFlag(&flagset.BoolFlag{
		Name:  "crd",
		Value: &c.crd,
		Usage: "Filter by CRD artifacts",
	})

	f.StringFlag(&flagset.StringFlag{
		Name:  "service-account",
		Value: &c.serviceAcct,
		Usage: "Name of the service account to deploy the operator",
	})

	return f
}

// Synopsis implements the cli.Command interface
func (k *K8sArtifactsCommand) Synopsis() string {
	return "Display the YAML artifacts to deploy Ensemble"
}

// Run implements the cli.Command interface
func (k *K8sArtifactsCommand) Run(args []string) int {
	flags := k.Flags()
	if err := flags.Parse(args); err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	if !k.service && !k.crd {
		k.service, k.crd = true, true
	}

	cc, err := loader.LoadFiles(artifacts.GetArtifacts())
	if err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	chartVals := map[string]interface{}{
		"dev": k.dev,
		"serviceAccount": map[string]interface{}{
			"name": k.serviceAcct,
		},
	}
	opts := chart.ReleaseOptions{
		Name: "ensemble",
	}
	vals, err := chart.ToRenderValues(cc, chartVals, opts)
	if err != nil {
		k.UI.Error(err.Error())
		return 1
	}
	result, err := engine.Render(cc, vals)
	if err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	artifacts := artifactsOrder{}

	addArfifact := func(data string) {
		data = strings.TrimPrefix(data, "---")
		data = strings.Trim(data, "\n")
		artifacts = append(artifacts, data)
	}

	if k.crd {
		// crds
		for _, i := range cc.Files {
			if i.Name != ".helmignore" {
				addArfifact(string(i.Data))
			}
		}
	}
	if k.service {
		// operator
		for _, i := range result {
			addArfifact(i)
		}
	}

	sort.Sort(artifacts)
	k.UI.Output(strings.Join(artifacts, "\n---\n"))

	return 0
}

func convertJSONToYaml(in []byte) ([]byte, error) {
	var res map[string]interface{}
	if err := json.Unmarshal(in, &res); err != nil {
		return nil, err
	}
	out, err := yaml.Marshal(res)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var artifactOrders = []string{
	"ServiceAccount",
	"ClusterRole",
	"ClusterRoleBinding",
	"Deployment",
	"Service",
	"CustomResourceDefinition",
}

type artifactsOrder []string

func (a artifactsOrder) Len() int {
	return len(a)
}

func (a artifactsOrder) Less(i, j int) bool {
	return parseKind(a[i]) < parseKind(a[j])
}

func (a artifactsOrder) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func parseKind(s string) int {
	re := regexp.MustCompile(`kind: (\w*)`)
	match := re.FindStringSubmatch(s)
	kind := strings.TrimSpace(match[1])

	for indx, k := range artifactOrders {
		if k == kind {
			return indx
		}
	}
	return -1
}
