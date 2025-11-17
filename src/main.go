package main

import (
	"fmt"
	"os"

	"scriptkiller/src/detector"
	"scriptkiller/src/tools"

	"github.com/charmbracelet/log"
)

func main() {
	args := os.Args[1:]
	cwd := "."
	for i, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println("Usage: scriptkiller [options] [path]")
			fmt.Println("Options:")
			fmt.Println("  --help, -h: Show this help message")
			return
		}
		if i == len(args)-1 {
			cwd = arg
			log.Info("Using cwd", "path", cwd)
		}
	}

	log.Info("Detecting project languages", "path", cwd)
	languages, err := detector.DetectProjectLanguages(cwd)
	if err != nil {
		log.Fatal("Failed to detect languages", "error", err)
	}

	if len(languages) == 0 {
		log.Warn("No supported languages detected in project")
		return
	}

	log.Info("Languages detected", "languages", languages)

	log.Info("Running security scan")

	tool_arr := []tools.SecurityTool{
		tools.NewGosecTool(),
		tools.NewOSVScannerTool(),
	}
	tool_out_map, errs := tools.RunAllToolsForLanguage(tool_arr, languages, cwd)
	if len(errs) != 0 {
		log.Error("Errors occurred", "errors", errs)
	}

	for toolName, output := range tool_out_map {
		fmt.Printf("\n=== %s Results ===\n", toolName)
		fmt.Printf("Duration: %dms\n", output.Duration)
		fmt.Printf("Total Findings: %d\n\n", output.TotalFindings())

		if len(output.Critical) > 0 {
			fmt.Printf("CRITICAL (%d):\n", len(output.Critical))
			for _, f := range output.Critical {
				fmt.Printf("  [%s] %s\n", f.ID, f.Message)
				fmt.Printf("    Location: %s\n", f.Location.String())
				if f.Suggestion != "" {
					fmt.Printf("    %s\n", f.Suggestion)
				}
			}
			fmt.Println()
		}

		if len(output.Warnings) > 0 {
			fmt.Printf("WARNINGS (%d):\n", len(output.Warnings))
			for _, f := range output.Warnings {
				fmt.Printf("  [%s] %s\n", f.ID, f.Message)
				fmt.Printf("    Location: %s\n", f.Location.String())
				if f.Suggestion != "" {
					fmt.Printf("    %s\n", f.Suggestion)
				}
			}
			fmt.Println()
		}

		if len(output.Info) > 0 {
			fmt.Printf("INFO (%d):\n", len(output.Info))
			for _, f := range output.Info {
				fmt.Printf("  [%s] %s\n", f.ID, f.Message)
				fmt.Printf("    Location: %s\n", f.Location.String())
				if f.Suggestion != "" {
					fmt.Printf("    %s\n", f.Suggestion)
				}
			}
			fmt.Println()
		}
	}
}
