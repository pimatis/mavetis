package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Pimatis/mavetis/src/engine"
	"github.com/Pimatis/mavetis/src/git"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/wizard"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateMenu state = iota
	stateReviewing
	stateFindings
	stateDetail
	stateAbout
	stateSecretsScanning
	stateSecretsFindings
	stateBaselineScanning
	stateBaselineDone
	stateRuleList
	stateRuleDetail
	stateWizard
	stateWizardDone
)

type reviewMode int

const (
	reviewStaged reviewMode = iota
	reviewAllFiles
	reviewSecrets
	reviewBaseline
	reviewWizard
)

type reviewMsg struct {
	report model.Report
	err    error
}

type modelImpl struct {
	state  state
	ready  bool
	width  int
	height int

	// Menu
	menuCursor int

	// Review
	reviewMode reviewMode
	spinner    spinner.Model
	reviewErr  error
	report     *model.Report

	// Findings list
	findingsCursor int
	viewport       viewport.Model

	// Detail
	detailViewport viewport.Model
	detailIndex    int

	// Rules
	ruleList   []model.RuleInfo
	ruleCursor int
	ruleDetail string

	// Wizard
	wizardProject wizard.Project
	wizardDone    bool
	wizardPath    string

	// Baseline
	baselinePath string
	baselineDone bool
}

var menuItems = []string{
	"Review Staged Changes",
	"Review All Files",
	"Secrets Scan",
	"Create Baseline",
	"Rule Explain",
	"Config Wizard",
	"About",
	"Quit",
}

func (m *modelImpl) Init() tea.Cmd {
	return nil
}

func (m *modelImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-8)
			m.detailViewport = viewport.New(msg.Width-4, msg.Height-8)
			m.ready = true
		}
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
		m.detailViewport.Width = msg.Width - 4
		m.detailViewport.Height = msg.Height - 8
		return m, nil

	case tea.KeyMsg:
		if m.state == stateReviewing {
			return m, nil
		}
		return m.handleKey(msg)

	case spinner.TickMsg:
		if m.state == stateReviewing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case reviewMsg:
		if msg.err != nil {
			m.reviewErr = msg.err
			m.state = stateMenu
			return m, nil
		}
		if m.state == stateBaselineScanning {
			if err := baselineCreate(".mavetis-baseline.yaml", msg.report); err != nil {
				m.reviewErr = err
				m.state = stateMenu
				return m, nil
			}
			m.baselinePath = ".mavetis-baseline.yaml"
			m.baselineDone = true
			m.state = stateBaselineDone
			return m, nil
		}
		m.report = &msg.report
		m.findingsCursor = 0
		m.viewport.SetContent(m.renderFindingsList())
		m.viewport.GotoTop()
		if m.state == stateReviewing {
			m.state = stateFindings
		} else if m.state == stateSecretsScanning {
			m.state = stateSecretsFindings
		}
		return m, nil
	}

	cmds = append(cmds, nil)
	return m, tea.Batch(cmds...)
}

func (m *modelImpl) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == stateDetail {
		return m.handleDetailKey(msg)
	}
	if m.state == stateFindings || m.state == stateSecretsFindings {
		return m.handleFindingsKey(msg)
	}
	if m.state == stateAbout {
		return m.handleAboutKey(msg)
	}
	if m.state == stateRuleList {
		return m.handleRuleListKey(msg)
	}
	if m.state == stateRuleDetail {
		return m.handleRuleDetailKey(msg)
	}
	if m.state == stateBaselineDone || m.state == stateWizardDone {
		return m.handleDoneKey(msg)
	}
	if m.state == stateWizard {
		return m.handleWizardKey(msg)
	}
	return m.handleMenuKey(msg)
}

func (m *modelImpl) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.menuCursor--
		if m.menuCursor < 0 {
			m.menuCursor = len(menuItems) - 1
		}

	case "down", "j":
		m.menuCursor++
		if m.menuCursor >= len(menuItems) {
			m.menuCursor = 0
		}

	case "enter":
		return m.handleMenuSelect()
	}

	return m, nil
}

