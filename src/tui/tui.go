package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Pimatis/mavetis/src/engine"
	"github.com/Pimatis/mavetis/src/git"
	"github.com/Pimatis/mavetis/src/model"

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
	spinner   spinner.Model
	reviewErr error
	report    *model.Report

	// Findings list
	findingsCursor int
	viewport       viewport.Model

	// Detail
	detailViewport viewport.Model
	detailIndex    int
}

var menuItems = []string{
	"Run Review (staged changes)",
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
		m.state = stateFindings
		if msg.err != nil {
			m.reviewErr = msg.err
			m.state = stateMenu
			return m, nil
		}
		m.report = &msg.report
		m.findingsCursor = 0
		m.viewport.SetContent(m.renderFindingsList())
		m.viewport.GotoTop()
		return m, nil
	}

	cmds = append(cmds, nil)
	return m, tea.Batch(cmds...)
}

func (m *modelImpl) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == stateDetail {
		return m.handleDetailKey(msg)
	}
	if m.state == stateFindings {
		return m.handleFindingsKey(msg)
	}
	if m.state == stateAbout {
		return m.handleAboutKey(msg)
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

	if strings.HasPrefix(choice, "Run Review") {
		m.state = stateReviewing
		m.spinner = spinner.New()
		m.spinner.Spinner = spinner.Dot
		m.spinner.Style = spinnerStyle
		return m, tea.Batch(m.spinner.Tick, m.runReviewCmd())
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

func (m *modelImpl) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var content string
	var footer string

	switch m.state {
	case stateReviewing:
		content = m.renderReviewing()
		footer = ""
	case stateFindings:
		return m.renderFindings()
	case stateDetail:
		return m.renderDetailView()
	case stateAbout:
		content = m.renderAbout()
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
	var b strings.Builder
	b.WriteString(titleStyle.Render("mavetis") + "  " + helpStyle.Render("reviewing..."))
	b.WriteString("\n\n\n")
	b.WriteString("   " + m.spinner.View() + " Running security review on staged changes...")
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
