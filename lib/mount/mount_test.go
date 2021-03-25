package mount

import (
	"reflect"
	"testing"

	"github.com/teseraio/ensemble/operator/proto"
)

func TestMountGroup(t *testing.T) {
	cases := []struct {
		files []string
		res   []*MountPoint
	}{
		{
			[]string{
				"/a/b/c",
				"/a/b/c/d",
				"/a/b/e",
				"/b/c",
				"/a/d",
			},
			[]*MountPoint{
				{
					Path: "/a",
					Files: map[string]string{
						"/a/b/c":   "",
						"/a/b/c/d": "",
						"/a/b/e":   "",
						"/a/d":     "",
					},
				},
				{
					Path: "/b",
					Files: map[string]string{
						"/b/c": "",
					},
				},
			},
		},
	}

	for _, c := range cases {
		input := []*proto.NodeSpec_File{}
		for _, file := range c.files {
			input = append(input, &proto.NodeSpec_File{
				Name:    file,
				Content: "",
			})
		}
		found, err := CreateMountPoints(input)
		if err != nil {
			t.Fatal(err)
		}
		/*
			for indx := range found {
				fmt.Println(found[indx])
				fmt.Println(c.res[indx])
			}
		*/
		if !reflect.DeepEqual(found, c.res) {
			t.Fatal("bad")
		}
	}
}
