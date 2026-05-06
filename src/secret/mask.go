package secret

import "math"

func mask(value string) string {
	if len(value) <= 8 {
		return "********"
	}
	return value[:4] + "****" + value[len(value)-4:]
}

func entropy(value string) float64 {
	counts := map[rune]float64{}
	total := 0.0
	for _, char := range value {
		counts[char]++
		total++
	}
	if total == 0 {
		return 0
	}
	score := 0.0
	for _, count := range counts {
		probability := count / total
		score -= probability * math.Log2(probability)
	}
	return score
}
