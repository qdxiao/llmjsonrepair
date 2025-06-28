# llmjsonrepair

专门用于修复大模型输出 JSON 格式错误的 Go 语言库

## 简介

`llmjsonrepair` 是一个强大的 Go 语言库，专门用于修复大语言模型（LLM）输出的不完整或格式错误的 JSON 数据。当 LLM 生成的 JSON 因为各种原因（如输出截断、格式错误、缺少引号等）无法正常解析时，这个库可以智能地修复这些问题。

## 特性

- 🔧 **智能修复**: 自动修复各种 JSON 格式错误
- 🎯 **上下文感知**: 基于解析上下文智能推断缺失的结构
- 🚀 **高容错性**: 处理深度嵌套、混合格式等复杂情况
- 📦 **简单易用**: 提供简洁的 API 接口
- 🔄 **多对象支持**: 支持处理连续的 JSON 对象流
- 🧠 **LLM 友好**: 专门针对大模型输出特点设计
- ⚡ **高性能**: 基于状态机的解析器，性能优异
- 🛡️ **安全可靠**: 经过充分测试，处理各种边界情况

## 支持修复的问题类型

- ✅ 缺少结尾括号/方括号
- ✅ 对象键缺少引号
- ✅ 字符串未正确闭合
- ✅ 深度嵌套结构错误
- ✅ 混乱的引号和转义字符
- ✅ 多 JSON 对象流被截断
- ✅ 缺少分隔符（逗号、冒号）
- ✅ LLM 思考过程文本残留
- ✅ 数值格式错误
- ✅ 布尔值和 null 值错误

## 安装

```bash
go get github.com/qdxiao/llmjsonrepair
```

## 快速开始

