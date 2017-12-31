// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations and
// renders the defined HTML document.
package quill

import (
	"bytes"
	"encoding/json"
)

func Render(ops []byte) ([]byte, error) {

	var ro []rawOp
	err := json.Unmarshal(ops, &ro)
	if err != nil {
		return nil, err
	}

	html := new(bytes.Buffer)

	oo := make([]*Op, 0, len(ro))
	for i := range ro {
		oo[i], err = rawOpToOp(ro[i])
		if err != nil {
			return nil, err
		}
		var prev *Op
		if i > 0 {
			prev = oo[i-1]
		}
		handlers[oo[i].Type](prev, oo[i], html)
	}

	return html.Bytes(), nil

}

type Op struct {
	Data string            // the string to insert or the value of the single item in the embed object
	Type   string            // the type of op (typically "string", but user can register any other type)
	Attrs  map[string]string // key is attribute name; value is either value string or "y" (indicating true)
}

type rawOp map[string]interface{}

func rawOpToOp(ro rawOp) (*Op, error) {
	o := &Op{}
	for k := range ro {
		switch k {
		case "insert":
			if str, ok := ro[k].(string); ok {
				// This op is a simple string insert.
				o.Type = "string"
				o.Data = str
			} else if mapStrIntf, ok := ro[k].(map[string]interface{}); ok {

			}
			//switch ro[k].(type) {
			//case string:
			//	// This op is a simple string insert.
			//	o.Insert = ro[k].(string)
			//}
		case "attributes":
			switch ro[k].(type)
		}
	}
	return o, nil
}

func extractString(intf interface{}) string {

}

type OpHandler func(prev, cur *Op, buf *bytes.Buffer)

var handlers = map[string]OpHandler{
	"string": func(prev, cur *Op, buf *bytes.Buffer) {
		return
	},
}
