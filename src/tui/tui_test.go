package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialStateIsMenu(t *testing.T) {
	m := &modelImpl{state: stateMenu}
	if m.state != stateMenu {
		t.Errorf("expected initial state menu, got %v", m.state)
	}

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestMenuQuitOnQ(t *testing.T) {
	m := &modelImpl{state: stateMenu}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected quit command on 'q'")
	}
	_ = updated
}

func TestMenuQuitOnCtrlC(t *testing.T) {
	m := &modelImpl{state: stateMenu}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command on ctrl+c")
	}
	_ = updated
}

func TestMenuNavigationDown(t *testing.T) {
	m := &modelImpl{state: stateMenu, menuCursor: 0}
	initial := m.menuCursor
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	after := updated.(*modelImpl)
	if after.menuCursor == initial {
		t.Error("expected cursor to move down on 'j'")
	}
}

func TestMenuNavigationUp(t *testing.T) {
	m := &modelImpl{state: stateMenu, menuCursor: 1}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	after := updated.(*modelImpl)
	if after.menuCursor != 0 {
		t.Errorf("expected cursor to move up on 'k', got %d", after.menuCursor)
	}
}

func TestMenuNavigationUpWraps(t *testing.T) {
	m := &modelImpl{state: stateMenu, menuCursor: 0}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	after := updated.(*modelImpl)
	if after.menuCursor != len(menuItems)-1 {
		t.Errorf("expected cursor to wrap to last item, got %d", after.menuCursor)
	}
}

func TestMenuNavigationDownWraps(t *testing.T) {
	m := &modelImpl{state: stateMenu, menuCursor: len(menuItems) - 1}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	after := updated.(*modelImpl)
	if after.menuCursor != 0 {
		t.Errorf("expected cursor to wrap to first item, got %d", after.menuCursor)
	}
}

func TestMenuArrowNavigation(t *testing.T) {
	m := &modelImpl{state: stateMenu, menuCursor: 0}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	after := updated.(*modelImpl)
	if after.menuCursor != 1 {
		t.Errorf("expected cursor to move down on arrow down, got %d", after.menuCursor)
	}

	m2 := &modelImpl{state: stateMenu, menuCursor: 1}
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyUp})
	after2 := updated2.(*modelImpl)
	if after2.menuCursor != 0 {
		t.Errorf("expected cursor to move up on arrow up, got %d", after2.menuCursor)
	}
}

func TestMenuEnterRunReviewTransitionsToReviewing(t *testing.T) {
	m := &modelImpl{state: stateMenu, menuCursor: 0}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	after := updated.(*modelImpl)

	if after.state != stateReviewing {
		t.Errorf("expected state reviewing after selecting Run Review, got %v", after.state)
	}
	if cmd == nil {
		t.Error("expected command for review")
	}
}

func TestMenuEnterQuitReturnsQuit(t *testing.T) {
	quitIdx := len(menuItems) - 1
	m := &modelImpl{state: stateMenu, menuCursor: quitIdx}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected quit command on Quit select")
	}
	_ = updated
}

func TestFindingsQuitOnQ(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "test-1", Title: "Test Finding", Severity: "low", Path: "test.go", Line: 1},
		},
	}
	m := &modelImpl{state: stateFindings, report: &r}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected quit command on 'q' in findings")
	}
}

func TestFindingsBackToMenu(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "test-1", Title: "Test Finding", Severity: "low", Path: "test.go", Line: 1},
		},
	}
	m := &modelImpl{state: stateFindings, report: &r}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	after := updated.(*modelImpl)

	if after.state != stateMenu {
		t.Errorf("expected state menu after escape in findings, got %v", after.state)
	}
	if after.report != nil {
		t.Error("expected report to be cleared on back to menu")
	}
}

func TestFindingsBackToMenuWithB(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "test-1", Title: "Test Finding", Severity: "low", Path: "test.go", Line: 1},
		},
	}
	m := &modelImpl{state: stateFindings, report: &r}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	after := updated.(*modelImpl)

	if after.state != stateMenu {
		t.Errorf("expected state menu after 'b' in findings, got %v", after.state)
	}
}

