package main

import (
	"fmt"
	"log"

	"github.com/qdxiao/llmjsonrepair/pkg"
)

type testCase struct {
	name        string
	malformed   string
	description string
}

func main() {
	testCases := []testCase{
		{
			name:        "简单修复：缺少结尾括号",
			malformed:   `{"name": "John Doe", "age": 30, "courses": ["Math", "Science"`,
			description: "测试最常见的情况，数组和对象都未闭合。",
		},
		{
			name:        "深度嵌套错误",
			malformed:   `{"id": 1, "user": {name: "Alice", details: { "email": "alice@example.com", affiliations: ["Org1", "Org2`,
			description: "在深层嵌套的对象中，键缺少引号，数组未闭合，整个结构也未闭合。",
		},
		{
			name:        "混乱的引号和转义",
			malformed:   `{"quote": "He said, \"This is a test.", "message": "Here's another quote: 'Hello World'`,
			description: "字符串内部包含转义引号，但整个字符串本身未正确闭合。还包含单引号。",
		},
		{
			name:        "JSON 对象流被截断",
			malformed:   `{"event": "start", "id": 1}{"event": "update", "id": 1, "payload": {"status": "in_progress"`,
			description: "模拟多个 JSON 对象流，但在第二个对象的中间被截断。",
		},
		{
			name:        "结构混乱的键值对",
			malformed:   `{user "John" age 30 city "New York" valid true`,
			description: "键和值之间缺少冒号和逗号，需要解析器进行大量猜测。",
		},
		{
			name:        "值是未闭合的对象",
			malformed:   `{"data": {"key1": "value1", "key2": {"nested_key": "nested_value"`,
			description: "一个对象的值是另一个未闭合的对象。",
		},
		{
			name:        "空键和缺失的值",
			malformed:   `{"": "empty key", "key_with_missing_value":, "another_key": "value"}`,
			description: "测试空字符串作为键，以及一个键后面只有逗号没有值的情况。",
		},
		{
			name:        "数组中混合了未加引号的字符串和数字",
			malformed:   `["string1", item2, 3, "item4`,
			description: "数组中包含未加引号的字符串字面量。",
		},
		{
			name:        "LLM 思考过程残留",
			malformed:   `Here is the JSON: {"reasoning": "The user wants a summary.", "result": {"summary": "This is a summary text...`,
			description: "JSON 前面有非 JSON 的文本，解析器应该能跳过它并找到 JSON 的开始。",
		},
	}
	for _, tc := range testCases {
		repairedData, err := pkg.Loads(tc.malformed)
		if err != nil {
			log.Printf("Load failed: %v\n", err)
			return
		}
		// 使用 Repair 函数来获取格式化的 JSON 字符串
		repairedString, err := pkg.Repair(tc.malformed)
		if err != nil {
			log.Printf("Parsing failed: %v\n", err)
			return
		}
		fmt.Printf("Parsing successful：%s-%s", repairedString, repairedData)
	}
}
