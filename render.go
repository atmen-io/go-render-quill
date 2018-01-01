// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations
// and renders the defined HTML document.
package quill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Render takes the Delta array of insert operations and returns the rendered HTML using the default settings of this package.
func Render(ops []byte) ([]byte, error) {
	return RenderExtended(ops, nil, nil)
}

// RenderExtended takes the Delta array of insert operations and, optionally, a function that provides a BlockWriter for block
// elements (text, header, blockquote, etc.) to customize how those elements are rendered, and, optionally, a function that
// may define an InlineWriter for certain types of inline attributes. Neither of these two functions must always have to give
// a non-nil value. The provided value will be used (and override the default functionality) only if it is not nil.
// The returned byte slice is the rendered HTML.
func RenderExtended(ops []byte, bws func(blockType string) BlockWriter, aws func(attrType string) InlineWriter) ([]byte, error) {

	var ro []rawOp
	err := json.Unmarshal(ops, &ro)
	if err != nil {
		return nil, err
	}

	var (
		//attrStates = make([]string, 0, 2) // the tags currently open in the order in which they were opened
		attrs   = new(AttrState)
		html    = new(bytes.Buffer) // the final output
		tempBuf = new(bytes.Buffer) // temporary buffer reused for each block element
		o       *Op
		bw      BlockWriter
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

		o.closePrevAttrs(tempBuf, attrs)

		// Open the last block element, write its body and close it to move on only when the "\n" of the
		// last block element is reached.
		if strings.IndexByte(o.Data, '\n') != -1 {

			split := strings.Split(o.Data, "\n")

			for i := range split {

				bw.Open(o, attrs)

				html.Write(tempBuf.Bytes())
				html.WriteString(split[i])

				bw.Close(o, attrs)

				tempBuf.Reset()

			}

		} else {

			tempBuf.WriteString(o.Data)

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
func (o *Op) closePrevAttrs(buf *bytes.Buffer, st *AttrState) {
	for i := len(st.t) - 1; i >= 0; i-- { // Start with the last attribute opened.
		if !o.HasAttr(st.t[i]) {
		}
	}
}

func (o *Op) OpenAttrs(buf *bytes.Buffer) {
}

type rawOp struct {
	Insert interface{}            `json:"insert"`
	Attrs  map[string]interface{} `json:"attributes"`
}

// rawOpToOp takes a raw Delta op as extracted from the JSON and turns it into an Op to make it usable for rendering.
func rawOpToOp(ro rawOp) (*Op, error) {
	if ro.Insert == nil {
		return nil, fmt.Errorf("op %q lacks an insert", ro)
	}
	o := new(Op)
	if str, ok := ro.Insert.(string); ok {
		o.Data = str
		if str == "\n" { // This op is a block element.
			// The first element in the attributes map should be the only element (giving the block type).
			for k := range ro.Attrs {
				o.Type = k
				break
			}
		} else { // This op is a simple string.
			o.Type = "text"
		}
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

// An OpHandler takes the previous Op (which is nil if the current Op is the first) and the current Op and writes the
// current Op to buf. Each handler should check the previous Op to see if it has attributes that are not set on the current
// Op and close the appropriate HTML tags before writing the current Op; also the handler should not needlessly open up a
// tag for an attribute if it was already opened for the previous Op. This ensures that the rendered HTML is lean.

// A BlockWriter defines how an insert of block type gets rendered. The opening HTML tag of a block element is written to the
// main buffer only after the "\n" character terminating the block is reached (the Op with the "\n" character holds the information
// about the block element).
type BlockWriter interface {
	Open(*Op, *AttrState)
	Close(*Op, *AttrState)
	//Write(*Op, io.Writer)
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

//type InlineWriter interface {
//	TagName() string
//	Write(*Op, *bytes.Buffer)
//}

type InlineWriter interface {
	//Attr() string // the attribute key that identifies this inline format
	Open(*Op, *AttrState)
	Close(*Op, io.Writer)
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

type AttrState struct {
	t    []string  // the list of currently open attribute tags
	temp io.Writer // the temporary buffer (for the block element)
}

// Add adds an inline attribute state to the end of the list of open states.
func (as *AttrState) Add(s string) {
	as.t = append(as.t, s)
	as.temp.Write([]byte(s))
}

// Pop removes the last attribute state from the list of states if the last is s.
func (as *AttrState) Pop(s string) {
	if as.t[len(as.t)-1] == s {
		as.t = as.t[:len(as.t)-1]
	}
}

func AttrsToClasses(attrs map[string]string) (classes []string) {
	for k, v := range attrs {
		switch k {
		case "align":
			classes = append(classes, "text-align-"+v)
		}
	}
	return
}

//var attrClasses = [...]string{"align"}

func ClassesList(cl []string) (classAttr string) {
	if len(cl) > 0 {
		classAttr = " class=" + strconv.Quote(strings.Join(cl, " "))
	}
	return
}

//func writeClasses(cl []string, buf *bytes.Buffer) {
//	if len(cl) > 0 {
//		buf.WriteString(" class=")
//		buf.WriteString(strconv.Quote(strings.Join(cl, " ")))
//	}
//}

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
