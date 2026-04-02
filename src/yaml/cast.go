package yaml

import "fmt"

func Map(value any) (map[string]any, error) {
	mapped, ok := value.(map[string]any)
	if ok {
		return mapped, nil
	}
	return nil, fmt.Errorf("expected map")
}

func List(value any) ([]any, error) {
	listed, ok := value.([]any)
	if ok {
		return listed, nil
	}
	return nil, fmt.Errorf("expected list")
}

func String(value any) (string, bool) {
	text, ok := value.(string)
	return text, ok
}

func Strings(value any) []string {
	listed, ok := value.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(listed))
	for _, item := range listed {
		text, ok := item.(string)
		if !ok {
			continue
		}
		values = append(values, text)
	}
	return values
}

func Bool(value any) (bool, bool) {
	flag, ok := value.(bool)
	return flag, ok
}

func Float(value any) (float64, bool) {
	floating, ok := value.(float64)
	if ok {
		return floating, true
	}
	integer, ok := value.(int)
	if ok {
		return float64(integer), true
	}
	return 0, false
}
