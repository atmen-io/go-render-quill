package quill

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSimple(t *testing.T) {

	cases := []string{
		`[{"insert": "\n"}]`,
		`[{"insert": "line1\nline2\n"}]`,
		`[{"insert": "line1\n\nline3\n"}]`,
		`[{"insert": "bkqt"}, {"attributes": {"blockquote": true}, "insert": "\n"}]`,
		`[{"attributes": {"color": "#a10000"}, "insert": "colored"}, {"insert": "\n"}]`,
	}

	want := []string{
		"<p><br></p>",
		"<p>line1</p><p>line2</p>",
		"<p>line1</p><p><br></p><p>line3</p>",
		"<blockquote>bkqt</blockquote>",
		`<p><span style="color:#a10000;">colored</span></p>`,
	}

	for i := range cases {

		bts, err := Render([]byte(cases[i]))
		if err != nil {
			t.Errorf("%s", err)
			t.FailNow()
		}
		if string(bts) != want[i] {
			t.Errorf("bad rendering (index %d); got: %s", i, bts)
		}

	}

}

func TestOps1(t *testing.T) {
	if err := testPair("ops1.json", "ops1.html"); err != nil {
		t.Errorf("%s", err)
	}
}

func TestNested(t *testing.T) {
	if err := testPair("nested.json", "nested.html"); err != nil {
		t.Errorf("%s", err)
	}
}

func testPair(opsFile, htmlFile string) error {
	opsArr, err := ioutil.ReadFile("./tests/" + opsFile)
	if err != nil {
		return fmt.Errorf("could not read %s; %s\n", opsFile, err)
	}
	desired, err := ioutil.ReadFile("./tests/" + htmlFile)
	if err != nil {
		return fmt.Errorf("could not read %s; %s\n", htmlFile, err)
	}
	got, err := Render(opsArr)
	if err != nil {
		return fmt.Errorf("error rendering; %s\n", err)
	}
	if !bytes.Equal(desired, got) {
		return fmt.Errorf("bad rendering; \nwanted: \n%s\ngot: \n%s\n", desired, got)
	}
	return nil
}

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
			want: []*Format{
				{ // Stay the same.
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

func TestOp_ClosePrevFormats(t *testing.T) {

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

		o.closePrevFormats(buf, &cases[i], nil)
		got := buf.String()
		if got != want[i] {
			t.Errorf("closed formats wrong (index %d); wanted %q; got %q\n", i, want[i], got)
		}

		buf.Reset()

	}

}
