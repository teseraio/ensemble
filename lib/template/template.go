package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

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
		return nil, fmt.Errorf("failed to template: %v", err)
	}

	// fmt.Println(string(buf1.Bytes()))

	// convert quotes
	out1 := string(buf1.Bytes())
	out1 = strings.Replace(out1, "&#34;", "\"", -1)

	// if the content is json, prettify it
	if strings.HasPrefix(out1, "{") {
		output, err := prettifyJSON([]byte(out1))
		if err != nil {
			return nil, fmt.Errorf("failed to prettify: %v", err)
		}
		out1 = string(output)
	}
	return []byte(out1), nil
}

func prettifyJSON(buf []byte) ([]byte, error) {
	var dst bytes.Buffer
	if err := json.Indent(&dst, buf, "", "\t"); err != nil {
		return nil, err
	}
	return dst.Bytes(), nil
}
