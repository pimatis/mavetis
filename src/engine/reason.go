package engine

import (
	"fmt"

	"github.com/Pimatis/mavetis/src/model"
)

func explain(item compiled, path string, line model.DiffLine, hunkText string) []string {
	reasons := make([]string, 0)
	if len(item.rule.Paths) != 0 {
		reasons = append(reasons, fmt.Sprintf("path matched scoped rule execution for %s", path))
	}
	if len(item.require) != 0 {
		reasons = append(reasons, fmt.Sprintf("matched %d required pattern checks on the diff line", len(item.require)))
	}
	if len(item.any) != 0 {
		reasons = append(reasons, fmt.Sprintf("matched at least one of %d alternative pattern checks", len(item.any)))
	}
	if len(item.near) != 0 {
		reasons = append(reasons, fmt.Sprintf("matched %d nearby context checks inside the same hunk", len(item.near)))
	}
	if len(item.absent) != 0 {
		reasons = append(reasons, fmt.Sprintf("no mitigation pattern from %d suppression checks was found nearby", len(item.absent)))
	}
	if item.rule.Entropy > 0 {
		reasons = append(reasons, fmt.Sprintf("entropy threshold %.2f was satisfied by the diff line", item.rule.Entropy))
	}
	if hunkText != "" && len(reasons) == 0 {
		reasons = append(reasons, "the hunk satisfied the rule conditions")
	}
	return reasons
}

func confidence(item compiled, value string, matchedContext bool) string {
	rank := confidenceRank(item.rule.Confidence)
	if matchedContext {
		rank++
	}
	if len(item.rule.Paths) != 0 {
		rank++
	}
	if item.rule.Entropy > 0 && entropy(value) >= item.rule.Entropy+0.4 {
		rank++
	}
	if rank > 3 {
		rank = 3
	}
	return confidenceValue(rank)
}

func confidenceRank(value string) int {
	if value == "high" {
		return 3
	}
	if value == "medium" {
		return 2
	}
	return 1
}

func confidenceValue(rank int) string {
	if rank >= 3 {
		return "high"
	}
	if rank == 2 {
		return "medium"
	}
	return "low"
}
