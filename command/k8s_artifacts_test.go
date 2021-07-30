package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
)

func TestCommands_K8sArtifacts_Filter(t *testing.T) {
	ui := cli.NewMockUi()
	cmd := &K8sArtifactsCommand{Meta: Meta{UI: ui}}

	cases := []struct {
		args []string
		num  int
	}{
		{
			[]string{""},
			7,
		},
		{
			[]string{"--crd"},
			2,
		},
		{
			[]string{"--service"},
			5,
		},
		{
			[]string{"--crd", "--service"},
			7,
		},
	}
	for _, c := range cases {
		code := cmd.Run(c.args)
		assert.Equal(t, code, 0)

		num := len(strings.Split(ui.OutputWriter.String(), "---"))
		assert.Equal(t, num, c.num)

		ui.OutputWriter.Reset()
	}
}
