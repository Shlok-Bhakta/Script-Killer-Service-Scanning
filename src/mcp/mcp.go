package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"scriptkiller/src/detector"
	"scriptkiller/src/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	server     *server.MCPServer
	targetPath string
	tools      []tools.SecurityTool
}

func New(targetPath string) *MCPServer {
	s := server.NewMCPServer(
		"ScriptKiller Security Scanner",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(true, false),
	)

	m := &MCPServer{
		server:     s,
		targetPath: targetPath,
		tools: []tools.SecurityTool{
			tools.NewGosecTool(),
			tools.NewOSVScannerTool(),
			tools.NewGrypeTool(),
			tools.NewBanditPyTool(),
			tools.NewGitleaksTool(),
			tools.NewCppcheckTool(),
			tools.NewESLintSecurityTool(),
		},
	}

	m.registerTools()
	m.registerResources()

	return m
}

func (m *MCPServer) registerTools() {
	scanTool := mcp.NewTool("scan",
		mcp.WithDescription("Run security scan on the target directory"),
		mcp.WithString("path",
			mcp.Description("Path to scan (defaults to target path)"),
		),
	)
	m.server.AddTool(scanTool, m.handleScan)

	listFindingsTool := mcp.NewTool("list_findings",
		mcp.WithDescription("List all security findings from the last scan"),
		mcp.WithString("severity",
			mcp.Description("Filter by severity: critical, warning, info, other"),
			mcp.Enum("critical", "warning", "info", "other", "all"),
		),
	)
	m.server.AddTool(listFindingsTool, m.handleListFindings)

	detectLanguagesTool := mcp.NewTool("detect_languages",
		mcp.WithDescription("Detect programming languages in the target directory"),
	)
	m.server.AddTool(detectLanguagesTool, m.handleDetectLanguages)
}

func (m *MCPServer) registerResources() {
	findingsResource := mcp.NewResource(
		"scriptkiller://findings",
		"Security Findings",
		mcp.WithResourceDescription("Current security scan findings"),
		mcp.WithMIMEType("application/json"),
	)
	m.server.AddResource(findingsResource, m.handleFindingsResource)

	languagesResource := mcp.NewResource(
		"scriptkiller://languages",
		"Detected Languages",
		mcp.WithResourceDescription("Programming languages detected in the project"),
		mcp.WithMIMEType("application/json"),
	)
	m.server.AddResource(languagesResource, m.handleLanguagesResource)
}

func (m *MCPServer) handleScan(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := m.targetPath
	if p, ok := request.GetArguments()["path"].(string); ok && p != "" {
		path = p
	}

	languages, err := detector.DetectProjectLanguages(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to detect languages: %v", err)), nil
	}

	toolOutputs, errs := tools.RunAllToolsForLanguage(m.tools, languages, path)

	var allFindings []tools.Finding
	for toolName, output := range toolOutputs {
		for _, f := range output.Critical {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			allFindings = append(allFindings, f)
		}
		for _, f := range output.Warnings {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			allFindings = append(allFindings, f)
		}
		for _, f := range output.Info {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			allFindings = append(allFindings, f)
		}
		for _, f := range output.Other {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			allFindings = append(allFindings, f)
		}
	}

	collapsedFindings := tools.CollapseFindingsToFindings(allFindings)

	result := map[string]any{
		"path":           path,
		"languages":      languages,
		"total_findings": len(collapsedFindings),
		"findings":       collapsedFindings,
		"errors":         len(errs),
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (m *MCPServer) handleListFindings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	severityFilter := "all"
	if s, ok := request.GetArguments()["severity"].(string); ok && s != "" {
		severityFilter = s
	}

	languages, err := detector.DetectProjectLanguages(m.targetPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to detect languages: %v", err)), nil
	}

	toolOutputs, _ := tools.RunAllToolsForLanguage(m.tools, languages, m.targetPath)

	var findings []tools.Finding
	for _, output := range toolOutputs {
		switch severityFilter {
		case "critical":
			findings = append(findings, output.Critical...)
		case "warning":
			findings = append(findings, output.Warnings...)
		case "info":
			findings = append(findings, output.Info...)
		case "other":
			findings = append(findings, output.Other...)
		default:
			findings = append(findings, output.Critical...)
			findings = append(findings, output.Warnings...)
			findings = append(findings, output.Info...)
			findings = append(findings, output.Other...)
		}
	}

	jsonBytes, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal findings: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (m *MCPServer) handleDetectLanguages(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	languages, err := detector.DetectProjectLanguages(m.targetPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to detect languages: %v", err)), nil
	}

	jsonBytes, err := json.MarshalIndent(languages, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal languages: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func (m *MCPServer) handleFindingsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	languages, err := detector.DetectProjectLanguages(m.targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect languages: %w", err)
	}

	toolOutputs, _ := tools.RunAllToolsForLanguage(m.tools, languages, m.targetPath)

	var findings []tools.Finding
	for _, output := range toolOutputs {
		findings = append(findings, output.Critical...)
		findings = append(findings, output.Warnings...)
		findings = append(findings, output.Info...)
		findings = append(findings, output.Other...)
	}

	jsonBytes, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal findings: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "scriptkiller://findings",
			MIMEType: "application/json",
			Text:     string(jsonBytes),
		},
	}, nil
}

func (m *MCPServer) handleLanguagesResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	languages, err := detector.DetectProjectLanguages(m.targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect languages: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(languages, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal languages: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "scriptkiller://languages",
			MIMEType: "application/json",
			Text:     string(jsonBytes),
		},
	}, nil
}

func (m *MCPServer) ServeStdio() error {
	return server.ServeStdio(m.server)
}
