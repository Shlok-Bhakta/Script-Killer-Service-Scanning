package scanner

import (
	"context"
	"scriptkiller/src/detector"
	"scriptkiller/src/tools"
	"sync"
	"time"
)

type ScanType int

const (
	Directory = iota
	Endpoint
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
	dirTools   []tools.SecurityTool
	mu         sync.RWMutex
	lastResult *ScanResult
}

func New(targetPath string) *Scanner {
	return &Scanner{
		targetPath: targetPath,
		dirTools: []tools.SecurityTool{
			tools.NewGosecTool(),
			tools.NewOSVScannerTool(),
			tools.NewGrypeTool(),
		},
	}
}

func (s *Scanner) Scan(ctx context.Context, scanType ScanType) (*ScanResult, error) {
	startTime := time.Now()
	var languages map[string]int
	if scanType == Directory {
		language, err := detector.DetectProjectLanguages(s.targetPath)
		languages = language
		if err != nil {
			return nil, err
		}
	}

	selectedTools := s.dirTools
	switch scanType {
	case Directory:
		selectedTools = s.dirTools
	}

	toolOutputs, errs := tools.RunAllToolsForLanguage(selectedTools, languages, s.targetPath)

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
	for _, output := range s.lastResult.ToolOutputs {
		findings = append(findings, output.Critical...)
		findings = append(findings, output.Warnings...)
		findings = append(findings, output.Info...)
		findings = append(findings, output.Other...)
	}
	return findings
}

func (s *Scanner) GetTargetPath() string {
	return s.targetPath
}

func (s *Scanner) SetTargetPath(path string) {
	s.targetPath = path
}
