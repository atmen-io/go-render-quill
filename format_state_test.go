package quill

import (
	"bytes"
	"testing"
)

func TestFormatState_addFormat(t *testing.T) {

	cases := []struct {
		current []*Format
		keyword string
		o       *Op
		want    []*Format
	}{
		{
			current: []*Format{}, // no formats
			keyword: "italic",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"italic": "y"},
			},
			want: []*Format{
				{
					Val:   "em",
					Place: Tag,
				},
			},
		},
		{
			current: []*Format{
				{ // One format already set.
					Val:   "em",
					Place: Tag,
				},
			},
			keyword: "italic",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"italic": "y"},
			},
			want: []*Format{ // The way addFormat works, it does not check if the format is already added.
				{
					Val:   "em",
					Place: Tag,
				},
				{
					Val:   "em",
					Place: Tag,
				},
			},
		},
	}

	fs := new(formatState)   // reuse
	buf := new(bytes.Buffer) // reuse

	for i, ca := range cases {

		fs.open = ca.current

		fmTer := ca.o.getFormatter(ca.keyword, nil)
		fm := fmTer.Fmt()
		fm.fm = fmTer

		fs.addFormat(fm, buf)

		if len(ca.want) != len(fs.open) {
			t.Errorf("unequal count of formats (index %d); got %s", i, fs.open)
			t.FailNow()
		}

		for j := range fs.open {
			if ca.want[j].Val != fs.open[j].Val {
				t.Errorf("did not add format Val correctly (index %d); got %q", i, fs.open[j].Val)
			}
			if ca.want[j].Place != fs.open[j].Place {
				t.Errorf("did not add format Place correctly (index %d); got %v", i, fs.open[j].Place)
			}
			if ca.want[j].Block != fs.open[j].Block {
				t.Errorf("did not add format Block correctly (index %d); got %v", i, fs.open[j].Block)
			}
		}

	}

}

func TestFormatState_ClosePrevFormats(t *testing.T) {

	o := &Op{
		Data: "stuff",
		Type: "text",
		// no attributes set
	}

	cases := []formatState{
		{[]*Format{
			{"em", Tag, false, o.getFormatter("italic", nil)},
			{"strong", Tag, false, o.getFormatter("bold", nil)},
		}},
		{[]*Format{
			{"<ul>", Tag, false, o.getFormatter("list", nil)}, // wrapped by FormatWrapper
			{"li", Tag, true, o.getFormatter("list", nil)},
			{"strong", Tag, false, o.getFormatter("bold", nil)},
		}},
	}

	want := []string{"</strong></em>", "</strong></li></ul>"}

	buf := new(bytes.Buffer)

	for i := range cases {

		cases[i].closePrevious(buf, o)
		got := buf.String()
		if got != want[i] {
			t.Errorf("closed formats wrong (index %d); wanted %q; got %q\n", i, want[i], got)
		}

		buf.Reset()

	}

}
