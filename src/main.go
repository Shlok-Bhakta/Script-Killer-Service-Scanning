package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"scriptkiller/src/mcp"
	"scriptkiller/src/tui"
	"scriptkiller/src/tui/scanner"

	"github.com/charmbracelet/log"
)

func main() {
	log.SetOutput(io.Discard)
	log.SetLevel(log.ErrorLevel)

	args := os.Args[1:]
	cwd := "."
	noTUI := false
	debug := false
	mcpMode := false

	for i, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: scriptkiller [options] [path]")
			fmt.Println("Options:")
			fmt.Println("  --help, -h: Show this help message")
			fmt.Println("  --no-tui: Scan and print results without TUI")
			fmt.Println("  --mcp: Start as MCP server (stdio transport)")
			fmt.Println("  --debug: Enable debug logging to stdout")
			return
		}
		if arg == "--no-tui" {
			noTUI = true
			continue
		}
		if arg == "--mcp" {
			mcpMode = true
			continue
		}
		if arg == "--debug" {
			debug = true
			continue
		}
		if i == len(args)-1 && arg != "--no-tui" && arg != "--debug" && arg != "--mcp" {
			cwd = arg
		}
	}

	if mcpMode {
		m := mcp.New(cwd)
		if err := m.ServeStdio(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if noTUI {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.InfoLevel)
		if err := runScan(cwd); err != nil {
			log.Fatal("Scan failed", "error", err)
		}
		return
	}

	if debug {
		logDir := "/tmp/scriptkillerlogs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
			os.Exit(1)
		}

		timestamp := time.Now().Format("20060102-150405")
		logPath := fmt.Sprintf("%s/scriptkiller-%s.log", logDir, timestamp)

		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		log.SetOutput(f)
		log.SetLevel(log.DebugLevel)

		fmt.Printf("Debug logging enabled: %s\n", logPath)
	}

	if err := tui.StartTUI(cwd); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}

func runScan(path string) error {
	s := scanner.New(path)
	ctx := context.Background()

	log.Info("Starting security scan", "path", path)
	result, err := s.Scan(ctx)
	if err != nil {
		return err
	}

	findings := s.GetAllFindings()
	log.Info("Scan completed", "duration", result.Duration, "findings", len(findings))

	if len(findings) == 0 {
		log.Info("No security issues found")
		return nil
	}

	critCount := 0
	warnCount := 0
	infoCount := 0

	for _, f := range findings {
		switch f.Severity {
		case "critical":
			critCount++
		case "warning":
			warnCount++
		case "info":
			infoCount++
		}
	}

	if critCount > 0 {
		log.Error("Critical issues found", "count", critCount)
	}
	if warnCount > 0 {
		log.Warn("Warnings found", "count", warnCount)
	}
	if infoCount > 0 {
		log.Info("Info items found", "count", infoCount)
	}

	fmt.Println()

	for _, f := range findings {
		var logFunc func(msg interface{}, keyvals ...interface{})
		switch f.Severity {
		case "critical":
			logFunc = log.Error
		case "warning":
			logFunc = log.Warn
		default:
			logFunc = log.Info
		}

		loc := f.Location.File
		if f.Location.Line > 0 {
			loc = fmt.Sprintf("%s:%d:%d", f.Location.File, f.Location.Line, f.Location.Column)
		}

		logFunc(f.Message,
			"id", f.ID,
			"location", loc,
		)

		if f.Suggestion != "" {
			log.Info("  → " + f.Suggestion)
		}
	}

	return nil
}
