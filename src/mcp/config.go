package mcp

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type OpencodeConfig struct {
	Schema string                          `json:"$schema"`
	MCP    map[string]OpencodeServerConfig `json:"mcp,omitempty"`
}

type OpencodeServerConfig struct {
	Type    string `json:"type"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

func FindAvailablePort(startPort int) int {
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port
		}
	}
	return 0
}

func InstallToOpenCode(port int) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(cwd, "opencode.jsonc")

	var config OpencodeConfig

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read opencode.jsonc: %w", err)
		}
		config = OpencodeConfig{
			Schema: "https://opencode.ai/config.json",
			MCP:    make(map[string]OpencodeServerConfig),
		}
	} else {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse opencode.jsonc: %w", err)
		}
		if config.MCP == nil {
			config.MCP = make(map[string]OpencodeServerConfig)
		}
	}

	config.MCP["script-killer"] = OpencodeServerConfig{
		Type:    "remote",
		URL:     fmt.Sprintf("http://localhost:%d/mcp", port),
		Enabled: true,
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write opencode.jsonc: %w", err)
	}

	return nil
}

func UninstallFromOpenCode() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(cwd, "opencode.jsonc")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read opencode.jsonc: %w", err)
	}

	var config OpencodeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse opencode.jsonc: %w", err)
	}

	if config.MCP != nil {
		delete(config.MCP, "script-killer")
	}

	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write opencode.jsonc: %w", err)
	}

	return nil
}