func (m *modelImpl) handleMenuSelect() (tea.Model, tea.Cmd) {
	choice := menuItems[m.menuCursor]

	if strings.HasPrefix(choice, "Review Staged Changes") {
		m.state = stateReviewing
		m.reviewMode = reviewStaged
		m.spinner = spinner.New()
		m.spinner.Spinner = spinner.Dot
		m.spinner.Style = spinnerStyle
		return m, tea.Batch(m.spinner.Tick, m.runReviewCmd())
	}

	if strings.HasPrefix(choice, "Review All Files") {
		m.state = stateReviewing
		m.reviewMode = reviewAllFiles
		m.spinner = spinner.New()
		m.spinner.Spinner = spinner.Dot
		m.spinner.Style = spinnerStyle
		return m, tea.Batch(m.spinner.Tick, m.runReviewAllCmd())
	}

	if strings.HasPrefix(choice, "Secrets Scan") {
		m.state = stateSecretsScanning
		m.reviewMode = reviewSecrets
		m.spinner = spinner.New()
		m.spinner.Spinner = spinner.Dot
		m.spinner.Style = spinnerStyle
		return m, tea.Batch(m.spinner.Tick, m.runSecretsScanCmd())
	}

	if strings.HasPrefix(choice, "Create Baseline") {
		m.state = stateBaselineScanning
		m.reviewMode = reviewBaseline
		m.spinner = spinner.New()
		m.spinner.Spinner = spinner.Dot
		m.spinner.Style = spinnerStyle
		return m, tea.Batch(m.spinner.Tick, m.runBaselineScanCmd())
	}

	if strings.HasPrefix(choice, "Rule Explain") {
		cfg, err := loadConfig("")
		if err != nil {
			m.reviewErr = err
			return m, nil
		}
		m.ruleList = ruleList(cfg)
		m.ruleCursor = 0
		m.state = stateRuleList
		m.viewport.SetContent(m.renderRuleList())
		m.viewport.GotoTop()
		return m, nil
	}

	if strings.HasPrefix(choice, "Config Wizard") {
		root, err := scanRoot()
		if err != nil {
			m.reviewErr = err
			return m, nil
		}
		m.wizardProject = wizardDetect(root)
		m.wizardDone = false
		m.state = stateWizard
		return m, nil
	}

	if strings.HasPrefix(choice, "About") {
		m.state = stateAbout
		return m, nil
	}

	if strings.HasPrefix(choice, "Quit") {
		return m, tea.Quit
	}

	return m, nil
}

func (m *modelImpl) handleFindingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "b":
		m.state = stateMenu
		m.report = nil
		m.reviewErr = nil
		return m, nil

	case "up", "k":
		m.moveFindingsCursor(-1)

	case "down", "j":
		m.moveFindingsCursor(1)

	case "home":
		m.findingsCursor = 0
		m.viewport.GotoTop()
		m.viewport.SetContent(m.renderFindingsList())

	case "end":
		m.findingsCursor = len(m.report.Findings) - 1
		m.scrollToBottom()
		m.viewport.SetContent(m.renderFindingsList())

	case "enter":
		m.state = stateDetail
		m.detailIndex = m.findingsCursor
		m.detailViewport.SetContent(m.renderDetail())
		m.detailViewport.GotoTop()
		return m, nil

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *modelImpl) moveFindingsCursor(delta int) {
	count := len(m.report.Findings)
	if count == 0 {
		return
	}
	prev := m.findingsCursor
	m.findingsCursor += delta
	if m.findingsCursor < 0 {
		m.findingsCursor = 0
	}
	if m.findingsCursor >= count {
		m.findingsCursor = count - 1
	}
	if m.findingsCursor == prev {
		return
	}

	m.viewport.SetContent(m.renderFindingsList())

	cursorLine := m.findingsCursor
	visibleStart := m.viewport.YOffset
	visibleEnd := m.viewport.YOffset + m.viewport.Height

	if cursorLine < visibleStart {
		m.viewport.YOffset = cursorLine
	}
	if cursorLine >= visibleEnd {
		m.viewport.YOffset = cursorLine - m.viewport.Height + 1
	}
	if m.viewport.YOffset < 0 {
		m.viewport.YOffset = 0
	}
}

func (m *modelImpl) scrollToBottom() {
	findings := len(m.report.Findings)
	if findings > 0 {
		m.viewport.YOffset = findings - m.viewport.Height + 1
		if m.viewport.YOffset < 0 {
			m.viewport.YOffset = 0
		}
	}
}

func (m *modelImpl) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "b":
		m.state = stateFindings
		return m, nil

	default:
		var cmd tea.Cmd
		m.detailViewport, cmd = m.detailViewport.Update(msg)
		return m, cmd
	}
}

