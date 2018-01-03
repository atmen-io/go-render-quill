package quill

type textFormat struct{}

func (*textFormat) Format() (string, StyleFormat) { return "p", Tag }

type blockQuoteFormat struct{}

func (*blockQuoteFormat) Format() (string, StyleFormat) { return "blockquote", Tag }

type headerFormat struct {
	h string // the string "h1", "h2", ...
}

func (hf *headerFormat) Format() (string, StyleFormat) { return hf.h, Tag }

type listFormat struct {
	lType  string // either "ul" or "ol"
	indent uint8  // the number of nested
}

func (lf *listFormat) Format() (string, StyleFormat) { return "li", Tag }

// listFormat implements the FormatWrapper interface.
func (lf *listFormat) PreWrap(openedTags []string) string {
	var count uint8
	for i := range openedTags {
		if openedTags[i] == lf.lType {
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
	if o.HasAttr("list") {
		return ""
	}
	return "</" + lf.lType + ">" // TODO: too simplistic, check for nested lists
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

func (af *alignFormat) Format() (string, StyleFormat) { return af.align, Class }
