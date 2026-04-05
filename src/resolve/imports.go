package resolve

import (
	"bufio"
	"go/parser"
	"go/token"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
)

type ImportRef struct {
	Module string
	Kind   string
	Line   int
}

type indexedImport struct {
	ImportRef
	start int
}

var jsImportFromPattern = regexp.MustCompile(`(?s)import\s+.*?\s+from\s+['"]([^'"]+)['"]`)
var jsRequirePattern = regexp.MustCompile(`require\(\s*['"]([^'"]+)['"]\s*\)`)
var jsExportFromPattern = regexp.MustCompile(`(?s)export\s+.*?\s+from\s+['"]([^'"]+)['"]`)
var jsDynamicImportPattern = regexp.MustCompile(`import\(\s*['"]([^'"]+)['"]\s*\)`)
var pyFromPattern = regexp.MustCompile(`^\s*from\s+([.A-Za-z_][A-Za-z0-9_.]*)\s+import\s+(.+)$`)
var pyImportPattern = regexp.MustCompile(`^\s*import\s+(.+)$`)
var jvmImportPattern = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([A-Za-z_][A-Za-z0-9_.]*)\s*;`)

func Imports(path string, content string) []ImportRef {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	language := analyze.Language(path)
	if language == "go" {
		return capImports(goImports(path, content))
	}
	if language == "javascript" || language == "typescript" {
		return capImports(jsImports(content))
	}
	if language == "python" {
		return capImports(pyImports(content))
	}
	if language == "jvm" {
		return capImports(jvmImports(content))
	}
	return nil
}

func goImports(path string, content string) []ImportRef {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ImportsOnly)
	if err != nil {
		return nil
	}
	refs := make([]ImportRef, 0, len(file.Imports))
	for _, spec := range file.Imports {
		module := strings.Trim(spec.Path.Value, `"`)
		if module == "" {
			continue
		}
		refs = append(refs, ImportRef{Module: module, Kind: "import", Line: fset.Position(spec.Pos()).Line})
	}
	return refs
}

func jsImports(content string) []ImportRef {
	matches := make([]indexedImport, 0)
	matches = append(matches, collectJS(content, jsImportFromPattern, "import")...)
	matches = append(matches, collectJS(content, jsRequirePattern, "require")...)
	matches = append(matches, collectJS(content, jsExportFromPattern, "export-from")...)
	matches = append(matches, collectJS(content, jsDynamicImportPattern, "dynamic-import")...)
	sort.Slice(matches, func(left int, right int) bool {
		if matches[left].start != matches[right].start {
			return matches[left].start < matches[right].start
		}
		if matches[left].Line != matches[right].Line {
			return matches[left].Line < matches[right].Line
		}
		return matches[left].Module < matches[right].Module
	})
	refs := make([]ImportRef, 0, len(matches))
	for _, match := range matches {
		refs = append(refs, match.ImportRef)
	}
	return refs
}

func collectJS(content string, pattern *regexp.Regexp, kind string) []indexedImport {
	indices := pattern.FindAllStringSubmatchIndex(content, -1)
	matches := make([]indexedImport, 0, len(indices))
	for _, index := range indices {
		if len(index) < 4 {
			continue
		}
		start := index[2]
		end := index[3]
		if start < 0 || end < 0 {
			continue
		}
		module := strings.TrimSpace(content[start:end])
		if module == "" {
			continue
		}
		matches = append(matches, indexedImport{ImportRef: ImportRef{Module: module, Kind: kind, Line: lineNumber(content, index[0])}, start: index[0]})
	}
	return matches
}

func pyImports(content string) []ImportRef {
	scanner := bufio.NewScanner(strings.NewReader(content))
	refs := make([]ImportRef, 0)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		if match := pyFromPattern.FindStringSubmatch(text); len(match) == 3 {
			module := normalizePyModule(match[1], match[2])
			if module != "" {
				refs = append(refs, ImportRef{Module: module, Kind: "from-import", Line: line})
			}
			continue
		}
		if match := pyImportPattern.FindStringSubmatch(text); len(match) == 2 {
			refs = append(refs, splitPyImports(match[1], line)...)
		}
	}
	return refs
}

func normalizePyModule(module string, names string) string {
	trimmed := strings.TrimSpace(module)
	if trimmed == "" {
		return ""
	}
	if strings.Trim(trimmed, ".") != "" {
		return strings.TrimSuffix(trimmed, ";")
	}
	name := firstPyName(names)
	if name == "" {
		return ""
	}
	return trimmed + name
}

func firstPyName(value string) string {
	cleaned := strings.TrimSpace(strings.Split(value, "#")[0])
	cleaned = strings.Trim(cleaned, "()")
	parts := strings.Split(cleaned, ",")
	if len(parts) == 0 {
		return ""
	}
	name := strings.TrimSpace(parts[0])
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return ""
	}
	return strings.TrimSpace(fields[0])
}

func splitPyImports(value string, line int) []ImportRef {
	cleaned := strings.TrimSpace(strings.Split(value, "#")[0])
	parts := strings.Split(cleaned, ",")
	refs := make([]ImportRef, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		refs = append(refs, ImportRef{Module: fields[0], Kind: "import", Line: line})
	}
	return refs
}

func jvmImports(content string) []ImportRef {
	scanner := bufio.NewScanner(strings.NewReader(content))
	refs := make([]ImportRef, 0)
	line := 0
	for scanner.Scan() {
		line++
		match := jvmImportPattern.FindStringSubmatch(scanner.Text())
		if len(match) != 2 {
			continue
		}
		refs = append(refs, ImportRef{Module: match[1], Kind: "import", Line: line})
	}
	return refs
}

func lineNumber(content string, offset int) int {
	if offset <= 0 {
		return 1
	}
	if offset > len(content) {
		offset = len(content)
	}
	return strings.Count(content[:offset], "\n") + 1
}

func capImports(refs []ImportRef) []ImportRef {
	if len(refs) == 0 {
		return nil
	}
	result := make([]ImportRef, 0, len(refs))
	seen := map[string]struct{}{}
	for _, ref := range refs {
		key := ref.Kind + "|" + ref.Module + "|" + strconv.Itoa(ref.Line)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, ref)
		if len(result) == 64 {
			return result
		}
	}
	return result
}
