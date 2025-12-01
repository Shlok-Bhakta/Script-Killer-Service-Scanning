package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"scriptkiller/src/nix"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type BanditOutput struct {
	Metrics map[string]any `json:"metrics"`
	Results []BanditIssue  `json:"results"`
}

type BanditIssue struct {
	Code            string `json:"code"`
	Filename        string `json:"filename"`
	IssueConfidence string `json:"issue_confidence"`
	IssueSeverity   string `json:"issue_severity"`
	IssueText       string `json:"issue_text"`
	LineNumber      int    `json:"line_number"`
	TestID          string `json:"test_id"`
	TestName        string `json:"test_name"`
}

type BanditScannerTool struct {
	info   ToolInfo
	logger *log.Logger
}

func NewBanditPyTool() *BanditScannerTool {
	return &BanditScannerTool{
		info: ToolInfo{
			Name: "bandit",
			Desc: "Bandit is a tool designed to find common security issues in Python code.",
			Url:  "https://github.com/PyCQA/bandit",
		},
		logger: log.WithPrefix("bandit"),
	}
}

func (b *BanditScannerTool) GetToolInfo() ToolInfo {
	return b.info
}

func (b *BanditScannerTool) Run(targetPath string) (ToolOutput, error) {
	startTime := time.Now()
	toolOut := ToolOutput{
		ToolName:  b.info.Name,
		StartTime: startTime,
		RawOutput: "",
		RawStderr: "",
		Critical:  []Finding{},
		Warnings:  []Finding{},
		Info:      []Finding{},
		Other:     []Finding{},
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return toolOut, fmt.Errorf("target path does not exist: %s", targetPath)
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		b.logger.Error("failed to get absolute path", "error", err)
		return toolOut, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	b.logger.Info("Running bandit on Python code", "path", absPath)

	output, err := nix.RunNixShellWithOutput(
		[]string{"python3", "python3Packages.bandit"},
		fmt.Sprintf("cd '%s' && bandit -r -f json . 2>/dev/null", absPath),
	)

	if err != nil && len(output) == 0 {
		b.logger.Error("bandit failed", "error", err)
		return toolOut, fmt.Errorf("bandit failed: %w", err)
	}

	// Bandit returns non-zero exit codes when issues are found, which is expected
	if err != nil {
		b.logger.Warn("Bandit exited with non-zero status (issues found)", "error", err)
	}

	toolOut.RawOutput = string(output)

	if len(output) == 0 {
		b.logger.Warn("No output from bandit")
		return toolOut, nil
	}

	b.logger.Debug("Bandit raw output", "output", string(output)[:min(500, len(output))])

	var parsed BanditOutput
	if jsonErr := json.Unmarshal(output, &parsed); jsonErr != nil {
		b.logger.Warn("Failed to parse bandit JSON output", "error", jsonErr, "output_preview", string(output)[:min(200, len(output))])
		return toolOut, nil
	}

	// Convert Bandit issues to Findings
	for _, issue := range parsed.Results {
		relPath, err := filepath.Rel(absPath, issue.Filename)
		if err != nil {
			relPath = issue.Filename
		}

		finding := Finding{
			ID:       fmt.Sprintf("%s-%s-%d", issue.TestID, relPath, issue.LineNumber),
			Severity: SeverityInfo,
			Message:  fmt.Sprintf("[%s] %s", issue.TestID, issue.IssueText),
			Location: Location{
				File: relPath,
				Line: issue.LineNumber,
			},
			Suggestion: strings.TrimSpace(issue.Code),
			Metadata: map[string]string{
				"test_id":    issue.TestID,
				"test_name":  issue.TestName,
				"confidence": issue.IssueConfidence,
			},
			Suppressed: false,
		}

		switch strings.ToLower(issue.IssueSeverity) {
		case "high":
			finding.Severity = SeverityCritical
			toolOut.Critical = append(toolOut.Critical, finding)
		case "medium":
			finding.Severity = SeverityWarning
			toolOut.Warnings = append(toolOut.Warnings, finding)
		case "low":
			finding.Severity = SeverityInfo
			toolOut.Info = append(toolOut.Info, finding)
		default:
			finding.Severity = SeverityOther
			toolOut.Other = append(toolOut.Other, finding)
		}
	}

	b.logger.Info("Bandit scan complete", "critical", len(toolOut.Critical), "warnings", len(toolOut.Warnings), "info", len(toolOut.Info))

	toolOut.Duration = time.Since(startTime).Milliseconds()
	return toolOut, nil
}

func (b *BanditScannerTool) IsApplicable(language string) bool {
	lang := strings.ToLower(language)
	return lang == "python" || lang == "py"
}

func (b *BanditScannerTool) Validate() error {
	output, err := nix.RunNixShellWithOutput(
		[]string{"python3", "python3Packages.bandit"},
		"bandit --version",
	)
	if err != nil {
		return fmt.Errorf("bandit is not available: %w", err)
	}
	b.logger.Debug("bandit version", "output", string(output))
	return nil
}

func (b *BanditScannerTool) GetConfigSchema() map[string]any {
	return map[string]any{
		"excluded_tests": map[string]any{
			"type":        "array",
			"description": "List of bandit test IDs to exclude (e.g., B101, B102)",
			"items":       map[string]string{"type": "string"},
		},
		"severity_threshold": map[string]any{
			"type":        "string",
			"description": "Minimum severity level to report",
			"enum":        []string{"low", "medium", "high"},
			"default":     "low",
		},
		"confidence_threshold": map[string]any{
			"type":        "string",
			"description": "Minimum confidence level to report",
			"enum":        []string{"low", "medium", "high"},
			"default":     "low",
		},
	}
}

// Helper function for line number conversion if needed
func (b *BanditScannerTool) parseLineNumber(line interface{}) int {
	switch v := line.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}