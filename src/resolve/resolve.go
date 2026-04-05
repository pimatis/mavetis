package resolve

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Pimatis/mavetis/src/analyze"
)

var goModuleCache sync.Map

func ResolveLocal(root string, fromPath string, module string, maxGoPackageFiles int) []string {
	module = strings.TrimSpace(module)
	if module == "" {
		return nil
	}
	language := analyze.Language(fromPath)
	if language == "go" {
		return resolveGo(root, module, maxGoPackageFiles)
	}
	if language == "javascript" || language == "typescript" {
		return resolveScript(root, fromPath, module)
	}
	if language == "python" {
		return resolvePython(root, fromPath, module)
	}
	if language == "jvm" {
		return resolveJVM(root, module)
	}
	return nil
}

func resolveGo(root string, module string, maxGoPackageFiles int) []string {
	modulePath := goModulePath(root)
	if modulePath == "" {
		return nil
	}
	if !strings.HasPrefix(module, modulePath) {
		return nil
	}
	rel := strings.TrimPrefix(module, modulePath)
	rel = strings.TrimPrefix(rel, "/")
	directory := root
	if rel != "" {
		directory = filepath.Join(root, filepath.FromSlash(rel))
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		candidate := filepath.Join(directory, name)
		relPath := relativeFile(root, candidate)
		if relPath == "" {
			continue
		}
		files = append(files, relPath)
	}
	sort.Strings(files)
	if len(files) > maxGoPackageFiles {
		return files[:maxGoPackageFiles]
	}
	return files
}

func resolveScript(root string, fromPath string, module string) []string {
	if !strings.HasPrefix(module, "./") && !strings.HasPrefix(module, "../") {
		return nil
	}
	base := filepath.Join(filepath.Dir(filepath.Join(root, filepath.FromSlash(fromPath))), filepath.FromSlash(module))
	probes := []string{base}
	if filepath.Ext(base) == "" {
		for _, ext := range []string{".js", ".ts", ".tsx", ".jsx", ".mjs", ".cjs"} {
			probes = append(probes, base+ext)
		}
		for _, ext := range []string{"index.js", "index.ts", "index.tsx", "index.jsx", "index.mjs", "index.cjs"} {
			probes = append(probes, filepath.Join(base, ext))
		}
	}
	return firstExisting(root, probes)
}

func resolvePython(root string, fromPath string, module string) []string {
	base := root
	trimmed := module
	if strings.HasPrefix(trimmed, ".") {
		current := filepath.Dir(filepath.Join(root, filepath.FromSlash(fromPath)))
		depth := leadingDots(trimmed)
		for index := 1; index < depth; index++ {
			current = filepath.Dir(current)
		}
		base = current
		trimmed = strings.TrimLeft(trimmed, ".")
	}
	path := ""
	if trimmed != "" {
		path = filepath.FromSlash(strings.ReplaceAll(trimmed, ".", "/"))
	}
	probes := make([]string, 0, 2)
	if path != "" {
		probes = append(probes, filepath.Join(base, path+".py"))
		probes = append(probes, filepath.Join(base, path, "__init__.py"))
	}
	if path == "" {
		probes = append(probes, filepath.Join(base, "__init__.py"))
	}
	return firstExisting(root, probes)
}

func resolveJVM(root string, module string) []string {
	relative := filepath.FromSlash(strings.ReplaceAll(module, ".", "/"))
	roots := []string{
		filepath.Join(root, "src", "main", "java"),
		filepath.Join(root, "src", "main", "kotlin"),
		filepath.Join(root, "src", "test", "java"),
		filepath.Join(root, "src", "test", "kotlin"),
		root,
	}
	probes := make([]string, 0, len(roots)*2)
	for _, sourceRoot := range roots {
		probes = append(probes, filepath.Join(sourceRoot, relative+".java"))
		probes = append(probes, filepath.Join(sourceRoot, relative+".kt"))
	}
	return firstExisting(root, probes)
}

func firstExisting(root string, probes []string) []string {
	seen := map[string]struct{}{}
	for _, probe := range probes {
		rel := relativeFile(root, probe)
		if rel == "" {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		seen[rel] = struct{}{}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		return []string{rel}
	}
	return nil
}

func relativeFile(root string, path string) string {
	cleaned := filepath.Clean(path)
	rel, err := filepath.Rel(root, cleaned)
	if err != nil {
		return ""
	}
	if rel == ".." {
		return ""
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ""
	}
	return filepath.ToSlash(rel)
}

func goModulePath(root string) string {
	if cached, ok := goModuleCache.Load(root); ok {
		return cached.(string)
	}
	path := filepath.Join(root, "go.mod")
	content, err := os.ReadFile(path)
	if err != nil {
		goModuleCache.Store(root, "")
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		module := strings.TrimSpace(strings.TrimPrefix(line, "module "))
		goModuleCache.Store(root, module)
		return module
	}
	goModuleCache.Store(root, "")
	return ""
}

func leadingDots(value string) int {
	count := 0
	for _, char := range value {
		if char != '.' {
			break
		}
		count++
	}
	if count == 0 {
		return 1
	}
	return count
}
