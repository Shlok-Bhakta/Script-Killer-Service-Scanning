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

type scanMode int

const (
	scanModeAll scanMode = iota
	scanModeCode
	scanModeDeps
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
	mode   scanMode
}

type Model struct {
	scanner    *scanner.Scanner
	watcher    *watcher.Watcher
	width      int
	height     int
	codeList   list.Model
	depsList   list.Model
	activeTab  int
	scanning   bool
	targetPath string
	ctx        context.Context
	cancelCtx  context.CancelFunc

	codeCritCount int
	codeWarnCount int
	codeInfoCount int
	depsCritCount int
	depsWarnCount int
	depsInfoCount int
	scanTime      string
}

func NewModel(targetPath string) Model {
	absPath, _ := filepath.Abs(targetPath)

	s := scanner.New(targetPath)
	w, _ := watcher.New(targetPath)

	delegate := list.NewDefaultDelegate()
	codeList := list.New([]list.Item{}, delegate, 0, 0)
	codeList.Title = "Code Security Findings"
	codeList.SetShowStatusBar(true)
	codeList.SetFilteringEnabled(true)

	depsList := list.New([]list.Item{}, delegate, 0, 0)
	depsList.Title = "Dependency Vulnerabilities"
	depsList.SetShowStatusBar(true)
	depsList.SetFilteringEnabled(true)

	theme := styles.CurrentTheme()
	codeList.Styles.Title = theme.S().Title
	codeList.Styles.TitleBar = lipgloss.NewStyle().Background(theme.BgBase)
	depsList.Styles.Title = theme.S().Title
	depsList.Styles.TitleBar = lipgloss.NewStyle().Background(theme.BgBase)

	zone.NewGlobal()

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		scanner:    s,
		watcher:    w,
		codeList:   codeList,
		depsList:   depsList,
		activeTab:  0,
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
	return scanCompleteMsg{result: result, err: err, mode: scanModeAll}
}

