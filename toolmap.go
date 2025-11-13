package main

type BaseTool {
	name string
	desc string
	url string
}

type SecurityTool interface {
    BaseTool
    Run(targetPath string) (ToolOutput, error)
    IsApplicable(language string) bool
}



struct toolOutput {
	output string

}


toolMap := map[string][]tool{
	"go": [tool{
		toolName: "gosec",
		toolPkg: "gosec",
		toolDesc: "Inspects source code for security problems by scanning the Go AST and SSA code representation.",
		toolURL: "https://github.com/securego/gosec",

		}]


func (t *tool) runTool(): string {

}