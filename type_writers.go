package quill

import (
	"bytes"
	"strconv"
)

type textWriter struct {
	classes []string
}

func (tw *textWriter) Write(o *Op, buf *bytes.Buffer) {



	buf.WriteString("<p>")
	buf.WriteString(o.Data)

}

func (tw *textWriter) TagName() string {
	return "p"
}

func (tw *textWriter) SetClass(class string) {
	for _, c := range tw.classes {
		// Avoiding adding a class twice.
		if c == class {
			return
		}
	}
	tw.classes = append(tw.classes, class)
}

func (tw *textWriter) GetClasses() []string {
	return tw.classes
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

	o.ClosePrevAttrs(buf)

	buf.WriteString("<img src=")
	buf.WriteString(strconv.Quote(o.Data))
	buf.WriteString(">")

}
