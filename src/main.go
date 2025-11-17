package main

import (
	"fmt"
	"os"
	"path/filepath"

	// "scriptkiller/src/nix"
	"scriptkiller/src/tools"

	"github.com/charmbracelet/log"
	// "github.com/charmbracelet/log"
)

func Detect_Lang(fName string) string {
	Lang := filepath.Ext(fName)
	switch Lang {
	case ".go":
		fmt.Println("Code Language Detected: Golang")
		return "Golang"
	case ".py":
		fmt.Println("Code Language Detected: Python")
		return "Python"
	case ".cpp":
		fmt.Println("Code Language Detected: C++")
		return "C++"
	case ".js":
		fmt.Println("Code Language Detected: Javascript")
		return "Javascript"
	default:
		fmt.Println("FileType/Code Language Not Supported")
		return "Not Availiable"
	}
}

func main() {
	// We need to think of some args.
	// Ideas:
	// 1. A flag to specify the working directory
	// 2. A flag specifying specific tools regardless of language

	// var fPath string
	// var fType string

	// fmt.Print("Insert Filepath Here: ")
	// fmt.Scan(&fPath)

	// fType = Detect_Lang(fPath)

	// if err := nix.RunNixShell([]string{"lolcat", "cowsay"}, "cowsay \"", fType, "\" | lolcat"); err != nil {
	// 	log.Fatal("Failed to run nix-shell", "error", err)
	// }
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

	tool_arr := []tools.SecurityTool{
		tools.NewGosecTool(),
		tools.NewGrypeTool(),
		tools.NewOSVScannerTool(),
	}
	tool_out_map, errs := tools.RunAllToolsForLanguage(tool_arr, "go", cwd)
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