func TestFindingsNavigation(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "f1", Title: "Finding 1", Severity: "low", Path: "a.go", Line: 1},
			{ID: "f2", Title: "Finding 2", Severity: "high", Path: "b.go", Line: 10},
			{ID: "f3", Title: "Finding 3", Severity: "critical", Path: "c.go", Line: 20},
		},
	}
	m := &modelImpl{state: stateFindings, report: &r, findingsCursor: 0}

	// Down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	after := updated.(*modelImpl)
	if after.findingsCursor != 1 {
		t.Errorf("expected cursor 1 after down, got %d", after.findingsCursor)
	}

	// Down again
	updated2, _ := after.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	after2 := updated2.(*modelImpl)
	if after2.findingsCursor != 2 {
		t.Errorf("expected cursor 2 after second down, got %d", after2.findingsCursor)
	}

	// Down at end stays at end
	updated3, _ := after2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	after3 := updated3.(*modelImpl)
	if after3.findingsCursor != 2 {
		t.Errorf("expected cursor to stay at end, got %d", after3.findingsCursor)
	}

	// Up
	updated4, _ := after2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	after4 := updated4.(*modelImpl)
	if after4.findingsCursor != 1 {
		t.Errorf("expected cursor 1 after up, got %d", after4.findingsCursor)
	}

	// Up at top stays at top
	m2 := &modelImpl{state: stateFindings, report: &r, findingsCursor: 0}
	updated5, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	after5 := updated5.(*modelImpl)
	if after5.findingsCursor != 0 {
		t.Errorf("expected cursor to stay at top, got %d", after5.findingsCursor)
	}
}

func TestHomeEndNavigation(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "f1", Title: "F1", Severity: "low", Path: "a.go", Line: 1},
			{ID: "f2", Title: "F2", Severity: "medium", Path: "b.go", Line: 2},
			{ID: "f3", Title: "F3", Severity: "high", Path: "c.go", Line: 3},
		},
	}
	m := &modelImpl{state: stateFindings, report: &r, findingsCursor: 1}

	// Home
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	after := updated.(*modelImpl)
	if after.findingsCursor != 0 {
		t.Errorf("expected cursor 0 after home, got %d", after.findingsCursor)
	}

	// End
	updated2, _ := after.Update(tea.KeyMsg{Type: tea.KeyEnd})
	after2 := updated2.(*modelImpl)
	if after2.findingsCursor != 2 {
		t.Errorf("expected cursor at end (2), got %d", after2.findingsCursor)
	}
}

func TestDetailViewEnterAndBack(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "f1", Title: "Finding 1", Severity: "critical", Category: "auth", Confidence: "high",
				Path: "src/auth.go", Line: 42, Side: "added", Message: "Missing auth check.",
				Snippet: "func handle() {", Remediation: "Add auth middleware.", Reasons: []string{"No auth found"}},
		},
	}

	// Test Enter transitions to detail
	m1 := &modelImpl{state: stateFindings, report: &r, findingsCursor: 0}
	updated, _ := m1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	after := updated.(*modelImpl)
	if after.state != stateDetail {
		t.Errorf("expected state detail after enter, got %v", after.state)
	}
	if after.detailIndex != 0 {
		t.Errorf("expected detailIndex 0, got %d", after.detailIndex)
	}

	// Test 'b' transitions back from detail to findings
	m2 := &modelImpl{state: stateDetail, report: &r, detailIndex: 0}
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	after2 := updated2.(*modelImpl)
	if after2.state != stateFindings {
		t.Errorf("expected state findings after 'b' in detail, got %v", after2.state)
	}

	// Test escape transitions back from detail to findings
	m3 := &modelImpl{state: stateDetail, report: &r, detailIndex: 0}
	updated3, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEsc})
	after3 := updated3.(*modelImpl)
	if after3.state != stateFindings {
		t.Errorf("expected state findings after escape in detail, got %v", after3.state)
	}
}

func TestDetailQuitOnQ(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "f1", Title: "F1", Severity: "low", Path: "a.go", Line: 1},
		},
	}
	m := &modelImpl{state: stateDetail, report: &r, detailIndex: 0}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected quit command on 'q' in detail")
	}
}

func TestReviewMsgErrorReturnsToMenu(t *testing.T) {
	m := &modelImpl{state: stateReviewing}
	updated, _ := m.Update(reviewMsg{err: fmt.Errorf("git not found")})
	after := updated.(*modelImpl)

	if after.state != stateMenu {
		t.Errorf("expected state menu on review error, got %v", after.state)
	}
	if after.reviewErr == nil {
		t.Error("expected review error to be set")
	}
}

