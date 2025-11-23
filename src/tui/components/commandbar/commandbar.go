package commandbar

import (
	"scriptkiller/src/tui/commands"
	"scriptkiller/src/tui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	textInput     textinput.Model
	statusMessage string
	statusError   bool
	focused       bool
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.CharLimit = 256
	ti.Prompt = ":"

	return Model{
		textInput: ti,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case commands.StatusMsg:
		m.statusMessage = msg.Message
		m.statusError = msg.IsError
		return m, nil

	case commands.ClearStatusMsg:
		m.statusMessage = ""
		return m, nil

	case tea.KeyMsg:
		if m.focused {
			m.textInput, cmd = m.textInput.Update(msg)

			if msg.Type == tea.KeyEnter {
				value := m.textInput.Value()
				m.textInput.SetValue("")
				return m, commands.HandleCommand(value)
			}

			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View(width int, focused bool) string {
	theme := styles.CurrentTheme()

	style := lipgloss.NewStyle().
		Foreground(theme.FgMuted).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(theme.Border).
		Width(width).
		Padding(0, 1)

	style = style.Foreground(theme.Primary)

	if focused {
		style = style.BorderForeground(theme.Accent)
	}

	text := ""
	if focused {
		if m.textInput.Value() == "" && m.statusMessage != "" {
			if m.statusError {
				style = style.Foreground(theme.Error)
			}
			text = m.statusMessage
		} else {
			text = m.textInput.View()
		}
	} else {
		helpText := theme.S().Subtle.Render(" q | quit • r | rescan • ↑/↓ | navigate • / | filter • : | Enter Command")
		text = helpText
	}

	return style.Render(text)
}

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	if focused {
		m.textInput.Focus()
	} else {
		m.textInput.Blur()
	}
}
