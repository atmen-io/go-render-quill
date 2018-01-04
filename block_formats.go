package quill

import "strings"

// paragraph
type textFormat struct{}

func (*textFormat) Fmt() *Format {
	return &Format{
		Val:   "p",
		Place: Tag,
		Block: true,
	}
}

func (*textFormat) HasFormat(o *Op) bool {
	return o.Type == "text"
}

// block quote
type blockQuoteFormat struct{}

func (*blockQuoteFormat) Fmt() *Format {
	return &Format{
		Val:   "blockquote",
		Place: Tag,
		Block: true,
	}
}

func (*blockQuoteFormat) HasFormat(o *Op) bool {
	return o.HasAttr("blockquote")
}

// header
type headerFormat struct {
	level string // the string "1", "2", "3", ...
}

func (hf *headerFormat) Fmt() *Format {
	return &Format{
		Val:   "h" + hf.level,
		Place: Tag,
		Block: true,
	}
}

func (hf *headerFormat) HasFormat(o *Op) bool {
	return o.Attrs["header"] == hf.level
}

// list
type listFormat struct {
	lType  string // either "ul" or "ol"
	indent uint8  // the number of nested
}

func (lf *listFormat) Fmt() *Format {
	return &Format{
		Val:   "li",
		Place: Tag,
		Block: true,
	}
}

func (lf *listFormat) HasFormat(o *Op) bool {
	return o.HasAttr("list")
}

// listFormat implements the FormatWrapper interface.
func (lf *listFormat) Wrap() (string, string) {
	return "<" + lf.lType + ">", "</" + lf.lType + ">"
}

// listFormat implements the FormatWrapper interface.
func (lf *listFormat) Open(open []*Format, o *Op) bool {
	var count uint8
	for i := range open {
		if open[i].Place == Tag && open[i].Val == "<"+lf.lType+">" {
			count++
		}
	}
	if count <= lf.indent {
		return true
	}
	return false
}

// listFormat implements the FormatWrapper interface.
func (lf *listFormat) Close(open []*Format, o *Op) bool {
	if !o.HasAttr("list") && strings.IndexByte(o.Data, '\n') != -1 {
		return true
	}
	var count uint8
	for i := range open {
		if open[i].Place == Tag && open[i].Val == lf.lType {
			count++
		}
	} // TODO
	return false
}

// indentDepths gives either the indent amount of a list or 0 if there is no indenting.
var indentDepths = map[string]uint8{
	"1": 1,
	"2": 2,
	"3": 3,
	"4": 4,
	"5": 5,
}

// text alignment
type alignFormat struct {
	val string
}

func (af *alignFormat) Fmt() *Format {
	return &Format{
		Val:   "align-" + af.val,
		Place: Class,
		Block: true,
	}
}

func (af *alignFormat) HasFormat(o *Op) bool {
	return o.Attrs["align"] == af.val
}
