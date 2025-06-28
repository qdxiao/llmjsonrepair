package pkg

import (
	"log"
	"strconv"
	"strings"
	"unicode"
)

const (
	inArray contextValue = iota
	inObjectKey
	inObjectValue
)

type Option func(p *parser)

// parser 是核心的 JSON 解析和修复结构体
type parser struct {
	jsonStr []rune
	index   int
	context *jsonContext
	logger  Logger
}

// NewParser 创建一个新的解析器实例
func NewParser(jsonStr string, opts ...Option) *parser {
	p := &parser{
		logger:  &log.Logger{},
		jsonStr: []rune(jsonStr),
		index:   0,
		context: &jsonContext{},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// WithLogger 设置日志
func WithLogger(l Logger) Option {
	return func(p *parser) {
		p.logger = l
	}
}

// getChar 安全地获取当前索引或偏移处的字符
func (p *parser) getChar(offset int) (rune, bool) {
	if p.index+offset >= len(p.jsonStr) || p.index+offset < 0 {
		return 0, false
	}
	return p.jsonStr[p.index+offset], true
}

// skipWhitespace 跳过所有空白字符
func (p *parser) skipWhitespace() {
	for {
		char, ok := p.getChar(0)
		if !ok || !unicode.IsSpace(char) {
			break
		}
		p.index++
	}
}

// Parse 解析器的启动方法
func (p *parser) Parse() (interface{}, error) {
	json, err := p.parseJSON()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if p.index < len(p.jsonStr) {
		// 如果解析提前结束但仍有内容，则尝试将它们作为多JSON对象处理
		results := []interface{}{json}
		for p.index < len(p.jsonStr) {
			p.skipWhitespace()
			if p.index == len(p.jsonStr) {
				break
			}
			nextJSON, err := p.parseJSON()
			if err == nil && nextJSON != nil {
				results = append(results, nextJSON)
			} else {
				p.index++
			}
		}
		if len(results) == 1 {
			return results[0], nil
		}
		return results, nil
	}

	return json, nil
}

// parseJSON 根据当前字符决定调用哪个具体的解析函数
func (p *parser) parseJSON() (interface{}, error) {
	p.skipWhitespace()
	char, ok := p.getChar(0)
	if !ok {
		return nil, nil
	}

	switch {
	case char == '{':
		p.index++
		return p.parseObject()
	case char == '[':
		p.index++
		return p.parseArray()
	case char == '"' || char == '\'':
		return p.parseString()
	case unicode.IsDigit(char) || char == '-':
		return p.parseNumber()
	case char == 't' || char == 'f' || char == 'n':
		return p.parseBooleanOrNull()
	case unicode.IsLetter(char):
		ctx, inCtx := p.context.current()
		if inCtx && (ctx == inObjectValue || ctx == inArray || ctx == inObjectKey) {
			return p.parseString()
		}
	}
	// 如果所有情况都不匹配，则前进一个字符并重试，以跳过垃圾字符
	p.index++
	return p.parseJSON()
}

// parseObject 解析一个JSON对象
func (p *parser) parseObject() (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	p.context.push(inObjectKey)
	defer p.context.pop()

	for {
		p.skipWhitespace()
		char, ok := p.getChar(0)
		if !ok || char == '}' {
			break
		}

		if char == ',' {
			p.index++
			continue
		}

		// 解析键
		p.context.stack[len(p.context.stack)-1] = inObjectKey
		key, err := p.parseString()
		if err != nil {
			// 如果键解析失败，可能是因为对象结束了
			p.skipWhitespace()
			if c, ok := p.getChar(0); ok && c == '}' {
				break
			}
			// 否则跳过一个字符继续尝试
			p.index++
			continue
		}

		p.skipWhitespace()
		if c, ok := p.getChar(0); ok && c == ':' {
			p.index++
		}

		// 解析值
		p.context.stack[len(p.context.stack)-1] = inObjectValue
		value, err := p.parseJSON()
		if err != nil {
			value = ""
		}
		obj[key] = value

		p.skipWhitespace()
		if c, ok := p.getChar(0); ok && c == ',' {
			p.index++
		} else if ok && c == '}' {
			// 找到结束符，可以中断循环
			break
		}
	}

	if char, ok := p.getChar(0); ok && char == '}' {
		p.index++
	}
	return obj, nil
}

// parseArray 解析一个JSON数组
func (p *parser) parseArray() ([]interface{}, error) {
	arr := make([]interface{}, 0)
	p.context.push(inArray)
	defer p.context.pop()

	for {
		p.skipWhitespace()
		char, ok := p.getChar(0)
		if !ok || char == ']' {
			break
		}

		if char == ',' { // 跳过多余的逗号
			p.index++
			continue
		}

		value, err := p.parseJSON()
		if err != nil {
			// 如果解析失败，可能是数组结束了
			p.skipWhitespace()
			if c, ok := p.getChar(0); ok && c == ']' {
				break
			}
			p.index++
			continue
		}
		arr = append(arr, value)

		p.skipWhitespace()
		if c, ok := p.getChar(0); ok && c == ',' {
			p.index++
		} else if ok && c == ']' {
			break
		}
	}

	if char, ok := p.getChar(0); ok && char == ']' {
		p.index++
	}
	return arr, nil
}

// parseString 解析一个JSON字符串
func (p *parser) parseString() (string, error) {
	p.skipWhitespace()
	var startQuote rune
	char, ok := p.getChar(0)
	if !ok {
		return "", nil
	}

	missingQuotes := false
	if char == '"' || char == '\'' {
		startQuote = char
		p.index++
	} else {
		missingQuotes = true
	}

	var sb strings.Builder
	for {
		char, ok := p.getChar(0)
		if !ok {
			break // 字符串未闭合
		}

		// 处理转义字符
		if char == '\\' {
			p.index++
			nextChar, nextOk := p.getChar(0)
			if !nextOk {
				break // 转义符在末尾
			}
			switch nextChar {
			case '"', '\\', '/', '\'':
				sb.WriteRune(nextChar)
			case 'b':
				sb.WriteRune('\b')
			case 'f':
				sb.WriteRune('\f')
			case 'n':
				sb.WriteRune('\n')
			case 'r':
				sb.WriteRune('\r')
			case 't':
				sb.WriteRune('\t')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(nextChar)
			}
			p.index++
			continue
		}

		// 检查字符串结束条件
		if !missingQuotes && char == startQuote {
			p.index++
			return sb.String(), nil
		}

		// 如果引号缺失，需要根据上下文决定何时结束
		if missingQuotes {
			ctx, inCtx := p.context.current()
			if inCtx {
				if ctx == inObjectKey && char == ':' {
					break
				}
				if (ctx == inObjectValue || ctx == inArray) && (char == ',' || char == '}' || char == ']') {
					break
				}
			} else if char == ',' || char == '}' || char == ']' || char == ':' {
				// 如果没有上下文，但遇到了分隔符，也认为字符串结束
				break
			}
		}

		sb.WriteRune(char)
		p.index++
	}

	// 对于未加引号的字符串，修剪尾部空格
	if missingQuotes {
		return strings.TrimRight(sb.String(), " \t\n\r"), nil
	}
	return sb.String(), nil
}

// parseNumber 解析一个数字
func (p *parser) parseNumber() (interface{}, error) {
	var sb strings.Builder
	for {
		char, ok := p.getChar(0)
		if !ok || (!unicode.IsDigit(char) && char != '.' && char != '-' && char != 'e' && char != 'E') {
			break
		}
		sb.WriteRune(char)
		p.index++
	}
	numStr := sb.String()
	if strings.Contains(numStr, ".") || strings.Contains(numStr, "e") || strings.Contains(numStr, "E") {
		f, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return numStr, nil // 如果转换失败，则作为字符串返回
		}
		return f, nil
	}
	i, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return numStr, nil // 如果转换失败，则作为字符串返回
	}
	return i, nil
}

// parseBooleanOrNull 解析 true, false, 或 null
func (p *parser) parseBooleanOrNull() (interface{}, error) {
	if strings.HasPrefix(string(p.jsonStr[p.index:]), "true") {
		p.index += 4
		return true, nil
	}
	if strings.HasPrefix(string(p.jsonStr[p.index:]), "false") {
		p.index += 5
		return false, nil
	}
	if strings.HasPrefix(string(p.jsonStr[p.index:]), "null") {
		p.index += 4
		return nil, nil
	}
	// 如果不是这些，则当作一个未加引号的字符串来解析
	return p.parseString()
}

// 定义解析器上下文中的状态
type contextValue int

// jsonContext 用于跟踪解析器的当前状态
type jsonContext struct {
	stack []contextValue
}

func (c *jsonContext) push(val contextValue) {
	c.stack = append(c.stack, val)
}

func (c *jsonContext) pop() {
	if len(c.stack) > 0 {
		c.stack = c.stack[:len(c.stack)-1]
	}
}

func (c *jsonContext) current() (contextValue, bool) {
	if len(c.stack) == 0 {
		return 0, false
	}
	return c.stack[len(c.stack)-1], true
}
