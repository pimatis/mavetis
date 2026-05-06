package risk

import (
	"fmt"
	"math"

	"github.com/Pimatis/mavetis/src/model"
)

const (
	criticalWeight = 10.0
	highWeight     = 7.0
	mediumWeight   = 4.0
	lowWeight      = 1.0
)

type Score struct {
	Value  float64 `json:"value"`
	Rating string  `json:"rating"`
}

func Calculate(summary model.Summary) Score {
	total := float64(summary.Critical)*criticalWeight +
		float64(summary.High)*highWeight +
		float64(summary.Medium)*mediumWeight +
		float64(summary.Low)*lowWeight

	denominator := summary.Files
	if denominator < 1 {
		denominator = 1
	}

	value := total / float64(denominator)
	return Score{Value: round(value), Rating: rating(value)}
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

func rating(value float64) string {
	if value == 0 {
		return "none"
	}
	if value <= 1.5 {
		return "low"
	}
	if value <= 3.5 {
		return "medium"
	}
	if value <= 6.0 {
		return "high"
	}
	return "critical"
}

func Format(score Score) string {
	return fmt.Sprintf("%.2f (%s)", score.Value, score.Rating)
}
