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

type GosecTool struct {
	info   ToolInfo
	logger *log.Logger
}

type gosecCWE struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type gosecIssue struct {
	Severity   string   `json:"severity"`
	Confidence string   `json:"confidence"`
	CWE        gosecCWE `json:"cwe"`
	RuleID     string   `json:"rule_id"`
	Details    string   `json:"details"`
	File       string   `json:"file"`
	Code       string   `json:"code"`
	Line       string   `json:"line"`
	Column     string   `json:"column"`
}

type gosecOutput struct {
	Issues []gosecIssue `json:"Issues"`
}

func NewGosecTool() *GosecTool {
	return &GosecTool{
		info: ToolInfo{
			Name: "gosec",
			Desc: "Inspects source code for security problems by scanning the Go AST and SSA code representation.",
			Url:  "https://github.com/securego/gosec",
		},
		logger: log.WithPrefix("gosec"),
	}
}

func (g *GosecTool) GetToolInfo() ToolInfo {
	return g.info
}

func (g *GosecTool) Run(targetPath string) (ToolOutput, error) {
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

	if _, err := os.Stat(filepath.Join(absPath, "go.mod")); os.IsNotExist(err) {
		g.logger.Error("target directory is not a Go module")
		return toolOut, fmt.Errorf("target directory is not a Go module (no go.mod found)")
	}

	g.logger.Info("Running gosec on Go module", "path", absPath)

	output, err := nix.RunNixShellWithOutput(
		[]string{"gosec", "go"},
		fmt.Sprintf("cd '%s' && gosec -fmt=json -no-fail -quiet ./... 2>/dev/null || gosec -fmt=json -no-fail -quiet ./...", absPath),
	)
	if err != nil && len(output) == 0 {
		g.logger.Error("gosec failed", "error", err)
		return toolOut, fmt.Errorf("gosec failed: %w", err)
	}

	toolOut.RawOutput = string(output)

	if len(output) == 0 {
		g.logger.Warn("No output from gosec")
		return toolOut, nil
	}

	g.logger.Debug("Gosec raw output", "output", string(output)[:min(500, len(output))])

	var gosecResult gosecOutput
	if err := json.Unmarshal(output, &gosecResult); err != nil {
		g.logger.Warn("Failed to parse gosec JSON output", "error", err, "output_preview", string(output)[:min(200, len(output))])
		return toolOut, nil
	}

	for _, issue := range gosecResult.Issues {
		cweInfo := ""
		if issue.CWE.ID != "" {
			cweInfo = fmt.Sprintf(" (CWE-%s)", issue.CWE.ID)
		}

		relPath, err := filepath.Rel(absPath, issue.File)
		if err != nil {
			relPath = issue.File
		}

		line, _ := strconv.Atoi(issue.Line)
		col, _ := strconv.Atoi(issue.Column)

		finding := Finding{
			ID:         fmt.Sprintf("%s-%s-%s-%s", issue.RuleID, relPath, issue.Line, issue.Column),
			Severity:   SeverityInfo,
			Message:    fmt.Sprintf("[%s] %s%s", issue.RuleID, issue.Details, cweInfo),
			Location:   Location{File: relPath, Line: line, Column: col},
			Suggestion: strings.TrimSpace(issue.Code),
			Metadata: map[string]string{
				"rule_id":    issue.RuleID,
				"cwe_id":     issue.CWE.ID,
				"confidence": issue.Confidence,
			},
			Suppressed: false,
		}

		switch strings.ToLower(issue.Severity) {
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

	g.logger.Info("Gosec scan complete", "critical", len(toolOut.Critical), "warnings", len(toolOut.Warnings), "info", len(toolOut.Info))

	toolOut.Duration = time.Since(startTime).Milliseconds()
	return toolOut, nil
}

func (g *GosecTool) IsApplicable(language string) bool {
	lang := strings.ToLower(language)
	return lang == "go" || lang == "golang"
}

func (g *GosecTool) Validate() error {
	output, err := nix.RunNixShellWithOutput(
		[]string{"gosec"},
		"gosec --version",
	)
	if err != nil {
		return fmt.Errorf("gosec is not available: %w", err)
	}
	g.logger.Debug("gosec version", "output", string(output))
	return nil
}

func (g *GosecTool) GetConfigSchema() map[string]any {
	return map[string]any{
		"excluded_rules": map[string]any{
			"type":        "array",
			"description": "List of gosec rule IDs to exclude (e.g., G101, G204)",
			"items":       map[string]string{"type": "string"},
		},
		"severity_threshold": map[string]any{
			"type":        "string",
			"description": "Minimum severity level to report",
			"enum":        []string{"low", "medium", "high"},
			"default":     "low",
		},
	}
}
