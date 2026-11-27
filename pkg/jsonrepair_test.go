package pkg_test

import (
	"bytes"
	"encoding/json"
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
