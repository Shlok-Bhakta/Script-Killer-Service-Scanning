package tools

import (
	"fmt"
	"sort"
	"strings"
)

type PackageKey struct {
	Package string
	Version string
}

type CollapsedFinding struct {
	Package        string
	CurrentVersion string
	FixVersion     string
	Severity       Severity
	VulnCount      int
	VulnIDs        []string
	Location       Location
	Sources        map[string]bool
}

func CollapseFindings(findings []Finding) []CollapsedFinding {
	groups := make(map[PackageKey]*CollapsedFinding)

	for _, f := range findings {
		pkg := f.Metadata["package"]
		ver := f.Metadata["version"]
		if pkg == "" {
			continue
		}

		normalizedPkg := normalizePackageName(pkg)
		key := PackageKey{Package: normalizedPkg, Version: ver}
		group, exists := groups[key]
		if !exists {
			group = &CollapsedFinding{
				Package:        pkg,
				CurrentVersion: ver,
				Severity:       f.Severity,
				Location:       f.Location,
				VulnIDs:        []string{},
				Sources:        make(map[string]bool),
			}
			groups[key] = group
		}

		if src := f.Metadata["source"]; src != "" {
			group.Sources[src] = true
		}

		vulnID := f.Metadata["vuln_id"]
		if vulnID == "" {
			vulnID = f.ID
		}
		group.VulnIDs = append(group.VulnIDs, vulnID)
		group.VulnCount++

		if severityRank(f.Severity) > severityRank(group.Severity) {
			group.Severity = f.Severity
		}

		fixVer := extractFixVersion(f)
		if fixVer != "" && (group.FixVersion == "" || compareVersions(fixVer, group.FixVersion) > 0) {
			group.FixVersion = fixVer
		}
	}

	result := make([]CollapsedFinding, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Severity != result[j].Severity {
			return severityRank(result[i].Severity) > severityRank(result[j].Severity)
		}
		return result[i].VulnCount > result[j].VulnCount
	})

	return result
}

func CollapseFindingsToFindings(findings []Finding) []Finding {
	collapsed := CollapseFindings(findings)
	result := make([]Finding, 0, len(collapsed))

	for _, c := range collapsed {
		msg := fmt.Sprintf("%s@%s has %d vulnerabilities", c.Package, c.CurrentVersion, c.VulnCount)

		suggestion := ""
		if c.FixVersion != "" {
			suggestion = fmt.Sprintf("Upgrade to %s to fix %d vulnerabilities", c.FixVersion, c.VulnCount)
		}

		vulnIDsDisplay := c.VulnIDs
		if len(vulnIDsDisplay) > 5 {
			vulnIDsDisplay = append(vulnIDsDisplay[:5], fmt.Sprintf("... and %d more", len(c.VulnIDs)-5))
		}

		sources := make([]string, 0, len(c.Sources))
		for src := range c.Sources {
			sources = append(sources, src)
		}
		sort.Strings(sources)

		f := Finding{
			ID:         fmt.Sprintf("%s-%s", c.Package, c.CurrentVersion),
			Severity:   c.Severity,
			Message:    msg,
			Location:   c.Location,
			Suggestion: suggestion,
			Metadata: map[string]string{
				"package":    c.Package,
				"version":    c.CurrentVersion,
				"fixed":      c.FixVersion,
				"vuln_count": fmt.Sprintf("%d", c.VulnCount),
				"vuln_ids":   strings.Join(vulnIDsDisplay, ", "),
				"source":     strings.Join(sources, ", "),
			},
			Suppressed: false,
		}
		result = append(result, f)
	}

	return result
}

func severityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityWarning:
		return 3
	case SeverityInfo:
		return 2
	case SeverityOther:
		return 1
	default:
		return 0
	}
}

func extractFixVersion(f Finding) string {
	if fix := f.Metadata["fixed"]; fix != "" {
		return fix
	}

	if strings.HasPrefix(f.Suggestion, "upgrade to ") {
		return strings.TrimPrefix(f.Suggestion, "upgrade to ")
	}
	if strings.HasPrefix(f.Suggestion, "Upgrade to ") {
		parts := strings.Fields(f.Suggestion)
		if len(parts) >= 3 {
			return parts[2]
		}
	}

	return ""
}

func compareVersions(a, b string) int {
	return strings.Compare(a, b)
}

func normalizePackageName(pkg string) string {
	if idx := strings.LastIndex(pkg, ":"); idx != -1 {
		return pkg[idx+1:]
	}
	if idx := strings.LastIndex(pkg, "/"); idx != -1 {
		return pkg[idx+1:]
	}
	return pkg
}
