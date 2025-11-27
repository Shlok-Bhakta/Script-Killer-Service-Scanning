package endpointlist

import (
	"scriptkiller/src/tui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type endpoint string

func (d endpoint) Title() string       { return string(d) }
func (d endpoint) Description() string { return "" }
func (d endpoint) FilterValue() string { return string(d) }

type EndpointAddedMsg struct {
	Address string
}

type Model struct {
	list      list.Model
	endpoints []string
	focused   bool
}

func New() Model {
	endpoints := []list.Item{}

	endpointList := list.New(endpoints, list.NewDefaultDelegate(), 0, 0)
	endpointList.Title = "Directories"
	endpointList.SetShowFilter(false)
	endpointList.SetShowHelp(false)
	endpointList.SetShowStatusBar(false)

	endpointDelegate := list.NewDefaultDelegate()
	endpointDelegate.ShowDescription = false
	endpointDelegate.SetSpacing(0)
	endpointDelegate.SetHeight(1)
	endpointList.SetDelegate(endpointDelegate)

	return Model{
		list:      endpointList,
		endpoints: []string{},
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case EndpointAddedMsg:
		m.endpoints = append(m.endpoints, msg.Address)
		addr := endpoint(msg.Address)
		m.list.InsertItem(len(m.list.Items()), addr)
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
		Width(width-2).
		Height(height).
		Padding(0, 1)

	if focused {
		style = style.BorderForeground(theme.Accent)
	}

	m.list.SetHeight(height)
	m.list.SetWidth(width - 4)

	return style.Render(m.list.View())
}

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m Model) GetDirectories() []string {
	return m.endpoints
}
