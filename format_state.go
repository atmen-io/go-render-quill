package quill

import (
	"bytes"
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

// addFormat adds a format that the string that will be written to buf right after this will have.
// Before calling addFormat, check if the Format is already opened up earlier.
// Do not use addFormat to write block-level styles (those are written by o.writeBlock after being merged).
func (fs *formatState) addFormat(fm *Format, buf *bytes.Buffer) {

	// Check that the place where the format is supposed to be is valid.
	if fm.Place < 0 || fm.Place > 2 {
		return
	}

	fs.doFormatWrapper("open", fm.fm, nil, buf)

	fs.open = append(fs.open, fm)

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

// Pop removes the last state from the list of open states.
func (fs *formatState) pop() {
	fs.open = fs.open[:len(fs.open)-1]
}

func (fs *formatState) doFormatWrapper(openClose string, fmTer Formatter, o *Op, buf *bytes.Buffer) (wrote bool) {
	if openClose == "open" {
		if fw, ok := fmTer.(FormatWrapper); ok {
			if wrapOpen := fw.PreWrap(fs.open); wrapOpen != "" {
				fs.open = append(fs.open, &Format{
					Val:   wrapOpen,
					Place: Tag,
					fm:    fmTer,
				})
				buf.WriteString(wrapOpen)
				wrote = true
			}
		}
	} else if openClose == "close" {
		if fw, ok := fmTer.(FormatWrapper); ok {
			if wrapClose := fw.PostWrap(fs.open, o); wrapClose != "" {
				fs.pop()                   // TODO ???
				buf.WriteString(wrapClose) // The complete closing wrap is given in wrapClose.
				wrote = true
			}
		}
	}
	return
}
