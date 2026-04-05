package scan

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/git"
	"github.com/Pimatis/mavetis/src/match"
)

const (
	maxExplicitFiles    = 128
	maxExplicitFileSize = 1 << 20
)

var ignoredWalkDirs = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"vendor":       {},
	".venv":        {},
	"__pycache__":  {},
	"dist":         {},
	"build":        {},
	".next":        {},
	"target":       {},
	"bin":          {},
	"obj":          {},
}

type ScannedFile struct {
	Path    string
	Content string
}

func Root() (string, error) {
	root, err := git.Root()
	if err == nil {
		return realPath(root)
	}
	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		return "", cwdErr
	}
	return realPath(cwd)
}

func LoadFiles(root string, targets []string) ([]ScannedFile, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	if len(targets) > maxExplicitFiles {
		return nil, fmt.Errorf("too many @file targets: max %d", maxExplicitFiles)
	}
	rootReal, err := realPath(root)
	if err != nil {
		return nil, err
	}
	files := make([]ScannedFile, 0, len(targets))
	seen := map[string]struct{}{}
	for _, target := range targets {
		expanded, err := expandTargets(rootReal, target)
		if err != nil {
			return nil, err
		}
		for _, candidate := range expanded {
			file, real, err := loadFile(rootReal, candidate)
			if err != nil {
				return nil, err
			}
			if _, ok := seen[real]; ok {
				continue
			}
			seen[real] = struct{}{}
			files = append(files, file)
			if len(files) > maxExplicitFiles {
				return nil, fmt.Errorf("too many @file targets: max %d", maxExplicitFiles)
			}
		}
	}
	return files, nil
}

func expandTargets(root string, target string) ([]string, error) {
	if strings.TrimSpace(target) == "" {
		return nil, errors.New("empty @file target")
	}
	if strings.ContainsRune(target, rune(0)) {
		return nil, fmt.Errorf("review target contains NUL byte: %q", target)
	}
	if !hasPattern(target) {
		directory, ok, err := expandDirectoryTarget(root, target)
		if err != nil {
			return nil, err
		}
		if ok {
			return directory, nil
		}
		return []string{target}, nil
	}
	if filepath.IsAbs(target) {
		return nil, fmt.Errorf("absolute @file globs are not supported: %s", target)
	}
	pattern := filepath.ToSlash(filepath.Clean(target))
	base := filepath.Join(root, staticPrefix(pattern))
	info, err := os.Stat(base)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("review target not found: %s", target)
		}
		return nil, fmt.Errorf("expand review target %q: %w", target, err)
	}
	if !info.IsDir() {
		base = filepath.Dir(base)
	}
	matches := make([]string, 0)
	err = filepath.WalkDir(base, func(path string, item fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if item.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if outsideRoot(rel) {
			return nil
		}
		if !match.Glob(pattern, filepath.ToSlash(rel)) {
			return nil
		}
		matches = append(matches, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("expand review target %q: %w", target, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("review target not found: %s", target)
	}
	sort.Strings(matches)
	return matches, nil
}

func expandDirectoryTarget(root string, target string) ([]string, bool, error) {
	absolute := target
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(root, target)
	}
	absolute = filepath.Clean(absolute)
	real, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("resolve review target %q: %w", target, err)
	}
	info, err := os.Stat(real)
	if err != nil {
		return nil, false, fmt.Errorf("stat review target %q: %w", target, err)
	}
	if !info.IsDir() {
		return nil, false, nil
	}
	rel, err := filepath.Rel(root, real)
	if err != nil {
		return nil, false, fmt.Errorf("resolve review target %q: %w", target, err)
	}
	if outsideRoot(rel) {
		return nil, false, fmt.Errorf("review target escapes repository root: %s", target)
	}
	matches := make([]string, 0)
	err = filepath.WalkDir(real, func(path string, item fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if item.IsDir() {
			if path != real && skipWalkDir(item.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if outsideRoot(relPath) {
			return nil
		}
		matches = append(matches, path)
		if len(matches) > maxExplicitFiles {
			return errors.New("too many @file targets")
		}
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("expand review target %q: %w", target, err)
	}
	if len(matches) == 0 {
		return nil, false, fmt.Errorf("review target directory is empty: %s", target)
	}
	sort.Strings(matches)
	return matches, true, nil
}

func skipWalkDir(name string) bool {
	_, ok := ignoredWalkDirs[name]
	return ok
}

func loadFile(root string, target string) (ScannedFile, string, error) {
	absolute := target
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(root, target)
	}
	absolute = filepath.Clean(absolute)
	real, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return ScannedFile{}, "", fmt.Errorf("resolve review target %q: %w", target, err)
	}
	info, err := os.Stat(real)
	if err != nil {
		return ScannedFile{}, "", fmt.Errorf("stat review target %q: %w", target, err)
	}
	if !info.Mode().IsRegular() {
		return ScannedFile{}, "", fmt.Errorf("review target must be a regular file: %s", target)
	}
	if info.Size() > maxExplicitFileSize {
		return ScannedFile{}, "", fmt.Errorf("review target exceeds 1 MiB limit: %s", target)
	}
	rel, err := filepath.Rel(root, real)
	if err != nil {
		return ScannedFile{}, "", fmt.Errorf("resolve review target %q: %w", target, err)
	}
	if outsideRoot(rel) {
		return ScannedFile{}, "", fmt.Errorf("review target escapes repository root: %s", target)
	}
	content, err := os.ReadFile(real)
	if err != nil {
		return ScannedFile{}, "", fmt.Errorf("read review target %q: %w", target, err)
	}
	if bytes.IndexByte(content, 0) >= 0 {
		return ScannedFile{}, "", fmt.Errorf("review target must be a text file: %s", target)
	}
	return ScannedFile{Path: filepath.ToSlash(rel), Content: string(content)}, real, nil
}

func realPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func hasPattern(value string) bool {
	return strings.ContainsAny(value, "*?[")
}

func staticPrefix(pattern string) string {
	segments := strings.Split(filepath.ToSlash(pattern), "/")
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" || segment == "." {
			continue
		}
		if segment == ".." {
			break
		}
		if hasPattern(segment) {
			break
		}
		parts = append(parts, segment)
	}
	if len(parts) == 0 {
		return "."
	}
	return filepath.Join(parts...)
}

func outsideRoot(rel string) bool {
	if rel == ".." {
		return true
	}
	return strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
