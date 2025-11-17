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

type GrypeTool struct {
    info   ToolInfo
    logger *log.Logger
}

type grypeFix struct {
    Versions []string `json:"versions"`
}

type grypeVulnerability struct {
    ID       string    `json:"id"`
    Severity string    `json:"severity"`
    Fix      grypeFix  `json:"fix,omitempty"`
    URLs     []string  `json:"urls,omitempty"`
}

type grypeArtifact struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Type    string `json:"type,omitempty"`
}

type grypeMatch struct {
    Vulnerability grypeVulnerability `json:"vulnerability"`
    Artifact      grypeArtifact      `json:"artifact"`
    Location      map[string]any     `json:"location,omitempty"`
}

type grypeOutput struct {
    Matches []grypeMatch `json:"matches"`
}

func NewGrypeTool() *GrypeTool {
    return &GrypeTool{
        info: ToolInfo{
            Name: "grype",
            Desc: "Scans container images and filesystems for vulnerabilities (Anchore Grype)",
            Url:  "https://github.com/anchore/grype",
        },
        logger: log.WithPrefix("grype"),
    }
}

func (g *GrypeTool) GetToolInfo() ToolInfo {
    return g.info
}

func (g *GrypeTool) Run(targetPath string) (ToolOutput, error) {
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

    g.logger.Info("Running grype", "path", absPath)

    output, err := nix.RunNixShellWithOutput(
        []string{"grype"},
        fmt.Sprintf("cd '%s' && grype -o json '%s' 2>/dev/null || true", absPath, absPath),
    )
    if err != nil && len(output) == 0 {
        g.logger.Error("grype failed", "error", err)
        return toolOut, fmt.Errorf("grype failed: %w", err)
    }

    toolOut.RawOutput = string(output)

    if len(output) == 0 {
        g.logger.Warn("No output from grype")
        return toolOut, nil
    }

    previewLen := 500
    if len(output) < previewLen {
        previewLen = len(output)
    }
    g.logger.Debug("Grype raw output", "output", string(output)[:previewLen])

    var gr grypeOutput
    if err := json.Unmarshal(output, &gr); err != nil {
        g.logger.Warn("Failed to parse grype JSON output", "error", err)
        return toolOut, nil
    }

    for _, m := range gr.Matches {
        vuln := m.Vulnerability
        art := m.Artifact

        id := vuln.ID
        if id == "" {
            id = fmt.Sprintf("%s@%s", art.Name, art.Version)
        }

        messageParts := []string{fmt.Sprintf("%s: %s@%s", id, art.Name, art.Version)}
        if len(vuln.URLs) > 0 {
            messageParts = append(messageParts, vuln.URLs[0])
        }
        message := strings.Join(messageParts, " ")

        fixed := ""
        if len(vuln.Fix.Versions) > 0 {
            fixed = vuln.Fix.Versions[0]
        }

        finding := Finding{
            ID:       fmt.Sprintf("%s-%s-%s", id, art.Name, art.Version),
            Severity: SeverityInfo,
            Message:  message,
            Location: Location{File: absPath},
            Suggestion: func() string {
                if fixed != "" {
                    return fmt.Sprintf("upgrade to %s", fixed)
                }
                return ""
            }(),
            Metadata: map[string]string{
                "vuln_id": id,
                "package": art.Name,
                "version": art.Version,
            },
            Suppressed: false,
        }

        switch strings.ToLower(vuln.Severity) {
        case "critical", "high":
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

    g.logger.Info("Grype scan complete", "critical", len(toolOut.Critical), "warnings", len(toolOut.Warnings), "info", len(toolOut.Info))

    toolOut.Duration = time.Since(startTime).Milliseconds()
    return toolOut, nil
}

func (g *GrypeTool) IsApplicable(language string) bool {
    lang := strings.ToLower(language)
    supported := []string{
        "c", "c++", "cpp",
        "dart",
        "elixir",
        "go", "golang",
        "java",
        "javascript", "js", "typescript", "ts",
        "php",
        "python", "py",
        "r",
        "ruby", "rb",
        "rust",
    }
    for _, s := range supported {
        if lang == s {
            return true
        }
    }
    return false
}

func (g *GrypeTool) Validate() error {
    output, err := nix.RunNixShellWithOutput(
        []string{"grype"},
        "grype --version",
    )
    if err != nil {
        return fmt.Errorf("grype is not available: %w", err)
    }
    g.logger.Debug("grype version", "output", string(output))
    return nil
}

func (g *GrypeTool) GetConfigSchema() map[string]any {
    return map[string]any{
        "severity_threshold": map[string]any{
            "type":        "string",
            "description": "Minimum severity level to report",
            "enum":        []string{"low", "medium", "high", "critical"},
            "default":     "low",
        },
    }
}
