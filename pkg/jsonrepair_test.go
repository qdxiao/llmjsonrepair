package pkg_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/qdxiao/llmjsonrepair/pkg"
)

func TestRepairQuickstartAndBoundaries(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"closes truncated containers", `{"items":["a","b"`, `{"items":["a","b"]}`},
		{"valid JSON", `{"value":[true,false,null,2]}`, `{"value":[true,false,null,2]}`},
		{"bare keys and values", `{name: Ada, active: true}`, `{"active":true,"name":"Ada"}`},
		{"extra commas", `{,x:[,1,2,],}`, `{"x":[1,2]}`},
		{"quoted string escapes", `{'text':'a\b\f\n\r\t\/\q',}`, `{"text":"a\b\f\n\r\t/\\q"}`},
		{"unfinished quoted string", `{'text':'hello`, `{"text":"hello"}`},
		{"false and null in malformed array", `[false,null,`, `[false,null]`},
		{"float and integer", `[-12,1.25,2e2,`, `[-12,1.25,200]`},
		{"concatenated objects", `{"first":1} {"second":2`, `[{"first":1},{"second":2}]`},
		{"missing value becomes null", `{"total":`, `{"total":null}`},
		{"prefixed text can change root type", `Here is the JSON: {"ok":true`, `["the JSON",{"ok":true}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkg.Repair(tt.raw)
			if err != nil {
				t.Fatal(err)
			}
			var compact bytes.Buffer
			if err := json.Compact(&compact, []byte(got)); err != nil {
				t.Fatalf("Repair returned invalid JSON %q: %v", got, err)
			}
			if compact.String() != tt.want {
				t.Errorf("Repair = %s, want %s", &compact, tt.want)
			}
		})
	}
}

func TestLoadsReturnsParsedValues(t *testing.T) {
	tests := []struct {
		raw  string
		want any
	}{
		{`{"active":true,"labels":[]}`, map[string]any{"active": true, "labels": []any{}}},
		{`[true,false,null`, []any{true, false, nil}},
		{`{title:你好`, map[string]any{"title": "你好"}},
	}
	for _, tt := range tests {
		got, err := pkg.Loads(tt.raw)
		if err != nil || !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Loads(%q) = %#v, %v; want %#v", tt.raw, got, err, tt.want)
		}
	}
}

func ExampleRepair() {
	fixed, err := pkg.Repair(`{"ok":true`)
	if err != nil {
		panic(err)
	}
	fmt.Println(fixed)
	// Output:
	// {
	//   "ok": true
	// }
}
