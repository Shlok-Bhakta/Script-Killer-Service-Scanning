package tools

import (
	"sync"
	"time"
)

type ToolInfo struct {
	Name string
	Desc string
	Url  string
}

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
	SeverityOther    Severity = "other"
)

func (s Severity) Color() string {
	switch s {
	case SeverityCritical:
		return "\033[1;31m"
	case SeverityWarning:
		return "\033[1;33m"
	case SeverityInfo:
		return "\033[1;36m"
	default:
		return "\033[0m"
	}
}

func (s Severity) String() string {
	return string(s)
}

type Location struct {
	File      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

func (l Location) String() string {
	if l.EndLine > 0 && l.EndColumn > 0 {
		return l.File + ":" + string(rune(l.Line)) + ":" + string(rune(l.Column)) + "-" + string(rune(l.EndLine)) + ":" + string(rune(l.EndColumn))
	}
	if l.Line > 0 && l.Column > 0 {
		return l.File + ":" + string(rune(l.Line)) + ":" + string(rune(l.Column))
	}
	if l.Line > 0 {
		return l.File + ":" + string(rune(l.Line))
	}
	return l.File
}

type Finding struct {
	ID         string
	Severity   Severity
	Message    string
	Location   Location
	Suggestion string
	Metadata   map[string]string
	Suppressed bool
}

type ToolOutput struct {
	ToolName  string
	Duration  int64
	ExitCode  int
	Error     error
	StartTime time.Time
	RawOutput string
	RawStderr string
	Critical  []Finding
	Warnings  []Finding
	Info      []Finding
	Other     []Finding
}

func (t *ToolOutput) TotalFindings() int {
	return len(t.Critical) + len(t.Warnings) + len(t.Info) + len(t.Other)
}

func (t *ToolOutput) HasIssues() bool {
	return t.TotalFindings() > 0
}

type SecurityTool interface {
	GetToolInfo() ToolInfo
	Run(targetPath string) (ToolOutput, error)
	IsApplicable(language string) bool
	Validate() error
	GetConfigSchema() map[string]any
}

func RunAllToolsForLanguage(tools []SecurityTool, detectedLanguages map[string]int, targetPath string) (map[string]ToolOutput, []error) {
	results := make(map[string]ToolOutput)
	var errors []error
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, tool := range tools {
		applicable := false
		for lang := range detectedLanguages {
			if tool.IsApplicable(lang) {
				applicable = true
				break
			}
		}

		if applicable {
			wg.Add(1)
			go func(t SecurityTool) {
				defer wg.Done()
				output, err := t.Run(targetPath)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errors = append(errors, err)
				}
				results[t.GetToolInfo().Name] = output
			}(tool)
		}
	}

	wg.Wait()
	return results, errors
}
