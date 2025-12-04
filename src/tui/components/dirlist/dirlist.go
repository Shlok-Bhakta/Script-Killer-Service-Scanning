package dirlist

import (
	"path/filepath"

	"scriptkiller/src/tui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type directoryItem string

func (d directoryItem) Title() string       { return string(d) }
func (d directoryItem) Description() string { return "" }
func (d directoryItem) FilterValue() string { return string(d) }

type DirectoryAddedMsg struct {
	Path string
}

type DirectorySelectedMsg struct {
	Path string
}

type Model struct {
	list        list.Model
	directories []string
	focused     bool
}

func New(targetPath string) Model {
	absPath, _ := filepath.Abs(targetPath)

	dirs := []list.Item{}
	dirs = append(dirs, directoryItem(absPath))

	directoryList := list.New(dirs, list.NewDefaultDelegate(), 0, 0)
	directoryList.Title = "Directories"
	directoryList.SetShowFilter(false)
	directoryList.SetShowHelp(false)
	directoryList.SetShowStatusBar(false)

	dirDelegate := list.NewDefaultDelegate()
	dirDelegate.ShowDescription = false
	dirDelegate.SetSpacing(0)
	dirDelegate.SetHeight(1)
	directoryList.SetDelegate(dirDelegate)

	return Model{
		list:        directoryList,
		directories: []string{absPath},
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case DirectoryAddedMsg:
		m.directories = append(m.directories, msg.Path)
		dir := directoryItem(msg.Path)
		m.list.InsertItem(len(m.list.Items()), dir)
		return m, nil
	}

	if m.focused {
		m.list, cmd = m.list.Update(msg)

		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			i, ok := m.list.SelectedItem().(directoryItem)
			if ok {
				return m, func() tea.Msg {
					return DirectorySelectedMsg{Path: string(i)}
				}
			}
		}
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
	return m.directories
}
