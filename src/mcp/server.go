package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"scriptkiller/src/tools"
	"scriptkiller/src/tui/scanner"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"script-killer",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	checkVulnsTool := mcp.NewTool("check_vulnerabilities",
		mcp.WithDescription("Scan a directory for security vulnerabilities using multiple tools (gosec, grype, osv-scanner)"),
		mcp.WithString("directory",
			mcp.Required(),
			mcp.Description("Path to the directory to scan for vulnerabilities"),
		),
	)

	s.AddTool(checkVulnsTool, handleCheckVulnerabilities)

	return s
}

func handleCheckVulnerabilities(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory, err := request.RequireString("directory")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid directory parameter: %v", err)), nil
	}

	s := scanner.New(directory)
	result, err := s.Scan(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Scan failed: %v", err)), nil
	}

	findings := s.GetAllFindings()

	response := map[string]interface{}{
		"directory": directory,
		"duration":  result.Duration.String(),
		"findings":  findings,
		"summary": map[string]int{
			"total":    len(findings),
			"critical": countBySeverity(findings, "critical"),
			"warning":  countBySeverity(findings, "warning"),
			"info":     countBySeverity(findings, "info"),
		},
	}

	jsonResult, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

func countBySeverity(findings []tools.Finding, severity string) int {
	count := 0
	for _, f := range findings {
		if string(f.Severity) == severity {
			count++
		}
	}
	return count
}
