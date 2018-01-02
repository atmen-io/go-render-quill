// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations
// and renders the defined HTML document.
package quill

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
)

// Render takes the Delta array of insert operations and returns the rendered HTML using the default settings of this package.
func Render(ops []byte) ([]byte, error) {
	return RenderExtended(ops, nil)
}

// RenderExtended takes the Delta array of insert operations and, optionally, a function that may provide a Formatter to
// customize the way certain kinds of inserts are rendered. If the given Formatter is nil, then the default one that is built
// in is used. The returned byte slice is the rendered HTML.
func RenderExtended(ops []byte, customFormats func(string, *Op) Formatter) ([]byte, error) {

	var raw []rawOp
	err := json.Unmarshal(ops, &raw)
	if err != nil {
		return nil, err
	}

	var (
		html    = new(bytes.Buffer) // the final output
		tempBuf = new(bytes.Buffer) // temporary buffer reused for each block element
		fs      = new(formatState)  // the tags currently open in the order in which they were opened
		o       = new(Op)           // allocate memory for an Op to reuse for all iterations
		fm      Formatter
	)

	o.Attrs = make(map[string]string, 3) // initialize once here only

	for i := range raw {

		if err = raw[i].makeOp(o); err != nil {
			return nil, err
		}

		fm = o.getFormatter(o.Type, customFormats)
		if fm == nil {
			continue // not returning an error
		}

		// If the op has a Write method defined (based on its type of attributes), we just write the body.
		if bw, ok := fm.(FormatWriter); ok {
			bw.Write(tempBuf)
			continue
		}

		// Open the last block element, write its body and close it to move on only when the "\n" of the
		// last block element is reached.
		if strings.IndexByte(o.Data, '\n') != -1 {

			if o.Data == "\n" { // Write a block element and flush the temporary buffer.

				html.Write(o.writeBlock(fs, tempBuf, customFormats)) // Copy the temporary buffer into the final output.

				tempBuf.Reset()

			} else {

				split := strings.Split(o.Data, "\n")

				for i := range split {

					if split[i] == "" { // If we're dealing with a blank line split at \n\n or a \n is at the end.
						o.Data = "\n"
						o.writeBlock(fs, tempBuf, customFormats)
						tempBuf.Reset()
					} else {
						o.Data = split[i]
						o.writeInline(fs, tempBuf, customFormats)
					}

				}

			}

		} else { // We are just adding stuff inline.

			o.writeInline(fs, tempBuf, customFormats)

		}

	}

	return html.Bytes(), nil

}

type Op struct {
	Data  string            // the text to insert or the value of the embed object (http://quilljs.com/docs/delta/#embeds)
	Type  string            // the type of the op (typically "text", but any other type can be registered)
	Attrs map[string]string // key is attribute name; value is either value string or "y" (meaning true) or "n" (meaning false)
}

// writeBlock writes a block element (which may be nested inside another block element if it is a FormatWrapper).
// The opening HTML tag of a block element is written to the main buffer only after the "\n" character terminating the
// block is reached (the Op with the "\n" character holds the information about the block element).
func (o *Op) writeBlock(fs *formatState, buf *bytes.Buffer, customFormats func(string, *Op) Formatter) []byte {

	o.closePrevAttrs(buf, fs, customFormats)

	// Open the tag for the Op if the Op has a format that uses tags.
	openTagOrNot(buf, fm.TagName())

	for attr := range o.Attrs {
		attrFm := o.getFormatter(attr, customFormats)
		if attrFm == nil {
			continue // not returning an error
		}
		if bw, ok := attrFm.(FormatWriter); ok {
			bw.Write(buf)
		}
		o.addAttr(fs, attrFm, buf)
	}

	if fm.TagName() != "" {
		buf.WriteByte('>')
	}

	if o.Data == "\n" {
		o.Data = "<br>" // Avoid having empty <p></p>.
	}

	buf.WriteString(o.Data) // Copy the data of the current Op.

	closeTagOrNot(buf, fm.TagName())

	return buf.Bytes()

}

func (o *Op) writeInline(fs *formatState, buf *bytes.Buffer, customFormats func(string, *Op) Formatter) {

	o.closePrevAttrs(buf, fs, customFormats)

	var fm Formatter // reuse in loop

	for attr := range o.Attrs {
		fm = o.getFormatter(attr, customFormats)
		if fm != nil {
			if bw, ok := fm.(FormatWriter); ok {
				bw.Write(buf)
				break
			} else {
				buf.WriteString(o.Data)
			}
		}
	}

	buf.WriteString(o.Data)

}

// HasAttr says if the Op is not nil and either has the attribute set to a non-blank value.
func (o *Op) HasAttr(attr string) bool {
	return o != nil && o.Attrs[attr] != ""
}

// getFormatter returns a formatter based on the keyword (either "text" or "" or an attribute name) and the Op settings.
// For every Op, first its Type is passed through here as the keyword, and then its attributes.
func (o *Op) getFormatter(keyword string, customFormats func(string, *Op) Formatter) Formatter {

	if customFormats != nil {
		if custom := customFormats(keyword, o); custom != nil {
			return custom
		}
	}

	switch keyword { // This is the list of currently recognized "keywords".
	case "text":
		return new(textFormat)
	case "header":
		return &headerFormat{
			h: "h" + o.Attrs["header"],
		}
	case "list":
		var lt string
		if o.Attrs["list"] == "bullet" {
			lt = "ul"
		} else {
			lt = "ol"
		}
		return &listFormat{
			lType:  lt,
			indent: indentDepths[o.Attrs["indent"]],
		}
	case "blockquote":
		return new(blockQuoteFormat)
	case "image":
		return new(imageFormat)
	case "link":
		return new(linkFormat)
	case "bold":
		return new(boldFormat)
	case "italic":
		return new(italicFormat)
	case "color":
		return new(colorFormat)
	}

	return nil

}

