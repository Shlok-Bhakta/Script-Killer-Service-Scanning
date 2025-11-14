package tools

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
	ServerityOther   Severity = "other"
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

type Finding struct {
	Severity   Severity
	Message    string
	Location   string
	Suggestion string
	Metadata   map[string]string
}

type ToolOutput struct {
	ToolName  string
	Duration  int64
	RawOutput string
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

// This is the interface that all security tools should implement
// getToolInfo() is going to be used on the UI side to display tool info and maybe even some settings??
// Run() is the main function that will be called by the scriptkiller service
// IsApplicable() is used to determine if the tool is applicable to the language of the file so go tools only run in go projects
type SecurityTool interface {
	GetToolInfo() ToolInfo
	Run(targetPath string) (ToolOutput, error)
	IsApplicable(language string) bool
}

func RunAllToolsForLanguage(tools []SecurityTool, language string, targetPath string) (map[string]ToolOutput, error) {
	results := make(map[string]ToolOutput)

	for _, tool := range tools {
		if tool.IsApplicable(language) {
			output, err := tool.Run(targetPath)
			if err != nil {
				return results, err
			}
			results[tool.GetToolInfo().Name] = output
		}
	}

	return results, nil
}