func (m *modelImpl) handleAboutKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "b", "enter":
		m.state = stateMenu
		return m, nil
	}
	return m, nil
}

func (m *modelImpl) handleRuleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "b":
		m.state = stateMenu
		m.ruleList = nil
		return m, nil

	case "up", "k":
		m.ruleCursor--
		if m.ruleCursor < 0 {
			m.ruleCursor = len(m.ruleList) - 1
		}
		m.viewport.SetContent(m.renderRuleList())

	case "down", "j":
		m.ruleCursor++
		if m.ruleCursor >= len(m.ruleList) {
			m.ruleCursor = 0
		}
		m.viewport.SetContent(m.renderRuleList())

	case "enter":
		if len(m.ruleList) == 0 {
			return m, nil
		}
		cfg, err := loadConfig("")
		if err != nil {
			m.reviewErr = err
			return m, nil
		}
		item := m.ruleList[m.ruleCursor]
		explanation, ok := ruleExplain(item.ID, allRulesFor(cfg))
		if !ok {
			m.ruleDetail = "Rule not found: " + item.ID
		} else {
			m.ruleDetail = outputExplain(explanation)
		}
		m.state = stateRuleDetail
		m.detailViewport.SetContent(m.ruleDetail)
		m.detailViewport.GotoTop()
		return m, nil

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *modelImpl) handleRuleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "b":
		m.state = stateRuleList
		return m, nil

	default:
		var cmd tea.Cmd
		m.detailViewport, cmd = m.detailViewport.Update(msg)
		return m, cmd
	}
}

func (m *modelImpl) handleDoneKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "b", "enter":
		m.state = stateMenu
		m.reviewErr = nil
		return m, nil
	}
	return m, nil
}

func (m *modelImpl) handleWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "b":
		m.state = stateMenu
		return m, nil

	case "enter":
		root, err := scanRoot()
		if err != nil {
			m.reviewErr = err
			m.state = stateMenu
			return m, nil
		}
		path := ".mavetis.yaml"
		if _, err := os.Stat(path); err == nil {
			m.reviewErr = fmt.Errorf("%s already exists; remove it first or use the CLI with --force", path)
			m.state = stateMenu
			return m, nil
		}
		project := wizardDetect(root)
		template := wizard.ConfigTemplate{
			Profile:    project.Profile,
			Severity:   "low",
			FailOn:     "high",
			Output:     "text",
			Ignore:     project.Ignore,
			Critical:   project.Critical,
			Restricted: project.Restricted,
		}
		content := wizardGenerate(template)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			m.reviewErr = err
			m.state = stateMenu
			return m, nil
		}
		_ = appendGitignore(root, ".mavetis.yaml")
		m.wizardPath = path
		m.wizardDone = true
		m.state = stateWizardDone
		return m, nil
	}
	return m, nil
}

// fillScreen wraps content with vertical/horizontal padding to fill the terminal.
func (m *modelImpl) fillScreen(content string, footer string) string {
	contentLines := strings.Count(content, "\n") + 1
	footerLines := 1
	if footer != "" {
		footerLines = strings.Count(footer, "\n") + 1
	}
	totalContent := contentLines + footerLines
	padTop := 0
	if totalContent < m.height {
		padTop = (m.height - totalContent) / 2
	}
	padding := strings.Repeat("\n", padTop)

	body := padding + content
	if footer != "" {
		body = body + "\n" + footer
	}

	return lipgloss.NewStyle().Width(m.width).Render(body)
}

func (m *modelImpl) runReviewCmd() tea.Cmd {
	return func() tea.Msg {
		raw, meta, err := git.Review(model.Review{Staged: true})
		if err != nil {
			return reviewMsg{err: err}
		}
		parsed, err := diffParse(raw, meta)
		if err != nil {
			return reviewMsg{err: err}
		}
		cfg, err := loadConfig("")
		if err != nil {
			return reviewMsg{err: err}
		}
		rules := allRulesFor(cfg)
		report, err := engine.Review(parsed, cfg, rules)
		if err != nil {
			return reviewMsg{err: err}
		}
		report.Meta = meta
		if cfg.Baseline.Path != "" {
			baselineFile, err := baselineLoad(cfg.Baseline.Path)
			if err != nil {
				return reviewMsg{err: err}
			}
			report = baselineFilter(report, baselineFile)
		}
		score := riskCalculate(report.Summary)
		report.Score = &model.Score{Value: score.Value, Rating: score.Rating}
		return reviewMsg{report: report}
	}
}

