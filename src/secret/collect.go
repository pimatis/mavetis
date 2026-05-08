package secret

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/scan"
)

const (
	maxFileSize  = 1 << 20
	maxScanFiles = 20000
)

var skippedDirs = map[string]struct{}{
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

func collect(root string, targets []string, pathPattern string) ([]file, error) {
	files := make([]file, 0)
	seen := map[string]struct{}{}
	for _, target := range targets {
		items, err := collectTarget(root, target)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if pathPattern != "" && !match.Glob(pathPattern, item.path) {
				continue
			}
			if _, ok := seen[item.real]; ok {
				continue
			}
			seen[item.real] = struct{}{}
			files = append(files, item)
			if len(files) > maxScanFiles {
				return nil, fmt.Errorf("too many files to scan: max %d", maxScanFiles)
			}
		}
	}
	sort.Slice(files, func(left int, right int) bool {
		return files[left].path < files[right].path
	})
	return files, nil
}

func collectTarget(root string, target string) ([]file, error) {
	if strings.TrimSpace(target) == "" {
		return nil, errors.New("empty secrets scan target")
	}
	if strings.ContainsRune(target, rune(0)) {
		return nil, fmt.Errorf("secrets scan target contains NUL byte: %q", target)
	}
	absolute := target
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(root, target)
	}
	absolute = filepath.Clean(absolute)
	real, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, fmt.Errorf("resolve secrets scan target %q: %w", target, err)
	}
	rel, err := filepath.Rel(root, real)
	if err != nil {
		return nil, fmt.Errorf("resolve secrets scan target %q: %w", target, err)
	}
	if outsideRoot(rel) {
		return nil, fmt.Errorf("secrets scan target escapes repository root: %s", target)
	}
	info, err := os.Stat(real)
	if err != nil {
		return nil, fmt.Errorf("stat secrets scan target %q: %w", target, err)
	}
	if info.Mode().IsRegular() {
		return []file{{path: filepath.ToSlash(rel), real: real, size: info.Size(), modTime: info.ModTime().UnixNano()}}, nil
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("secrets scan target must be a regular file or directory: %s", target)
	}
	return walk(root, real)
}

func walk(root string, directory string) ([]file, error) {
	files := make([]file, 0)
	ignorePatterns := scan.LoadGitignorePatterns(root)
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != directory && skipDir(entry.Name()) {
				return filepath.SkipDir
			}
			if path != directory {
				relDir, relErr := filepath.Rel(root, path)
				if relErr == nil && scan.IsGitignored(ignorePatterns, relDir) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() > maxFileSize {
			return nil
		}
		real, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(root, real)
		if err != nil {
			return err
		}
		if outsideRoot(rel) {
			return nil
		}
		if scan.IsGitignored(ignorePatterns, rel) {
			return nil
		}
		files = append(files, file{path: filepath.ToSlash(rel), real: real, size: info.Size(), modTime: info.ModTime().UnixNano()})
		if len(files) > maxScanFiles {
			return fmt.Errorf("too many files to scan: max %d", maxScanFiles)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func skipDir(name string) bool {
	_, ok := skippedDirs[name]
	return ok
}

func realPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(absolute)
}

func outsideRoot(rel string) bool {
	if rel == ".." {
		return true
	}
	return strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
