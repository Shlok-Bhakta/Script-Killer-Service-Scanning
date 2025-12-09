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
			tools.NewBanditPyTool(),
			tools.NewGitleaksTool(),
			tools.NewESLintSecurityTool(),
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

	var findings []tools.Finding
	for toolName, output := range s.lastResult.ToolOutputs {
		for _, f := range output.Critical {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			findings = append(findings, f)
		}
		for _, f := range output.Warnings {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			findings = append(findings, f)
		}
		for _, f := range output.Info {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			findings = append(findings, f)
		}
		for _, f := range output.Other {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["source"] = toolName
			findings = append(findings, f)
		}
	}
	return tools.CollapseFindingsToFindings(findings)
}

func (s *Scanner) GetTargetPath() string {
	return s.targetPath
}
