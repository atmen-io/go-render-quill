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
	lType string // either "ul" or "ol"
}

func (lf *listFormat) TagName() string { return lf.lType }

func (*listFormat) Class() string { return "" }

func (*listFormat) Style() string { return "" }
