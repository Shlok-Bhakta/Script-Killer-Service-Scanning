package orchestrator

import (
	"context"
	"fmt"
	"scriptkiller/src/tui/components/dirlist"
	"scriptkiller/src/tui/scanner"
	"scriptkiller/src/tui/watcher"

	tea "github.com/charmbracelet/bubbletea"
)

type ScanCompleteMsg struct {
	Result *scanner.ScanResult
	Err    error
}

type ScanStartedMsg struct{}

type Model struct {
	scanner  *scanner.Scanner
	watcher  *watcher.Watcher
	scanning bool
	scanTime string
	ctx      context.Context
	cancel   context.CancelFunc
	scanType scanner.ScanType
}

func New(targetPath string) Model {
	s := scanner.New(targetPath)
	w, _ := watcher.New(targetPath)
	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		scanner:  s,
		watcher:  w,
		ctx:      ctx,
		cancel:   cancel,
		scanType: scanner.Directory,
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
	result, err := m.scanner.Scan(m.ctx, m.scanType)
	return ScanCompleteMsg{Result: result, Err: err}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case watcher.FileChangeMsg:
		if !m.scanning {
			m.scanning = true
			return m, tea.Batch(
				func() tea.Msg { return ScanStartedMsg{} },
				m.doScan,
				m.watcher.Start(m.ctx),
			)
		}
		return m, m.watcher.Start(m.ctx)

	case ScanCompleteMsg:
		m.scanning = false
		if msg.Result != nil {
			m.scanTime = fmt.Sprintf("%v", msg.Result.Duration)
		}
		return m, nil
	case dirlist.DirectorySelectedMsg:
		if m.cancel != nil {
			m.cancel()
		}
		m.scanType = scanner.Directory
		m.scanner.SetTargetPath(msg.Path)
	}

	return m, nil
}

func (m Model) IsScanning() bool {
	return m.scanning
}

func (m Model) GetScanTime() string {
	return m.scanTime
}

func (m Model) GetScanner() *scanner.Scanner {
	return m.scanner
}

func (m Model) TriggerScan() tea.Cmd {
	return func() tea.Msg {
		return tea.Batch(
			func() tea.Msg { return ScanStartedMsg{} },
			m.doScan,
		)()
	}
}

func (m Model) Cleanup() {
	if m.cancel != nil {
		m.cancel()
	}
}
