package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Pimatis/mavetis/src/baseline"
	"github.com/Pimatis/mavetis/src/scan"
)

func runBaseline(arguments []string) int {
	create := false
	output := ".mavetis-baseline.yaml"
	reviewArgs := make([]string, 0, len(arguments))
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		if arg == "--create" {
			create = true
			continue
		}
		if strings.HasPrefix(arg, "--output=") {
			output = strings.TrimPrefix(arg, "--output=")
			continue
		}
		if arg == "--output" && i+1 < len(arguments) {
			output = arguments[i+1]
			i++
			continue
		}
		reviewArgs = append(reviewArgs, arg)
	}
	if !create {
		return fail(fmt.Errorf("baseline requires --create"))
	}
	root, err := scan.Root()
	if err != nil {
		return fail(err)
	}
	outPath := output
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(root, outPath)
	}
	spec, cfg, rules, err := prepareReview(reviewArgs, false)
	if err != nil {
		return fail(err)
	}
	report, err := buildReport(spec, cfg, rules)
	if err != nil {
		return fail(err)
	}
	if err := baseline.Create(outPath, report); err != nil {
		return fail(err)
	}
	fmt.Printf("Created baseline with %d findings: %s\n", len(report.Findings), outPath)
	return 0
}
