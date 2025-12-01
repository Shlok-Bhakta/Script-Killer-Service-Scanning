package endpointlist

import (
	"scriptkiller/src/tui/styles"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
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
	showPopup bool
	input     textinput.Model
}

func New() Model {
	endpoints := []list.Item{}

	endpointList := list.New(endpoints, list.NewDefaultDelegate(), 0, 0)
	endpointList.Title = "Endpoints"
	endpointList.SetShowFilter(false)
	endpointList.SetShowHelp(false)
	endpointList.SetShowStatusBar(false)

	endpointDelegate := list.NewDefaultDelegate()
	endpointDelegate.ShowDescription = false
	endpointDelegate.SetSpacing(0)
	endpointDelegate.SetHeight(1)
	endpointList.SetDelegate(endpointDelegate)

	ti := textinput.New()
	ti.Placeholder = "http://localhost:8080"
	ti.CharLimit = 256
	ti.Width = 40

	return Model{
		list:      endpointList,
		endpoints: []string{},
		input:     ti,
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

	case tea.KeyMsg:
		if m.showPopup {
			switch msg.String() {
			case "enter":
				if m.input.Value() != "" {
					m.endpoints = append(m.endpoints, m.input.Value())
					addr := endpoint(m.input.Value())
					m.list.InsertItem(len(m.list.Items()), addr)
					m.input.SetValue("")
				}
				m.showPopup = false
				m.input.Blur()
				return m, nil
			case "esc":
				m.showPopup = false
				m.input.Blur()
				m.input.SetValue("")
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		if m.focused && msg.String() == "a" {
			m.showPopup = true
			m.input.Focus()
			return m, textinput.Blink
		}
	}

	if m.focused && !m.showPopup {
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
		m.list.Title = "Endpoints [a]dd"
	} else {
		m.list.Title = "Endpoints"
	}

	m.list.SetHeight(height)
	m.list.SetWidth(width - 4)

	return style.Render(m.list.View())
}

func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

func (m Model) GetEndpoints() []string {
	return m.endpoints
}

func (m Model) IsPopupOpen() bool {
	return m.showPopup
}

func (m Model) RenderWithOverlay(baseView string) string {
	if !m.showPopup {
		return baseView
	}

	theme := styles.CurrentTheme()

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent).
		MarginBottom(1)

	hintStyle := lipgloss.NewStyle().
		Foreground(theme.FgMuted).
		MarginTop(1)

	popupContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Add Endpoint"),
		m.input.View(),
		hintStyle.Render("enter: confirm • esc: cancel"),
	)

	fg := wrapModel{content: popupStyle.Render(popupContent)}
	bg := wrapModel{content: baseView}
	overlayModel := overlay.New(&fg, &bg, overlay.Center, overlay.Center, 0, 0)
	return overlayModel.View()
}

type wrapModel struct {
	content string
}

func (w wrapModel) Init() tea.Cmd                           { return nil }
func (w wrapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return w, nil }
func (w wrapModel) View() string                            { return w.content }
