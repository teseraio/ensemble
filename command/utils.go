package command

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func readComponents(filename string, recursive bool) ([]string, error) {
	var content []byte
	if filename == "-" {
		// stdin
		lines := []string{}
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if scanner.Err() != nil {
			return nil, fmt.Errorf("failed to read from stdin: %v", scanner.Err())
		}
		content = []byte(strings.Join(lines, "\n"))
	} else if _, err := os.Stat(filename); !os.IsNotExist(err) {
		// directory
		files, err := readDir(filename, recursive)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %v", err)
		}
		raws := []string{}
		for _, file := range files {
			elem, err := ioutil.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read filename: %v", file)
			}
			raws = append(raws, string(elem))
		}
		content = []byte(strings.Join(raws, "---"))
	} else {
		// filename
		if content, err = ioutil.ReadFile(filename); err != nil {
			return nil, fmt.Errorf("failed to read filename: %v", err)
		}
	}
	comps := strings.Split(string(content), "---")
	return comps, nil
}

func readDir(dirname string, recursive bool) ([]string, error) {
	result := []string{}
	var err error

	isGoodFile := func(i os.FileInfo) bool {
		if i.IsDir() {
			return false
		}
		ext := filepath.Ext(i.Name())
		return ext == ".yaml" || ext == ".yml"
	}

	if recursive {
		err = filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
			if isGoodFile(info) {
				result = append(result, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		files, err := ioutil.ReadDir(dirname)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if isGoodFile(f) {
				result = append(result, filepath.Join(dirname, f.Name()))
			}
		}
	}

	return result, nil
}
