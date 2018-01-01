package quill

import (
	"bytes"
	"strconv"
)

type boldWriter struct{}

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

	buf.WriteString("<img src=")
	buf.WriteString(strconv.Quote(o.Data))
	if o.Attrs["alt"] != "" {
		buf.WriteString(" alt=")
		buf.WriteString(strconv.Quote(o.Attrs["alt"]))
	}
	buf.WriteString(">")

}

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

	buf.WriteString("<img src=")
	buf.WriteString(strconv.Quote(o.Data))
	if o.Attrs["alt"] != "" {
		buf.WriteString(" alt=")
		buf.WriteString(strconv.Quote(o.Attrs["alt"]))
	}
	buf.WriteString(">")

}

type italicWriter struct{}

func (iw *italicWriter) TagName() string {
	return "" // The body contains the entire element.
}

func (iw *italicWriter) SetClass(class string) {
	for _, c := range iw.classes {
		// Avoiding adding a class twice.
		if c == class {
			return
		}
	}
	iw.classes = append(iw.classes, class)
}

func (iw *italicWriter) GetClasses() []string {
	return iw.classes
}

func (iw *italicWriter) Write(o *Op, buf *bytes.Buffer) {

	buf.WriteString("<img src=")
	buf.WriteString(strconv.Quote(o.Data))
	if o.Attrs["alt"] != "" {
		buf.WriteString(" alt=")
		buf.WriteString(strconv.Quote(o.Attrs["alt"]))
	}
	buf.WriteString(">")

}
