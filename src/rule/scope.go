package rule

var codeSet = []string{
	"**/*.go",
	"**/*.js",
	"**/*.jsx",
	"**/*.ts",
	"**/*.tsx",
	"**/*.mjs",
	"**/*.cjs",
	"**/*.py",
	"**/*.rb",
	"**/*.java",
	"**/*.kt",
	"**/*.php",
	"**/*.cs",
	"**/*.rs",
	"**/*.swift",
}

var webSet = []string{
	"**/*.js",
	"**/*.jsx",
	"**/*.ts",
	"**/*.tsx",
	"**/*.vue",
	"**/*.svelte",
	"**/*.html",
}

var configSet = []string{
	"**/*.json",
	"**/*.yaml",
	"**/*.yml",
	"**/*.toml",
	"**/*.ini",
	"**/*.conf",
	"**/*.env",
	"**/.env",
	"**/.env.*",
}

var workflowSet = []string{
	".github/workflows/**",
	".gitlab-ci.yml",
	"**/Jenkinsfile",
	"**/*.yaml",
	"**/*.yml",
}

var deploymentSet = []string{
	"Dockerfile",
	"**/Dockerfile",
	"**/docker-compose.yml",
	"**/docker-compose.yaml",
	"**/compose.yml",
	"**/compose.yaml",
	"**/*.conf",
	"**/*.cfg",
	"**/*.cnf",
	"**/*.properties",
	"**/nginx.conf",
	"**/Caddyfile",
	"**/*.env",
	"**/.env",
	"**/.env.*",
}

var manifestSet = []string{
	"**/package.json",
	"**/package-lock.json",
	"**/pnpm-lock.yaml",
	"**/yarn.lock",
	"**/bun.lock",
	"**/bun.lockb",
	"**/requirements.txt",
	"**/pyproject.toml",
	"**/Pipfile",
	"**/go.mod",
	"**/go.sum",
	"**/Cargo.toml",
	"**/Cargo.lock",
	"**/pom.xml",
	"**/build.gradle",
}

func codeFiles() []string {
	return clone(codeSet)
}

func webFiles() []string {
	return clone(webSet)
}

func configFiles() []string {
	return clone(configSet)
}

func workflowFiles() []string {
	return clone(workflowSet)
}

func deploymentFiles() []string {
	return clone(deploymentSet)
}

func manifestFiles() []string {
	return clone(manifestSet)
}

func codeAndConfigFiles() []string {
	result := clone(codeSet)
	result = append(result, configSet...)
	return result
}

func webAndConfigFiles() []string {
	result := clone(webSet)
	result = append(result, configSet...)
	return result
}

func clone(values []string) []string {
	return append([]string{}, values...)
}
