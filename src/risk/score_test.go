package risk

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestCalculateZeroFindings(t *testing.T) {
	score := Calculate(model.Summary{Files: 10})
	if score.Value != 0 {
		t.Fatalf("expected zero value, got %f", score.Value)
	}
	if score.Rating != "none" {
		t.Fatalf("expected none rating, got %s", score.Rating)
	}
}

func TestCalculateLowRating(t *testing.T) {
	score := Calculate(model.Summary{Files: 10, Low: 10})
	if score.Value != 1.0 {
		t.Fatalf("expected 1.0, got %f", score.Value)
	}
	if score.Rating != "low" {
		t.Fatalf("expected low, got %s", score.Rating)
	}
}

func TestCalculateMediumRating(t *testing.T) {
	score := Calculate(model.Summary{Files: 10, Medium: 5, Low: 5})
	want := (5.0*4.0 + 5.0*1.0) / 10.0
	if score.Value != want {
		t.Fatalf("expected %f, got %f", want, score.Value)
	}
	if score.Rating != "medium" {
		t.Fatalf("expected medium, got %s", score.Rating)
	}
}

func TestCalculateHighRating(t *testing.T) {
	score := Calculate(model.Summary{Files: 10, High: 5, Medium: 5})
	want := (5.0*7.0 + 5.0*4.0) / 10.0
	if score.Value != want {
		t.Fatalf("expected %f, got %f", want, score.Value)
	}
	if score.Rating != "high" {
		t.Fatalf("expected high, got %s", score.Rating)
	}
}

func TestCalculateCriticalRating(t *testing.T) {
	score := Calculate(model.Summary{Files: 10, Critical: 10})
	if score.Value != 10.0 {
		t.Fatalf("expected 10.0, got %f", score.Value)
	}
	if score.Rating != "critical" {
		t.Fatalf("expected critical, got %s", score.Rating)
	}
}

func TestCalculateUsesFilesAsDenominator(t *testing.T) {
	score := Calculate(model.Summary{Files: 5, High: 5})
	want := (5.0 * 7.0) / 5.0
	if score.Value != want {
		t.Fatalf("expected %f, got %f", want, score.Value)
	}
}

func TestCalculateFallsBackToOneWhenNoFiles(t *testing.T) {
	score := Calculate(model.Summary{Critical: 2})
	if score.Value != 20.0 {
		t.Fatalf("expected 20.0, got %f", score.Value)
	}
}

func TestFormatRendersValueAndRating(t *testing.T) {
	formatted := Format(Score{Value: 3.45, Rating: "medium"})
	if formatted != "3.45 (medium)" {
		t.Fatalf("unexpected format: %s", formatted)
	}
}
