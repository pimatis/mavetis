package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/yaml"
)

func Load(path string) (model.Config, error) {
	config := model.Config{Severity: "low", FailOn: "high", Output: "text"}
	if path == "" {
		path = detect()
	}
	if path == "" {
		return config, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("read config: %w", err)
	}
	value, err := yaml.Parse(string(content))
	if err != nil {
		return config, fmt.Errorf("parse config: %w", err)
	}
	mapped, err := yaml.Map(value)
	if err != nil {
		return config, fmt.Errorf("decode config: %w", err)
	}
	decodeConfig(mapped, &config)
	if err := validateConfig(config); err != nil {
		return config, err
	}
	snapshotPath := config.Snapshot.Path
	if snapshotPath != "" && !filepath.IsAbs(snapshotPath) {
		snapshotPath = filepath.Join(filepath.Dir(path), snapshotPath)
		config.Snapshot.Path = filepath.Clean(snapshotPath)
	}
	snapshots, err := LoadSnapshots(snapshotPath)
	if err != nil {
		return config, err
	}
	config.Snapshots = snapshots
	return config, nil
}

func detect() string {
	candidates := []string{".mavetis.yaml", ".mavetis.yml"}
	for _, candidate := range candidates {
		_, err := os.Stat(candidate)
		if err == nil {
			return filepath.Clean(candidate)
		}
	}
	return ""
}

func decodeConfig(mapped map[string]any, config *model.Config) {
	severity, ok := yaml.String(mapped["severity"])
	if ok {
		config.Severity = severity
	}
	failOn, ok := yaml.String(mapped["failon"])
	if ok {
		config.FailOn = failOn
	}
	failOn, ok = yaml.String(mapped["fail-on"])
	if ok {
		config.FailOn = failOn
	}
	output, ok := yaml.String(mapped["output"])
	if ok {
		config.Output = output
	}
	profile, ok := yaml.String(mapped["profile"])
	if ok {
		config.Profile = profile
	}
	config.Ignore = yaml.Strings(mapped["ignore"])
	allow, ok := mapped["allow"]
	if ok {
		allowMap, err := yaml.Map(allow)
		if err == nil {
			config.Allow.Paths = yaml.Strings(allowMap["paths"])
			config.Allow.Values = yaml.Strings(allowMap["values"])
			config.Allow.Regexes = yaml.Strings(allowMap["regexes"])
		}
	}
	company, ok := mapped["company"]
	if ok {
		companyMap, err := yaml.Map(company)
		if err == nil {
			config.Company.Prefixes = yaml.Strings(companyMap["prefixes"])
		}
	}
	supply, ok := mapped["supply"]
	if ok {
		supplyMap, err := yaml.Map(supply)
		if err == nil {
			config.Supply.AllowPackages = yaml.Strings(supplyMap["allow-packages"])
			config.Supply.DenyPackages = yaml.Strings(supplyMap["deny-packages"])
			config.Supply.TrustedRegistries = yaml.Strings(supplyMap["trusted-registries"])
		}
	}
	snapshot, ok := mapped["snapshot"]
	if ok {
		snapshotMap, err := yaml.Map(snapshot)
		if err == nil {
			config.Snapshot.Path, _ = yaml.String(snapshotMap["path"])
		}
	}
	zones, ok := mapped["zones"]
	if ok {
		zonesMap, err := yaml.Map(zones)
		if err == nil {
			config.Zones.Critical = yaml.Strings(zonesMap["critical"])
			config.Zones.Restricted = yaml.Strings(zonesMap["restricted"])
		}
	}
}