func (m Model) doCodeScan() tea.Msg {
	result, err := m.scanner.ScanCode(m.ctx)
	return scanCompleteMsg{result: result, err: err, mode: scanModeCode}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := m.width / 2
		listHeight := m.height - 9
		m.codeList.SetSize(listWidth-4, listHeight-4)
		m.depsList.SetSize(listWidth-4, listHeight-4)

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
		case "tab":
			m.activeTab = (m.activeTab + 1) % 2
			return m, nil
		case "1":
			m.activeTab = 0
			return m, nil
		case "2":
			m.activeTab = 1
			return m, nil
		}

	case watcher.FileChangeMsg:
		if !m.scanning {
			m.scanning = true
			var scanCmd tea.Cmd
			if msg.IsDependencyChange {
				scanCmd = m.doScan
			} else {
				scanCmd = m.doCodeScan
			}
			return m, tea.Batch(scanCmd, m.watcher.Start(m.ctx))
		}
		return m, m.watcher.Start(m.ctx)

	case scanCompleteMsg:
		m.scanning = false

		if msg.err != nil {
			return m, nil
		}

		codeFindings, depsFindings := m.scanner.GetFindingsByType()

		codeItems := make([]list.Item, len(codeFindings))
		for i, f := range codeFindings {
			codeItems[i] = FindingItem{finding: f}
		}
		m.codeList.SetItems(codeItems)

		depsItems := make([]list.Item, len(depsFindings))
		for i, f := range depsFindings {
			depsItems[i] = FindingItem{finding: f}
		}
		m.depsList.SetItems(depsItems)

		m.codeCritCount = 0
		m.codeWarnCount = 0
		m.codeInfoCount = 0
		for _, f := range codeFindings {
			switch f.Severity {
			case tools.SeverityCritical:
				m.codeCritCount++
			case tools.SeverityWarning:
				m.codeWarnCount++
			case tools.SeverityInfo:
				m.codeInfoCount++
			}
		}

		m.depsCritCount = 0
		m.depsWarnCount = 0
		m.depsInfoCount = 0
		for _, f := range depsFindings {
			switch f.Severity {
			case tools.SeverityCritical:
				m.depsCritCount++
			case tools.SeverityWarning:
				m.depsWarnCount++
			case tools.SeverityInfo:
				m.depsInfoCount++
			}
		}

		if msg.result != nil {
			m.scanTime = fmt.Sprintf("%v", msg.result.Duration)
		}

		return m, nil
	}

	if !m.scanning {
		if m.activeTab == 0 {
			m.codeList, cmd = m.codeList.Update(msg)
		} else {
			m.depsList, cmd = m.depsList.Update(msg)
		}
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

	tabs := m.renderTabs()

	var content string
	if m.scanning {
		scanStyle := theme.S().Base.
			Width(m.width).
			Height(m.height - 9).
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
			Height(m.height - 10)

		detailStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Width(detailWidth-2).
			Height(m.height-10).
			Padding(1, 2)

		var listView string
		var selectedFinding *tools.Finding

		if m.activeTab == 0 {
			listView = listStyle.Render(m.codeList.View())
			if item := m.codeList.SelectedItem(); item != nil {
				if fi, ok := item.(FindingItem); ok {
					selectedFinding = &fi.finding
				}
			}
		} else {
			listView = listStyle.Render(m.depsList.View())
			if item := m.depsList.SelectedItem(); item != nil {
				if fi, ok := item.(FindingItem); ok {
					selectedFinding = &fi.finding
				}
			}
		}

		detailView := ""
		if selectedFinding != nil {
			detailView = m.renderDetail(selectedFinding, detailWidth-6)
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
	helpText := theme.S().Subtle.Render(" q: quit • r: rescan • tab/1/2: switch tabs • ↑/↓: navigate • /: filter")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		statusBar,
		helpText,
	)
}

func (m Model) renderTabs() string {
	theme := styles.CurrentTheme()

	activeTabStyle := lipgloss.NewStyle().
		Foreground(theme.BgBase).
		Background(theme.Accent).
		Bold(true).
		Padding(0, 2)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(theme.FgMuted).
		Background(theme.BgBase).
		Padding(0, 2)

	codeTab := ""
	depsTab := ""

	if m.activeTab == 0 {
		codeTab = activeTabStyle.Render("Code")
		depsTab = inactiveTabStyle.Render("Dependencies")
	} else {
		codeTab = inactiveTabStyle.Render("Code")
		depsTab = activeTabStyle.Render("Dependencies")
	}

	tabsContainer := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border).
		Width(m.width).
		Padding(0, 1)

	return tabsContainer.Render(lipgloss.JoinHorizontal(lipgloss.Left, codeTab, " ", depsTab))
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

	if finding.CVE != nil {
		b.WriteString(theme.S().Subtitle.Render("CVE") + "\n")
		b.WriteString(theme.S().Text.Render(*finding.CVE) + "\n\n")
	}

	if finding.SeverityScore != nil {
		b.WriteString(theme.S().Subtitle.Render("CVSS Score") + "\n")
		b.WriteString(theme.S().Text.Render(fmt.Sprintf("%.1f", *finding.SeverityScore)) + "\n\n")
	}

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
		critCount := m.codeCritCount + m.depsCritCount
		warnCount := m.codeWarnCount + m.depsWarnCount
		infoCount := m.codeInfoCount + m.depsInfoCount

		badges := []string{}

		if critCount > 0 {
			badges = append(badges, lipgloss.NewStyle().
				Foreground(theme.BgBase).
				Background(theme.Error).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf(" %d ", critCount)))
		}
		if warnCount > 0 {
			badges = append(badges, lipgloss.NewStyle().
				Foreground(theme.BgBase).
				Background(theme.Warning).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf(" %d ", warnCount)))
		}
		if infoCount > 0 {
			badges = append(badges, lipgloss.NewStyle().
				Foreground(theme.BgBase).
				Background(theme.Info).
				Bold(true).
				Padding(0, 1).
				Render(fmt.Sprintf(" %d ", infoCount)))
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
