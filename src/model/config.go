package model

type Config struct {
	Severity string
	FailOn   string
	Output   string
	Profile  string
	Ignore   []string
	Allow    Allow
	Company  Company
	Zones    Zones
	Rules    []Rule
}

type Allow struct {
	Paths   []string
	Values  []string
	Regexes []string
}

type Company struct {
	Prefixes []string
}

type Zones struct {
	Critical   []string
	Restricted []string
}

type Review struct {
	Mode       string
	Base       string
	Head       string
	Format     string
	Severity   string
	FailOn     string
	Profile    string
	ConfigPath string
	RulesPath  string
	Path       string
	Explain    bool
	Staged     bool
}