func TestReviewMsgSuccessTransitionsToFindings(t *testing.T) {
	r := model.Report{
		Meta:    model.DiffMeta{Mode: "staged"},
		Summary: model.Summary{Findings: 1, Low: 1},
		Findings: []model.Finding{
			{ID: "f1", Title: "Test Finding", Severity: "low", Path: "test.go", Line: 1},
		},
	}
	m := &modelImpl{state: stateReviewing}
	updated, _ := m.Update(reviewMsg{report: r})
	after := updated.(*modelImpl)

	if after.state != stateFindings {
		t.Errorf("expected state findings on success, got %v", after.state)
	}
	if after.report == nil {
		t.Error("expected report to be set")
	}
}

func TestRenderFindingsList(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{ID: "f1", Title: "Critical Auth Bypass", Severity: "critical", Path: "src/auth/login.go", Line: 15},
			{ID: "f2", Title: "Hardcoded Secret", Severity: "high", Path: "src/config.go", Line: 3},
			{ID: "f3", Title: "Weak Hash Algorithm", Severity: "medium", Path: "src/crypto/hash.go", Line: 42},
			{ID: "f4", Title: "Missing Input Validation", Severity: "low", Path: "src/api/handler.go", Line: 100},
		},
	}
	m := &modelImpl{state: stateFindings, report: &r, findingsCursor: 0}
	m.viewport = viewport.New(80, 20)

	content := m.renderFindingsList()

	for _, f := range r.Findings {
		if !strings.Contains(content, f.Title) {
			t.Errorf("findings list missing finding Title: %s", f.Title)
		}
		if !strings.Contains(content, f.Path) {
			t.Errorf("findings list missing finding Path: %s", f.Path)
		}
	}
}

func TestRenderFindingsListEmpty(t *testing.T) {
	r := model.Report{Findings: nil}
	m := &modelImpl{state: stateFindings, report: &r}

	content := m.renderFindingsList()
	if content == "" {
		t.Error("expected non-empty content for empty findings")
	}
}

func TestRenderDetail(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{
				ID: "mavetis-auth-001", RuleID: "mavetis-auth-001",
				Title: "Missing Authentication Check",
				Category: "auth", Severity: "critical", Confidence: "high",
				Path: "src/api/handler.go", Line: 42, Side: "added",
				Message:     "No authentication middleware detected on this endpoint.",
				Snippet:     "func GetUsers(w http.ResponseWriter, r *http.Request) {",
				Remediation: "Add authentication middleware to this route.",
				Reasons:     []string{"No auth middleware found in handler chain", "Session validation absent"},
			},
		},
	}
	m := &modelImpl{state: stateDetail, report: &r, detailIndex: 0}

	content := m.renderDetail()

	checks := []string{
		"Missing Authentication Check",
		"mavetis-auth-001",
		"auth",
		"critical",
		"high",
		"src/api/handler.go",
		"42",
		"No authentication middleware detected",
		"Add authentication middleware",
		"No auth middleware found",
	}

	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("detail view missing: %s", check)
		}
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := &modelImpl{state: stateMenu}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	after := updated.(*modelImpl)

	if after.width != 100 {
		t.Errorf("expected width 100, got %d", after.width)
	}
	if after.height != 40 {
		t.Errorf("expected height 40, got %d", after.height)
	}
	if !after.ready {
		t.Error("expected ready after WindowSizeMsg")
	}
	if after.viewport.Width != 96 {
		t.Errorf("expected viewport width 96 (100-4), got %d", after.viewport.Width)
	}
}

func TestReviewingStateIgnoresKeypresses(t *testing.T) {
	m := &modelImpl{state: stateReviewing, menuCursor: 0}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	after := updated.(*modelImpl)

	if after.state != stateReviewing {
		t.Error("expected state to remain reviewing on keypress during review")
	}
}

func TestSeverityBadge(t *testing.T) {
	tests := []struct {
		severity string
		contains string
	}{
		{"critical", "critical"},
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"info", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			result := severityBadge(tt.severity)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("severityBadge(%q) missing %q in output: %s", tt.severity, tt.contains, result)
			}
		})
	}
}

func TestNewModel(t *testing.T) {
	m := NewModel()
	if m == nil {
		t.Fatal("NewModel() returned nil")
	}
	impl, ok := m.(*modelImpl)
	if !ok {
		t.Fatal("NewModel() did not return *modelImpl")
	}
	if impl.state != stateMenu {
		t.Errorf("expected state menu, got %v", impl.state)
	}
}

