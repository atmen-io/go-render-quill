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
func (ro *rawOp) makeOp() (*Op, error) {
	if ro.Insert == nil {
		return nil, fmt.Errorf("op %q lacks an insert", ro)
	}
	o := new(Op)
	if str, ok := ro.Insert.(string); ok {
		// This op is a simple string insert.
		o.Data = str
		o.Type = "text"
	} else if mapStrIntf, ok := ro.Insert.(map[string]interface{}); ok {
		if _, ok = mapStrIntf["insert"]; !ok {
			return nil, fmt.Errorf("op %q lacks an insert", ro)
		}
		for mk := range mapStrIntf {
			ins := make(map[string]string)
			ins[mk] = extractString(mapStrIntf[mk])
		}
	}
	if ro.Attrs != nil {
		o.Attrs = make(map[string]string, len(ro.Attrs))
		for attr := range ro.Attrs {
			o.Attrs[attr] = extractString(ro.Attrs[attr])
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
	case float64:
		return strconv.FormatFloat(val, 'f', 0, 64)
	}
	return ""
}
