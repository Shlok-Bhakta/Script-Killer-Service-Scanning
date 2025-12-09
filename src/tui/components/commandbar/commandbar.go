package commandbar

import (
	"scriptkiller/src/tui/commands"
	"scriptkiller/src/tui/styles"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Focus int

const (
	FocusFindings Focus = iota
	FocusDirectories
	FocusEndpoints
	FocusCommand
)

type findingsKeyMap struct{}

func (k findingsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rescan")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
	}
}
func (k findingsKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

type directoriesKeyMap struct{}

func (k directoriesKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
	}
}
func (k directoriesKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

type endpointsKeyMap struct{}

func (k endpointsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
	}
}
func (k endpointsKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

type commandKeyMap struct{}

func (k commandKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "run")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
	}
}
func (k commandKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

type Model struct {
	textInput     textinput.Model
	help          help.Model
	statusMessage string
	statusError   bool
	focused       bool
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.CharLimit = 256
	ti.Prompt = ":"

	h := help.New()

	return Model{
		textInput: ti,
		help:      h,
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

func (m Model) View(width int, focused bool, currentFocus Focus) string {
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
		m.help.Width = width - 4
		m.help.Styles.ShortKey = lipgloss.NewStyle().Foreground(theme.Accent)
		m.help.Styles.ShortDesc = lipgloss.NewStyle().Foreground(theme.FgMuted)
		m.help.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(theme.Border)

		var keyMap help.KeyMap
		switch currentFocus {
		case FocusFindings:
			keyMap = findingsKeyMap{}
		case FocusDirectories:
			keyMap = directoriesKeyMap{}
		case FocusEndpoints:
			keyMap = endpointsKeyMap{}
		case FocusCommand:
			keyMap = commandKeyMap{}
		default:
			keyMap = findingsKeyMap{}
		}
		text = m.help.View(keyMap)
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
