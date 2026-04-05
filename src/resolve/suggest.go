package resolve

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/scan"
)

type Limits struct {
	MaxDepth          int
	MaxFiles          int
	MaxEdgesPerFile   int
	MaxBytesPerFile   int64
	MaxTotalBytes     int64
	MaxGoPackageFiles int
}

type queueItem struct {
	file  scan.ScannedFile
	depth int
}

var ignoredDirs = map[string]struct{}{
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

func DefaultLimits() Limits {
	return Limits{
		MaxDepth:          2,
		MaxFiles:          50,
		MaxEdgesPerFile:   64,
		MaxBytesPerFile:   256 << 10,
		MaxTotalBytes:     2 << 20,
		MaxGoPackageFiles: 16,
	}
}

func Suggest(root string, seeds []scan.ScannedFile, limits Limits) ([]model.Suggestion, error) {
	_, suggestions, err := Discover(root, seeds, limits)
	if err != nil {
		return nil, err
	}
	return suggestions, nil
}

func Discover(root string, seeds []scan.ScannedFile, limits Limits) ([]scan.ScannedFile, []model.Suggestion, error) {
	if len(seeds) == 0 {
		return nil, nil, nil
	}
	rootReal, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, nil, err
	}
	queue := make([]queueItem, 0, len(seeds))
	visited := map[string]struct{}{}
	totalBytes := int64(0)
	for _, seed := range seeds {
		absolute := filepath.Join(rootReal, filepath.FromSlash(seed.Path))
		real, err := filepath.EvalSymlinks(absolute)
		if err == nil {
			visited[real] = struct{}{}
		}
		totalBytes += int64(len(seed.Content))
		queue = append(queue, queueItem{file: seed})
	}
	discovered := make([]scan.ScannedFile, 0)
	suggestions := make([]model.Suggestion, 0)
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if current.depth >= limits.MaxDepth {
			continue
		}
		refs := Imports(current.file.Path, current.file.Content)
		if len(refs) > limits.MaxEdgesPerFile {
			refs = refs[:limits.MaxEdgesPerFile]
		}
		for _, ref := range refs {
			resolved := ResolveLocal(rootReal, current.file.Path, ref.Module, limits.MaxGoPackageFiles)
			for _, path := range resolved {
				if len(suggestions) >= limits.MaxFiles {
					return discovered, suggestions, nil
				}
				if ignoredPath(path) {
					continue
				}
				file, real, ok := loadSuggestion(rootReal, path, totalBytes, limits)
				if !ok {
					continue
				}
				if _, ok := visited[real]; ok {
					continue
				}
				visited[real] = struct{}{}
				totalBytes += int64(len(file.Content))
				discovered = append(discovered, file)
				suggestions = append(suggestions, model.Suggestion{Path: path, From: current.file.Path, Reason: reasonText(ref.Kind), Depth: current.depth + 1})
				queue = append(queue, queueItem{file: file, depth: current.depth + 1})
			}
		}
	}
	return discovered, suggestions, nil
}

func loadSuggestion(root string, path string, totalBytes int64, limits Limits) (scan.ScannedFile, string, bool) {
	absolute := filepath.Join(root, filepath.FromSlash(path))
	real, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return scan.ScannedFile{}, "", false
	}
	if outsideRoot(root, real) {
		return scan.ScannedFile{}, "", false
	}
	info, err := os.Stat(real)
	if err != nil {
		return scan.ScannedFile{}, "", false
	}
	if !info.Mode().IsRegular() {
		return scan.ScannedFile{}, "", false
	}
	if info.Size() > limits.MaxBytesPerFile {
		return scan.ScannedFile{}, "", false
	}
	if totalBytes+info.Size() > limits.MaxTotalBytes {
		return scan.ScannedFile{}, "", false
	}
	content, err := os.ReadFile(real)
	if err != nil {
		return scan.ScannedFile{}, "", false
	}
	if bytes.IndexByte(content, 0) >= 0 {
		return scan.ScannedFile{}, "", false
	}
	return scan.ScannedFile{Path: path, Content: string(content)}, real, true
}

func ignoredPath(path string) bool {
	segments := strings.Split(filepath.ToSlash(path), "/")
	for _, segment := range segments {
		if _, ok := ignoredDirs[segment]; ok {
			return true
		}
	}
	return false
}

func outsideRoot(root string, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return true
	}
	if rel == ".." {
		return true
	}
	return strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func reasonText(kind string) string {
	if kind == "require" {
		return "required"
	}
	if kind == "export-from" {
		return "re-exported"
	}
	if kind == "dynamic-import" {
		return "dynamically imported"
	}
	return "imported"
}
