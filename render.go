// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations and
// renders the defined HTML document.
package quill

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func Render(ops []byte, customHandlers map[string]OpHandler) ([]byte, error) {

	var ro []rawOp
	err := json.Unmarshal(ops, &ro)
	if err != nil {
		return nil, err
	}

	var (
		html      = new(bytes.Buffer)
		prev, cur *Op
	)

	for i := range ro {
		cur, err = rawOpToOp(ro[i])
		if err != nil {
			return nil, err
		}
		if _, ok := customHandlers[cur.Type]; ok {
			customHandlers[cur.Type](prev, cur, html)
		} else {
			handlers[cur.Type](prev, cur, html)
		}
		prev = cur
	}

	return html.Bytes(), nil

}

type Op struct {
	Data  string            // the string to insert or the value of the single item in the embed object
	Type  string            // the type of op (typically "string", but user can register any other type)
	Attrs map[string]string // key is attribute name; value is either value string or "y" (indicating true)
}

type rawOp map[string]interface{}

func rawOpToOp(ro rawOp) (*Op, error) {
	if _, ok := ro["insert"]; !ok {
		return nil, fmt.Errorf("op %q lacks an insert", ro)
	}
	o := new(Op)
	if str, ok := ro["insert"].(string); ok {
		// This op is a simple string insert.
		o.Type = "string"
		o.Data = str
	} else if mapStrIntf, ok := ro["insert"].(map[string]interface{}); ok {
		for mk := range mapStrIntf {
			ins := make(map[string]string)
			ins[mk] = extractString(mapStrIntf[mk])
		}
	}
	if _, ok := ro["attributes"]; ok {
		o.Attrs = make(map[string]string)
		if attrs, ok := ro["attributes"].(map[string]interface{}); ok {
			for attr := range attrs {
				o.Attrs[attr] = extractString(attrs[attr])
			}
		}
	}
	return o, nil
}

func extractString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val == true {
			return "y"
		}
		return "n"
	default:
		return ""
	}
}

type OpHandler func(prev, cur *Op, buf *bytes.Buffer)

var handlers = map[string]OpHandler{
	"string": func(prev, cur *Op, buf *bytes.Buffer) {
		return
	},
}
