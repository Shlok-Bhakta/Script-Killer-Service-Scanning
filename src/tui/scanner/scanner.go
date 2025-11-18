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
	targetPath string
	tools      []tools.SecurityTool
	mu         sync.RWMutex
	lastResult *ScanResult
}

func New(targetPath string) *Scanner {
	return &Scanner{
		targetPath: targetPath,
		tools: []tools.SecurityTool{
			tools.NewGosecTool(),
			tools.NewOSVScannerTool(),
			tools.NewGrypeTool(),
		},
	}
}

func (s *Scanner) Scan(ctx context.Context) (*ScanResult, error) {
	startTime := time.Now()

	languages, err := detector.DetectProjectLanguages(s.targetPath)
	if err != nil {
		return nil, err
	}

	toolOutputs, errs := tools.RunAllToolsForLanguage(s.tools, languages, s.targetPath)

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

func (s *Scanner) GetTargetPath() string {
	return s.targetPath
}
