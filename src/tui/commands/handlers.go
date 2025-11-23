package commands

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"scriptkiller/src/tui/components/dirlist"
)

type StatusMsg struct {
	Message string
	IsError bool
}

type ClearStatusMsg struct{}

func HandleCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)

	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "add", "a":
		return handleAdd(parts)
	case "remove", "rm":
		return handleRemove(parts)
	case "list", "ls":
		return handleList(parts)
	default:
		return sendStatus("Unrecognized Command", true)
	}
}

func handleAdd(parts []string) tea.Cmd {
	if len(parts) >= 3 && parts[1] == "dir" {
		dir := parts[2]
		return tea.Batch(
			func() tea.Msg {
				return dirlist.DirectoryAddedMsg{Path: dir}
			},
			sendStatus(fmt.Sprintf("Added directory: %s", dir), false),
			clearStatusAfter(time.Second*3),
		)
	}
	return sendStatus("Usage: add dir <path>", true)
}

func handleRemove(parts []string) tea.Cmd {
	if len(parts) >= 3 && parts[1] == "dir" {
		return nil
	}
	return sendStatus("Usage: remove dir <path>", true)
}

func handleList(parts []string) tea.Cmd {
	if len(parts) >= 2 && parts[1] == "dirs" {
		return nil
	}
	return nil
}

func sendStatus(message string, isError bool) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Message: message, IsError: isError}
	}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return ClearStatusMsg{}
	}
}
