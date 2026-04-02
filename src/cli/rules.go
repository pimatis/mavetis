package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/engine"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/output"
)

func runRules(arguments []string) int {
	if len(arguments) == 0 {
		return fail(errors.New("rules command requires validate, list, show, test, or matrix"))
	}
	if arguments[0] == "validate" {
		return validateRules(arguments[1:])
	}
	if arguments[0] == "list" {
		return listRules(arguments[1:])
	}
	if arguments[0] == "show" {
		return showRule(arguments[1:])
	}
	if arguments[0] == "test" {
		return testRules(arguments[1:])
	}
	if arguments[0] == "matrix" {
		return matrixRules(arguments[1:])
	}
	return fail(errors.New("rules command requires validate, list, show, test, or matrix"))
}

func validateRules(arguments []string) int {
	flags := flag.NewFlagSet("rules validate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	path := flags.String("rules", "", "Path to custom rules file")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *path == "" {
		return fail(errors.New("rules validate requires --rules"))
	}
	rules, err := config.LoadRules(*path)
	if err != nil {
		return fail(err)
	}
	fmt.Printf("validated %d custom rules\n", len(rules))
	return 0
}

func listRules(arguments []string) int {
	rules, err := loadRules(arguments)
	if err != nil {
		return fail(err)
	}
	sort.Slice(rules, func(left int, right int) bool {
		return rules[left].ID < rules[right].ID
	})
	for _, item := range rules {
		fmt.Printf("%s\t%s\t%s\t%s\n", item.ID, item.Category, item.Severity, item.Title)
	}
	return 0
}

func showRule(arguments []string) int {
	flags := flag.NewFlagSet("rules show", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	path := flags.String("rules", "", "Path to custom rules file")
	id := flags.String("id", "", "Rule identifier")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *id == "" {
		return fail(errors.New("rules show requires --id"))
	}
	rules, err := allRules(*path)
	if err != nil {
		return fail(err)
	}
	for _, item := range rules {
		if item.ID != *id {
			continue
		}
		fmt.Printf("id: %s\n", item.ID)
		fmt.Printf("title: %s\n", item.Title)
		fmt.Printf("category: %s\n", item.Category)
		fmt.Printf("severity: %s\n", item.Severity)
		fmt.Printf("confidence: %s\n", item.Confidence)
		fmt.Printf("target: %s\n", item.Target)
		fmt.Printf("require: %v\n", item.Require)
		fmt.Printf("any: %v\n", item.Any)
		fmt.Printf("near: %v\n", item.Near)
		fmt.Printf("absent: %v\n", item.Absent)
		fmt.Printf("paths: %v\n", item.Paths)
		fmt.Printf("standards: %v\n", item.Standards)
		return 0
	}
	return fail(fmt.Errorf("rule not found: %s", *id))
}

func testRules(arguments []string) int {
	flags := flag.NewFlagSet("rules test", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	rulesPath := flags.String("rules", "", "Path to custom rules file")
	diffPath := flags.String("diff", "", "Path to unified diff file")
	format := flags.String("format", "text", "Output format")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *diffPath == "" {
		return fail(errors.New("rules test requires --diff"))
	}
	content, err := os.ReadFile(*diffPath)
	if err != nil {
		return fail(err)
	}
	parsed, err := diff.Parse(string(content), model.DiffMeta{Mode: "test"})
	if err != nil {
		return fail(err)
	}
	rules, err := allRules(*rulesPath)
	if err != nil {
		return fail(err)
	}
	report, err := engine.Review(parsed, model.Config{Severity: "low", Output: *format}, rules)
	if err != nil {
		return fail(err)
	}
	if *format == "json" {
		body, err := output.JSON(report)
		if err != nil {
			return fail(err)
		}
		fmt.Println(body)
		return 0
	}
	if *format == "sarif" {
		body, err := output.SARIF(report)
		if err != nil {
			return fail(err)
		}
		fmt.Println(body)
		return 0
	}
	fmt.Print(output.TextExplain(report, true))
	return 0
}

func matrixRules(arguments []string) int {
	rules, err := loadRules(arguments)
	if err != nil {
		return fail(err)
	}
	rows := engine.Matrix(engine.MatrixInfos(rules))
	for _, row := range rows {
		fmt.Printf("%s\t%s\n", row.Control, row.Rules)
	}
	return 0
}

func loadRules(arguments []string) ([]model.Rule, error) {
	flags := flag.NewFlagSet("rules list", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	path := flags.String("rules", "", "Path to custom rules file")
	if err := flags.Parse(arguments); err != nil {
		return nil, err
	}
	return allRules(*path)
}

func allRules(path string) ([]model.Rule, error) {
	return loadAllRules(model.Config{}, path)
}
