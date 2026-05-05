package scan

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func LoadExistingFiles(root string, targets []string) ([]ScannedFile, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	if len(targets) > maxExplicitFiles {
		return nil, fmt.Errorf("too many changed context targets: max %d", maxExplicitFiles)
	}
	rootReal, err := realPath(root)
	if err != nil {
		return nil, err
	}
	cleaned := uniqueTargets(targets)
	files := make([]ScannedFile, 0, len(cleaned))
	seen := map[string]struct{}{}
	for _, target := range cleaned {
		file, real, ok, err := loadExistingFile(rootReal, target)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if _, ok := seen[real]; ok {
			continue
		}
		seen[real] = struct{}{}
		files = append(files, file)
	}
	return files, nil
}

func uniqueTargets(targets []string) []string {
	seen := map[string]struct{}{}
	cleaned := make([]string, 0, len(targets))
	for _, target := range targets {
		value := strings.TrimSpace(target)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		cleaned = append(cleaned, value)
	}
	sort.Strings(cleaned)
	return cleaned
}

func loadExistingFile(root string, target string) (ScannedFile, string, bool, error) {
	if strings.ContainsRune(target, rune(0)) {
		return ScannedFile{}, "", false, fmt.Errorf("changed context target contains NUL byte: %q", target)
	}
	if filepath.IsAbs(target) {
		return ScannedFile{}, "", false, fmt.Errorf("absolute changed context targets are not supported: %s", target)
	}
	cleaned := filepath.Clean(target)
	if outsideRoot(cleaned) {
		return ScannedFile{}, "", false, fmt.Errorf("changed context target escapes repository root: %s", target)
	}
	absolute := filepath.Join(root, cleaned)
	real, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		if isMissing(err) {
			return ScannedFile{}, "", false, nil
		}
		return ScannedFile{}, "", false, fmt.Errorf("resolve changed context target %q: %w", target, err)
	}
	info, err := os.Stat(real)
	if err != nil {
		if isMissing(err) {
			return ScannedFile{}, "", false, nil
		}
		return ScannedFile{}, "", false, fmt.Errorf("stat changed context target %q: %w", target, err)
	}
	if !info.Mode().IsRegular() {
		return ScannedFile{}, "", false, nil
	}
	if info.Size() > maxExplicitFileSize {
		return ScannedFile{}, "", false, nil
	}
	rel, err := filepath.Rel(root, real)
	if err != nil {
		return ScannedFile{}, "", false, fmt.Errorf("resolve changed context target %q: %w", target, err)
	}
	if outsideRoot(rel) {
		return ScannedFile{}, "", false, fmt.Errorf("changed context target escapes repository root: %s", target)
	}
	content, err := os.ReadFile(real)
	if err != nil {
		return ScannedFile{}, "", false, fmt.Errorf("read changed context target %q: %w", target, err)
	}
	if bytes.IndexByte(content, 0) >= 0 {
		return ScannedFile{}, "", false, nil
	}
	return ScannedFile{Path: filepath.ToSlash(rel), Content: string(content)}, real, true, nil
}

func isMissing(err error) bool {
	if os.IsNotExist(err) {
		return true
	}
	return errors.Is(err, os.ErrNotExist)
}
