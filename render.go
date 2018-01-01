// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations
// and renders the defined HTML document.
package quill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func Render(ops []byte) ([]byte, error) {
	return RenderExtended(ops, nil, nil)
}

// RenderExtended takes the Delta array of insert operations and, optionally, a map of custom Op rendering functions that may
// customizing how operations of certain types are rendered. The returned byte slice is the rendered HTML.
func RenderExtended(ops []byte, bws func(string) BlockWriter, aws func(string) InlineWriter) ([]byte, error) {

	var ro []map[string]interface{}
	err := json.Unmarshal(ops, &ro)
	if err != nil {
		return nil, err
	}

	// attrStates lists the tags currently open in the order in which they were opened.
	var (
		attrStates = make([]string, 0, 2)
		html       = new(bytes.Buffer)
		tempBuf    = new(bytes.Buffer)
		o          *Op
		bw         BlockWriter
	)

	for i := range ro {

		o, err = rawOpToOp(ro[i])
		if err != nil {
			return nil, err
		}

		if bws != nil {
			if custom := bws(o.Type); custom != nil {
				bw = custom
			}
		} else {
			bw = blockWriterByType(o.Type)
		}
		if bw == nil {
			return html.Bytes(), fmt.Errorf("no type handler found for op %q", ro[i])
		}

		o.ClosePrevAttrs(tempBuf, attrStates)

		// Close the last block element, open a new one, and write the inner body of the new block element
		// only when the "\n" of the new block element is reached.
		if strings.IndexByte(o.Data, '\n') != -1 {

			split := strings.Split(o.Data, "\n")

			for i := range split {

				setUpClasses(o, bw, aws)

				tn := bw.TagName(o)

				// Some block elements have no closing tag (images, for example) but write everything in the body.
				if tn != "" {
					html.WriteString("<")
					html.WriteString(tn)
					writeClasses(bw.GetClasses(), html)
				}

				html.Write(tempBuf.Bytes())
				html.WriteString(split[i])

				if tn != "" {
					html.WriteString("</")
					html.WriteString(tn)
					html.WriteString(">")
				}

				tempBuf.Reset()

			}

		} else {

			bw.Write(o, tempBuf)

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
func (o *Op) ClosePrevAttrs(buf *bytes.Buffer, st []string) {
	for i := len(st) - 1; i >= 0; i-- { // Start with the last attribute opened.
		if !o.HasAttr(st[i]) {
		}
	}
}

func (o *Op) OpenAttrs(buf *bytes.Buffer) {
}

// rawOpToOp takes a raw Delta op as extracted from the JSON and turns it into an Op to make it usable for rendering.
func rawOpToOp(ro map[string]interface{}) (*Op, error) {
	if _, ok := ro["insert"]; !ok {
		return nil, fmt.Errorf("op %q lacks an insert", ro)
	}
	o := new(Op)
	if str, ok := ro["insert"].(string); ok {
		// This op is a simple string insert.
		o.Type = "text"
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

// A BlockWriter defines how an insert of block type gets rendered. The opening HTML tag of a block element is written to the
// main buffer only after the "\n" character terminating the block is reached (the Op with the "\n" character holds the information
// about the block element).
type BlockWriter interface {
	TagName(o *Op) string
	SetClass(string)
	GetClasses() []string
	Write(*Op, *bytes.Buffer)
}

type BlockData struct {
	TagName string
}

func blockWriterByType(t string) BlockWriter {
	switch t {
	case "text":
		return new(textWriter)
	case "blockquote":
		return new(blockQuoteWriter)
	case "header":
		return new(headerWriter)
	}
	return nil
}

type InlineWriter interface {
	TagName() string
	SetClass(string)
	GetClasses() []string
	Write(*Op, *bytes.Buffer)
}

func inlineWriterByType(t string) InlineWriter {
	switch t {
	case "bold":
		return new(boldWriter)
	case "image":
		return new(imageWriter)
	case "italic":
		return new(italicWriter)
	}
	return nil
}

func setUpClasses(o *Op, bw BlockWriter, aws func(string) InlineWriter) {
	var ar InlineWriter
	for attr := range o.Attrs {
		if aws != nil {
			if custom := aws(attr); custom != nil {
				ar = custom
			}
		} else {
			ar = inlineWriterByType(attr)
		}
		if ar == nil {
			// This attribute type is unknown.
			//return html.Bytes(), fmt.Errorf("no type handler found for op %q", ro[i])
			return
		}
	}
}

func writeClasses(cl []string, buf *bytes.Buffer) {
	if len(cl) > 0 {
		buf.WriteString(" class=")
		buf.WriteString(strconv.Quote(strings.Join(cl, " ")))
	}
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
