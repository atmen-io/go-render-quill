package quill

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func Example() {
	ops := []byte(`[
		{
			"insert": "Heading1"
		},
		{
			"attributes": {
				"header": 1
			},
			"insert": "\n"
		},
		{
			"insert": "Hello, this is text.\nAnd "
		},
		{
			"attributes": {
				"italic": true
			},
			"insert": "here is italic "
		},
		{
			"insert": "(and not).\nAnd "
		},
		{
			"attributes": {
				"bold": true
			},
			"insert": "here is bold"
		}
	]`)
	fmt.Println(Render(ops, nil))
	//-- Output: <h1>Heading1</h1><p>Hello, this is text.</p><p>And <em>here is italic </em>(and not).</p><p>And <strong>here is bold</strong>
}

func TestRawOpToOp(t *testing.T) {

	ro := rawOp{
		"insert": "string to insert.\n",
		"attributes": map[string]interface{}{
			"bold":   true,
			"link":   "https://widerwebs.com",
			"italic": false,
		},
	}

	desired := Op{
		Data: "string to insert.\n",
		Type: "string",
		Attrs: map[string]string{
			"bold":   "y",
			"italic": "n",
			"link":   "https://widerwebs.com",
		},
	}

	got, err := rawOpToOp(ro)
	if err != nil {
		fmt.Errorf("error: %s", err)
	}

	if !reflect.DeepEqual(*got, desired) {
		t.Errorf("failed rawOpToOp; wanted %v; got %v", desired, got)
	}

}

func TestExtractString(t *testing.T) {
	if extractString("random string") != "random string" {
		t.Errorf("failed stringc extract")
	}
	if extractString(true) != "y" {
		t.Errorf("failed bool true extract")
	}
	if extractString(false) != "n" {
		t.Errorf("failed bool false extract")
	}
}

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
