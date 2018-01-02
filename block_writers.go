package quill

type textFormat struct{}

func (*textFormat) TagName() string { return "p" }

func (*textFormat) Class() string { return "" }

func (*textFormat) Style() string { return "" }

type blockQuoteFormat struct{}

func (*blockQuoteFormat) TagName() string { return "blockquote" }

func (*blockQuoteFormat) Class() string { return "" }

func (*blockQuoteFormat) Style() string { return "" }

type headerFormat struct {
	h string // the string "h1", "h2", ...
}

func (hf *headerFormat) TagName() string { return hf.h }

func (*headerFormat) Class() string { return "" }

func (*headerFormat) Style() string { return "" }

type listFormat struct {
	lType  string // either "ul" or "ol"
	indent uint8  // the number of nested
}

func (lf *listFormat) TagName() string { return "li" }

func (*listFormat) Class() string { return "" }

func (*listFormat) Style() string { return "" }

func (lf *listFormat) Wrap(openedTags []string) string {
	var count uint8
	for i := range openedTags {
		if openedTags[i] == lf.lType {
			count++
		}
	}
	if count <= lf.indent {
		return lf.lType
	}
	return ""
}

// indentDepths gives either the indent amount of a list or 0 if there is no indenting.
var indentDepths = map[string]uint8{
	"1": 1,
	"2": 2,
	"3": 3,
	"4": 4,
	"5": 5,
}
