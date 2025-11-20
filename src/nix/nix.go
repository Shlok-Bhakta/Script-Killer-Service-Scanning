package nix

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"
)

//go:embed nix-portable-binary
var nixPortableBinary []byte

const nixPortableVersion = "v012"

var (
	extractMutex sync.Mutex
	initOnce     sync.Once
)

func getNixPortablePath() (string, error) {
	extractMutex.Lock()
	defer extractMutex.Unlock()

	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "scriptkiller")
	nixPortablePath := filepath.Join(cacheDir, fmt.Sprintf("nix-portable-%s", nixPortableVersion))

	if _, err := os.Stat(nixPortablePath); err != nil {
		log.Info("Extracting nix-portable to cache", "path", nixPortablePath)

		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			log.Fatal("Failed to create cache directory", "error", err)
			return "", fmt.Errorf("failed to create cache dir: %w", err)
		}

		if err := os.WriteFile(nixPortablePath, nixPortableBinary, 0755); err != nil {
			log.Fatal("Failed to write nix-portable binary", "error", err)
			return "", fmt.Errorf("failed to extract nix-portable: %w", err)
		}
		log.Info("Successfully extracted nix-portable", "path", nixPortablePath)
	} else {
		log.Debug("Found nix-portable in cache", "path", nixPortablePath)
	}

	initOnce.Do(func() {
		log.Info("Initializing nix-portable runtime (one-time setup)...")
		cmd := exec.Command(nixPortablePath, "nix-shell", "--version")
		cmd.Env = append(os.Environ(), "LC_ALL=C")
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Warn("Nix initialization warning", "error", err, "output", string(out))
		} else {
			log.Info("Nix portable runtime initialized successfully")
		}
	})

	return nixPortablePath, nil
}

func RunNixShell(packages []string, command string, args ...string) error {
	log.Debug("Getting nix-portable path")
	nixPath, err := getNixPortablePath()
	if err != nil {
		log.Fatal("Failed to get nix-portable", "error", err)
		return err
	}

	nixArgs := []string{"nix-shell", "-I", "nixpkgs=https://github.com/NixOS/nixpkgs/archive/91c9a64ce2a84e648d0cf9671274bb9c2fb9ba60.tar.gz", "-p"}
	nixArgs = append(nixArgs, packages...)
	nixArgs = append(nixArgs, "--run", command+" "+joinArgs(args))

	log.Info("Running nix-shell", "packages", packages, "command", command)

	cmd := exec.Command(nixPath, nixArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), "LC_ALL=C")

	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to run nix-shell", "error", err)
		return err
	}

	log.Debug("nix-shell completed successfully")
	return nil
}

func RunNixShellWithOutput(packages []string, command string, args ...string) ([]byte, error) {
	log.Debug("Getting nix-portable path")
	nixPath, err := getNixPortablePath()
	if err != nil {
		log.Error("Failed to get nix-portable", "error", err)
		return nil, err
	}

	nixArgs := []string{"nix-shell", "-I", "nixpkgs=https://github.com/NixOS/nixpkgs/archive/91c9a64ce2a84e648d0cf9671274bb9c2fb9ba60.tar.gz", "-p"}
	nixArgs = append(nixArgs, packages...)
	nixArgs = append(nixArgs, "--run", command+" "+joinArgs(args))

	log.Info("Running nix-shell with output capture", "packages", packages, "command", command)

	cmd := exec.Command(nixPath, nixArgs...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "GOTOOLCHAIN=local")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	
	if err != nil {
		log.Error("Failed to run nix-shell", "error", err, "stderr", stderr.String())
		return output, err
	}

	log.Debug("nix-shell completed successfully")
	return output, nil
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}