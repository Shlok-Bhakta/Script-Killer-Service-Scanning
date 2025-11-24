package statusbar

import (
	"fmt"
	"scriptkiller/src/tui/orchestrator"
	"scriptkiller/src/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	scanning  bool
	scanTime  string
	critCount int
	warnCount int
	infoCount int
}

func New() Model {
	return Model{}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg.(type) {
	case orchestrator.ScanStartedMsg:
		m.scanning = true
		return m, nil

	case orchestrator.ScanCompleteMsg:
		m.scanning = false
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

	return style.Render(leftContent)
}

func (m *Model) SetCounts(crit, warn, info int) {
	m.critCount = crit
	m.warnCount = warn
	m.infoCount = info
}

func (m *Model) SetScanTime(scanTime string) {
	m.scanTime = scanTime
}
