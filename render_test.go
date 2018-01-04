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
		`[{"insert":"abc "},{"attributes":{"bold":true},"insert":"bld"},{"attributes":{"list":"bullet"},"insert":"\n"}]`,
		`[  {"attributes":{"bold":true,"italic":true},"insert":"bbii "},
			{"attributes":{"italic":true},"insert":"ii"},{"insert":"\n"}]`, // ordering alphabetically
	}

	want := []string{
		"<p><br></p>",
		"<p>line1</p><p>line2</p>",
		"<p>line1</p><p><br></p><p>line3</p>",
		"<blockquote>bkqt</blockquote>",
		`<p><span style="color:#a10000;">colored</span></p>`,
		"<ul><li>abc <strong>bld</strong></li></ul>",
		"<p><em><strong>bbii </strong>ii</em></p>", // not the Quill.js style of "<p><strong><em>bbii </em></strong><em>ii</em></p>"
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
