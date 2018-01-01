package quill

import (
	"bytes"
)

type textWriter struct {
	classes []string
}

func (tw *textWriter) Write(o *Op, tempBuf *bytes.Buffer) {
	tempBuf.WriteString(o.Data)
}

func (tw *textWriter) Open(o *Op, as *AttrState) string {
	return "<p" + ClassesList(AttrsToClasses(o.Attrs)) + ">"
}

func (tw *textWriter) Close(o *Op, as *AttrState) string {
	return "</p>"
}

//func (tw *textWriter) setClass(class string) {
//	for _, c := range tw.classes {
//		// Avoiding adding a class twice.
//		if c == class {
//			return
//		}
//	}
//	tw.classes = append(tw.classes, class)
//}

func (tw *textWriter) Classes() []string {
	return tw.classes
}

type blockQuoteWriter struct {
	classes []string
}

func (bw *blockQuoteWriter) Write(o *Op, buf *bytes.Buffer) {
	buf.WriteString(o.Data)
}

func (bw *blockQuoteWriter) TagName(o *Op) string {
	return "blockquote"
}

func (bw *blockQuoteWriter) SetClass(class string) {
	for _, c := range bw.classes {
		// Avoiding adding a class twice.
		if c == class {
			return
		}
	}
	bw.classes = append(bw.classes, class)
}

func (bw *blockQuoteWriter) GetClasses() []string {
	return bw.classes
}

type headerWriter struct {
	classes []string
}

func (hw *headerWriter) Write(o *Op, buf *bytes.Buffer) {
	buf.WriteString(o.Data)
}

func (hw *headerWriter) TagName(o *Op) string {
	return "h" + o.Attrs["header"]
}

func (hw *headerWriter) SetClass(class string) {
	for _, c := range hw.classes {
		// Avoiding adding a class twice.
		if c == class {
			return
		}
	}
	hw.classes = append(hw.classes, class)
}

func (hw *headerWriter) GetClasses() []string {
	return hw.classes
}
