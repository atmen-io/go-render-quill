package quill

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRawOp_makeOp(t *testing.T) {

	rawOps := []rawOp{
		{
			Insert: "string to insert.",
			Attrs: map[string]interface{}{
				"bold":      true,
				"link":      "https://widerwebs.com",
				"italic":    false,
				"underline": nil, // nil value is set if JSON value is null
			},
		},
		{
			Insert: "\n",
			Attrs: map[string]interface{}{
				"align": "center",
			},
		},
		{
			Insert: "\n",
			Attrs: map[string]interface{}{
				"align":      "center",
				"blockquote": true,
			},
		},
	}

	want := []Op{
		{
			Data: "string to insert.\n",
			Type: "string",
			Attrs: map[string]string{
				"bold":      "y",
				"italic":    "",
				"link":      "https://widerwebs.com",
				"underline": "",
			},
		},
		{
			Data: "\n",
			Type: "text",
			Attrs: map[string]string{
				"align": "center",
			},
		},
		{
			Data: "\n",
			Type: "text",
			Attrs: map[string]string{
				"align":      "center",
				"blockquote": "y",
			},
		},
	}

	for i := range rawOps {

		got, err := rawOps[i].makeOp()
		if err != nil {
			fmt.Errorf("error: %s", err)
		}

		if !reflect.DeepEqual(*got, want[i]) {
			t.Errorf("failed rawOpToOp; got %v for index %s", want[i], i)
		}

	}

}

func TestExtractString(t *testing.T) {
	if extractString("random string") != "random string" {
		t.Errorf("failed stringc extract")
	}
	if extractString(true) != "y" {
		t.Errorf("failed bool true extract")
	}
	if extractString(false) != "" {
		t.Errorf("failed bool false extract")
	}
}
