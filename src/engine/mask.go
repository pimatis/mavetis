package engine

import "regexp"

var token = regexp.MustCompile(`[A-Za-z0-9_\-./+=]{12,}`)

func mask(value string) string {
	return token.ReplaceAllStringFunc(value, func(item string) string {
		if len(item) <= 8 {
			return "********"
		}
		return item[:4] + "****" + item[len(item)-4:]
	})
}
