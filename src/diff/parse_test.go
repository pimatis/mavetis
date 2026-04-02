package diff

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestParse(t *testing.T) {
	input := `diff --git a/app.js b/app.js
index 1111111..2222222 100644
--- a/app.js
+++ b/app.js
@@ -1,3 +1,4 @@
 const token = load();
-if (!session) return;
+if (!session) return;
+localStorage.setItem("token", token);
 run();
`
	parsed, err := Parse(input, model.DiffMeta{Mode: "staged"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(parsed.Files) != 1 {
		t.Fatalf("unexpected file count: %d", len(parsed.Files))
	}
	file := parsed.Files[0]
	if file.Path != "app.js" {
		t.Fatalf("unexpected path: %s", file.Path)
	}
	if len(file.Hunks) != 1 {
		t.Fatalf("unexpected hunk count: %d", len(file.Hunks))
	}
	lines := file.Flatten()
	if len(lines) != 5 {
		t.Fatalf("unexpected line count: %d", len(lines))
	}
	if lines[3].Kind != "added" || lines[3].NewNumber != 3 {
		t.Fatalf("unexpected added line: %#v", lines[3])
	}
}
