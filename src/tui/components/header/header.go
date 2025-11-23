package header

import (
	"fmt"
	"scriptkiller/src/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	path string
}

func New(path string) Model {
	return Model{path: path}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(width int) string {
	theme := styles.CurrentTheme()

	style := lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Background(theme.BgBase).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border).
		Width(width).
		Padding(0, 1)

	return style.Render(fmt.Sprintf("🔒 ScriptKiller Security Scanner - %s", m.path))
}

func (m *Model) SetPath(path string) {
	m.path = path
}
