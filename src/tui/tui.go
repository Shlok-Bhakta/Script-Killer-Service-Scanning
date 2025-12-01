package tui

import (
	"scriptkiller/src/tui/components/commandbar"
	"scriptkiller/src/tui/components/details"
	"scriptkiller/src/tui/components/dirlist"
	"scriptkiller/src/tui/components/endpointlist"
	"scriptkiller/src/tui/components/findings"
	"scriptkiller/src/tui/components/header"
	"scriptkiller/src/tui/components/statusbar"
	"scriptkiller/src/tui/orchestrator"
	"scriptkiller/src/tui/styles"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

type Focus int

const (
	FocusFindings Focus = iota
	FocusDirectories
	FocusEndpoints
	FocusCommand
)

type Model struct {
	width  int
	height int

	orchestrator          orchestrator.Model
	headerComponent       header.Model
	dirlistComponent      dirlist.Model
	endpointListComponent endpointlist.Model
	findingsComponent     findings.Model
	detailsComponent      details.Model
	statusbarComponent    statusbar.Model
	commandbarComponent   commandbar.Model

	focus Focus
}

func NewModel(targetPath string) Model {
	zone.NewGlobal()

	findingsComp := findings.New()
	findingsComp.SetFocused(true)

	return Model{
		orchestrator:          orchestrator.New(targetPath),
		headerComponent:       header.New(targetPath),
		dirlistComponent:      dirlist.New(targetPath),
		endpointListComponent: endpointlist.New(),
		findingsComponent:     findingsComp,
		detailsComponent:      details.New(),
		statusbarComponent:    statusbar.New(),
		commandbarComponent:   commandbar.New(),
		focus:                 FocusFindings,
	}
}

func (m Model) Init() tea.Cmd {
	return m.orchestrator.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "tab" {
			switch m.focus {
			case FocusDirectories:
				m.focus = FocusEndpoints
				m.findingsComponent.SetFocused(true)
			case FocusEndpoints:
				m.focus = FocusFindings
				m.endpointListComponent.SetFocused(true)
			case FocusFindings:
				m.focus = FocusCommand
				m.commandbarComponent.SetFocused(true)
			case FocusCommand:
				m.focus = FocusDirectories
				m.dirlistComponent.SetFocused(true)
			}
			return m, nil
		}

		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.orchestrator.Cleanup()
			return m, tea.Quit
		}

		if m.focus == FocusCommand {
			m.commandbarComponent, cmd = m.commandbarComponent.Update(msg)
			return m, cmd
		} else if m.focus == FocusFindings {
			if !m.orchestrator.IsScanning() {
				if msg.String() == "r" {
					return m, m.orchestrator.TriggerScan()
				}
				m.findingsComponent, cmd = m.findingsComponent.Update(msg)
				return m, cmd
			}
		} else if m.focus == FocusDirectories {
			m.dirlistComponent, cmd = m.dirlistComponent.Update(msg)
			return m, cmd
		}

	case orchestrator.ScanCompleteMsg:
		m.orchestrator, cmd = m.orchestrator.Update(msg)
		cmds = append(cmds, cmd)

		if msg.Err == nil && msg.Result != nil {
			allFindings := m.orchestrator.GetScanner().GetAllFindings()
			m.findingsComponent.SetFindings(allFindings)

			critCount, warnCount, infoCount := m.findingsComponent.GetCounts()
			m.statusbarComponent.SetCounts(critCount, warnCount, infoCount)
			m.statusbarComponent.SetScanTime(m.orchestrator.GetScanTime())
		}

		m.statusbarComponent, cmd = m.statusbarComponent.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case orchestrator.ScanStartedMsg:
		m.orchestrator, cmd = m.orchestrator.Update(msg)
		cmds = append(cmds, cmd)

		m.statusbarComponent, cmd = m.statusbarComponent.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case dirlist.DirectoryAddedMsg:
		m.dirlistComponent, cmd = m.dirlistComponent.Update(msg)
		cmds = append(cmds, cmd)

		dirs := m.dirlistComponent.GetDirectories()
		if len(dirs) > 0 {
			m.headerComponent.SetPath(dirs[0])
		}
		return m, tea.Batch(cmds...)
	}

	m.orchestrator, cmd = m.orchestrator.Update(msg)
	cmds = append(cmds, cmd)

	if !m.orchestrator.IsScanning() {
		m.findingsComponent, cmd = m.findingsComponent.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.commandbarComponent, cmd = m.commandbarComponent.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	theme := styles.CurrentTheme()

	dirHeight := 8
	statusBarHeight := 3
	commandBarHeight := 3
	headerHeight := 3
	contentHeight := m.height - dirHeight - statusBarHeight - commandBarHeight - headerHeight

	headerView := m.headerComponent.View(m.width)

	m.dirlistComponent.SetFocused(m.focus == FocusDirectories)

	halfWidth := m.width / 2
	dirView := m.dirlistComponent.View(halfWidth, dirHeight, m.focus == FocusDirectories)
	endpointView := m.endpointListComponent.View(halfWidth, dirHeight, m.focus == FocusEndpoints)

	scanPointView := lipgloss.JoinHorizontal(lipgloss.Top, dirView, endpointView)

	var content string
	if m.orchestrator.IsScanning() {
		scanStyle := theme.S().Base.
			Width(m.width).
			Height(contentHeight).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center)
		content = scanStyle.Render("⠋ Scanning project for security issues...\n\nThis may take a moment.")
	} else {
		listWidth := m.width / 2
		detailWidth := m.width / 2

		m.findingsComponent.SetFocused(m.focus == FocusFindings)
		listView := m.findingsComponent.View(listWidth, contentHeight, m.focus == FocusFindings)

		selectedFinding := m.findingsComponent.GetSelectedFinding()
		m.detailsComponent.SetFinding(selectedFinding)
		detailView := m.detailsComponent.View(detailWidth, contentHeight)

		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			listView,
			detailView,
		)
	}

	statusBarView := m.statusbarComponent.View(m.width)

	m.commandbarComponent.SetFocused(m.focus == FocusCommand)
	commandBarView := m.commandbarComponent.View(m.width, m.focus == FocusCommand)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerView,
		scanPointView,
		content,
		statusBarView,
		commandBarView,
	)
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