// closePrevAttrs checks if the previous Op opened any attribute tags that are not supposed to be set on the current Op and closes
// those tags in the opposite order in which they were opened.
func (o *Op) closePrevAttrs(buf *bytes.Buffer, fs *formatState, customFormats func(string, *Op) Formatter) {
	var f format                             // reused in the loop for convenience
	var tagsList []string                    // the currently open tags (used by WrappedFormatter)
	for i := len(fs.open) - 1; i >= 0; i-- { // Start with the last attribute opened.

		f = fs.open[i]

		if f.place == Tag {
			tagsList = append(tagsList, f.val)
		}

		fmter := o.getFormatter(f.keyword, customFormats)
		if fmter == nil {
			continue // Really this should never be the case.
		}

		fVal, fPlace := fmter.Format()

		if f.place == fPlace && f.val == fVal {
		}

		if fw, ok := fmter.(FormatWrapper); ok {
			if outer := fw.PostWrap(tagsList, o); outer != "" {
				fs.add(format{
					val:   outer,
					place: Tag,
				})
				buf.WriteByte('<')
				buf.WriteString(outer)
				buf.WriteByte('>')
			}
		}

	}
}

// addAttr adds an format that the string that will be written to buf right after this will have.
// The format is written only if it is not already opened up earlier.
func (o *Op) addAttr(fs *formatState, fm Formatter, buf *bytes.Buffer) {

	var tagsList []string // the currently open tags (used by WrappedFormatter)
	fVal, fPlace := fm.Format()

	// Check that the place where the format is supposed to be is valid.
	if fPlace < 0 || fPlace > 2 {
		return
	}

	// Check if this format is already opened.
	for i := range fs.open {
		if fs.open[i].place == fPlace && fs.open[i].val == fVal {
			return
		}
		if fs.open[i].place == Tag {
			tagsList = append(tagsList, fs.open[i].val)
		}
	}

	if fw, ok := fm.(FormatWrapper); ok {
		if outer := fw.PreWrap(tagsList); outer != "" {
			fs.add(format{
				val:   outer,
				place: Tag,
			})
			buf.WriteByte('<')
			buf.WriteString(outer)
			buf.WriteByte('>')
		}
	}

	fs.add(format{
		val:   fVal,
		place: fPlace,
	})

	buf.WriteByte('<')

	switch fPlace {
	case Tag:
		buf.WriteString(fVal)
	case Class:
		buf.WriteString("<span class=")
		buf.WriteString(strconv.Quote(fVal))
	case Style:
		buf.WriteString("<span style=")
		buf.WriteString(strconv.Quote(fVal))
	}

	buf.WriteByte('>')

}

// Each handler should check the previous Op to see if it has attributes that are not set on the current Op and close the
// appropriate HTML tags before writing the current Op; also the handler should not needlessly open up a  tag for an
// attribute if it was already opened for the previous Op. This ensures that the rendered HTML is lean.

// A StyleFormat is either an HTML tag name, a CSS class, or a style attribute value.
type StyleFormat uint8

const (
	Tag StyleFormat = iota
	Class
	Style
)

type Formatter interface {
	Format() (string, StyleFormat) // Format gives the string to write and where to place it.
	//HasFormat(*Op, []string) bool // Given the current Op and a list of currently open tags, say if the Op needs the format set.
}

// A Formatter may also be a FormatWriter if it wishes to write the body of the Op in some custom way (useful for embeds).
type FormatWriter interface {
	Formatter
	Write(io.Writer) // Write the entire body of the element.
}

// A FormatWrapper wraps text in additional HTML tags (such as "ul" for lists).
type FormatWrapper interface {
	Formatter
	PreWrap([]string) string       // Give Wrap a list of opened tag names and it'll say what tag name, if anything, should be written.
	PostWrap([]string, *Op) string // Give it the current Op so it says if it's now time to close the wrap.
}

type format struct {
	val, keyword string
	place        StyleFormat
}

type formatState struct {
	open []format // the list of currently open attribute tags
	//temp io.Writer // the temporary buffer (for the block element)
}

// Add adds an inline attribute state to the end of the list of open states.
func (fs *formatState) add(f format) {
	fs.open = append(fs.open, f)
}

// Pop removes the last attribute state from the list of states if the last is s.
func (fs *formatState) close(f format) {
	if fs.open[len(fs.open)-1] == f {
		fs.open = fs.open[:len(fs.open)-1]
	}
}

// If cl has something, then ClassesList returns the class attribute to add to an HTML element with a space before the
// "class" attribute and spaces between each class name.
func ClassesList(cl []string) (classAttr string) {
	if len(cl) > 0 {
		classAttr = " class=" + strconv.Quote(strings.Join(cl, " "))
	}
	return
}

func openTagOrNot(buf *bytes.Buffer, s string) {
	if s != "" {
		buf.WriteByte('<')
		buf.WriteString(s)
	}
}

func closeTagOrNot(buf *bytes.Buffer, s string) {
	if s != "" {
		buf.WriteString("</")
		buf.WriteString(s)
		buf.WriteByte('>')
	}
}
