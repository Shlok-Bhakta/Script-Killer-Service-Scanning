package tools

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"scriptkiller/src/nix"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	gocvss30 "github.com/pandatix/go-cvss/30"
	gocvss31 "github.com/pandatix/go-cvss/31"
)

type OSVScannerTool struct {
	info   ToolInfo
	logger *log.Logger
}

type osvSeverity struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}

type osvPackageInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem"`
}

type osvRangeEvent struct {
	Introduced string `json:"introduced,omitempty"`
	Fixed      string `json:"fixed,omitempty"`
}

type osvRange struct {
	Type   string          `json:"type"`
	Events []osvRangeEvent `json:"events"`
}

type osvAffected struct {
	Package osvPackageInfo `json:"package"`
	Ranges  []osvRange     `json:"ranges"`
}

type osvReference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type osvDatabaseSpecific struct {
	CWEID    []string `json:"cwe_ids,omitempty"`
	Severity string   `json:"severity,omitempty"`
}

type osvVulnerability struct {
	ID               string              `json:"id"`
	Aliases          []string            `json:"aliases"`
	Summary          string              `json:"summary"`
	Details          string              `json:"details"`
	Severity         []osvSeverity       `json:"severity,omitempty"`
	Affected         []osvAffected       `json:"affected"`
	References       []osvReference      `json:"references"`
	DatabaseSpecific osvDatabaseSpecific `json:"database_specific"`
}

type osvPackageResult struct {
	Package         osvPackageInfo     `json:"package"`
	Vulnerabilities []osvVulnerability `json:"vulnerabilities"`
}

type osvSource struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type osvResult struct {
	Source   osvSource          `json:"source"`
	Packages []osvPackageResult `json:"packages"`
}

type osvOutput struct {
	Results []osvResult `json:"results"`
}

func NewOSVScannerTool() *OSVScannerTool {
	return &OSVScannerTool{
		info: ToolInfo{
			Name: "osv-scanner",
			Desc: "Scans dependencies for known vulnerabilities using the OSV database",
			Url:  "https://github.com/google/osv-scanner",
		},
		logger: log.WithPrefix("osv-scanner"),
	}
}

func (o *OSVScannerTool) GetToolInfo() ToolInfo {
	return o.info
}

func (o *OSVScannerTool) Run(targetPath string) (ToolOutput, error) {
	startTime := time.Now()
	toolOut := ToolOutput{
		ToolName:  o.info.Name,
		StartTime: startTime,
		RawOutput: "",
		RawStderr: "",
		Critical:  []Finding{},
		Warnings:  []Finding{},
		Info:      []Finding{},
		Other:     []Finding{},
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		o.logger.Error("failed to get absolute path", "error", err)
		return toolOut, fmt.Errorf("failed to get absolute path: %w", err)
	}

	o.logger.Info("Running osv-scanner", "path", absPath)

	output, err := nix.RunNixShellWithOutput(
		[]string{"osv-scanner"},
		fmt.Sprintf("cd '%s' && osv-scanner --format json -r . 2>/dev/null || true", absPath),
	)
	if err != nil && len(output) == 0 {
		o.logger.Error("osv-scanner failed", "error", err)
		return toolOut, fmt.Errorf("osv-scanner failed: %w", err)
	}

	toolOut.RawOutput = string(output)

	if len(output) == 0 {
		o.logger.Warn("No output from osv-scanner")
		return toolOut, nil
	}

	o.logger.Debug("OSV-Scanner raw output", "output", string(output)[:min(500, len(output))])

	var osvResult osvOutput
	if err := json.Unmarshal(output, &osvResult); err != nil {
		o.logger.Warn("Failed to parse osv-scanner JSON output", "error", err, "output_preview", string(output)[:min(200, len(output))])
		return toolOut, nil
	}

	for _, result := range osvResult.Results {
		for _, pkg := range result.Packages {
			for _, vuln := range pkg.Vulnerabilities {
				severity, cvssScore := o.calculateSeverity(vuln)

				fixedVersion := o.extractFixedVersion(vuln)
				fixInfo := ""
				if fixedVersion != "" {
					fixInfo = fmt.Sprintf(" (fix: upgrade to %s)", fixedVersion)
				}

				cveInfo := ""
				if len(vuln.Aliases) > 0 {
					cveInfo = fmt.Sprintf(" [%s]", strings.Join(vuln.Aliases, ", "))
				}

				scoreInfo := ""
				if cvssScore > 0 {
					scoreInfo = fmt.Sprintf(" CVSS:%.1f", cvssScore)
				}

				referenceURL := o.buildNISTURL(vuln)
				if referenceURL == "" {
					referenceURL = o.extractReferenceURL(vuln)
				}
				suggestion := ""
				if referenceURL != "" {
					suggestion = fmt.Sprintf("See: %s", referenceURL)
				}

				finding := Finding{
					ID:       fmt.Sprintf("%s-%s-%s", vuln.ID, pkg.Package.Name, pkg.Package.Version),
					Severity: severity,
					Message:  fmt.Sprintf("[%s]%s %s%s%s", vuln.ID, scoreInfo, vuln.Summary, cveInfo, fixInfo),
					Location: Location{
						File: result.Source.Path,
					},
					Suggestion: suggestion,
					Metadata: map[string]string{
						"vuln_id":   vuln.ID,
						"package":   pkg.Package.Name,
						"version":   pkg.Package.Version,
						"ecosystem": pkg.Package.Ecosystem,
						"fixed":     fixedVersion,
					},
					Suppressed: false,
				}

				if len(vuln.Aliases) > 0 {
					finding.Metadata["aliases"] = strings.Join(vuln.Aliases, ",")
				}
				if cvssScore > 0 {
					finding.Metadata["cvss_score"] = fmt.Sprintf("%.1f", cvssScore)
				}

				switch severity {
				case SeverityCritical:
					toolOut.Critical = append(toolOut.Critical, finding)
				case SeverityWarning:
					toolOut.Warnings = append(toolOut.Warnings, finding)
				case SeverityInfo:
					toolOut.Info = append(toolOut.Info, finding)
				default:
					toolOut.Other = append(toolOut.Other, finding)
				}
			}
		}
	}

	o.logger.Info("OSV-Scanner scan complete", "critical", len(toolOut.Critical), "warnings", len(toolOut.Warnings), "info", len(toolOut.Info))

	toolOut.Duration = time.Since(startTime).Milliseconds()
	return toolOut, nil
}