func (m *modelImpl) runReviewAllCmd() tea.Cmd {
	return func() tea.Msg {
		root, err := scanRoot()
		if err != nil {
			return reviewMsg{err: err}
		}
		files, err := loadAllFiles(root)
		if err != nil {
			return reviewMsg{err: err}
		}
		if len(files) == 0 {
			return reviewMsg{err: fmt.Errorf("no files found in repository")}
		}
		diff := fromFiles(files)
		cfg, err := loadConfig("")
		if err != nil {
			return reviewMsg{err: err}
		}
		rules := allRulesFor(cfg)
		report, err := engine.Review(diff, cfg, rules)
		if err != nil {
			return reviewMsg{err: err}
		}
		if cfg.Baseline.Path != "" {
			baselineFile, err := baselineLoad(cfg.Baseline.Path)
			if err != nil {
				return reviewMsg{err: err}
			}
			report = baselineFilter(report, baselineFile)
		}
		score := riskCalculate(report.Summary)
		report.Score = &model.Score{Value: score.Value, Rating: score.Rating}
		return reviewMsg{report: report}
	}
}

func (m *modelImpl) runSecretsScanCmd() tea.Cmd {
	return func() tea.Msg {
		root, err := scanRoot()
		if err != nil {
			return reviewMsg{err: err}
		}
		cfg, err := loadConfig("")
		if err != nil {
			return reviewMsg{err: err}
		}
		report, err := secretScan(root, cfg)
		if err != nil {
			return reviewMsg{err: err}
		}
		score := riskCalculate(report.Summary)
		report.Score = &model.Score{Value: score.Value, Rating: score.Rating}
		return reviewMsg{report: report}
	}
}

func (m *modelImpl) runBaselineScanCmd() tea.Cmd {
	return func() tea.Msg {
		root, err := scanRoot()
		if err != nil {
			return reviewMsg{err: err}
		}
		files, err := loadAllFiles(root)
		if err != nil {
			return reviewMsg{err: err}
		}
		if len(files) == 0 {
			return reviewMsg{err: fmt.Errorf("no files found in repository")}
		}
		diff := fromFiles(files)
		cfg, err := loadConfig("")
		if err != nil {
			return reviewMsg{err: err}
		}
		rules := allRulesFor(cfg)
		report, err := engine.Review(diff, cfg, rules)
		if err != nil {
			return reviewMsg{err: err}
		}
		score := riskCalculate(report.Summary)
		report.Score = &model.Score{Value: score.Value, Rating: score.Rating}
		return reviewMsg{report: report}
	}
}

func (m *modelImpl) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var content string
	var footer string

	switch m.state {
	case stateReviewing, stateSecretsScanning, stateBaselineScanning:
		content = m.renderReviewing()
		footer = ""
	case stateFindings, stateSecretsFindings:
		return m.renderFindings()
	case stateDetail:
		return m.renderDetailView()
	case stateAbout:
		content = m.renderAbout()
		footer = helpStyle.Render("esc/b enter  back  q quit")
	case stateRuleList:
		return m.renderRuleListView()
	case stateRuleDetail:
		return m.renderRuleDetailView()
	case stateBaselineDone:
		content = m.renderBaselineDone()
		footer = helpStyle.Render("esc/b enter  back  q quit")
	case stateWizard:
		content = m.renderWizard()
		footer = helpStyle.Render("esc/b enter  back  q quit")
	case stateWizardDone:
		content = m.renderWizardDone()
		footer = helpStyle.Render("esc/b enter  back  q quit")
	default:
		content = m.renderMenu()
		footer = helpStyle.Render("↑↓ navigate  ↵ select  q quit")
	}

	return m.fillScreen(content, footer)
}

func (m *modelImpl) renderMenu() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("mavetis") + "  " + helpStyle.Render("security review"))
	b.WriteString("\n\n")

	for i, item := range menuItems {
		if i == m.menuCursor {
			b.WriteString(menuSelectedStyle.Render(" > " + item))
		} else {
			b.WriteString(menuItemStyle.Render("   " + item))
		}
		b.WriteString("\n")
	}

	if m.reviewErr != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.reviewErr.Error()))
	}

	return b.String()
}

