package config

import (
	"fmt"
	"os"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/yaml"
)

func LoadSnapshots(path string) ([]model.Snapshot, error) {
	if path == "" {
		return nil, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshots: %w", err)
	}
	value, err := yaml.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse snapshots: %w", err)
	}
	mapped, err := yaml.Map(value)
	if err != nil {
		return nil, fmt.Errorf("decode snapshots: %w", err)
	}
	items, err := yaml.List(mapped["snapshots"])
	if err != nil {
		return nil, fmt.Errorf("decode snapshots: expected snapshots list")
	}
	snapshots := make([]model.Snapshot, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		entry, err := yaml.Map(item)
		if err != nil {
			return nil, fmt.Errorf("decode snapshot: %w", err)
		}
		snapshot := model.Snapshot{}
		snapshot.ID, _ = yaml.String(entry["id"])
		snapshot.Path, _ = yaml.String(entry["path"])
		snapshot.Anchor, _ = yaml.String(entry["anchor"])
		snapshot.Category, _ = yaml.String(entry["category"])
		snapshot.Severity, _ = yaml.String(entry["severity"])
		snapshot.Require = yaml.Strings(entry["require"])
		snapshot.Standards = yaml.Strings(entry["standards"])
		snapshot.Message, _ = yaml.String(entry["message"])
		snapshot.Remediation, _ = yaml.String(entry["remediation"])
		if snapshot.ID == "" {
			return nil, fmt.Errorf("decode snapshot: missing id")
		}
		if snapshot.Path == "" {
			return nil, fmt.Errorf("decode snapshot: missing path for %s", snapshot.ID)
		}
		if snapshot.Anchor == "" {
			return nil, fmt.Errorf("decode snapshot: missing anchor for %s", snapshot.ID)
		}
		if len(snapshot.Require) == 0 {
			return nil, fmt.Errorf("decode snapshot: missing require for %s", snapshot.ID)
		}
		if snapshot.Severity == "" {
			snapshot.Severity = "high"
		}
		if snapshot.Category == "" {
			snapshot.Category = "snapshot"
		}
		if snapshot.Message == "" {
			snapshot.Message = "The diff weakens a repository-specific security snapshot."
		}
		if snapshot.Remediation == "" {
			snapshot.Remediation = "Restore the required security behavior or refresh the snapshot only after review."
		}
		if _, ok := seen[snapshot.ID]; ok {
			return nil, fmt.Errorf("decode snapshot: duplicate id %s", snapshot.ID)
		}
		if err := ValidateSeverity(snapshot.Severity, "snapshot severity"); err != nil {
			return nil, err
		}
		seen[snapshot.ID] = struct{}{}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}
