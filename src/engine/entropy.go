package engine

import "math"

func entropy(value string) float64 {
	counts := map[rune]float64{}
	total := 0.0
	for _, char := range value {
		if char == ' ' || char == '\t' {
			continue
		}
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
