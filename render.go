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

// Render takes a Delta array of insert operations and returns the rendered HTML using the built-in settings.
func Render(ops []byte) ([]byte, error) {
	return RenderExtended(ops, nil)
}

// RenderExtended takes a Delta array of insert operations and, optionally, a function that may provide a Formatter to
// customize the way certain kinds of inserts are rendered. If the given Formatter is nil, then the default one that is
// built in is used. The returned byte slice is the rendered HTML.
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

				// Avoid empty paragraphs and "\n" in the output.
				if tempBuf.Len() == 0 {
					o.Data = "<br>"
				} else {
					o.Data = ""
				}

				o.writeBlock(fs, tempBuf, html, typeFm, customFormats)

			} else { // Extract the block-terminating line feeds and write each part as its own Op.

				split := strings.Split(o.Data, "\n")

				for i := range split {

					o.Data = split[i]

					// If the current o.Data still has an "\n" following (its not the last in split), then it ends a block.
					// If the last element in split is just "" then the last character in the rawOp was a "\n".
					if i < len(split)-1 {

						// Avoid having empty paragraphs.
						if tempBuf.Len() == 0 && o.Data == "" {
							o.Data = "<br>"
						}

						o.writeBlock(fs, tempBuf, html, typeFm, customFormats)

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

	//fmt.Printf("WRITING BLOCK: %q \n", *o)

	// Close the inline formats opened within the block.
	o.closePrevFormats(tempBuf, fs, customFormats)

	var blockWrap struct {
		tagName string
		classes []string
		style   string
		fs      formatState // the formats for the block element itself
	}

	// Check if the Op Type calls for a tag on the block.
	fm := typeFm.Fmt()
	if fm.Block && fm.Place == Tag && fm.Val != "" {
		blockWrap.tagName = fm.Val // Default block tag to format specified by the Type.
	}

	// If an opening tag has not been written, it may be specified by an attribute.
	for attr := range o.Attrs {
		attrFm := o.getFormatter(attr, customFormats)
		if attrFm == nil {
			continue // not returning an error
		}
		if fw, ok := attrFm.(FormatWriter); ok {
			// If an attribute format wants to write the entire body, let it write the body.
			fw.Write(tempBuf)
		}
		// Save the desired attributes without writing them anywhere.
		blockWrap.fs.addFormat(attr, attrFm, &bytes.Buffer{})
	}

	// Merge all formats into a single tag.
	for i := range blockWrap.fs.open {
		val := blockWrap.fs.open[i].Val
		switch blockWrap.fs.open[i].Place {
		case Tag:
			if fm.Block && val != "" {
				blockWrap.tagName = val // Override value set by Type.
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

	//if o.Data == "" {
	//	finalBuf.WriteString("<br>") // Avoid having empty "<p></p>" where a line break is desired.
	//}

	finalBuf.WriteString(o.Data) // Copy the data of the current Op (usually just "<br>" or nothing).

	finalBuf.Write(tempBuf.Bytes())

	closeTag(finalBuf, blockWrap.tagName)

	tempBuf.Reset()

}

func (o *Op) writeInline(fs *formatState, buf *bytes.Buffer, fmTer Formatter, customFormats func(string, *Op) Formatter) {

	//fmt.Printf("WRITING INLINE: %q \n", *o)

	o.closePrevFormats(buf, fs, customFormats)

	// The first fmTer (passed in as a parameter) is the fmTer given by the Type of the Op insert.
	if fmTer != nil {
		fs.addFormat(o.Type, fmTer, buf)
	}

	for attr := range o.Attrs {
		fmTer = o.getFormatter(attr, customFormats)
		if fmTer != nil {
			if bw, ok := fmTer.(FormatWriter); ok {
				bw.Write(buf)
				continue
			}
			fs.addFormat(attr, fmTer, buf)
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
	case "align":
		return &alignFormat{
			align: "align-" + o.Attrs["align"],
		}
	case "image":
		return new(imageFormat)
	case "link":
		return new(linkFormat)
	case "bold":
		return new(boldFormat)
	case "italic":
		return new(italicFormat)
	case "color":
		return &colorFormat{
			c: "color:" + o.Attrs["color"] + ";",
		}
	}

	return nil

}

// closePrevAttrs checks if the previous Op opened any attribute tags that are not supposed to be set on the current Op and closes
// those tags in the opposite order in which they were opened.
func (o *Op) closePrevFormats(buf *bytes.Buffer, fs *formatState, customFormats func(string, *Op) Formatter) {

	var f *Format // reused in the loop for convenience
	var tempClosed []*Format

	for i := len(fs.open) - 1; i >= 0; i-- { // Start with the last format opened.

		f = fs.open[i]

		fmTer := o.getFormatter(f.keyword, customFormats)
		if fmTer == nil {
			continue // Really this should never be the case.
		}

		fm := fmTer.Fmt()

		// If this format is not set on the current Op, close it.
		if fm.Val != f.Val && !fm.Block {

			// If we need to close a tag after which there are tags that should stay open, close the following tags for now.
			if i < len(fs.open)-1 {
				for ij := len(fs.open) - 1; ij > i; ij-- {
					tempClosed = append(tempClosed, fs.open[ij])
					fs.open[ij].close(buf)
					fs.pop()
					i--
				}
			}

			fm.close(buf)
			fs.pop()

		}

		// If a wrapping open tag was written, decrement i to reflect the shortened format state list.
		if fs.doFormatWrapper("close", f.keyword, fmTer, o, buf) {
			i--
		}

	}

	// Open back up the closed tags.
	for i := range tempClosed {
		fs.addFormat(tempClosed[i].keyword, o.getFormatter(tempClosed[i].keyword, customFormats), buf)
	}

}

// Each handler should check the previous Op to see if it has attributes that are not set on the current Op and close the
// appropriate HTML tags before writing the current Op; also the handler should not needlessly open up a  tag for an
// attribute if it was already opened for the previous Op. This ensures that the rendered HTML is lean.

// A FormatPlace is either an HTML tag name, a CSS class, or a style attribute value.
type FormatPlace uint8

const (
	Tag FormatPlace = iota
	Class
	Style
)

type Formatter interface {
	Fmt() *Format // Format gives the string to write and where to place it.
}

// A Formatter may also be a FormatWriter if it wishes to write the body of the Op in some custom way (useful for embeds).
type FormatWriter interface {
	Formatter
	Write(io.Writer) // Write the entire body of the element.
}

// A FormatWrapper wraps text in additional HTML tags (such as "ul" for lists).
type FormatWrapper interface {
	Formatter
	PreWrap([]*Format) string       // Given the currently open formats, say what to write to open the wrap (complete HTML tag).
	PostWrap([]*Format, *Op) string // Given the currently open formats and the current Op, say what to write to close the wrap.
}

type Format struct {
	Val     string      // the value to print
	Place   FormatPlace // where this format is placed in the text
	Block   bool        // indicate whether this is a block-level format (not printed until a "\n" is reached)
	//keyword string      // the format identifier (either an insert type or attribute name)
	fm Formatter
}

func (f *Format) close(buf *bytes.Buffer) {
	if f.Place == Tag {
		closeTag(buf, f.Val)
	} else {
		closeTag(buf, "span")
	}
}

// A formatState holds the current state of open tag, class, or style formats.
type formatState struct {
	open []*Format // the list of currently open attribute tags
}

// addFormat adds a format that the string that will be written to buf right after this will have.
// The format is written only if it is not already opened up earlier.
func (fs *formatState) addFormat(keyword string, fmTer Formatter, buf *bytes.Buffer) {

	fm := fmTer.Fmt()

	// Check that the place where the format is supposed to be is valid.
	if fm.Place < 0 || fm.Place > 2 {
		return
	}

	// Check if this format is already opened.
	for i := range fs.open {
		if fs.open[i].Place == fm.Place && fs.open[i].Val == fm.Val {
			return
		}
	}

	fs.doFormatWrapper("open", keyword, fmTer, nil, buf)

	fm.keyword = keyword

	fs.open = append(fs.open, fm)

	// Do not write block-level styles (those are written by o.writeBlock after being merged).
	if !fm.Block {

		buf.WriteByte('<')

		switch fm.Place {
		case Tag:
			buf.WriteString(fm.Val)
		case Class:
			buf.WriteString("span class=")
			buf.WriteString(strconv.Quote(fm.Val))
		case Style:
			buf.WriteString("span style=")
			buf.WriteString(strconv.Quote(fm.Val))
		}

		buf.WriteByte('>')

	}

}

// Pop removes the last state from the list of open states.
func (fs *formatState) pop() {
	fs.open = fs.open[:len(fs.open)-1]
}

func (fs *formatState) doFormatWrapper(openClose string, keyword string, fmTer Formatter, o *Op, buf *bytes.Buffer) bool {
	if openClose == "open" {
		if fw, ok := fmTer.(FormatWrapper); ok {
			if wrapOpen := fw.PreWrap(fs.open); wrapOpen != "" {
				fs.open = append(fs.open, &Format{
					Val:     wrapOpen,
					Place:   Tag,
					keyword: keyword,
				})
				buf.WriteString(wrapOpen)
				return true
			}
		}
	} else if openClose == "close" {
		if fw, ok := fmTer.(FormatWrapper); ok {
			if wrapClose := fw.PostWrap(fs.open, o); wrapClose != "" {
				fs.pop()                   // TODO ???
				buf.WriteString(wrapClose) // The complete closing wrap is given in wrapClose.
				return true
			}
		}
	}
	return false
}

// If cl has something, then ClassesList returns the class attribute to add to an HTML element with a space before the
// "class" attribute and spaces between each class name.
func ClassesList(cl []string) (classAttr string) {
	if len(cl) > 0 {
		classAttr = " class=" + strconv.Quote(strings.Join(cl, " "))
	}
	return
}

// closeTag writes a complete closing tag.
func closeTag(buf *bytes.Buffer, s string) {
	buf.WriteString("</")
	buf.WriteString(s)
	buf.WriteByte('>')
}
