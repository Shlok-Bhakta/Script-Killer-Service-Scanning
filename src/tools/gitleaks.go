package tools

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"scriptkiller/src/nix"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type GitleaksTool struct {
	info   ToolInfo
	logger *log.Logger
}

// Gitleaks JSON output structure
type gitleaksFinding struct {
	Description string `json:"Description"`
	StartLine   int    `json:"StartLine"`
	EndLine     int    `json:"EndLine"`
	StartColumn int    `json:"StartColumn"`
	EndColumn   int    `json:"EndColumn"`
	Match       string `json:"Match"`
	Secret      string `json:"Secret"`
	File        string `json:"File"`
	SymlinkFile string `json:"SymlinkFile,omitempty"`
	Commit      string `json:"Commit,omitempty"`
	Entropy     float64 `json:"Entropy"`
	Author      string `json:"Author,omitempty"`
	Email       string `json:"Email,omitempty"`
	Date        string `json:"Date,omitempty"`
	Message     string `json:"Message,omitempty"`
	Tags        []string `json:"Tags,omitempty"`
	RuleID      string `json:"RuleID"`
	Fingerprint string `json:"Fingerprint"`
}

func NewGitleaksTool() *GitleaksTool {
	return &GitleaksTool{
		info: ToolInfo{
			Name: "gitleaks",
			Desc: "Detects secrets like passwords, API keys, and tokens in git repos and files",
			Url:  "https://github.com/gitleaks/gitleaks",
		},
		logger: log.WithPrefix("gitleaks"),
	}
}

func (g *GitleaksTool) GetToolInfo() ToolInfo {
	return g.info
}

func (g *GitleaksTool) Run(targetPath string) (ToolOutput, error) {
	startTime := time.Now()
	toolOut := ToolOutput{
		ToolName:  g.info.Name,
		StartTime: startTime,
		RawOutput: "",
		RawStderr: "",
		Critical:  []Finding{},
		Warnings:  []Finding{},
		Info:      []Finding{},
		Other:     []Finding{},
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		g.logger.Error("failed to get absolute path", "error", err)
		return toolOut, fmt.Errorf("failed to get absolute path: %w", err)
	}

	g.logger.Info("Running gitleaks", "path", absPath)

	// Run gitleaks with dir command on the directory (v8.19.0+ uses 'dir' instead of 'detect')
	// Write to temp file then cat it to avoid stdout buffering issues
	// --report-format json for JSON output
	// --exit-code 0 prevents non-zero exit on findings
	// --no-banner suppresses the banner for cleaner output
	output, err := nix.RunNixShellWithOutput(
		[]string{"gitleaks"},
		fmt.Sprintf("gitleaks dir '%s' --report-format json --report-path /tmp/gitleaks-report.json --exit-code 0 --no-banner >&2; cat /tmp/gitleaks-report.json 2>/dev/null || echo '[]'", absPath),
	)
	if err != nil && len(output) == 0 {
		g.logger.Error("gitleaks failed", "error", err)
		return toolOut, fmt.Errorf("gitleaks failed: %w", err)
	}

	toolOut.RawOutput = string(output)

	// Log actual output length for debugging
	g.logger.Debug("Gitleaks output received", "length", len(output), "output", string(output))

	if len(output) == 0 {
		g.logger.Info("No secrets found by gitleaks (empty output)")
		return toolOut, nil
	}

	previewLen := 500
	if len(output) < previewLen {
		previewLen = len(output)
	}
	g.logger.Debug("Gitleaks raw output", "output", string(output)[:previewLen])

	var findings []gitleaksFinding
	if err := json.Unmarshal(output, &findings); err != nil {
		// If output is empty array or no findings, this is fine
		if strings.TrimSpace(string(output)) == "[]" || strings.TrimSpace(string(output)) == "" {
			g.logger.Info("No secrets found by gitleaks")
			return toolOut, nil
		}
		g.logger.Warn("Failed to parse gitleaks JSON output", "error", err, "output_preview", string(output)[:previewLen])
		return toolOut, nil
	}

	for _, f := range findings {
		// Mask the secret for display (show first 4 chars + asterisks)
		maskedSecret := maskSecret(f.Secret)

		message := fmt.Sprintf("%s: %s", f.RuleID, f.Description)
		if maskedSecret != "" {
			message = fmt.Sprintf("%s (secret: %s)", message, maskedSecret)
		}

		finding := Finding{
			ID:       f.Fingerprint,
			Severity: SeverityCritical, // Secrets are always critical
			Message:  message,
			Location: Location{
				File:      f.File,
				Line:      f.StartLine,
				Column:    f.StartColumn,
				EndLine:   f.EndLine,
				EndColumn: f.EndColumn,
			},
			Suggestion: "Remove the secret from source code and rotate the credential immediately",
			Metadata: map[string]string{
				"rule_id":     f.RuleID,
				"description": f.Description,
				"entropy":     fmt.Sprintf("%.2f", f.Entropy),
				"match":       truncateString(f.Match, 100),
			},
			Suppressed: false,
		}

		// Add git-related metadata if available
		if f.Commit != "" {
			finding.Metadata["commit"] = f.Commit
		}
		if f.Author != "" {
			finding.Metadata["author"] = f.Author
		}
		if f.Email != "" {
			finding.Metadata["email"] = f.Email
		}
		if f.Date != "" {
			finding.Metadata["date"] = f.Date
		}
		if len(f.Tags) > 0 {
			finding.Metadata["tags"] = strings.Join(f.Tags, ", ")
		}

		// All secret findings are critical
		toolOut.Critical = append(toolOut.Critical, finding)
	}

	g.logger.Info("Gitleaks scan complete", "secrets_found", len(toolOut.Critical))

	toolOut.Duration = time.Since(startTime).Milliseconds()
	return toolOut, nil
}

// IsApplicable returns true for all languages since secrets can be in any codebase
func (g *GitleaksTool) IsApplicable(language string) bool {
	// Gitleaks should scan any project for secrets regardless of language
	return true
}

func (g *GitleaksTool) Validate() error {
	output, err := nix.RunNixShellWithOutput(
		[]string{"gitleaks"},
		"gitleaks version",
	)
	if err != nil {
		return fmt.Errorf("gitleaks is not available: %w", err)
	}
	g.logger.Debug("gitleaks version", "output", string(output))
	return nil
}

func (g *GitleaksTool) GetConfigSchema() map[string]any {
	return map[string]any{
		"config_path": map[string]any{
			"type":        "string",
			"description": "Path to custom gitleaks config file (.gitleaks.toml)",
			"default":     "",
		},
		"baseline_path": map[string]any{
			"type":        "string",
			"description": "Path to baseline file to ignore known secrets",
			"default":     "",
		},
	}
}

// maskSecret masks a secret for safe display, showing only first few characters
func maskSecret(secret string) string {
	if len(secret) == 0 {
		return ""
	}
	if len(secret) <= 4 {
		return "****"
	}
	// Show first 4 characters, mask the rest
	return secret[:4] + strings.Repeat("*", min(len(secret)-4, 8))
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
