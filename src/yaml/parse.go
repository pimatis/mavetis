package yaml

import (
	"fmt"
	"strconv"
	"strings"
)

type line struct {
	indent int
	text   string
	number int
}

type Parser struct {
	lines []line
}

func Parse(input string) (any, error) {
	parser := Parser{}
	lines := make([]line, 0)
	raw := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	for index, value := range raw {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := 0
		for _, char := range value {
			if char != ' ' {
				break
			}
			indent++
		}
		if indent%2 != 0 {
			return nil, fmt.Errorf("line %d: indentation must use multiples of two spaces", index+1)
		}
		lines = append(lines, line{indent: indent, text: trimmed, number: index + 1})
	}
	parser.lines = lines
	if len(parser.lines) == 0 {
		return map[string]any{}, nil
	}
	value, next, err := parser.block(0, parser.lines[0].indent)
	if err != nil {
		return nil, err
	}
	if next != len(parser.lines) {
		return nil, fmt.Errorf("line %d: unexpected trailing content", parser.lines[next].number)
	}
	return value, nil
}

func (parser Parser) block(index int, indent int) (any, int, error) {
	if index >= len(parser.lines) {
		return map[string]any{}, index, nil
	}
	current := parser.lines[index]
	if current.indent < indent {
		return map[string]any{}, index, nil
	}
	if current.indent > indent {
		return nil, index, fmt.Errorf("line %d: invalid indentation", current.number)
	}
	if strings.HasPrefix(current.text, "- ") || current.text == "-" {
		return parser.list(index, indent)
	}
	return parser.mapv(index, indent)
}

func (parser Parser) list(index int, indent int) ([]any, int, error) {
	values := make([]any, 0)
	for index < len(parser.lines) {
		current := parser.lines[index]
		if current.indent < indent {
			return values, index, nil
		}
		if current.indent > indent {
			return nil, index, fmt.Errorf("line %d: invalid list indentation", current.number)
		}
		if !strings.HasPrefix(current.text, "- ") && current.text != "-" {
			return values, index, nil
		}
		item := strings.TrimSpace(strings.TrimPrefix(current.text, "-"))
		index++
		if item == "" {
			if index >= len(parser.lines) {
				values = append(values, map[string]any{})
				continue
			}
			next := parser.lines[index]
			if next.indent <= indent {
				values = append(values, map[string]any{})
				continue
			}
			value, nextIndex, err := parser.block(index, next.indent)
			if err != nil {
				return nil, nextIndex, err
			}
			values = append(values, value)
			index = nextIndex
			continue
		}
		if pair, ok := split(item); ok {
			entry := map[string]any{}
			key := pair[0]
			value := pair[1]
			if value != "" {
				entry[key] = scalar(value)
			}
			if value == "" {
				entry[key] = map[string]any{}
			}
			if index < len(parser.lines) {
				next := parser.lines[index]
				if next.indent > indent {
					nested, nextIndex, err := parser.block(index, next.indent)
					if err != nil {
						return nil, nextIndex, err
					}
					nestedMap, ok := nested.(map[string]any)
					if ok {
						nestedEntry, ok := entry[key].(map[string]any)
						if ok && len(nestedEntry) == 0 {
							entry[key] = nestedMap
						}
						for nestedKey, nestedValue := range nestedMap {
							entry[nestedKey] = nestedValue
						}
					}
					if !ok {
						entry[key] = nested
					}
					index = nextIndex
				}
			}
			values = append(values, entry)
			continue
		}
		values = append(values, scalar(item))
	}
	return values, index, nil
}

func (parser Parser) mapv(index int, indent int) (map[string]any, int, error) {
	values := map[string]any{}
	for index < len(parser.lines) {
		current := parser.lines[index]
		if current.indent < indent {
			return values, index, nil
		}
		if current.indent > indent {
			return nil, index, fmt.Errorf("line %d: invalid map indentation", current.number)
		}
		pair, ok := split(current.text)
		if !ok {
			return nil, index, fmt.Errorf("line %d: invalid mapping entry", current.number)
		}
		key := pair[0]
		value := pair[1]
		index++
		if value != "" {
			values[key] = scalar(value)
			continue
		}
		if index >= len(parser.lines) {
			values[key] = map[string]any{}
			continue
		}
		next := parser.lines[index]
		if next.indent <= indent {
			values[key] = map[string]any{}
			continue
		}
		nested, nextIndex, err := parser.block(index, next.indent)
		if err != nil {
			return nil, nextIndex, err
		}
		values[key] = nested
		index = nextIndex
	}
	return values, index, nil
}

func split(value string) ([2]string, bool) {
	pair := [2]string{}
	index := strings.Index(value, ":")
	if index <= 0 {
		return pair, false
	}
	pair[0] = strings.TrimSpace(value[:index])
	pair[1] = strings.TrimSpace(value[index+1:])
	if pair[0] == "" {
		return pair, false
	}
	return pair, true
}

func scalar(value string) any {
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
		decoded, err := strconv.Unquote(value)
		if err == nil {
			return decoded
		}
	}
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") && len(value) >= 2 {
		return value[1 : len(value)-1]
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		inner := strings.TrimSpace(value[1 : len(value)-1])
		if inner == "" {
			return []any{}
		}
		parts := strings.Split(inner, ",")
		items := make([]any, 0, len(parts))
		for _, part := range parts {
			items = append(items, scalar(strings.TrimSpace(part)))
		}
		return items
	}
	lower := strings.ToLower(value)
	if lower == "true" {
		return true
	}
	if lower == "false" {
		return false
	}
	if lower == "null" {
		return nil
	}
	integer, err := strconv.Atoi(value)
	if err == nil {
		return integer
	}
	floating, err := strconv.ParseFloat(value, 64)
	if err == nil {
		return floating
	}
	return value
}
