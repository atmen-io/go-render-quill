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

func (*imageFormat) Format() (_ string, _ StyleFormat) { return } // The body contains the entire element.

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

func (*italicFormat) Format() (string, StyleFormat) { return "em", Tag }

type colorFormat struct {
	c string
}

func (cf *colorFormat) Format() (string, StyleFormat) { return cf.c, Style }

type linkFormat struct {
	href string
}

func (*linkFormat) Format() (_ string, _ StyleFormat) { return } // Wrapper only.

func (lf *linkFormat) PreWrap(_ []string) string {
	return "<a href=" + strconv.Quote(lf.href) + ` target="_blank">`
}

func (lf *linkFormat) PostWrap(openedTags []string, o *Op) string {
	if o.HasAttr("link") {
		return ""
	}
	return "</a>"
}
