package quill

import (
	"bytes"
	"strconv"
	"strings"
)

type textWriter struct{}

//func (tw *textWriter) Open(o *Op, buf *bytes.Buffer) {
//}

func (tw *textWriter) Write(o *Op, buf *bytes.Buffer) {

	split := strings.Split(o.Data, "\n")
	if len(split) > 1 {
		buf.WriteString(split[0])
		buf.WriteString("</p>")
		o.ClosePrevAttrs(buf)
	} else {
		o.ClosePrevAttrs(buf)
	}

	buf.WriteString("<p>")
	buf.WriteString(o.Data)


}

//func (tw *textWriter) Close(o *Op, buf *bytes.Buffer) {
//}

type imageWriter struct{}

//func (iw *imageWriter) Open(o *Op, buf *bytes.Buffer) {
//}

func (iw *imageWriter) Write(o *Op, buf *bytes.Buffer) {

	o.ClosePrevAttrs(buf)

	buf.WriteString("<img src=")
	buf.WriteString(strconv.Quote(o.Data))
	buf.WriteString(">")

}

//func (iw *imageWriter) Close(o *Op, buf *bytes.Buffer) {
//}
