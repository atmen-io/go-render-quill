package quill

import (
	"bytes"
	"sort"
	"testing"
)

func TestFormatState_add(t *testing.T) {

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
			want: []*Format{ // The way add works, it does not check if the format is already added.
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

	fs := new(formatState) // reuse

	for i, ca := range cases {

		fs.open = ca.current

		fmTer := ca.o.getFormatter(ca.keyword, nil)
		fm := fmTer.Fmt()
		fm.fm = fmTer

		fs.add(fm)

		if len(ca.want) != len(fs.open) {
			t.Errorf("(index %d) unequal count of formats", i)
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

func TestFormatState_closePrevious(t *testing.T) {

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

func TestFormatState_Sort(t *testing.T) {

	o := &Op{
		Data: "stuff",
		Type: "text",
		// no attributes set
	}

	cases := []*formatState{
		{[]*Format{
			{"strong", Tag, false, o.getFormatter("bold", nil)},
			{"em", Tag, false, o.getFormatter("italic", nil)},
		}},
		{[]*Format{
			{"u", Tag, false, o.getFormatter("underline", nil)},
			{"align-center", Class, false, o.getFormatter("align", nil)},
			{"strong", Tag, false, o.getFormatter("bold", nil)},
		}},
		{[]*Format{
			{"color:#e0e0e0;", Style, false, o.getFormatter("color", nil)},
			{"em", Tag, false, o.getFormatter("italic", nil)},
		}},
		{[]*Format{
			{"em", Tag, false, o.getFormatter("italic", nil)},
			{`<a href="https://widerwebs.com" target="_blank">`, Tag, false, o.getFormatter("link", nil)}, // link wrapper
		}},
	}

	want := []*formatState{
		{[]*Format{
			{"em", Tag, false, o.getFormatter("italic", nil)},
			{"strong", Tag, false, o.getFormatter("bold", nil)},
		}},
		{[]*Format{
			{"strong", Tag, false, o.getFormatter("bold", nil)},
			{"u", Tag, false, o.getFormatter("underline", nil)},
			{"align-center", Class, false, o.getFormatter("align", nil)},
		}},
		{[]*Format{
			{"em", Tag, false, o.getFormatter("italic", nil)},
			{"color:#e0e0e0;", Style, false, o.getFormatter("color", nil)},
		}},
		{[]*Format{
			{`<a href="https://widerwebs.com" target="_blank">`, Tag, false, o.getFormatter("link", nil)}, // link wrapper
			{"em", Tag, false, o.getFormatter("italic", nil)},
		}},
	}

	for i := range cases {

		sort.Sort(cases[i])

		caseI := cases[i].open
		wantI := want[i].open

		ok := true
		for j := range caseI {
			if caseI[j].Val != wantI[j].Val {
				ok = false
			} else if caseI[j].Place != wantI[j].Place {
				ok = false
			}
		}
		if !ok {
			t.Errorf("bad sorting (index %d); got:\n", i)
			for k := range caseI {
				t.Errorf("  (%d) %+v\n", k, *caseI[k])
			}
		}

	}

}
