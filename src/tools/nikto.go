package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"scriptkiller/src/nix"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type NiktoScannerTool struct {
	info   ToolInfo
	logger *log.Logger
}

func NewNiktoScannerTool() *NiktoScannerTool {
	return &NiktoScannerTool{
		info: ToolInfo{
			Name: "nikto",
			Desc: "Web vulnerability scanner",
			Url:  "https://cirt.net/Nikto2",
		},
		logger: log.WithPrefix("nikto"),
	}
}

func (t *NiktoScannerTool) GetToolInfo() ToolInfo {
	return t.info
}

type niktoJSON struct {
	Host     string `json:"host"`
	Site     string `json:"site"`
	Port     int    `json:"port"`
	SSL      bool   `json:"ssl"`
	Findings []struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		URI         string `json:"uri"`
		OSVDB       string `json:"osvdb"`
	} `json:"findings"`
}

func (t *NiktoScannerTool) Run(targetURL string) (ToolOutput, error) {
	start := time.Now()
	out := ToolOutput{
		ToolName:  t.info.Name,
		StartTime: start,
		Critical:  []Finding{},
		Warnings:  []Finding{},
		Info:      []Finding{},
		Other:     []Finding{},
	}

	if _, err := url.ParseRequestURI(targetURL); err != nil {
		return out, fmt.Errorf("invalid URL: %s", targetURL)
	}

	t.logger.Info("Running Nikto", "target", targetURL)

	cmd := fmt.Sprintf("nikto -h %s -Format json 2>&1", targetURL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	raw, err := nix.RunNixShellWithOutputCtx(ctx, []string{"nikto"}, cmd)

	out.RawOutput = string(raw)
	out.RawStderr = ""

	if err != nil && len(raw) == 0 {
		t.logger.Error("nikto scan failed", "error", err)
		return out, fmt.Errorf("nikto failed: %w", err)
	}

	var parsed niktoJSON
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.logger.Warn("Failed to parse nikto JSON", "error", err)
		return out, nil // non-fatal — consistent with OSV scanner behavior
	}

	for _, f := range parsed.Findings {
		severity := t.mapSeverity(f.Description)

		msg := fmt.Sprintf("%s", f.Description)

		suggestion := ""
		if f.OSVDB != "" {
			suggestion = fmt.Sprintf("https://www.cvedetails.com/osvdb/%s", f.OSVDB)
		}

		finding := Finding{
			ID:       fmt.Sprintf("nikto-%s", f.ID),
			Severity: severity,
			Message:  msg,
			Location: Location{
				File: f.URI,
			},
			Suggestion: suggestion,
			Metadata: map[string]string{
				"uri":     f.URI,
				"osvdb":   f.OSVDB,
				"scanner": "nikto",
			},
		}

		switch severity {
		case SeverityCritical:
			out.Critical = append(out.Critical, finding)
		case SeverityWarning:
			out.Warnings = append(out.Warnings, finding)
		case SeverityInfo:
			out.Info = append(out.Info, finding)
		default:
			out.Other = append(out.Other, finding)
		}
	}

	out.Duration = time.Since(start).Milliseconds()
	t.logger.Info("Nikto scan complete", "critical", len(out.Critical), "warnings", len(out.Warnings), "info", len(out.Info))

	return out, nil
}

func (t *NiktoScannerTool) mapSeverity(desc string) Severity {
	desc = strings.ToLower(desc)

	if strings.Contains(desc, "remote code execution") ||
		strings.Contains(desc, "rce") ||
		strings.Contains(desc, "shell") ||
		strings.Contains(desc, "admin access") ||
		strings.Contains(desc, "critical") {

		return SeverityCritical
	}

	if strings.Contains(desc, "xss") ||
		strings.Contains(desc, "sqli") ||
		strings.Contains(desc, "directory traversal") ||
		strings.Contains(desc, "csrf") ||
		strings.Contains(desc, "authentication") {

		return SeverityWarning
	}

	if strings.Contains(desc, "info") ||
		strings.Contains(desc, "banner") ||
		strings.Contains(desc, "version") {

		return SeverityInfo
	}

	return SeverityOther
}

func (t *NiktoScannerTool) Validate() error {
	out, err := nix.RunNixShellWithOutput([]string{"nikto"}, "nikto -Version")
	if err != nil {
		return fmt.Errorf("nikto is not available: %w", err)
	}
	t.logger.Debug("nikto version", "output", string(out))
	return nil
}

func (t *NiktoScannerTool) GetConfigSchema() map[string]any {
	return map[string]any{
		"severity_threshold": map[string]any{
			"type":        "string",
			"description": "Minimum severity level to report",
			"enum":        []string{"info", "warning", "critical"},
			"default":     "info",
		},
	}
}

func (t *NiktoScannerTool) IsApplicable(target string) bool {
	return true
}
