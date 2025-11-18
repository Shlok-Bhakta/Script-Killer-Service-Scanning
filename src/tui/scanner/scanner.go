package scanner

import (
	"context"
	"scriptkiller/src/detector"
	"scriptkiller/src/tools"
	"sync"
	"time"
)

type ScanResult struct {
	Languages   map[string]int
	ToolOutputs map[string]tools.ToolOutput
	Errors      []error
	Duration    time.Duration
	Timestamp   time.Time
}

type Scanner struct {
	targetPath      string
	codeTools       []tools.SecurityTool
	dependencyTools []tools.SecurityTool
	mu              sync.RWMutex
	lastResult      *ScanResult
}

func New(targetPath string) *Scanner {
	return &Scanner{
		targetPath: targetPath,
		codeTools: []tools.SecurityTool{
			tools.NewGosecTool(),
		},
		dependencyTools: []tools.SecurityTool{
			tools.NewOSVScannerTool(),
			tools.NewGrypeTool(),
		},
	}
}

func (s *Scanner) Scan(ctx context.Context) (*ScanResult, error) {
	return s.scan(ctx, true)
}

func (s *Scanner) ScanCode(ctx context.Context) (*ScanResult, error) {
	return s.scan(ctx, false)
}

func (s *Scanner) scan(ctx context.Context, includeDependencies bool) (*ScanResult, error) {
	startTime := time.Now()

	languages, err := detector.DetectProjectLanguages(s.targetPath)
	if err != nil {
		return nil, err
	}

	var toolsToRun []tools.SecurityTool
	if includeDependencies {
		toolsToRun = append(toolsToRun, s.codeTools...)
		toolsToRun = append(toolsToRun, s.dependencyTools...)
	} else {
		toolsToRun = s.codeTools
	}

	toolOutputs, errs := tools.RunAllToolsForLanguage(toolsToRun, languages, s.targetPath)

	if !includeDependencies && s.lastResult != nil {
		s.mu.RLock()
		for toolName, output := range s.lastResult.ToolOutputs {
			if _, exists := toolOutputs[toolName]; !exists {
				toolOutputs[toolName] = output
			}
		}
		s.mu.RUnlock()
	}

	result := &ScanResult{
		Languages:   languages,
		ToolOutputs: toolOutputs,
		Errors:      errs,
		Duration:    time.Since(startTime),
		Timestamp:   time.Now(),
	}

	s.mu.Lock()
	s.lastResult = result
	s.mu.Unlock()

	return result, nil
}

func (s *Scanner) LastResult() *ScanResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastResult
}

func (s *Scanner) GetAllFindings() []tools.Finding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastResult == nil {
		return nil
	}

	seenCVEs := make(map[string]bool)

	var critical, warnings, info, other []tools.Finding

	for _, output := range s.lastResult.ToolOutputs {
		critical = append(critical, filterDuplicateCVEs(output.Critical, seenCVEs)...)
		warnings = append(warnings, filterDuplicateCVEs(output.Warnings, seenCVEs)...)
		info = append(info, filterDuplicateCVEs(output.Info, seenCVEs)...)
		other = append(other, filterDuplicateCVEs(output.Other, seenCVEs)...)
	}

	var findings []tools.Finding
	findings = append(findings, critical...)
	findings = append(findings, warnings...)
	findings = append(findings, info...)
	findings = append(findings, other...)

	return findings
}

func filterDuplicateCVEs(findings []tools.Finding, seenCVEs map[string]bool) []tools.Finding {
	var filtered []tools.Finding
	for _, f := range findings {
		if f.CVE != nil {
			if seenCVEs[*f.CVE] {
				continue
			}
			seenCVEs[*f.CVE] = true
		}
		filtered = append(filtered, f)
	}
	return filtered
}

func (s *Scanner) GetFindingsByType() (codeFindings []tools.Finding, depsFindings []tools.Finding) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastResult == nil {
		return nil, nil
	}

	codeSeenCVEs := make(map[string]bool)
	depsSeenCVEs := make(map[string]bool)

	var codeCritical, codeWarnings, codeInfo, codeOther []tools.Finding
	var depsCritical, depsWarnings, depsInfo, depsOther []tools.Finding

	codeToolNames := make(map[string]bool)
	for _, tool := range s.codeTools {
		codeToolNames[tool.GetToolInfo().Name] = true
	}

	for toolName, output := range s.lastResult.ToolOutputs {
		if codeToolNames[toolName] {
			codeCritical = append(codeCritical, filterDuplicateCVEs(output.Critical, codeSeenCVEs)...)
			codeWarnings = append(codeWarnings, filterDuplicateCVEs(output.Warnings, codeSeenCVEs)...)
			codeInfo = append(codeInfo, filterDuplicateCVEs(output.Info, codeSeenCVEs)...)
			codeOther = append(codeOther, filterDuplicateCVEs(output.Other, codeSeenCVEs)...)
		} else {
			depsCritical = append(depsCritical, filterDuplicateCVEs(output.Critical, depsSeenCVEs)...)
			depsWarnings = append(depsWarnings, filterDuplicateCVEs(output.Warnings, depsSeenCVEs)...)
			depsInfo = append(depsInfo, filterDuplicateCVEs(output.Info, depsSeenCVEs)...)
			depsOther = append(depsOther, filterDuplicateCVEs(output.Other, depsSeenCVEs)...)
		}
	}

	codeFindings = append(codeFindings, codeCritical...)
	codeFindings = append(codeFindings, codeWarnings...)
	codeFindings = append(codeFindings, codeInfo...)
	codeFindings = append(codeFindings, codeOther...)

	depsFindings = append(depsFindings, depsCritical...)
	depsFindings = append(depsFindings, depsWarnings...)
	depsFindings = append(depsFindings, depsInfo...)
	depsFindings = append(depsFindings, depsOther...)

	return codeFindings, depsFindings
}

func (s *Scanner) GetTargetPath() string {
	return s.targetPath
}
