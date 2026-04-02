package engine

import (
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

type MatrixRow struct {
	Control string   `json:"control"`
	Rules   []string `json:"rules"`
}

func Matrix(rows []model.RuleInfo) []MatrixRow {
	grouped := map[string][]string{}
	for _, row := range rows {
		for _, standard := range row.Standards {
			if !strings.HasPrefix(standard, "OWASP-ASVS-") {
				continue
			}
			grouped[standard] = append(grouped[standard], row.ID)
		}
	}
	result := make([]MatrixRow, 0, len(grouped))
	for control, rules := range grouped {
		sort.Strings(rules)
		result = append(result, MatrixRow{Control: control, Rules: rules})
	}
	sort.Slice(result, func(left int, right int) bool {
		return result[left].Control < result[right].Control
	})
	return result
}
