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
	)
	o.Attrs = make(map[string]string, 3) // initialize once here only

	for i := range raw {

		if err = raw[i].makeOp(o); err != nil {
			return nil, err
		}

		// Get Formatter to check if the Op will just be writing a body.
		typeFm := o.getFormatter(o.Type, customFormats)
		if typeFm == nil {
			// Each Op should have a Formatter given by its type.
			return nil, fmt.Errorf("no formatter found for type %q of op %q", o.Type, raw[i])
		}

		// If the op has a Write method defined (based on its type of attributes), we just write the body.
		if bw, ok := typeFm.(FormatWriter); ok {
			bw.Write(tempBuf)
			continue
		}

		// Open the a block element, write its body, and close it to move on only when the ending "\n" of the block is reached.
		if strings.IndexByte(o.Data, '\n') != -1 {

			if o.Data == "\n" { // Write a block element and flush the temporary buffer.

				// Copy the temporary block buffer to the final output and reset the temporary buffer.
				o.Data = ""
				o.writeBlock(fs, tempBuf, html, typeFm, customFormats)

			} else { // Extract the block-terminating line feeds and write each part as its own Op.

				split := strings.Split(o.Data, "\n")

				for i := range split {

					if split[i] == "" { // We're dealing with a blank line (split at \n\n) or a \n is at the very beginning or end.

						if i > 0 && i != len(split)-1 { // If the empty string represents an empty paragraph.
							o.Data = "<br>"
						}

						o.writeBlock(fs, tempBuf, html, typeFm, customFormats)

					} else {

						o.Data = split[i]
						o.writeInline(fs, tempBuf, typeFm, customFormats)

					}

				}

			}

		} else { // We are just adding stuff inline.

			o.writeInline(fs, tempBuf, typeFm, customFormats)

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
func (o *Op) writeBlock(fs *formatState, tempBuf *bytes.Buffer, finalBuf *bytes.Buffer, typeFm Formatter, customFormats func(string, *Op) Formatter) {

	// Close the inline formats opened within the block.
	o.closePrevFormats(tempBuf, fs, customFormats)

	var blockWrap struct {
		tagName string
		classes []string
		style   string
		fs      formatState // the formats for the block element itself
	}

	// Open the tag for the Op if the Op Type calls for a tag.
	tVal, tPlace := typeFm.Format()
	if tPlace == Tag && tVal != "" {
		blockWrap.tagName = tVal
	}

	// If an opening tag has not been written, it may be specified in an attribute.
	for attr := range o.Attrs {
		attrFm := o.getFormatter(attr, customFormats)
		if attrFm == nil {
			continue // not returning an error
		}
		if fw, ok := attrFm.(FormatWriter); ok {
			// If an attribute format wants to write the entire body, let it write the body.
			fw.Write(tempBuf)
		}
		o.addAttr(&blockWrap.fs, attrFm, tempBuf)
	}

	// Merge all formats into a single tag.
	for i := range blockWrap.fs.open {
		val := blockWrap.fs.open[i].val
		switch blockWrap.fs.open[i].place {
		case Tag:
			if blockWrap.tagName == "" {
				blockWrap.tagName = val
			}
		case Class:
			blockWrap.classes = append(blockWrap.classes, val)
		case Style:
			blockWrap.style += val
		}

	}

	finalBuf.WriteByte('<')
	finalBuf.WriteString(blockWrap.tagName)
	finalBuf.WriteString(ClassesList(blockWrap.classes))
	if blockWrap.style != "" {
		finalBuf.WriteString(" style=")
		finalBuf.WriteString(strconv.Quote(blockWrap.style))
	}
	finalBuf.WriteByte('>')

	finalBuf.WriteString(o.Data) // Copy the data of the current Op (usually just "<br>" or nothing).

	finalBuf.Write(tempBuf.Bytes())

	o.closePrevFormats(finalBuf, &blockWrap.fs, customFormats)

	closeTag(finalBuf, blockWrap.tagName)

	tempBuf.Reset()

}

func (o *Op) writeInline(fs *formatState, buf *bytes.Buffer, fm Formatter, customFormats func(string, *Op) Formatter) {

	o.closePrevFormats(buf, fs, customFormats)

	if fm != nil {
		o.addAttr(fs, fm, buf)
	}

	for attr := range o.Attrs {
		fm = o.getFormatter(attr, customFormats)
		if fm != nil {
			if bw, ok := fm.(FormatWriter); ok {
				bw.Write(buf)
				continue
			}
			o.addAttr(fs, fm, buf)
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
		lf := &listFormat{
			indent: indentDepths[o.Attrs["indent"]],
		}
		if o.Attrs["list"] == "bullet" {
			lf.lType = "ul"
		} else {
			lf.lType = "ol"
		}
		return lf
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
func (o *Op) closePrevFormats(buf *bytes.Buffer, fs *formatState, customFormats func(string, *Op) Formatter) {
	var f format                             // reused in the loop for convenience
	var tagsList []string                    // the currently open tags (used by FormatWrapper)
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

		// If this format is not set on the current Op, close it.
		if !o.HasAttr(f.keyword) || (fVal != f.val) {

			fs.pop()

			if fPlace == Tag {
				closeTag(buf, fVal)
			} else {
				closeTag(buf, "span")
			}

		}

		if fw, ok := fmter.(FormatWrapper); ok {
			if wrapClose := fw.PostWrap(tagsList, o); wrapClose != "" {
				//fs.add(format{
				//	val:   outer,
				//	place: Tag,
				//})
				fs.pop() // TODO ???
				closeTag(buf, wrapClose)
			}
		}

	}
}

// addAttr adds an format that the string that will be written to buf right after this will have.
// The format is written only if it is not already opened up earlier.
func (o *Op) addAttr(fs *formatState, fm Formatter, buf *bytes.Buffer) {

	var tagsList []string // the currently open tags (used by FormatWrapper)
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
		if wrapOpen := fw.PreWrap(tagsList); wrapOpen != "" {
			fs.add(format{
				val:   wrapOpen,
				place: Tag,
			})
			buf.WriteString(wrapOpen)
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
	PreWrap([]string) string       // Give Wrap a list of opened tag names and it'll say what complete tag, if anything, to write.
	PostWrap([]string, *Op) string // Give it the current Op so it says if it's now time to close the wrap.
}

type format struct {
	val, keyword string
	place        StyleFormat
}

// A formatState holds the current state of open tag, class, or style formats.
type formatState struct {
	open []format // the list of currently open attribute tags
}

// Add adds an inline attribute state to the end of the list of open states.
func (fs *formatState) add(f format) {
	fs.open = append(fs.open, f)
}

// Pop removes the last state from the list of open states.
func (fs *formatState) pop() {
	fs.open = fs.open[:len(fs.open)-1]
}

// If cl has something, then ClassesList returns the class attribute to add to an HTML element with a space before the
// "class" attribute and spaces between each class name.
func ClassesList(cl []string) (classAttr string) {
	if len(cl) > 0 {
		classAttr = " class=" + strconv.Quote(strings.Join(cl, " "))
	}
	return
}

// openTagOrNot writes an "<" and string s to the buffer if s is not blank.
func openTagOrNot(buf *bytes.Buffer, s string) {
	if s != "" {
		buf.WriteByte('<')
		buf.WriteString(s)
	}
}

// closeTag writes a complete closing tag.
func closeTag(buf *bytes.Buffer, s string) {
	buf.WriteString("</")
	buf.WriteString(s)
	buf.WriteByte('>')
}
