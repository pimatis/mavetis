package analyze

import (
	"regexp"
	"sort"
	"strings"
)

var word = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

func Tokens(value string) []string {
	matches := word.FindAllString(value, -1)
	if len(matches) == 0 {
		return nil
	}
	set := map[string]struct{}{}
	for _, item := range matches {
		lower := strings.ToLower(item)
		if len(lower) < 2 {
			continue
		}
		set[lower] = struct{}{}
	}
	values := make([]string, 0, len(set))
	for item := range set {
		values = append(values, item)
	}
	sort.Strings(values)
	return values
}

func HasToken(tokens []string, want string) bool {
	for _, item := range tokens {
		if item == strings.ToLower(want) {
			return true
		}
	}
	return false
}

func SharesAny(tokens []string, wants []string) bool {
	for _, want := range wants {
		if HasToken(tokens, want) {
			return true
		}
	}
	return false
}
