package analyze

import "strings"

var popular = []string{
	"react",
	"express",
	"lodash",
	"axios",
	"requests",
	"flask",
	"django",
	"jsonwebtoken",
	"chalk",
	"typescript",
}

func SuspiciousPackage(name string) string {
	clean := normalizePackage(name)
	if clean == "" {
		return ""
	}
	for _, item := range popular {
		if clean == item {
			return ""
		}
		if distance(clean, item) == 1 {
			return item
		}
	}
	return ""
}

func normalizePackage(name string) string {
	value := strings.TrimSpace(strings.ToLower(name))
	value = strings.Trim(value, `"'`)
	if strings.HasPrefix(value, "@") {
		parts := strings.Split(value, "/")
		if len(parts) == 2 {
			value = parts[1]
		}
	}
	if index := strings.IndexAny(value, " =:<>@"); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

func distance(left string, right string) int {
	if left == right {
		return 0
	}
	if left == "" {
		return len(right)
	}
	if right == "" {
		return len(left)
	}
	prev := make([]int, len(right)+1)
	for index := range prev {
		prev[index] = index
	}
	for i := 1; i <= len(left); i++ {
		curr := make([]int, len(right)+1)
		curr[0] = i
		for j := 1; j <= len(right); j++ {
			cost := 0
			if left[i-1] != right[j-1] {
				cost = 1
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = curr
	}
	return prev[len(right)]
}

func min(values ...int) int {
	best := values[0]
	for _, item := range values[1:] {
		if item < best {
			best = item
		}
	}
	return best
}