```go
package main

import (
    "fmt"
    "log"
    "github.com/qdxiao/llmjsonrepair/pkg"
)

func main() {
    // 示例：修复缺少结尾括号的 JSON
    malformedJSON := `{"name": "John", "age": 30, "skills": ["Go", "Python"`
    
    // 方法1: 修复并返回格式化的 JSON 字符串
    repairedJSON, err := pkg.Repair(malformedJSON)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("修复后的 JSON:")
    fmt.Println(repairedJSON)
    
    // 方法2: 修复并返回解析后的数据结构
    data, err := pkg.Loads(malformedJSON)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("解析后的数据: %+v\n", data)
}
```

## API 文档

### `Repair(jsonStr string) (string, error)`

修复 JSON 字符串并返回格式化的 JSON 字符串。

**参数:**
- `jsonStr`: 需要修复的 JSON 字符串

**返回:**
- `string`: 修复并格式化后的 JSON 字符串
- `error`: 错误信息（如果修复失败）

**示例:**
```go
malformed := `{"name": "Alice", "age": 25`
repaired, err := pkg.Repair(malformed)
// 输出: {"name": "Alice", "age": 25}
```

### `Loads(jsonStr string) (interface{}, error)`

修复 JSON 字符串并返回解析后的数据结构。

**参数:**
- `jsonStr`: 需要修复的 JSON 字符串

**返回:**
- `interface{}`: 解析后的数据结构（`map[string]interface{}` 或 `[]interface{}`）
- `error`: 错误信息（如果修复失败）

**示例:**
```go
malformed := `[1, 2, 3`
data, err := pkg.Loads(malformed)
// 返回: []interface{}{1, 2, 3}
```

### `NewParser(jsonStr string, opts ...Option) *parser`

创建一个新的解析器实例，支持自定义选项。

**参数:**
- `jsonStr`: 需要解析的 JSON 字符串
- `opts`: 可选的配置选项

**可用选项:**
- `WithLogger(logger Logger)`: 设置自定义日志记录器

## 详细使用示例

### 1. 基本修复示例

```go
// 缺少结尾括号
malformed := `{"user": {"name": "John", "email": "john@example.com"`
repaired, _ := pkg.Repair(malformed)
fmt.Println(repaired)
// 输出: {"user": {"name": "John", "email": "john@example.com"}}
```

### 2. 处理未加引号的键

```go
// 对象键缺少引号
malformed := `{name: "Alice", age: 30, active: true}`
repaired, _ := pkg.Repair(malformed)
fmt.Println(repaired)
// 输出: {"name": "Alice", "age": 30, "active": true}
```

### 3. 修复字符串闭合问题

```go
// 字符串未正确闭合
malformed := `{"message": "Hello World, "status": "ok"}`
repaired, _ := pkg.Repair(malformed)
fmt.Println(repaired)
// 输出: {"message": "Hello World", "status": "ok"}
```

### 4. 处理多 JSON 对象流

```go
// 多个 JSON 对象，第二个被截断
malformed := `{"id": 1, "type": "start"}{"id": 2, "type": "update", "data": {"progress": 50`
data, _ := pkg.Loads(malformed)
fmt.Printf("%+v\n", data)
// 输出: [map[id:1 type:start] map[id:2 type:update data:map[progress:50]]]
```

### 5. 跳过非 JSON 文本

```go
// JSON 前有说明文字
malformed := `Here is the response: {"result": "success", "code": 200`
repaired, _ := pkg.Repair(malformed)
fmt.Println(repaired)
// 输出: {"result": "success", "code": 200}
```

## 高级用法

### 自定义日志记录

```go
import "log"

// 创建自定义日志记录器
logger := log.New(os.Stdout, "[JSON-REPAIR] ", log.LstdFlags)

// 使用自定义日志记录器创建解析器
parser := pkg.NewParser(malformedJSON, pkg.WithLogger(logger))
result, err := parser.Parse()
```

## 性能特点

- **内存效率**: 使用 rune 切片处理 Unicode 字符，避免字符串频繁分配
- **解析速度**: 基于状态机的单遍解析，时间复杂度 O(n)
- **容错能力**: 智能跳过无效字符，最大化数据恢复

## 适用场景

- **AI 应用开发**: 处理 ChatGPT、Claude 等大模型的 JSON 输出
- **API 集成**: 修复第三方 API 返回的格式错误 JSON
- **数据处理**: 清理和修复日志文件中的 JSON 数据
- **爬虫开发**: 处理网页中提取的不完整 JSON 数据
- **配置文件**: 修复手动编辑导致的配置文件格式错误

## 测试用例

项目包含了全面的测试用例，涵盖以下场景：

1. 简单修复：缺少结尾括号
2. 深度嵌套错误
3. 混乱的引号和转义
4. JSON 对象流被截断
5. 结构混乱的键值对
6. 值是未闭合的对象
7. 空键和缺失的值
8. 数组中混合了未加引号的字符串和数字
9. LLM 思考过程残留

运行测试：
```bash
go test ./pkg -v
```

## 常见问题

### Q: 这个库能处理所有类型的 JSON 错误吗？
A: 该库专门针对 LLM 输出的常见错误进行优化，能处理大部分格式问题。对于极端复杂的错误，可能需要人工干预。

### Q: 修复后的 JSON 是否保持原有的数据类型？
A: 是的，库会尽力保持原有的数据类型（字符串、数字、布尔值、null）。

### Q: 性能如何？
A: 基于状态机的单遍解析，对于大多数应用场景性能表现良好。具体性能取决于输入数据的复杂度。

### Q: 是否支持自定义修复规则？
A: 当前版本提供了通用的修复策略。如需特定的修复规则，可以通过 Issue 提出需求。

## 贡献指南

我们欢迎各种形式的贡献！

### 如何贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 报告问题

如果发现 bug 或有功能建议，请通过 GitHub Issues 报告。

### 开发环境

- Go 1.18.0 或更高版本
- 推荐使用 Go Modules

## 更新日志

### v1.0.0
- 初始版本发布
- 支持基本的 JSON 修复功能
- 提供 Repair 和 Loads 两个主要 API
- 支持多种常见的 JSON 格式错误修复

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 致谢

感谢所有为这个项目做出贡献的开发者！

---

如果这个项目对你有帮助，请给我们一个 ⭐️！
