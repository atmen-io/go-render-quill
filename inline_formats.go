package quill

import (
	"io"
	"strconv"
)

type boldFormat struct{}

func (*boldFormat) Format() (string, StyleFormat) { return "strong", Tag }

type imageFormat struct {
	src, alt string
}

func (*imageFormat) Format() (string, StyleFormat) { return "", Tag } // The body contains the entire element.

// imageFormat implements the FormatWriter interface.
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

func (*italicFormat) TagName() (string, StyleFormat) { return "em", Tag }

type colorFormat struct {
	c string
}

func (cf *colorFormat) Format() (string, StyleFormat) { return cf.c, Style }
