package secret

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/cache"
	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/risk"
)

const (
	maxLineSize = 64 << 10
)

type Options struct {
	Targets []string
	Path    string
	Cache   string
	NoCache bool
}

type file struct {
	path    string
	real    string
	size    int64
	modTime int64
}

type allowance struct {
	values  []string
	regexes []*regexp.Regexp
}

func Scan(root string, config model.Config, options Options) (model.Report, error) {
	rootReal, err := realPath(root)
	if err != nil {
		return model.Report{}, err
	}
	targets := options.Targets
	if len(targets) == 0 {
		targets = []string{"."}
	}
	files, err := collect(rootReal, targets, options.Path)
	if err != nil {
		return model.Report{}, err
	}
	allowlist, err := compileAllow(config)
	if err != nil {
		return model.Report{}, err
	}
	cacheData := cache.Data{}
	cachePath := ""
	cacheEnabled := !options.NoCache
	if cacheEnabled {
		cachePath, cacheData, err = cache.Load(rootReal, "secrets", options.Cache, cacheKey(config))
		if err != nil {
			return model.Report{}, err
		}
		cache.Prune(cacheData, cacheFiles(files))
	}
	report := model.Report{Meta: model.DiffMeta{Mode: "secrets"}}
	report.Policy = &model.Policy{FailOn: config.FailOn}
	report.Rules = infos()
	report.Summary.Files = len(files)
	seen := map[string]struct{}{}
	findings := make([]model.Finding, 0)
	for _, item := range files {
		if match.Any(config.Ignore, item.path) {
			continue
		}
		if match.Any(config.Allow.Paths, item.path) {
			continue
		}
		fileFindings, cacheHit := cache.Findings(cacheData, cacheFile(item))
		if !cacheHit {
			var scanErr error
			fileFindings, scanErr = scanFile(item, allowlist)
			if scanErr != nil {
				return report, scanErr
			}
		}
		if cacheEnabled && !cacheHit {
			cache.Put(cacheData, cacheFile(item), fileFindings)
		}
		for _, finding := range fileFindings {
			if _, exists := seen[finding.ID]; exists {
				continue
			}
			seen[finding.ID] = struct{}{}
			if model.SeverityRank(finding.Severity) < model.SeverityRank(config.Severity) {
				continue
			}
			findings = append(findings, finding)
		}
	}
	sortFindings(findings)
	report.Findings = findings
	for _, finding := range findings {
		report.Summary.Add(finding)
	}
	if cacheEnabled {
		_ = cache.Save(cachePath, cacheData)
	}
	score := risk.Calculate(report.Summary)
	report.Score = &model.Score{Value: score.Value, Rating: score.Rating}
	return report, nil
}

func cacheFile(item file) cache.File {
	return cache.File{Path: item.path, Size: item.size, ModTime: item.modTime}
}

func cacheFiles(files []file) []cache.File {
	items := make([]cache.File, 0, len(files))
	for _, item := range files {
		items = append(items, cacheFile(item))
	}
	return items
}

func scanFile(item file, allowlist allowance) ([]model.Finding, error) {
	content, err := os.ReadFile(item.real)
	if err != nil {
		return nil, fmt.Errorf("read secrets scan target %q: %w", item.path, err)
	}
	if bytes.IndexByte(content, 0) >= 0 {
		return nil, nil
	}
	lines := strings.Split(string(content), "\n")
	findings := make([]model.Finding, 0)
	items := patterns()
	for index, line := range lines {
		if len(line) > maxLineSize {
			continue
		}
		lineMatchedSpecific := false
		for _, itemPattern := range items {
			if itemPattern.id == "secret.scan.generic" && lineMatchedSpecific {
				continue
			}
			finding, ok := evaluate(itemPattern, item.path, index+1, line, allowlist)
			if !ok {
				continue
			}
			if itemPattern.id != "secret.scan.generic" {
				lineMatchedSpecific = true
			}
			findings = append(findings, finding)
		}
	}
	return findings, nil
}

func evaluate(item pattern, path string, lineNumber int, line string, allowlist allowance) (model.Finding, bool) {
	if item.path != nil && !item.path.MatchString(path) {
		return model.Finding{}, false
	}
	matches := item.re.FindStringSubmatch(line)
	if len(matches) == 0 {
		return model.Finding{}, false
	}
	secretValue := matches[0]
	if item.group > 0 && item.group < len(matches) {
		secretValue = matches[item.group]
	}
	if allowed(allowlist, secretValue) || allowed(allowlist, line) {
		return model.Finding{}, false
	}
	if item.entropy > 0 && entropy(secretValue) < item.entropy {
		return model.Finding{}, false
	}
	snippet := strings.Replace(line, secretValue, mask(secretValue), 1)
	finding := model.Finding{
		RuleID:      item.id,
		Title:       item.title,
		Category:    "secret",
		Severity:    item.severity,
		Confidence:  item.confidence,
		Path:        path,
		Line:        lineNumber,
		Side:        "file",
		Message:     item.message,
		Snippet:     strings.TrimSpace(snippet),
		Remediation: item.remediation,
		Standards:   append([]string{}, item.standards...),
		Reasons:     []string{"matched a secret-specific pattern", fmt.Sprintf("candidate entropy %.2f", entropy(secretValue))},
	}
	finding.ID = identity(finding.RuleID, finding.Path, finding.Line, finding.Snippet)
	return finding, true
}

func infos() []model.RuleInfo {
	items := patterns()
	infos := make([]model.RuleInfo, 0, len(items))
	for _, item := range items {
		infos = append(infos, model.RuleInfo{ID: item.id, Title: item.title, Category: "secret", Severity: item.severity, Standards: item.standards})
	}
	return infos
}

func compileAllow(config model.Config) (allowance, error) {
	result := allowance{values: append([]string{}, config.Allow.Values...)}
	for _, pattern := range config.Allow.Regexes {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return result, fmt.Errorf("compile allow regex %q: %w", pattern, err)
		}
		result.regexes = append(result.regexes, re)
	}
	return result, nil
}

func allowed(allowlist allowance, value string) bool {
	for _, allowedValue := range allowlist.values {
		if allowedValue == "" {
			continue
		}
		if strings.Contains(value, allowedValue) {
			return true
		}
	}
	for _, re := range allowlist.regexes {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func identity(ruleID string, path string, line int, snippet string) string {
	sum := sha256.Sum256([]byte(ruleID + "|" + path + "|" + fmt.Sprintf("%d", line) + "|" + snippet))
	return hex.EncodeToString(sum[:8])
}

func sortFindings(findings []model.Finding) {
	sort.Slice(findings, func(left int, right int) bool {
		if model.SeverityRank(findings[left].Severity) != model.SeverityRank(findings[right].Severity) {
			return model.SeverityRank(findings[left].Severity) > model.SeverityRank(findings[right].Severity)
		}
		if findings[left].Path != findings[right].Path {
			return findings[left].Path < findings[right].Path
		}
		if findings[left].Line != findings[right].Line {
			return findings[left].Line < findings[right].Line
		}
		return findings[left].RuleID < findings[right].RuleID
	})
}
