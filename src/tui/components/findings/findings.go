package findings

import (
	"fmt"
	"scriptkiller/src/tools"
	"scriptkiller/src/tui/scanner"
	"scriptkiller/src/tui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FindingItem struct {
	finding tools.Finding
}

func (i FindingItem) FilterValue() string {
	return i.finding.Message + " " + i.finding.Location.File
}

func (i FindingItem) Title() string {
	icon := ""
	switch i.finding.Severity {
	case tools.SeverityCritical:
		icon = styles.CurrentTheme().ItemErrorIcon.String()
	case tools.SeverityWarning:
		icon = styles.CurrentTheme().ItemBusyIcon.String()
	case tools.SeverityInfo:
		icon = styles.CurrentTheme().ItemOnlineIcon.String()
	}
	return fmt.Sprintf("%s %s", icon, i.finding.Message)
}

func (i FindingItem) Description() string {
	loc := ""
	if i.finding.Location.Line > 0 {
		loc = fmt.Sprintf("%s:%d", i.finding.Location.File, i.finding.Location.Line)
	} else {
		loc = i.finding.Location.File
	}
	return loc
}

type ScanCompleteMsg struct {
	Result *scanner.ScanResult
	Err    error
}

type Model struct {
	list      list.Model
	findings  []tools.Finding
	focused   bool
	critCount int
	warnCount int
	infoCount int
}

func New() Model {
	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Security Findings"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	theme := styles.CurrentTheme()
	l.Styles.Title = theme.S().Title
	l.Styles.TitleBar = lipgloss.NewStyle().Background(theme.BgBase)

	delegate.ShowDescription = false

	return Model{
		list: l,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ScanCompleteMsg:
		if msg.Err != nil {
			return m, nil
		}

		items := make([]list.Item, len(m.findings))
		for i, f := range m.findings {
			items[i] = FindingItem{finding: f}
		}
		m.list.SetItems(items)

		m.critCount = 0
		m.warnCount = 0
		m.infoCount = 0
		for _, f := range m.findings {
			switch f.Severity {
			case tools.SeverityCritical:
				m.critCount++
			case tools.SeverityWarning:
				m.warnCount++
			case tools.SeverityInfo:
				m.infoCount++
			}
		}

		return m, nil
	}

	if m.focused {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m Model) View(width, height int, focused bool) string {
	theme := styles.CurrentTheme()

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(width - 2).
		Height(height)

	if focused {
		style = style.BorderForeground(theme.Accent)
	}

	m.list.SetSize(width-2, height-2)

	return style.Render(m.list.View())
}

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m *Model) SetFindings(findings []tools.Finding) {
	m.findings = findings

	items := make([]list.Item, len(findings))
	for i, f := range findings {
		items[i] = FindingItem{finding: f}
	}
	m.list.SetItems(items)

	m.critCount = 0
	m.warnCount = 0
	m.infoCount = 0
	for _, f := range findings {
		switch f.Severity {
		case tools.SeverityCritical:
			m.critCount++
		case tools.SeverityWarning:
			m.warnCount++
		case tools.SeverityInfo:
			m.infoCount++
		}
	}
}

func (m Model) GetSelectedFinding() *tools.Finding {
	if item := m.list.SelectedItem(); item != nil {
		if fi, ok := item.(FindingItem); ok {
			return &fi.finding
		}
	}
	return nil
}

func (m Model) GetCounts() (crit, warn, info int) {
	return m.critCount, m.warnCount, m.infoCount
}
