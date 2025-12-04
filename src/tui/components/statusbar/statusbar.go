package statusbar

import (
	"fmt"
	"scriptkiller/src/tui/components/dirlist"
	"scriptkiller/src/tui/components/endpointlist"
	"scriptkiller/src/tui/orchestrator"
	"scriptkiller/src/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	scanning     bool
	scanTime     string
	critCount    int
	warnCount    int
	infoCount    int
	selectedScan string
	errorMsg     string
}

func New(selected string) Model {
	return Model{selectedScan: selected}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case orchestrator.ScanStartedMsg:
		m.scanning = true
		return m, nil

	case orchestrator.ScanCompleteMsg:
		m.scanning = false

		if msg.Err != nil {
			m.errorMsg = msg.Err.Error()
		} else {
			m.errorMsg = "" // clear previous errors on success
		}

		return m, nil

	case dirlist.DirectorySelectedMsg:
		m.selectedScan = msg.Path
		return m, nil
	case endpointlist.EndpointSelectedMsg:
		m.selectedScan = msg.Address
		return m, nil
	}

	return m, nil
}

func (m Model) View(width int) string {
	theme := styles.CurrentTheme()

	style := lipgloss.NewStyle().
		Foreground(theme.FgMuted).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(theme.Border).
		Width(width).
		Padding(0, 1)

	leftContent := ""

	rightContent := fmt.Sprintf("Selected Scanner: %s", m.selectedScan)

	if m.scanning {
		leftContent = "⠋ Scanning..."
	} else if m.errorMsg != "" {
		leftContent = lipgloss.NewStyle().
			Foreground(theme.Error).
			Render("Scan failed: " + m.errorMsg)
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

	left := lipgloss.NewStyle().
		Width(width / 2).
		Align(lipgloss.Left).
		Render(leftContent)

	right := lipgloss.NewStyle().
		Width(width / 2).
		Align(lipgloss.Left).
		MaxHeight(1). // or remove to allow multiple lines
		Render(rightContent)

	combined := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return style.Render(combined)
}

func (m *Model) SetCounts(crit, warn, info int) {
	m.critCount = crit
	m.warnCount = warn
	m.infoCount = info
}

func (m *Model) SetScanTime(scanTime string) {
	m.scanTime = scanTime
}
