package quill

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestOps1(t *testing.T) {
	if err := testPair("ops1.json", "ops1.html"); err != nil {
		t.Errorf("%s", err)
	}
}

func TestEmpty(t *testing.T) {
	if err := testPair("empty.json", "empty.html"); err != nil {
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
	got, err := Render(opsArr, nil)
	if err != nil {
		return fmt.Errorf("error rendering; %s\n", err)
	}
	if !bytes.Equal(desired, got) {
		return fmt.Errorf("bad rendering; wanted: \n%s\n got: \n%s\n", desired, got)
	}
	return nil
}

func TestOp_ClosePrevAttrs(t *testing.T) {
	attrStates = []string{"italic", "bold"}
	o := &Op{
		Data: "stuff",
		// no attributes set
	}
	desired := "</b></em>"
	buf := new(bytes.Buffer)
	o.ClosePrevAttrs(buf)
	got := buf.String()
	if got != desired {
		t.Errorf("closed attributes wrong; wanted %q; got %q\n", desired, got)
	}
}
