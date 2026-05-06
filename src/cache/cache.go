package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

const Version = 1

type File struct {
	Path    string
	Size    int64
	ModTime int64
}

type Data struct {
	Version int              `json:"version"`
	Key     string           `json:"key"`
	Files   map[string]Entry `json:"files"`
}

type Entry struct {
	Size     int64           `json:"size"`
	ModTime  int64           `json:"modTime"`
	Findings []model.Finding `json:"findings"`
}

func Load(root string, namespace string, configured string, key string) (string, Data, error) {
	path, err := Path(root, namespace, configured)
	if err != nil {
		return "", Data{}, err
	}
	data := Data{Version: Version, Key: key, Files: map[string]Entry{}}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return path, data, nil
		}
		return "", Data{}, fmt.Errorf("read cache: %w", err)
	}
	loaded := Data{}
	if err := json.Unmarshal(content, &loaded); err != nil {
		return path, data, nil
	}
	if loaded.Version != Version || loaded.Key != key || loaded.Files == nil {
		return path, data, nil
	}
	return path, loaded, nil
}

func Save(path string, data Data) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

func Path(root string, namespace string, configured string) (string, error) {
	if configured != "" {
		if filepath.IsAbs(configured) {
			return filepath.Clean(configured), nil
		}
		return filepath.Join(root, configured), nil
	}
	directory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(root))
	name := hex.EncodeToString(sum[:8]) + ".json"
	return filepath.Join(directory, "mavetis", namespace, name), nil
}

func Findings(data Data, file File) ([]model.Finding, bool) {
	entry, ok := data.Files[file.Path]
	if !ok {
		return nil, false
	}
	if entry.Size != file.Size || entry.ModTime != file.ModTime {
		return nil, false
	}
	return CloneFindings(entry.Findings), true
}

func Put(data Data, file File, findings []model.Finding) {
	data.Files[file.Path] = Entry{Size: file.Size, ModTime: file.ModTime, Findings: CloneFindings(findings)}
}

func Prune(data Data, files []File) {
	current := map[string]struct{}{}
	for _, file := range files {
		current[file.Path] = struct{}{}
	}
	for path := range data.Files {
		if _, ok := current[path]; ok {
			continue
		}
		delete(data.Files, path)
	}
}

func CloneFindings(findings []model.Finding) []model.Finding {
	cloned := make([]model.Finding, 0, len(findings))
	for _, finding := range findings {
		finding.Reasons = append([]string{}, finding.Reasons...)
		finding.Standards = append([]string{}, finding.Standards...)
		cloned = append(cloned, finding)
	}
	return cloned
}

func Key(parts ...string) string {
	values := append([]string{}, parts...)
	sort.Strings(values)
	sum := sha256.Sum256([]byte(strings.Join(values, "\n")))
	return hex.EncodeToString(sum[:])
}
