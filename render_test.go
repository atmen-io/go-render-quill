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
	}

	want := []string{
		"<p><br></p>",
		"<p>line1</p><p>line2</p>",
		"<p>line1</p><p><br></p><p>line3</p>",
	}

	for i := range cases {

		bts, err := Render([]byte(cases[i]))
		if err != nil {
			t.Errorf("%s", err)
			t.FailNow()
		}
		if string(bts) != want[i] {
			t.Errorf("bad rendering; got: %s", bts)
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
		fms     []*Format
		keyword string
		o       *Op
		want    []*Format
	}{
		{
			fms: []*Format{}, // no formats
			keyword: "header",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"header":"1"},
			},
			want: []*Format{
				{
					Val: "h1",
					Place: Tag,
					Block: true,
					keyword: "header",
				},
			},
		},
		{
			fms: []*Format{
				{ // One format already set.
					Val: "h1",
					Place: Tag,
					Block: true,
					keyword: "header",
				},
			},
			keyword: "header",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"header":"1"},
			},
			want: []*Format{
				{ // Stay the same.
					Val: "h1",
					Place: Tag,
					Block: true,
					keyword: "header",
				},
			},
		},
	}

	fs := new(formatState) // reused
	buf := new(bytes.Buffer)

	for i, ca := range cases {

		fs.open = ca.fms

		fs.addFormat(ca.keyword, ca.o.getFormatter(ca.keyword, nil), buf)

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
			if ca.want[j].keyword != fs.open[j].keyword {
				t.Errorf("did not add format keyword correctly (index %d); got %q", i, fs.open[j].keyword)
			}
		}

	}

}

func TestOp_ClosePrevFormats(t *testing.T) {
	cases := []formatState{
		{[]*Format{
			{"em", Tag, false, "italic"},
			{"strong", Tag, false, "bold"},
		}},
		{[]*Format{
			{"<ul>", Tag, false, "list"}, // wrapped by FormatWrapper
			{"li", Tag, true, "list"},
			{"strong", Tag, false, "bold"},
		}},
	}
	want := []string{"</strong></em>", "</strong></li></ul>"}
	o := &Op{
		Data: "stuff",
		Type: "text",
		// no attributes set
	}

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
