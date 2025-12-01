package commands

import (
	"fmt"
	"strings"
	"time"

	"scriptkiller/src/tui/components/dirlist"
	"scriptkiller/src/tui/components/endpointlist"

	tea "github.com/charmbracelet/bubbletea"
)

const STATUS_TIMEOUT = time.Second * 3

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
	if len(parts) >= 3 {
		switch parts[1] {
		case "endpoint":
			addr := parts[2]
			return tea.Batch(func() tea.Msg {
				return endpointlist.EndpointAddedMsg{Address: addr}
			},
				sendStatus(fmt.Sprintf("Added endpoint: %s", addr), false),
				clearStatusAfter(STATUS_TIMEOUT),
			)
		case "dir":
			dir := parts[2]
			return tea.Batch(
				func() tea.Msg {
					return dirlist.DirectoryAddedMsg{Path: dir}
				},
				sendStatus(fmt.Sprintf("Added directory: %s", dir), false),
				clearStatusAfter(STATUS_TIMEOUT),
			)
		}
	}
	return sendStatus("Usage: add <type> \n Available types: dir endpoint", true)
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
