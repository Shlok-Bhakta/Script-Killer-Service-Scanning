package nix

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/log"
)

//go:embed nix-portable-binary
var nixPortableBinary []byte

const nixPortableVersion = "v012"

func getNixPortablePath() (string, error) {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "scriptkiller")
	nixPortablePath := filepath.Join(cacheDir, fmt.Sprintf("nix-portable-%s", nixPortableVersion))

	if _, err := os.Stat(nixPortablePath); err == nil {
		log.Debug("Found nix-portable in cache", "path", nixPortablePath)
		return nixPortablePath, nil
	}

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
	return nixPortablePath, nil
}

func RunNixShell(packages []string, command string, args ...string) error {
	log.Debug("Getting nix-portable path")
	nixPath, err := getNixPortablePath()
	if err != nil {
		log.Fatal("Failed to get nix-portable", "error", err)
		return err
	}

	nixArgs := []string{"nix-shell", "-p"}
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

	nixArgs := []string{"nix-shell", "-p"}
	nixArgs = append(nixArgs, packages...)
	nixArgs = append(nixArgs, "--run", command+" "+joinArgs(args))

	log.Info("Running nix-shell with output capture", "packages", packages, "command", command)

	cmd := exec.Command(nixPath, nixArgs...)
	cmd.Env = append(os.Environ(), "LC_ALL=C")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Failed to run nix-shell", "error", err)
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
