// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations
// and renders the defined HTML document.
package quill

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Render takes the Delta array of insert operations and, optionally, a map of custom Op rendering functions that may
// customizing how operations of certain types are rendered. The returned byte slice is the rendered HTML.
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
	Data  string            // the text to insert or the value of the embed object (http://quilljs.com/docs/delta/#embeds)
	Type  string            // the type of the op (typically "string", but you can register any other type)
	Attrs map[string]string // key is attribute name; value is either value string or "y" (meaning true) or "n" (meaning false)
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

// An OpHandler takes the previous Op (which is nil if the current Op is the first) and the current Op and writes
// the current Op to buf. The handler may need to know the previous Op to decide whether to begin writing the current
// Op data only after closing the HTML tag that set attributes on the previous Op; all attributes of the previous Op
// should be checked before writing the current one.
type OpHandler func(prev *Op, cur *Op, buf *bytes.Buffer)

var handlers = map[string]OpHandler{
	"string": func(prev *Op, cur *Op, buf *bytes.Buffer) {
		return
	},
}

func prevHas(o *Op, attr string) bool {
	if o == nil {
		return false
	}
	return o.Attrs[attr] == "y"
}