func (m *modelImpl) renderReviewing() string {
	var label string
	switch m.reviewMode {
	case reviewAllFiles:
		label = "Running security review on all files..."
	case reviewSecrets:
		label = "Scanning for secrets..."
	case reviewBaseline:
		label = "Creating baseline from all files..."
	default:
		label = "Running security review on staged changes..."
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("mavetis") + "  " + helpStyle.Render("reviewing..."))
	b.WriteString("\n\n\n")
	b.WriteString("   " + m.spinner.View() + " " + label)
	return b.String()
}

func (m *modelImpl) renderAbout() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("mavetis") + "  " + helpStyle.Render(model.Version))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("local-first security review tool"))
	b.WriteString("\n")
	b.WriteString(findingPathStyle.Render("  " + model.Repository))
	b.WriteString("\n\n")

	b.WriteString(sectionStyle.Render("Commands"))
	b.WriteString("\n")
	cmdItems := []string{
		"review --staged    review staged git changes",
		"review --base      review branch diff",
		"review src/        review specific files",
		"ci                 CI mode with stricter defaults",
		"init               project setup wizard",
		"baseline --create  capture known findings",
		"secrets scan       full filesystem secret scan",
		"hooks install      install git pre-commit hook",
		"rules explain      show rule documentation",
	}
	for _, item := range cmdItems {
		b.WriteString(helpStyle.Render("   " + item))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Links"))
	b.WriteString("\n")
	b.WriteString(findingPathStyle.Render("  " + model.Repository))
	b.WriteString("\n")

	return b.String()
}

func (m *modelImpl) renderFindings() string {
	header := m.renderFindingsHeader()
	content := lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View())

	footer := helpStyle.Render("↑↓ navigate  ↵ detail  b back  q quit")
	return lipgloss.JoinVertical(lipgloss.Left, content, "\n"+footer)
}

func (m *modelImpl) renderFindingsHeader() string {
	if m.report == nil {
		return ""
	}
	s := m.report.Summary
	summary := fmt.Sprintf("Files: %d  Findings: %d  C:%d H:%d M:%d L:%d",
		s.Files, s.Findings, s.Critical, s.High, s.Medium, s.Low)
	return lipgloss.NewStyle().Padding(0, 2).Render(titleStyle.Render("Findings") + "  " + helpStyle.Render(summary))
}

