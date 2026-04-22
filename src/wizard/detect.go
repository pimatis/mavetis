package wizard

import (
	"os"
	"path/filepath"
	"strings"
)

const maxDetectDepth = 3

var skipDetectDirs = map[string]struct{}{
	".git": {}, "node_modules": {}, "vendor": {}, ".venv": {},
	"__pycache__": {}, "dist": {}, "build": {}, ".next": {},
	"target": {}, "bin": {}, "obj": {}, ".idea": {}, ".vscode": {},
}

type Project struct {
	Language   string
	Profile    string
	Ignore     []string
	Critical   []string
	Restricted []string
}

func Detect(root string) Project {
	project := Project{}
	detectProject(root, &project, 0, "")
	if project.Language == "javascript" {
		detectNodeProfile(root, &project)
	}
	if project.Profile == "" {
		project.Profile = "auth"
	}
	dedupeStrings(&project.Ignore)
	dedupeStrings(&project.Critical)
	dedupeStrings(&project.Restricted)
	return project
}

func detectProject(root string, project *Project, depth int, prefix string) {
	if depth > maxDetectDepth {
		return
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() {
			detectFile(name, project)
			continue
		}
		if _, skip := skipDetectDirs[name]; skip {
			continue
		}
		rel := name
		if prefix != "" {
			rel = prefix + "/" + name
		}
		detectDir(rel, project)
		detectProject(filepath.Join(root, name), project, depth+1, rel)
	}
}

func detectFile(name string, project *Project) {
	lower := strings.ToLower(name)
	switch lower {
	case "go.mod":
		project.Language = "go"
		project.Profile = "backend"
		appendIgnore(project, "vendor/**")
	case "package.json":
		project.Language = "javascript"
		appendIgnore(project, "node_modules/**")
	case "cargo.toml":
		project.Language = "rust"
		project.Profile = "backend"
		appendIgnore(project, "target/**")
	case "requirements.txt", "pyproject.toml", "pipfile":
		project.Language = "python"
		project.Profile = "backend"
		appendIgnore(project, "__pycache__/**")
	case "pom.xml", "build.gradle", "build.gradle.kts":
		project.Language = "java"
		project.Profile = "backend"
	case "composer.json":
		project.Language = "php"
		project.Profile = "backend"
	case "pubspec.yaml":
		project.Language = "dart"
		project.Profile = "frontend"
	}
}

func detectDir(rel string, project *Project) {
	lower := strings.ToLower(rel)
	segments := strings.Split(lower, "/")
	last := segments[len(segments)-1]
	switch last {
	case "auth", "authentication", "session", "login", "oauth", "identity", "guard":
		appendZone(project, "critical", rel+"/**")
	case "api", "handler", "controller", "routes", "endpoint", "middleware", "service", "gateway":
		appendZone(project, "restricted", rel+"/**")
	case "config", "configuration", "settings", "env", "secrets":
		appendZone(project, "restricted", rel+"/**")
	}
}

func detectNodeProfile(root string, project *Project) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return
	}
	content := string(data)
	if strings.Contains(content, `"react"`) || strings.Contains(content, `"vue"`) ||
		strings.Contains(content, `"angular"`) || strings.Contains(content, `"svelte"`) ||
		strings.Contains(content, `"next"`) || strings.Contains(content, `"nuxt"`) {
		project.Profile = "frontend"
		appendIgnore(project, ".next/**")
		appendIgnore(project, "dist/**")
	}
	if strings.Contains(content, `"express"`) || strings.Contains(content, `"fastify"`) ||
		strings.Contains(content, `"nestjs"`) || strings.Contains(content, `"koa"`) ||
		strings.Contains(content, `"hapi"`) || strings.Contains(content, `"fastify"`) {
		project.Profile = "backend"
	}
}

func appendIgnore(project *Project, pattern string) {
	project.Ignore = append(project.Ignore, pattern)
}

func appendZone(project *Project, zone string, pattern string) {
	if zone == "critical" {
		project.Critical = append(project.Critical, pattern)
	}
	if zone == "restricted" {
		project.Restricted = append(project.Restricted, pattern)
	}
}

func dedupeStrings(list *[]string) {
	if len(*list) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(*list))
	result := make([]string, 0, len(*list))
	for _, item := range *list {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	*list = result
}
