package diff

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestFilterByPath(t *testing.T) {
	input := model.Diff{Files: []model.DiffFile{{Path: "src/app.go"}, {Path: "README.md"}}}
	filtered := Filter(input, "src/**")
	if len(filtered.Files) != 1 {
		t.Fatalf("unexpected file count: %d", len(filtered.Files))
	}
	if filtered.Files[0].Path != "src/app.go" {
		t.Fatalf("unexpected path: %s", filtered.Files[0].Path)
	}
}
