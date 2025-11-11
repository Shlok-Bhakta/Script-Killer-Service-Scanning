package main

import (
	"github.com/charmbracelet/log"
)

func main() {
	log.Debug("Cookie 🍪")
	log.Info("Hello World!")

	if err := runNixShell([]string{"lolcat", "cowsay"}, "cowsay \"hi\" | lolcat"); err != nil {
		log.Fatal("Failed to run nix-shell", "error", err)
	}
}
