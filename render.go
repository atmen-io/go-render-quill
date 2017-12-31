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
func Render(ops []byte, th map[string]TypeHandler, ah map[string]AttrHandler) ([]byte, error) {

	var ro []map[string]interface{}
	err := json.Unmarshal(ops, &ro)
	if err != nil {
		return nil, err
	}

	var html = new(bytes.Buffer)
	var o *Op

	for i := range ro {
		o, err = rawOpToOp(ro[i])
		if err != nil {
			return nil, err
		}
		if _, ok := th[o.Type]; ok {
			th[o.Type](o, html)
		} else {
			opHandlers[o.Type](o, html)
		}
	}

	return html.Bytes(), nil

}

type Op struct {
	Data  string            // the text to insert or the value of the embed object (http://quilljs.com/docs/delta/#embeds)
	Type  string            // the type of the op (typically "string", but you can register any other type)
	Attrs map[string]string // key is attribute name; value is either value string or "y" (meaning true) or "n" (meaning false)
}

// HasAttr says if the Op is not nil and has the attribute set to the value "y".
func (o *Op) HasAttr(attr string) bool {
	return o != nil && o.Attrs[attr] == "y"
}

// ClosePrev checks if the previous Op opened any attribute tags that are not supposed to be set on the current Op and closes
// those tags in the opposite order in which they were opened.
func (o *Op) ClosePrevAttrs(buf *bytes.Buffer) {
	for i := len(attrStates); i > 0; i-- {
		indx := i - 1
		if !o.HasAttr(attrStates[indx]) {
		}
	}
}

func (o *Op) OpenAttrs(buf *bytes.Buffer) {
}

// attrStates lists the tags currently open in the order in which they were opened.
var attrStates = make([]string, 0, 2)

// rawOpToOp takes a raw Delta op as extracted from the JSON and turns it into an Op to make it usable for rendering.
func rawOpToOp(ro map[string]interface{}) (*Op, error) {
	if _, ok := ro["insert"]; !ok {
		return nil, fmt.Errorf("op %q lacks an insert", ro)
	}
	o := new(Op)
	if str, ok := ro["insert"].(string); ok {
		// This op is a simple string insert.
		o.Type = "string"
		o.Data = str
	} else if mapStrIntf, ok := ro["insert"].(map[string]interface{}); ok {
		if _, ok = mapStrIntf["insert"]; !ok {
			return nil, fmt.Errorf("op %q lacks an insert", ro)
		}
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

// An OpHandler takes the previous Op (which is nil if the current Op is the first) and the current Op and writes the
// current Op to buf. Each handler should check the previous Op to see if it has attributes that are not set on the current
// Op and close the appropriate HTML tags before writing the current Op; also the handler should not needlessly open up a
// tag for an attribute if it was already opened for the previous Op. This ensures that the rendered HTML is lean.
//type OpHandler func(o *Op, buf *bytes.Buffer)

type OpHandler struct {
	Open, Write, Close func(o *Op, buf *bytes.Buffer)
}

var Text = OpHandler{
	Open: func(o *Op, buf *bytes.Buffer) {
	},
	Write: func(o *Op, buf *bytes.Buffer) {
	},
	Close: func(o *Op, buf *bytes.Buffer) {
	},
}

var opHandlers = map[string]OpHandler{
	"text": Text,
}

func extractString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val == true {
			return "y"
		}
	}
	return ""
}
