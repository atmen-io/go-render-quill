package quill

import (
	"bytes"
	"strconv"
)

type boldWriter struct{}

type imageWriter struct {
	classes []string
}

func (iw *imageWriter) TagName() string {
	return "" // The body contains the entire element.
}

func (iw *imageWriter) SetClass(class string) {
	for _, c := range iw.classes {
		// Avoiding adding a class twice.
		if c == class {
			return
		}
	}
	iw.classes = append(iw.classes, class)
}

func (iw *imageWriter) GetClasses() []string {
	return iw.classes
}

func (iw *imageWriter) Write(o *Op, buf *bytes.Buffer) {

	o.ClosePrevAttrs(buf)

	buf.WriteString("<img src=")
	buf.WriteString(strconv.Quote(o.Data))
	if o.Attrs["alt"] != "" {
		buf.WriteString(" alt=")
		buf.WriteString(strconv.Quote(o.Attrs["alt"]))
	}
	buf.WriteString(">")

}

type italicWriter struct{}
