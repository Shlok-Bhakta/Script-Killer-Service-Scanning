package main

import (
	"fmt"
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

	tools := []tools.SecurityTool{
		tools.NewGosecTool(),
	}

	for _, tool := range tools {
		// toolInfo := tool.GetToolInfo()
		log.Info(tool.Run("."))
	}
}