func (m *modelImpl) renderFindingsList() string {
	if m.report == nil || len(m.report.Findings) == 0 {
		return helpStyle.Render("No findings detected.")
	}

	var lines []string
	for i, f := range m.report.Findings {
		cursor := "  "
		if i == m.findingsCursor {
			cursor = "> "
		}

		badge := severityBadge(f.Severity)
		location := findingPathStyle.Render(f.Path) + findingLineStyle.Render(fmt.Sprintf(":%d", f.Line))
		title := lipgloss.NewStyle().Padding(0, 1).Render(f.Title)

		row := fmt.Sprintf("%s%s %s %s", cursor, badge, title, location)
		row = lipgloss.NewStyle().MaxWidth(m.viewport.Width - 4).Render(row)

		if i == m.findingsCursor {
			row = lipgloss.NewStyle().
				Background(lipgloss.Color("#333333")).
				Render(row)
		}
		lines = append(lines, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *modelImpl) renderDetailView() string {
	header := lipgloss.NewStyle().Padding(0, 2).Render(titleStyle.Render("Finding Detail"))
	content := lipgloss.JoinVertical(lipgloss.Left, header, m.detailViewport.View())

	footer := helpStyle.Render("↑↓ scroll  b back  q quit")
	return lipgloss.JoinVertical(lipgloss.Left, content, "\n"+footer)
}

func (m *modelImpl) renderDetail() string {
	if m.report == nil || m.detailIndex >= len(m.report.Findings) {
		return ""
	}
	f := m.report.Findings[m.detailIndex]

	var b strings.Builder
	pad := lipgloss.NewStyle().PaddingLeft(2)

	b.WriteString(pad.Render(severityBadge(f.Severity) + " " + f.Title))
	b.WriteString("\n\n")

	b.WriteString(pad.Render(sectionStyle.Render("Details")))
	b.WriteString("\n")

	detailPairs := [][2]string{
		{"Rule", f.RuleID},
		{"Category", f.Category},
		{"Confidence", f.Confidence},
		{"Path", f.Path},
		{"Line", fmt.Sprintf("%d", f.Line)},
		{"Side", f.Side},
	}

	for _, pair := range detailPairs {
		label := helpStyle.Render(pair[0] + ":")
		value := lipgloss.NewStyle().PaddingLeft(1).Render(pair[1])
		b.WriteString(pad.Render(label + value))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(pad.Render(sectionStyle.Render("Message")))
	b.WriteString("\n")
	b.WriteString(pad.Render(lipgloss.NewStyle().PaddingLeft(2).Render(f.Message)))
	b.WriteString("\n\n")

	b.WriteString(pad.Render(sectionStyle.Render("Snippet")))
	b.WriteString("\n")
	b.WriteString(pad.Render(lipgloss.NewStyle().PaddingLeft(2).Foreground(subtle).Render(f.Snippet)))
	b.WriteString("\n\n")

	b.WriteString(pad.Render(sectionStyle.Render("Remediation")))
	b.WriteString("\n")
	b.WriteString(pad.Render(lipgloss.NewStyle().PaddingLeft(2).Render(f.Remediation)))
	b.WriteString("\n\n")

	if len(f.Reasons) > 0 {
		b.WriteString(pad.Render(sectionStyle.Render("Reasons")))
		b.WriteString("\n")
		for _, reason := range f.Reasons {
			b.WriteString(pad.Render(lipgloss.NewStyle().PaddingLeft(4).Render("• " + reason)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	return b.String()
}

func (m *modelImpl) renderRuleListView() string {
	header := lipgloss.NewStyle().Padding(0, 2).Render(titleStyle.Render("Rules") + "  " + helpStyle.Render(fmt.Sprintf("%d rules", len(m.ruleList))))
	content := lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View())
	footer := helpStyle.Render("↑↓ navigate  ↵ explain  b back  q quit")
	return lipgloss.JoinVertical(lipgloss.Left, content, "\n"+footer)
}

func (m *modelImpl) renderRuleList() string {
	if len(m.ruleList) == 0 {
		return helpStyle.Render("No rules available.")
	}
	var lines []string
	for i, r := range m.ruleList {
		cursor := "  "
		if i == m.ruleCursor {
			cursor = "> "
		}
		badge := severityBadge(r.Severity)
		row := fmt.Sprintf("%s%s %s %s", cursor, badge, r.ID, r.Title)
		row = lipgloss.NewStyle().MaxWidth(m.viewport.Width - 4).Render(row)
		if i == m.ruleCursor {
			row = lipgloss.NewStyle().
				Background(lipgloss.Color("#333333")).
				Render(row)
		}
		lines = append(lines, row)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *modelImpl) renderRuleDetailView() string {
	header := lipgloss.NewStyle().Padding(0, 2).Render(titleStyle.Render("Rule Detail"))
	content := lipgloss.JoinVertical(lipgloss.Left, header, m.detailViewport.View())
	footer := helpStyle.Render("↑↓ scroll  b back  q quit")
	return lipgloss.JoinVertical(lipgloss.Left, content, "\n"+footer)
}

func (m *modelImpl) renderBaselineDone() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("mavetis") + "  " + helpStyle.Render("baseline created"))
	b.WriteString("\n\n")
	if m.baselineDone && m.report != nil {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("Baseline saved to: " + m.baselinePath))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(fmt.Sprintf("Findings captured: %d", len(m.report.Findings))))
	} else {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("Baseline creation failed."))
	}
	if m.reviewErr != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.reviewErr.Error()))
	}
	return b.String()
}

func (m *modelImpl) renderWizard() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("mavetis") + "  " + helpStyle.Render("config wizard"))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("Detected project profile: " + m.wizardProject.Profile))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("Language: " + m.wizardProject.Language))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("Press Enter to generate .mavetis.yaml"))
	return b.String()
}

func (m *modelImpl) renderWizardDone() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("mavetis") + "  " + helpStyle.Render("config created"))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("Configuration saved to: " + m.wizardPath))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render("Profile: " + m.wizardProject.Profile))
	if m.reviewErr != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.reviewErr.Error()))
	}
	return b.String()
}

func NewModel() tea.Model {
	return &modelImpl{
		state: stateMenu,
	}
}

var ErrNoTerminal = errors.New("not a terminal")

func isTerminal(fd int) bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func Run() error {
	if !isTerminal(0) {
		return ErrNoTerminal
	}
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
