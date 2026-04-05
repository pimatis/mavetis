package model

type Config struct {
	Severity  string
	FailOn    string
	Output    string
	Profile   string
	Ignore    []string
	Allow     Allow
	Company   Company
	Zones     Zones
	Supply    Supply
	Snapshot  SnapshotConfig
	Rules     []Rule
	Snapshots []Snapshot
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

type Supply struct {
	AllowPackages     []string
	DenyPackages      []string
	TrustedRegistries []string
}

type SnapshotConfig struct {
	Path string
}

type Snapshot struct {
	ID          string
	Path        string
	Anchor      string
	Category    string
	Severity    string
	Require     []string
	Standards   []string
	Message     string
	Remediation string
}

type Review struct {
	Mode          string
	Base          string
	Head          string
	Format        string
	Severity      string
	FailOn        string
	Profile       string
	ConfigPath    string
	RulesPath     string
	Path          string
	Explain       bool
	Staged        bool
	WithSuggested bool
	StdinTargets  bool
	Files         []string
}
