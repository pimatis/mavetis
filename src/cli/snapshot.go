package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
)

func snapshotRules(arguments []string) int {
	flags := flag.NewFlagSet("rules snapshot", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	output := flags.String("output", "", "Output path for generated snapshots")
	path := flags.String("path", "", "Optional path glob to limit snapshot generation")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *output == "" {
		return fail(errors.New("rules snapshot requires --output"))
	}
	root, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	files, err := analyze.ScanSecurityAnchors(root, *path)
	if err != nil {
		return fail(err)
	}
	content := renderSnapshots(files)
	if err := os.WriteFile(*output, []byte(content), 0o600); err != nil {
		return fail(err)
	}
	count := strings.Count(content, "- id:")
	fmt.Printf("wrote %d snapshots to %s\n", count, filepath.Clean(*output))
	return 0
}

func renderSnapshots(files []analyze.AnchorFile) string {
	builder := strings.Builder{}
	builder.WriteString("snapshots:\n")
	sort.Slice(files, func(left int, right int) bool {
		return files[left].Path < files[right].Path
	})
	for _, file := range files {
		for _, anchor := range file.Anchors {
			builder.WriteString("  - id: ")
			builder.WriteString(snapshotID(file.Path, anchor.Name))
			builder.WriteString("\n")
			builder.WriteString("    path: ")
			builder.WriteString(file.Path)
			builder.WriteString("\n")
			builder.WriteString("    anchor: ")
			builder.WriteString(anchor.Name)
			builder.WriteString("\n")
			builder.WriteString("    category: ")
			builder.WriteString(anchor.Category)
			builder.WriteString("\n")
			builder.WriteString("    severity: ")
			builder.WriteString(anchor.Severity)
			builder.WriteString("\n")
			builder.WriteString("    message: Repository security snapshot regressed for ")
			builder.WriteString(anchor.Name)
			builder.WriteString(".\n")
			builder.WriteString("    remediation: Restore the required security behavior or refresh the snapshot only after review.\n")
			builder.WriteString("    require:\n")
			for _, token := range anchor.Expected {
				builder.WriteString("      - ")
				builder.WriteString(token)
				builder.WriteString("\n")
			}
			builder.WriteString("    standards:\n")
			builder.WriteString("      - OWASP-ASVS\n")
		}
	}
	return builder.String()
}

func snapshotID(path string, anchor string) string {
	value := strings.ToLower(filepath.ToSlash(path) + "." + anchor)
	value = strings.ReplaceAll(value, "/", ".")
	value = strings.ReplaceAll(value, "-", ".")
	value = strings.ReplaceAll(value, "_", ".")
	return value
}
