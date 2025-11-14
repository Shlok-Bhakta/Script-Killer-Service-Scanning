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

type Finding struct {
	Severity   Severity
	Message    string
	Location   string
	Suggestion string
}

type ToolOutput struct {
	RawOutput string
	Critical  []Finding
	Warnings  []Finding
	Info      []Finding
	Other     []Finding
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

// TODO: Add a way to run all tools of x lang and format output into pretty table.
