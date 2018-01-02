package quill

import (
	"io"
	"strconv"
)

type boldFormat struct{}

func (*boldFormat) TagName() string {
	return "strong"
}

func (*boldFormat) Class() string { return "" }

func (*boldFormat) Style() string { return "" }

type imageFormat struct {
	src, alt string
}

func (*imageFormat) TagName() string {
	return "" // The body contains the entire element.
}

func (*imageFormat) Class() string { return "" }

func (*imageFormat) Style() string { return "" }

func (iw *imageFormat) Write(buf io.Writer) {

	buf.Write([]byte("<img src="))
	buf.Write([]byte(strconv.Quote(iw.src)))
	if iw.alt != "" {
		buf.Write([]byte(" alt="))
		buf.Write([]byte(strconv.Quote(iw.alt)))
	}
	buf.Write([]byte{'>'})

}

type italicFormat struct{}

func (*italicFormat) TagName() string {
	return "em"
}

func (*italicFormat) Class() string { return "" }

func (*italicFormat) Style() string { return "" }

type colorFormat struct {
	c string //
}

func (*colorFormat) TagName() string {
	return ""
}

func (*colorFormat) Class() string { return "" }

func (cf *colorFormat) Style() string { return cf.c }