func TestMenuViewContainsItems(t *testing.T) {
	m := &modelImpl{state: stateMenu, ready: true, width: 80, height: 30}
	view := m.View()
	for _, item := range menuItems {
		if !strings.Contains(view, item) {
			t.Errorf("menu view missing item: %s", item)
		}
	}
}

func TestFindingsViewEmptyReport(t *testing.T) {
	r := model.Report{}
	m := &modelImpl{state: stateFindings, report: &r, ready: true}
	m.viewport = viewport.New(80, 20)
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view for empty report")
	}
}

func TestDetailViewRendersCorrectly(t *testing.T) {
	r := model.Report{
		Findings: []model.Finding{
			{
				ID: "f1", Title: "Detail Test", RuleID: "rule-001",
				Category: "injection", Severity: "high", Confidence: "medium",
				Path: "src/main.go", Line: 10, Side: "deleted",
				Message: "SQL injection risk detected.", Snippet: "query := \"SELECT * FROM \" + input",
				Remediation: "Use parameterized queries.", Reasons: []string{"String concatenation in SQL"},
			},
		},
	}
	m := &modelImpl{state: stateDetail, report: &r, detailIndex: 0, ready: true}
	m.detailViewport = viewport.New(80, 20)
	view := m.View()
	if !strings.Contains(view, "Finding Detail") {
		t.Error("detail view missing header")
	}
}

func TestErrNoTerminal(t *testing.T) {
	err := ErrNoTerminal
	if err.Error() == "" {
		t.Error("ErrNoTerminal should have a message")
	}
}

func TestIsTerminalNonTTY(t *testing.T) {
	if isTerminal(0) {
		t.Log("stdin is a TTY (expected in interactive mode)")
	}
}

func TestMenuEnterAboutTransitionsToAbout(t *testing.T) {
	m := &modelImpl{state: stateMenu, menuCursor: 1}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	after := updated.(*modelImpl)

	if after.state != stateAbout {
		t.Errorf("expected state about after selecting About, got %v", after.state)
	}
}

func TestAboutKeyEscReturnsToMenu(t *testing.T) {
	m := &modelImpl{state: stateAbout}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	after := updated.(*modelImpl)

	if after.state != stateMenu {
		t.Errorf("expected state menu after esc in about, got %v", after.state)
	}
}

func TestAboutKeyBReturnsToMenu(t *testing.T) {
	m := &modelImpl{state: stateAbout}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	after := updated.(*modelImpl)

	if after.state != stateMenu {
		t.Errorf("expected state menu after 'b' in about, got %v", after.state)
	}
}

func TestAboutKeyEnterReturnsToMenu(t *testing.T) {
	m := &modelImpl{state: stateAbout}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	after := updated.(*modelImpl)

	if after.state != stateMenu {
		t.Errorf("expected state menu after enter in about, got %v", after.state)
	}
}

func TestAboutQuitOnQ(t *testing.T) {
	m := &modelImpl{state: stateAbout}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected quit command on 'q' in about")
	}
}

func TestAboutViewContainsContent(t *testing.T) {
	m := &modelImpl{state: stateAbout, ready: true, width: 80, height: 30}
	view := m.View()
	checks := []string{"mavetis", "0.2.0", "Commands", "review", "init", "secrets", "Links"}
	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("about view missing: %s", check)
		}
	}
}

func TestFillScreenPadsProperly(t *testing.T) {
	m := &modelImpl{width: 80, height: 20}
	result := m.fillScreen("hello", "footer")
	if result == "" {
		t.Error("fillScreen should return non-empty string")
	}
	if !strings.Contains(result, "hello") {
		t.Error("fillScreen missing content")
	}
	if !strings.Contains(result, "footer") {
		t.Error("fillScreen missing footer")
	}
}

func TestViewForAboutState(t *testing.T) {
	m := &modelImpl{state: stateAbout, ready: true, width: 80, height: 30}
	view := m.View()
	if !strings.Contains(view, "mavetis") {
		t.Error("about view missing title")
	}
}

func TestTallScreenDoesNotCrash(t *testing.T) {
	m := &modelImpl{state: stateMenu, ready: true, width: 80, height: 10}
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view for small screen")
	}
}

func TestFillScreenNoFooter(t *testing.T) {
	m := &modelImpl{width: 80, height: 20}
	result := m.fillScreen("content only", "")
	if !strings.Contains(result, "content only") {
		t.Error("fillScreen without footer should still contain content")
	}
}
