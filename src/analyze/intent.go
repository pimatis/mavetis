package analyze

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Intent struct {
	Name     string
	Category string
	Severity string
	Expected []string
	Bypass   []string
}

type Anchor struct {
	Name     string
	Line     int
	Category string
	Severity string
	Expected []string
	Bypass   []string
}

var functionName = regexp.MustCompile(`(?i)(?:func\s+([A-Za-z_][A-Za-z0-9_]*)|function\s+([A-Za-z_][A-Za-z0-9_]*)|const\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*\(|let\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*\(|var\s+([A-Za-z_][A-Za-z0-9_]*)\s*=\s*\()`)

func SecurityIntent(name string) (Intent, bool) {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "ownership") || strings.Contains(lower, "authorize") || strings.Contains(lower, "permission") {
		return Intent{Name: name, Category: "authorization", Severity: "critical", Expected: []string{"owner", "tenant", "org", "permission", "policy", "query", "find", "verify"}, Bypass: []string{"skip", "allow", "permitall", "return", "true"}}, true
	}
	if strings.Contains(lower, "sanitize") || strings.Contains(lower, "escape") {
		return Intent{Name: name, Category: "input", Severity: "high", Expected: []string{"sanitize", "escape", "strip", "encode", "allowlist", "validate"}, Bypass: []string{"noop", "identity", "raw", "unsafe", "return"}}, true
	}
	if strings.Contains(lower, "mfa") || strings.Contains(lower, "otp") || strings.Contains(lower, "totp") || strings.Contains(lower, "webauthn") {
		return Intent{Name: name, Category: "auth", Severity: "critical", Expected: []string{"verify", "challenge", "totp", "otp", "webauthn", "mfa"}, Bypass: []string{"skip", "disable", "optional", "allow", "return"}}, true
	}
	if strings.Contains(lower, "token") || strings.Contains(lower, "jwt") || strings.Contains(lower, "validate") || strings.Contains(lower, "verify") {
		return Intent{Name: name, Category: "token", Severity: "high", Expected: []string{"verify", "signature", "issuer", "audience", "claims", "expiry", "exp", "nbf"}, Bypass: []string{"decode", "skip", "allowunsigned", "acceptinvalid", "return"}}, true
	}
	return Intent{}, false
}

func SecurityAnchorsFromText(text string) []Anchor {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	anchors := make([]Anchor, 0)
	for index, line := range lines {
		name := declaredName(line)
		if name == "" {
			continue
		}
		intent, ok := SecurityIntent(name)
		if !ok {
			continue
		}
		anchors = append(anchors, Anchor{Name: intent.Name, Line: index + 1, Category: intent.Category, Severity: intent.Severity, Expected: append([]string{}, intent.Expected...), Bypass: append([]string{}, intent.Bypass...)})
	}
	return anchors
}

func ScanSecurityAnchors(root string, glob string) ([]AnchorFile, error) {
	files := make([]AnchorFile, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if glob != "" && !matchGlob(glob, rel) {
			return nil
		}
		if !Executable(rel) || Fixture(rel) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		anchors := SecurityAnchorsFromText(string(content))
		if len(anchors) == 0 {
			return nil
		}
		files = append(files, AnchorFile{Path: rel, Anchors: anchors})
		return nil
	})
	return files, err
}

type AnchorFile struct {
	Path    string
	Anchors []Anchor
}

func declaredName(line string) string {
	match := functionName.FindStringSubmatch(line)
	for index := 1; index < len(match); index++ {
		if strings.TrimSpace(match[index]) != "" {
			return strings.TrimSpace(match[index])
		}
	}
	return ""
}

func SecurityWindowTokens(path string, line int, radius int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	index := 0
	values := make([]string, 0)
	for scanner.Scan() {
		index++
		if index < line-radius {
			continue
		}
		if index > line+radius {
			break
		}
		values = append(values, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return Tokens(strings.Join(values, "\n")), nil
}

func matchGlob(pattern string, value string) bool {
	ok, err := filepath.Match(pattern, value)
	if err == nil && ok {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "**")
		return strings.HasPrefix(value, strings.TrimSuffix(prefix, "/"))
	}
	return false
}
