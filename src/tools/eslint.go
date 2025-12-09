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

// ESLint finding structure
type eslintFinding struct {
    FilePath string `json:"filePath"`
    Messages []struct {
        RuleId    string `json:"ruleId"`
        Severity  int    `json:"severity"`
        Message   string `json:"message"`
        Line      int    `json:"line"`
        Column    int    `json:"column"`
        EndLine   int    `json:"endLine"`
        EndColumn int    `json:"endColumn"`
    } `json:"messages"`
}

type ESLintSecurityTool struct {
    info   ToolInfo
    logger *log.Logger
}

func NewESLintSecurityTool() *ESLintSecurityTool {
    return &ESLintSecurityTool{
        info: ToolInfo{
            Name: "eslint-security",
            Desc: "Scans JavaScript/TypeScript for security issues using ESLint and eslint-plugin-security",
            Url:  "https://github.com/nodesecurity/eslint-plugin-security",
        },
        logger: log.WithPrefix("eslint-security"),
    }
}

func (e *ESLintSecurityTool) GetToolInfo() ToolInfo {
    return e.info
}

func (e *ESLintSecurityTool) Run(targetPath string) (ToolOutput, error) {
    startTime := time.Now()
    toolOut := ToolOutput{
        ToolName:  e.info.Name,
        StartTime: startTime,
        RawOutput: "",
        Critical:  []Finding{},
        Warnings:  []Finding{},
        Info:      []Finding{},
        Other:     []Finding{},
    }

    absPath, err := filepath.Abs(targetPath)
    if err != nil {
        e.logger.Error("failed to get absolute path", "error", err)
        return toolOut, fmt.Errorf("failed to get absolute path: %w", err)
    }

    e.logger.Info("Running ESLint on JS/TS code", "path", absPath)

    // Command: make sure plugin is available, then run scan
    cmdStr := "npm install --no-audit --no-fund eslint-plugin-security && eslint . --ext .js,.jsx,.ts,.tsx -c .eslintrc.json -f json 2>/dev/null || true"

    output, err := nix.RunNixShellWithOutput(
        []string{"nodejs", "eslint"},
        fmt.Sprintf("cd '%s' && %s", absPath, cmdStr),
    )
    toolOut.RawOutput = string(output)
    if err != nil && len(output) == 0 {
        e.logger.Error("ESLint failed", "error", err)
        return toolOut, fmt.Errorf("eslint failed: %w", err)
    }

    // Debug output preview
    previewLen := 600
    if len(output) < previewLen {
        previewLen = len(output)
    }
    e.logger.Debug("ESLint raw output", "output", string(output)[:previewLen])

    // Parse output JSON
    var files []eslintFinding
    if err := json.Unmarshal(output, &files); err != nil {
        e.logger.Error("Failed to parse ESLint JSON", "error", err, "json", truncateStrings(string(output), 400))
        return toolOut, nil
    }

    for _, file := range files {
        for _, msg := range file.Messages {
            severity := SeverityInfo
            // Treat high-severity rules from plugin as critical
            if msg.Severity == 2 {
                switch msg.RuleId {
                case "security/detect-eval-with-expression",
                    "security/detect-object-injection",
                    "security/detect-child-process",
                    "security/detect-buffer-noassert":
                    severity = SeverityCritical
                default:
                    severity = SeverityWarning
                }
            } else if msg.Severity == 1 {
                severity = SeverityWarning
            }

            finding := Finding{
                ID: fmt.Sprintf("%s:%s:%d:%d", file.FilePath, msg.RuleId, msg.Line, msg.Column),
                Severity: severity,
                Message:  fmt.Sprintf("[%s] %s", msg.RuleId, msg.Message),
                Location: Location{
                    File:      file.FilePath,
                    Line:      msg.Line,
                    Column:    msg.Column,
                    EndLine:   msg.EndLine,
                    EndColumn: msg.EndColumn,
                },
                Suggestion: "Review this code and refactor to eliminate this security issue.",
                Metadata: map[string]string{
                    "rule_id": msg.RuleId,
                },
                Suppressed: false,
            }

            switch severity {
            case SeverityCritical:
                toolOut.Critical = append(toolOut.Critical, finding)
            case SeverityWarning:
                toolOut.Warnings = append(toolOut.Warnings, finding)
            case SeverityInfo, SeverityOther:
                toolOut.Info = append(toolOut.Info, finding)
            }
        }
    }

    toolOut.Duration = time.Since(startTime).Milliseconds()
    e.logger.Info("ESLint scan complete", "critical", len(toolOut.Critical), "warnings", len(toolOut.Warnings), "info", len(toolOut.Info))
    return toolOut, nil
}

func (e *ESLintSecurityTool) IsApplicable(language string) bool {
    lang := strings.ToLower(language)
    return lang == "javascript" || lang == "js" ||
        lang == "typescript" || lang == "ts"
}

func (e *ESLintSecurityTool) Validate() error {
    // Just test that ESLint can run inside nix shell, and plugin is installed
    output, err := nix.RunNixShellWithOutput(
        []string{"nodejs", "eslint"},
        "npm install --no-audit --no-fund eslint-plugin-security && eslint --version",
    )
    if err != nil {
        return fmt.Errorf("eslint is not available: %w", err)
    }
    e.logger.Debug("eslint version", "output", string(output))
    return nil
}

func (e *ESLintSecurityTool) GetConfigSchema() map[string]any {
    return map[string]any{
        "config_path": map[string]any{
            "type":        "string",
            "description": "Path to .eslintrc config file",
            "default":     ".eslintrc.json",
        },
    }
}

// Utility for truncating long output in logs
func truncateStrings(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}