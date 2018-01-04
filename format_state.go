package quill

import (
	"bytes"
	"sort"
	"strconv"
)

// A formatState holds the current state of open tag, class, or style formats.
type formatState struct {
	open []*Format // the list of currently open attribute tags
}

// hasSet says if the given format is already opened.
func (fs *formatState) hasSet(fm *Format) bool {
	for i := range fs.open {
		if fs.open[i].Place == fm.Place && fs.open[i].Val == fm.Val {
			return true
		}
	}
	return false
}

// closePrevious checks if the previous ops opened any formats that are not set on the current Op and closes those formats
// in the opposite order in which they were opened.
func (fs *formatState) closePrevious(buf *bytes.Buffer, o *Op) {

	closedTemp := &formatState{}

	for i := len(fs.open) - 1; i >= 0; i-- { // Start with the last format opened.

		f := fs.open[i]

		if f.wrap {
			if f.fm.(FormatWrapper).Close(fs.open, o) {
				buf.WriteString(f.wrapPost)
				fs.pop()
			}
			continue
		}

		// If this format is not set on the current Op, close it.
		if !f.fm.HasFormat(o) {

			// If we need to close a tag after which there are tags that should stay open, close the following tags for now.
			if i < len(fs.open)-1 {
				for ij := len(fs.open) - 1; ij > i; ij-- {
					closedTemp.add(fs.open[ij])
					if f.wrap {
						buf.WriteString(f.wrapPost)
					} else if f.Place == Tag {
						closeTag(buf, f.Val)
					} else {
						closeTag(buf, "span")
					}
					fs.pop()
				}
			}

			if f.Place == Tag {
				closeTag(buf, f.Val)
			} else {
				closeTag(buf, "span")
			}
			fs.pop()

		}

	}

	// Re-open the temporarily closed formats.
	closedTemp.writeFormats(buf)
	fs.open = append(fs.open, closedTemp.open...) // Copy after the sorting.

}

// add adds a format that the string that will be written to buf right after this will have.
// Before calling add, check if the Format is already opened up earlier.
// Do not use add to write block-level styles (those are written by o.writeBlock after being merged).
func (fs *formatState) add(f *Format) {
	if f.Place < 3 { // Check if the Place is valid.
		fs.open = append(fs.open, f)
	}
}

// writeFormats sorts the formats in the current formatState and writes them all out to buf. If a format implements
// the FormatWrapper interface, that format's opening wrap is printed.
func (fs *formatState) writeFormats(buf *bytes.Buffer) {

	sort.Sort(fs) // Ensure that the serialization is consistent even if attribute ordering in a map changes.

	for i := range fs.open {

		if fs.open[i].wrap {
			buf.WriteString(fs.open[i].wrapPre) // The complete opening or closing wrap is given.
			continue
		}

		buf.WriteByte('<')

		switch fs.open[i].Place {
		case Tag:
			buf.WriteString(fs.open[i].Val)
		case Class:
			buf.WriteString("span class=")
			buf.WriteString(strconv.Quote(fs.open[i].Val))
		case Style:
			buf.WriteString("span style=")
			buf.WriteString(strconv.Quote(fs.open[i].Val))
		}

		buf.WriteByte('>')

	}

}

// pop removes the last format state from the list of open states.
func (fs *formatState) pop() {
	fs.open = fs.open[:len(fs.open)-1]
}

// Implement the sort.Interface interface.

func (fs *formatState) Len() int {
	return len(fs.open)
}

func (fs *formatState) Less(i, j int) bool {

	// Formats that implement the FormatWrapper interface are written first.
	if _, ok := fs.open[i].fm.(FormatWrapper); ok {
		return true
	} else if _, ok := fs.open[j].fm.(FormatWrapper); ok {
		return false
	}

	// Tags are written first, then classes, and then style attributes.
	if fs.open[i].Place != fs.open[j].Place {
		return fs.open[i].Place < fs.open[j].Place
	}

	// Simply check values.
	return fs.open[i].Val < fs.open[j].Val

}

func (fs *formatState) Swap(i, j int) {
	fs.open[i], fs.open[j] = fs.open[j], fs.open[i]
}
