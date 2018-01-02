package quill

import (
	"fmt"
	"strconv"
)

type rawOp struct {
	Insert interface{}            `json:"insert"`
	Attrs  map[string]interface{} `json:"attributes"`
}

// makeOp takes a raw Delta op as extracted from the JSON and turns it into an Op to make it usable for rendering.
func (ro *rawOp) makeOp(o *Op) error {

	if ro.Insert == nil {
		return fmt.Errorf("op %q lacks an insert", ro)
	}

	switch ins := ro.Insert.(type) {
	case string:
		// This op is a simple string insert.
		o.Type = "text"
		o.Data = ins
	case map[string]interface{}:
		if _, ok := ins["insert"]; !ok {
			return fmt.Errorf("op %q lacks an insert", ro)
		}
		// There should be only one item in the map (the element's key being the insert type).
		for mk := range ins {
			o.Type = mk
			o.Data = extractString(ins[mk])
			break
		}
	default:
		return fmt.Errorf("op %q lacks an insert", ro)
	}

	if ro.Attrs != nil {
		for attr := range ro.Attrs { // the map was already made
			o.Attrs[attr] = extractString(ro.Attrs[attr])
		}
	} else {
		// Clear the map for later reuse.
		for k := range o.Attrs {
			delete(o.Attrs, k)
		}
	}

	return nil

}

func extractString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val == true {
			return "y"
		}
	case float64:
		return strconv.FormatFloat(val, 'f', 0, 64)
	}
	return ""
}
