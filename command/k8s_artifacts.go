package command

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/teseraio/ensemble/k8s"
	"github.com/teseraio/ensemble/lib/template"
	"gopkg.in/yaml.v2"
)

// K8sArtifactsCommand is the command for kubernetes
type K8sArtifactsCommand struct {
	Meta

	service bool
	crd     bool
	dev     bool
}

// Help implements the cli.Command interface
func (k *K8sArtifactsCommand) Help() string {
	return ""
}

// Synopsis implements the cli.Command interface
func (k *K8sArtifactsCommand) Synopsis() string {
	return ""
}

// Run implements the cli.Command interface
func (k *K8sArtifactsCommand) Run(args []string) int {
	flags := k.Meta.FlagSet("k8s artifacts")
	flags.Usage = func() {}

	flags.BoolVar(&k.dev, "dev", false, "")
	flags.BoolVar(&k.service, "service", false, "")
	flags.BoolVar(&k.crd, "crd", false, "")

	if err := flags.Parse(args); err != nil {
		k.UI.Error(err.Error())
		return 1
	}

	if !k.service && !k.crd {
		k.service, k.crd = true, true
	}

	// TODO: By default load the docker image that represents the
	// version of the binary.

	var image, pullPolicy string
	if k.dev {
		// load from local repository
		image = "ensemble:dev"
		pullPolicy = "Never"
	} else {
		// load the latest docker image
		image = "teseraio/ensemble:latest"
		pullPolicy = "Always"
	}

	artifacts := []string{}

	if k.crd {
		crdObjs, err := listAssetsByPrefix("crd-", nil)
		if err != nil {
			k.UI.Error(err.Error())
			return 1
		}
		artifacts = append(artifacts, crdObjs...)
	}

	if k.service {
		objs := map[string]interface{}{
			"Image":           image,
			"ImagePullPolicy": pullPolicy,
		}
		srvObjs, err := listAssetsByPrefix("srv-", objs)
		if err != nil {
			k.UI.Error(err.Error())
			return 1
		}
		artifacts = append(artifacts, srvObjs...)
	}

	k.UI.Output(strings.Join(artifacts, "---\n"))
	return 0
}

func listAssetsByPrefix(prefix string, obj interface{}) ([]string, error) {
	artifacts := []string{}

	// sort to return the files always in the same order
	assetNames := k8s.AssetNames()
	sort.Strings(assetNames)

	for _, n := range assetNames {
		if strings.Contains(n, prefix) {
			asset := k8s.MustAsset(n)
			if strings.HasSuffix(n, ".template") {
				// render the template
				res, err := template.RunTmpl(string(asset), obj)
				if err != nil {
					return nil, err
				}
				asset = res
			}
			yaml, err := convertJSONToYaml(asset)
			if err != nil {
				return nil, err
			}
			artifacts = append(artifacts, string(yaml))
		}
	}
	return artifacts, nil
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
