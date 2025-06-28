package pkg

import (
	"encoding/json"
	"fmt"
)

// Repair 尝试修复并解析JSON字符串
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

// Loads 修复JSON并返回一个数据结构 (map[string]interface{} 或 []interface{})
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
