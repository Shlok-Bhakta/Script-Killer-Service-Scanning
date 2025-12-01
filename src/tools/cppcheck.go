package tools

import (
    "encoding/xml"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
    "scriptkiller/src/nix"

    "github.com/charmbracelet/log"
)

type cppcheckLocation struct {
    File   string `xml:"file,attr"`
    Line   string `xml:"line,attr"`
    Column string `xml:"column,attr"`
}

type cppcheckError struct {
    ID       string             `xml:"id,attr"`
    Severity string             `xml:"severity,attr"`
    Msg      string             `xml:"msg,attr"`
    Verbose  string             `xml:"verbose,attr"`
    Locs     []cppcheckLocation `xml:"location"`
}

type cppcheckOutput struct {
    Errors []cppcheckError `xml:"errors>error"`
}

type CppcheckTool struct {
    info   ToolInfo
    logger *log.Logger
}

func NewCppcheckTool() *CppcheckTool {
    return &CppcheckTool{
        info: ToolInfo{
            Name: "cppcheck",
            Desc: "Static analysis tool for C/C++ to find bugs and misuses.",
            Url:  "https://cppcheck.sourceforge.io/",
        },
        logger: log.WithPrefix("cppcheck"),
    }
}

func (c *CppcheckTool) GetToolInfo() ToolInfo {
    return c.info
}

func (c *CppcheckTool) Run(targetPath string) (ToolOutput, error) {
    startTime := time.Now()
    toolOut := ToolOutput{
        ToolName:  c.info.Name,
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
        c.logger.Error("failed to get absolute path", "error", err)
        return toolOut, fmt.Errorf("failed to resolve absolute path: %w", err)
    }

    c.logger.Info("Running cppcheck", "path", absPath)

    // According to official cppcheck manual, XML output is produced with --xml
    // and --xml-version=2; cppcheck writes XML to stderr so write to a temp
    // file and cat it. Use the same nix-shell approach as other tool wrappers.
    cmd := fmt.Sprintf("cd '%s' && cppcheck --enable=all --inconclusive --xml --xml-version=2 . 2>/tmp/cppcheck.xml || true; cat /tmp/cppcheck.xml 2>/dev/null || echo ''", absPath)
    output, err := nix.RunNixShellWithOutput(
        []string{"cppcheck"},
        cmd,
    )

    if err != nil && len(output) == 0 {
        c.logger.Error("cppcheck failed", "error", err)
        return toolOut, fmt.Errorf("cppcheck failed: %w", err)
    }

    toolOut.RawOutput = string(output)

    if len(output) == 0 {
        c.logger.Info("No output from cppcheck")
        toolOut.Duration = time.Since(startTime).Milliseconds()
        return toolOut, nil
    }

    previewLen := 500
    if len(output) < previewLen {
        previewLen = len(output)
    }
    c.logger.Debug("Cppcheck raw output", "output", string(output)[:previewLen])

    var parsed cppcheckOutput
    if err := xml.Unmarshal(output, &parsed); err != nil {
        c.logger.Warn("Failed to parse cppcheck XML output", "error", err)
        return toolOut, nil
    }

    for _, e := range parsed.Errors {
        locFile := ""
        line := 0
        col := 0
        if len(e.Locs) > 0 {
            loc := e.Locs[0]
            locFile = loc.File
            if n, perr := strconv.Atoi(loc.Line); perr == nil {
                line = n
            }
            if n, perr := strconv.Atoi(loc.Column); perr == nil {
                col = n
            }
        }

        relPath := locFile
        if locFile != "" {
            if rp, rerr := filepath.Rel(absPath, locFile); rerr == nil {
                relPath = rp
            }
        }

        id := e.ID
        if id == "" {
            id = fmt.Sprintf("cppcheck-%s-%d", strings.ReplaceAll(e.Msg, " ", "-"), line)
        }

        finding := Finding{
            ID:       fmt.Sprintf("%s-%s-%d", id, relPath, line),
            Severity: SeverityInfo,
            Message:  e.Msg,
            Location: Location{File: relPath, Line: line, Column: col},
            Suggestion: strings.TrimSpace(e.Verbose),
            Metadata: map[string]string{
                "cppcheck_id": id,
            },
            Suppressed: false,
        }

        switch strings.ToLower(e.Severity) {
        case "error":
            finding.Severity = SeverityCritical
            toolOut.Critical = append(toolOut.Critical, finding)
        case "warning":
            finding.Severity = SeverityWarning
            toolOut.Warnings = append(toolOut.Warnings, finding)
        case "style", "performance", "portability", "information":
            finding.Severity = SeverityInfo
            toolOut.Info = append(toolOut.Info, finding)
        default:
            finding.Severity = SeverityOther
            toolOut.Other = append(toolOut.Other, finding)
        }
    }

    c.logger.Info("Cppcheck scan complete", "critical", len(toolOut.Critical), "warnings", len(toolOut.Warnings), "info", len(toolOut.Info))
    toolOut.Duration = time.Since(startTime).Milliseconds()
    return toolOut, nil
}

func (c *CppcheckTool) IsApplicable(language string) bool {
    lang := strings.ToLower(language)
    supported := []string{"c", "c++", "cpp", "c/c++", "h", "hpp"}
    for _, s := range supported {
        if lang == s {
            return true
        }
    }
    return false
}

func (c *CppcheckTool) Validate() error {
    output, err := nix.RunNixShellWithOutput(
        []string{"cppcheck"},
        "cppcheck --version",
    )
    if err != nil {
        return fmt.Errorf("cppcheck is not available: %w", err)
    }
    c.logger.Debug("cppcheck version", "output", string(output))
    return nil
}

func (c *CppcheckTool) GetConfigSchema() map[string]any {
    return map[string]any{
        "enable_checks": map[string]any{
            "type":        "array",
            "description": "List of checks to enable (passed to --enable)",
            "items":       map[string]string{"type": "string"},
        },
        "exclude": map[string]any{
            "type":        "array",
            "description": "Paths to exclude from analysis",
            "items":       map[string]string{"type": "string"},
        },
    }
}