func (o *OSVScannerTool) calculateSeverity(vuln osvVulnerability) (Severity, float64) {
	if len(vuln.Severity) > 0 {
		for _, sev := range vuln.Severity {
			if sev.Type == "CVSS_V3" {
				score := o.parseCVSSScore(sev.Score)
				if score >= 9.0 {
					return SeverityCritical, score
				} else if score >= 7.0 {
					return SeverityWarning, score
				} else if score >= 4.0 {
					return SeverityInfo, score
				} else if score > 0 {
					return SeverityOther, score
				}
			}
		}
	}

	summaryLower := strings.ToLower(vuln.Summary)
	detailsLower := strings.ToLower(vuln.Details)
	combined := summaryLower + " " + detailsLower

	criticalKeywords := []string{
		"remote code execution",
		"arbitrary code execution",
		"rce",
		"privilege escalation",
		"authentication bypass",
	}
	for _, keyword := range criticalKeywords {
		if strings.Contains(combined, keyword) {
			return SeverityCritical, 0
		}
	}

	warningKeywords := []string{
		"denial of service",
		"dos",
		"information disclosure",
		"sql injection",
		"xss",
		"csrf",
	}
	for _, keyword := range warningKeywords {
		if strings.Contains(combined, keyword) {
			return SeverityWarning, 0
		}
	}

	return SeverityInfo, 0
}

func (o *OSVScannerTool) parseCVSSScore(cvssVector string) float64 {
	if !strings.HasPrefix(cvssVector, "CVSS:") {
		return 0.0
	}

	if strings.HasPrefix(cvssVector, "CVSS:3.1/") {
		if cvss31, err := gocvss31.ParseVector(cvssVector); err == nil {
			return cvss31.BaseScore()
		}
	} else if strings.HasPrefix(cvssVector, "CVSS:3.0/") {
		if cvss30, err := gocvss30.ParseVector(cvssVector); err == nil {
			return cvss30.BaseScore()
		}
	}

	return 0.0
}

func (o *OSVScannerTool) extractFixedVersion(vuln osvVulnerability) string {
	for _, affected := range vuln.Affected {
		for _, r := range affected.Ranges {
			for _, event := range r.Events {
				if event.Fixed != "" {
					return event.Fixed
				}
			}
		}
	}
	return ""
}

func (o *OSVScannerTool) buildNISTURL(vuln osvVulnerability) string {
	for _, alias := range vuln.Aliases {
		if strings.HasPrefix(alias, "CVE-") {
			return fmt.Sprintf("https://nvd.nist.gov/vuln/detail/%s", alias)
		}
	}
	return ""
}

func (o *OSVScannerTool) extractReferenceURL(vuln osvVulnerability) string {
	var advisoryURL, webURL, fallbackURL string

	for _, ref := range vuln.References {
		if strings.Contains(ref.URL, "nvd.nist.gov") {
			return ref.URL
		}
		if ref.Type == "ADVISORY" && advisoryURL == "" {
			advisoryURL = ref.URL
		}
		if ref.Type == "WEB" && webURL == "" {
			webURL = ref.URL
		}
		if fallbackURL == "" {
			fallbackURL = ref.URL
		}
	}

	if advisoryURL != "" {
		return advisoryURL
	}
	if webURL != "" {
		return webURL
	}
	return fallbackURL
}

func (o *OSVScannerTool) IsApplicable(language string) bool {
	lang := strings.ToLower(language)
	supportedLanguages := []string{
		"c", "c++", "cpp",
		"dart",
		"elixir",
		"go", "golang",
		"java",
		"javascript", "js", "typescript", "ts",
		"php",
		"python", "py",
		"r",
		"ruby", "rb",
		"rust",
	}

	for _, supported := range supportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

func (o *OSVScannerTool) Validate() error {
	output, err := nix.RunNixShellWithOutput(
		[]string{"osv-scanner"},
		"osv-scanner --version",
	)
	if err != nil {
		return fmt.Errorf("osv-scanner is not available: %w", err)
	}
	o.logger.Debug("osv-scanner version", "output", string(output))
	return nil
}

func (o *OSVScannerTool) GetConfigSchema() map[string]any {
	return map[string]any{
		"severity_threshold": map[string]any{
			"type":        "string",
			"description": "Minimum severity level to report",
			"enum":        []string{"low", "medium", "high", "critical"},
			"default":     "low",
		},
	}
}
