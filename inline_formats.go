package quill

import (
	"io"
	"strconv"
)

type boldFormat struct{}

func (*boldFormat) Fmt() *Format {
	return &Format{
		Val:   "strong",
		Place: Tag,
	}
}

type imageFormat struct {
	src, alt string
}

func (*imageFormat) Fmt() *Format { return new(Format) } // The body contains the entire element.

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

func (*italicFormat) Fmt() *Format {
	return &Format{
		Val:   "em",
		Place: Tag,
	}
}

type colorFormat struct {
	c string
}

func (cf *colorFormat) Fmt() *Format {
	return &Format{
		Val:   cf.c,
		Place: Style,
	}
}

type linkFormat struct {
	href string
}

func (*linkFormat) Fmt() *Format { return new(Format) } // a wrapper only

func (lf *linkFormat) PreWrap(_ []*Format) string {
	return `<a href=` + strconv.Quote(lf.href) + ` target="_blank">`
}

func (lf *linkFormat) PostWrap(_ []*Format, o *Op) string {
	if o.HasAttr("link") && o.Attrs["link"] == lf.href {
		return ""
	}
	return "</a>"
}
