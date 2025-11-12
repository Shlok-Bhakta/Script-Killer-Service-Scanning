package main

import (
	"github.com/charmbracelet/log"
	"fmt"
	"path/filepath"
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
	log.Debug("Cookie 🍪") // should not be seen in normal output
	log.Info("Hello World!")

	var fPath string
	var fType string 

	fmt.Print("Insert Filepath Here: ")
  	fmt.Scan(&fPath)

	fType = Detect_Lang(fPath)

	if err := runNixShell([]string{"lolcat", "cowsay"}, "cowsay \"",fType,"\" | lolcat"); err != nil {
		log.Fatal("Failed to run nix-shell", "error", err)
	}
}
