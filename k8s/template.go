package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

//go:generate go-bindata -pkg k8s -o ./bindata.go ./resources

func marshalToJSON(v interface{}, trim bool) string {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("BUG: %v", err))
	}
	res := string(buf)
	if trim {
		res = strings.TrimPrefix(res, "{")
		res = strings.TrimSuffix(res, "}")
	}
	return res
}

func RunTmpl2(name string, obj interface{}) ([]byte, error) {
	content, err := Asset(fmt.Sprintf("resources/%s.template", name))
	if err != nil {
		return nil, err
	}
	return RunTmpl(string(content), obj)
}

func separator(s string) func() string {
	i := -1
	return func() string {
		i++
		if i == 0 {
			return ""
		}
		return s
	}
}

func RunTmpl(tmpl string, obj interface{}) ([]byte, error) {
	funcMap := template.FuncMap{
		// map converts an object (i.e. struct, map) into json
		"map": func(v interface{}) string { return marshalToJSON(v, false) },

		// mapX is the same as map but removing the outer JSON brackets
		"mapX": func(v interface{}) string { return marshalToJSON(v, true) },

		"title":     func(v interface{}) string { return strings.Title(v.(string)) },
		"separator": separator,
	}
	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return nil, err
	}
	buf1 := new(bytes.Buffer)
	if err = t.Execute(buf1, obj); err != nil {
		return nil, err
	}

	// convert quotes
	out1 := string(buf1.Bytes())
	out1 = strings.Replace(out1, "&#34;", "\"", -1)

	output, err := prettifyJSON([]byte(out1))
	if err != nil {
		return nil, err
	}
	return output, nil
}

func prettifyJSON(buf []byte) ([]byte, error) {
	var dst bytes.Buffer
	if err := json.Indent(&dst, buf, "", "\t"); err != nil {
		return nil, err
	}
	return dst.Bytes(), nil
}
