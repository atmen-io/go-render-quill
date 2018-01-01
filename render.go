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
	return RenderExtended(ops, nil)
}

// RenderExtended takes the Delta array of insert operations and, optionally, a function that provides a BlockWriter for block
// elements (text, header, blockquote, etc.) to customize how those elements are rendered, and, optionally, a function that
// may define an InlineWriter for certain types of inline attributes. Neither of these two functions must always have to give
// a non-nil value. The provided value will be used (and override the default functionality) only if it is not nil.
// The returned byte slice is the rendered HTML.
func RenderExtended(ops []byte, customFormats func(*Op) Formatter) ([]byte, error) {

	var raw []rawOp
	err := json.Unmarshal(ops, &raw)
	if err != nil {
		return nil, err
	}

	var (
		html    = new(bytes.Buffer) // the final output
		tempBuf = new(bytes.Buffer) // temporary buffer reused for each block element
		o       *Op
		fm      Formatter
	)

	attrs := &AttrState{ // the tags currently open in the order in which they were opened
		temp: tempBuf,
	}

	for i := range raw {

		o, err = raw[i].makeOp()
		if err != nil {
			return nil, err
		}

		fm = o.getFormatter(customFormats)
		if fm == nil {
			return html.Bytes(), fmt.Errorf("no formatter found for op %q", raw[i])
		}

		o.closePrevAttrs(tempBuf, attrs)

		// Open the last block element, write its body and close it to move on only when the "\n" of the
		// last block element is reached.
		if strings.IndexByte(o.Data, '\n') != -1 {

			split := strings.Split(o.Data, "\n")

			for i := range split {

				if i > 0 {
					oi := &Op{
						Data:  split[i],
						Attrs: o.Attrs,
						// Type:
					}
					oi.getFormatter(customFormats)
					oi.write(tempBuf)
				}

				//bw.Open(o, attrs)

				o.write()

				html.Write(tempBuf.Bytes())
				html.WriteString(split[i])

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

func (o *Op) write(buf *bytes.Buffer, fm Formatter) {
	if fm.TagName(o) != "" {
		buf.WriteString("<")
		buf.WriteString(fm.TagName())
		buf.WriteString(">")
	}
	buf.WriteString(o.Data)
	if fm.TagName() != "" {
		buf.WriteString("</")
		buf.WriteString(fm.TagName())
		buf.WriteString(">")
	}
}

// HasAttr says if the Op is not nil and has the attribute set to a non-blank value.
func (o *Op) HasAttr(attr string) bool {
	return o != nil && o.Attrs[attr] != ""
}

// getFormatter returns a formatter based on the keyword (either "text" or "" or an attribute name) and the Op settings.
func (o *Op) getFormatter(keyword string, customFormats func(string, *Op) Formatter) Formatter {

	if customFormats != nil {
		if custom := customFormats(keyword, o); custom != nil {
			return custom
		}
	}

	switch keyword {
	case "text":
		return new(textFormat)
	case "header":
		return &headerFormat{
			h: "h"+o.Attrs["header"],
		}
	case "list":
		var lt string
		if o.Attrs["list"] == "bullet" {
			lt = "ul"
		} else {
			lt = "ol"
		}
		return &listFormat{
			lType: lt,
		}
	case "blockquote":
		return new(blockQuoteFormat)
	}

	return nil

}

// closePrevAttrs checks if the previous Op opened any attribute tags that are not supposed to be set on the current Op and closes
// those tags in the opposite order in which they were opened.
func (o *Op) closePrevAttrs(buf *bytes.Buffer, st *AttrState) {
	for i := len(st.t) - 1; i >= 0; i-- { // Start with the last attribute opened.
		if !o.HasAttr(st.t[i]) {
		}
	}
}

func (o *Op) OpenAttrs(buf *bytes.Buffer) {
}

// An OpHandler takes the previous Op (which is nil if the current Op is the first) and the current Op and writes the
// current Op to buf. Each handler should check the previous Op to see if it has attributes that are not set on the current
// Op and close the appropriate HTML tags before writing the current Op; also the handler should not needlessly open up a
// tag for an attribute if it was already opened for the previous Op. This ensures that the rendered HTML is lean.

// A BlockWriter defines how an insert of block type gets rendered. The opening HTML tag of a block element is written to the
// main buffer only after the "\n" character terminating the block is reached (the Op with the "\n" character holds the information
// about the block element).
//type BlockWriter interface {
//	Open(*Op, *AttrState)
//	Close(*Op, *AttrState)
//	//Write(*Op, io.Writer)
//}

//func blockWriterByType(t string) BlockWriter {
//	switch t {
//	case "text":
//		return new(textWriter)
//	case "blockquote":
//		return new(blockQuoteWriter)
//	case "header":
//		return new(headerWriter)
//	case "list":
//		return new(listWriter)
//	}
//	return nil
//}

//type InlineWriter interface {
//	TagName() string
//	Write(*Op, *bytes.Buffer)
//}

type Formatter interface {
	TagName() string // Optionally wrap the element with the tag (return empty string for no wrap).
	Class() string   // Optionally give a CSS class to set (return empty string for no class).
}

//type InlineWriter interface {
//	//Attr() string // the attribute key that identifies this inline format
//	Open(*Op, *AttrState)
//	Close(*Op, io.Writer)
//}

//func inlineWriterByType(t string) Formatter {
//	switch t {
//	case "bold":
//		return new(boldWriter)
//	case "image":
//		return new(imageWriter)
//	case "italic":
//		return new(italicWriter)
//	}
//	return nil
//}

//func setUpClasses(o *Op, bw BlockWriter, aws func(string) InlineWriter) {
//	var ar InlineWriter
//	for attr := range o.Attrs {
//		if aws != nil {
//			if custom := aws(attr); custom != nil {
//				ar = custom
//			}
//		} else {
//			ar = inlineWriterByType(attr)
//		}
//		if ar == nil {
//			// This attribute type is unknown.
//			//return html.Bytes(), fmt.Errorf("no type handler found for op %q", ro[i])
//			return
//		}
//	}
//}

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
