package quill

type textFormat struct{}

func (*textFormat) Fmt() *Format {
	return &Format{
		Val:   "p",
		Place: Tag,
		Block: true,
	}
}

type blockQuoteFormat struct{}

func (*blockQuoteFormat) Fmt() *Format {
	return &Format{
		Val:   "blockquote",
		Place: Tag,
		Block: true,
	}
}

type headerFormat struct {
	h string // the string "h1", "h2", ...
}

func (hf *headerFormat) Fmt() *Format {
	return &Format{
		Val:   hf.h,
		Place: Tag,
		Block: true,
	}
}

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

// listFormat implements the FormatWrapper interface.
func (lf *listFormat) PreWrap(openTags []*Format) string {
	var count uint8
	for i := range openTags {
		if openTags[i].Place == Tag && openTags[i].Val == lf.lType {
			count++
		}
	}
	if count <= lf.indent {
		return "<" + lf.lType + ">"
	}
	return ""
}

// listFormat implements the FormatWrapper interface.
func (lf *listFormat) PostWrap(openedTags []string, o *Op) string {
	if o.Attrs["list"] == lf.lType { // TODO: too simplistic; check for nested lists
		return ""
	}
	return "</" + lf.lType + ">"
}

// indentDepths gives either the indent amount of a list or 0 if there is no indenting.
var indentDepths = map[string]uint8{
	"1": 1,
	"2": 2,
	"3": 3,
	"4": 4,
	"5": 5,
}

type alignFormat struct {
	align string
}

func (af *alignFormat) Fmt() *Format {
	return &Format{
		Val:   af.align,
		Place: Class,
		Block: true,
	}
}
