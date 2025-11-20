package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"scriptkiller/src/tools"
	"scriptkiller/src/tui/scanner"
	"scriptkiller/src/tui/styles"
	"scriptkiller/src/tui/watcher"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
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

type scanCompleteMsg struct {
	result *scanner.ScanResult
	err    error
}

type Model struct {
	scanner    *scanner.Scanner
	watcher    *watcher.Watcher
	width      int
	height     int
	list       list.Model
	findings   []tools.Finding
	scanning   bool
	targetPath string
	ctx        context.Context
	cancelCtx  context.CancelFunc

	critCount int
	warnCount int
	infoCount int
	scanTime  string
}

func NewModel(targetPath string) Model {
	absPath, _ := filepath.Abs(targetPath)

	s := scanner.New(targetPath)
	w, _ := watcher.New(targetPath)

	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Security Findings"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	theme := styles.CurrentTheme()
	l.Styles.Title = theme.S().Title
	l.Styles.TitleBar = lipgloss.NewStyle().Background(theme.BgBase)

	zone.NewGlobal()

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		scanner:    s,
		watcher:    w,
		list:       l,
		targetPath: absPath,
		ctx:        ctx,
		cancelCtx:  cancel,
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.doScan)
	if m.watcher != nil {
		cmds = append(cmds, m.watcher.Start(m.ctx))
	}
	return tea.Batch(cmds...)
}

func (m Model) doScan() tea.Msg {
	result, err := m.scanner.Scan(m.ctx)
	return scanCompleteMsg{result: result, err: err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := m.width / 2
		listHeight := m.height - 6
		m.list.SetSize(listWidth-4, listHeight-4)

		return m, nil

	case tea.KeyMsg:
		if m.scanning {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				if m.cancelCtx != nil {
					m.cancelCtx()
				}
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if m.cancelCtx != nil {
				m.cancelCtx()
			}
			return m, tea.Quit
		case "r":
			m.scanning = true
			return m, m.doScan
		}

	case watcher.FileChangeMsg:
		if !m.scanning {
			m.scanning = true
			return m, tea.Batch(m.doScan, m.watcher.Start(m.ctx))
		}
		return m, m.watcher.Start(m.ctx)

	case scanCompleteMsg:
		m.scanning = false

		if msg.err != nil {
			return m, nil
		}

		m.findings = m.scanner.GetAllFindings()

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

		if msg.result != nil {
			m.scanTime = fmt.Sprintf("%v", msg.result.Duration)
		}

		return m, nil
	}

	if !m.scanning {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	theme := styles.CurrentTheme()

	headerStyle := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Background(theme.BgBase).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border).
		Width(m.width).
		Padding(0, 1)

	header := headerStyle.Render(fmt.Sprintf("🔒 ScriptKiller Security Scanner - %s", m.targetPath))

	var content string
	if m.scanning {
		scanStyle := theme.S().Base.
			Width(m.width).
			Height(m.height - 6).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center)
		content = scanStyle.Render("⠋ Scanning project for security issues...\n\nThis may take a moment.")
	} else {
		listWidth := m.width / 2
		detailWidth := m.width - listWidth

		listStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Width(listWidth - 2).
			Height(m.height - 7)

		detailStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Width(detailWidth-2).
			Height(m.height-7).
			Padding(1, 2)

		listView := listStyle.Render(m.list.View())

		detailView := ""
		if item := m.list.SelectedItem(); item != nil {
			if fi, ok := item.(FindingItem); ok {
				detailView = m.renderDetail(&fi.finding, detailWidth-6)
			}
		} else {
			detailView = theme.S().Subtle.Render("Select a finding to view details")
		}

		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			listView,
			detailStyle.Render(detailView),
		)
	}

	statusBar := m.renderStatusBar()
	helpText := theme.S().Subtle.Render(" q: quit • r: rescan • ↑/↓: navigate • /: filter")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
		helpText,
	)
}

func (m Model) renderDetail(finding *tools.Finding, width int) string {
	if finding == nil {
		return ""
	}

	theme := styles.CurrentTheme()
	var b strings.Builder

	badge := ""
	switch finding.Severity {
	case tools.SeverityCritical:
		badge = lipgloss.NewStyle().
			Foreground(theme.BgBase).
			Background(theme.Error).
			Bold(true).
			Padding(0, 1).
			Render(" CRITICAL ")
	case tools.SeverityWarning:
		badge = lipgloss.NewStyle().
			Foreground(theme.BgBase).
			Background(theme.Warning).
			Bold(true).
			Padding(0, 1).
			Render(" WARNING ")
	case tools.SeverityInfo:
		badge = lipgloss.NewStyle().
			Foreground(theme.BgBase).
			Background(theme.Info).
			Bold(true).
			Padding(0, 1).
			Render(" INFO ")
	}

	b.WriteString(badge + "\n\n")

	b.WriteString(theme.S().Subtitle.Render("Message") + "\n")
	b.WriteString(theme.S().Text.Render(finding.Message) + "\n\n")

	b.WriteString(theme.S().Subtitle.Render("Location") + "\n")
	loc := finding.Location.File
	if finding.Location.Line > 0 {
		loc = fmt.Sprintf("%s:%d:%d", finding.Location.File, finding.Location.Line, finding.Location.Column)
	}
	b.WriteString(theme.S().Muted.Render(loc) + "\n\n")

	if finding.ID != "" {
		b.WriteString(theme.S().Subtitle.Render("ID") + "\n")
		b.WriteString(theme.S().Text.Render(finding.ID) + "\n\n")
	}

	if finding.Suggestion != "" {
		b.WriteString(theme.S().Subtitle.Render("Suggestion") + "\n")
		b.WriteString(theme.S().Text.Render(finding.Suggestion) + "\n\n")
	}

	if len(finding.Metadata) > 0 {
		b.WriteString(theme.S().Subtitle.Render("Metadata") + "\n")
		for k, v := range finding.Metadata {
			b.WriteString(theme.S().Muted.Render(fmt.Sprintf("  %s: %v", k, v)) + "\n")
		}
	}

	return b.String()
}

func (m Model) renderStatusBar() string {
	theme := styles.CurrentTheme()

	statusStyle := lipgloss.NewStyle().
		Foreground(theme.FgMuted).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(theme.Border).
		Width(m.width).
		Padding(0, 1)

	leftContent := ""

	if m.scanning {
		leftContent = "⠋ Scanning..."
	} else {
		badges := []string{}

		if m.critCount > 0 {
			badges = append(badges, lipgloss.NewStyle().
				Foreground(theme.BgBase).
				Background(theme.Error).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf(" %d ", m.critCount)))
		}
		if m.warnCount > 0 {
			badges = append(badges, lipgloss.NewStyle().
				Foreground(theme.BgBase).
				Background(theme.Warning).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf(" %d ", m.warnCount)))
		}
		if m.infoCount > 0 {
			badges = append(badges, lipgloss.NewStyle().
				Foreground(theme.BgBase).
				Background(theme.Info).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf(" %d ", m.infoCount)))
		}

		for i, badge := range badges {
			leftContent += badge
			if i < len(badges)-1 {
				leftContent += " "
			}
		}

		if m.scanTime != "" {
			leftContent += "  " + theme.S().Subtle.Render("Scan: "+m.scanTime)
		}
	}

	return statusStyle.Render(leftContent)
}

func StartTUI(targetPath string) error {
	m := NewModel(targetPath)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
