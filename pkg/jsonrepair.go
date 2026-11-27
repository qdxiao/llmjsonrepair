// Package pkg repairs malformed JSON, including truncated output from language models.
// Repair is best effort; validate repaired data against the expected schema before using it.
package pkg

import (
	"encoding/json"
	"fmt"
)

// Repair returns formatted JSON, attempting to repair malformed input.
// Repair can guess missing structure, but cannot recover values that were never present.
func Repair(jsonStr string) (string, error) {
	// 尝试直接解析，如果成功就直接返回
	var out interface{}
	if err := json.Unmarshal([]byte(jsonStr), &out); err == nil {
		repaired, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to re-marshal already-valid json: %w", err)
		}
		return string(repaired), nil
	}

	// 如果直接解析失败，则启动修复程序
	parser := NewParser(jsonStr)
	parsedJSON, err := parser.Parse()
	if err != nil {
		return "", err
	}

	repaired, err := json.MarshalIndent(parsedJSON, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal repaired json: %w", err)
	}
	return string(repaired), nil
}

// Loads parses valid or repairable JSON into Go values. Objects become maps and
// arrays become slices. Input with multiple JSON values may produce a slice.
func Loads(jsonStr string) (interface{}, error) {
	// 尝试直接解析
	var out interface{}
	if err := json.Unmarshal([]byte(jsonStr), &out); err == nil {
		return out, nil
	}

	// 如果失败则修复
	parser := NewParser(jsonStr)
	return parser.Parse()
}
