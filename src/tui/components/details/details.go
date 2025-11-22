package details

import (
	"fmt"
	"scriptkiller/src/tools"
	"scriptkiller/src/tui/styles"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	selectedFinding *tools.Finding
}

func New() Model {
	return Model{}
}

func (m *Model) SetFinding(finding *tools.Finding) {
	m.selectedFinding = finding
}

func (m Model) View(width, height int) string {
	theme := styles.CurrentTheme()

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(width-2).
		Height(height).
		Padding(1, 2)

	if m.selectedFinding == nil {
		return style.Render(theme.S().Subtle.Render("Select a finding to view details"))
	}

	return style.Render(m.renderDetail(m.selectedFinding, width-6))
}

func (m Model) renderDetail(finding *tools.Finding, width int) string {
	if finding == nil {
		return ""
	}

	theme := styles.CurrentTheme()
	var b strings.Builder

	badge := ""
	switch finding.Severity {
	case tools.SeverityCritical:
		badge = lipgloss.NewStyle().
			Foreground(theme.BgBase).
			Background(theme.Error).
			Bold(true).
			Padding(0, 1).
			Render(" CRITICAL ")
	case tools.SeverityWarning:
		badge = lipgloss.NewStyle().
			Foreground(theme.BgBase).
			Background(theme.Warning).
			Bold(true).
			Padding(0, 1).
			Render(" WARNING ")
	case tools.SeverityInfo:
		badge = lipgloss.NewStyle().
			Foreground(theme.BgBase).
			Background(theme.Info).
			Bold(true).
			Padding(0, 1).
			Render(" INFO ")
	}

	b.WriteString(badge + "\n\n")

	b.WriteString(theme.S().Subtitle.Render("Message") + "\n")
	b.WriteString(theme.S().Text.Render(finding.Message) + "\n\n")

	b.WriteString(theme.S().Subtitle.Render("Location") + "\n")
	loc := finding.Location.File
	if finding.Location.Line > 0 {
		loc = fmt.Sprintf("%s:%d:%d", finding.Location.File, finding.Location.Line, finding.Location.Column)
	}
	b.WriteString(theme.S().Muted.Render(loc) + "\n\n")

	if finding.ID != "" {
		b.WriteString(theme.S().Subtitle.Render("ID") + "\n")
		b.WriteString(theme.S().Text.Render(finding.ID) + "\n\n")
	}

	if finding.Suggestion != "" {
		b.WriteString(theme.S().Subtitle.Render("Suggestion") + "\n")
		b.WriteString(theme.S().Text.Render(finding.Suggestion) + "\n\n")
	}

	if len(finding.Metadata) > 0 {
		b.WriteString(theme.S().Subtitle.Render("Metadata") + "\n")
		for k, v := range finding.Metadata {
			b.WriteString(theme.S().Muted.Render(fmt.Sprintf("  %s: %v", k, v)) + "\n")
		}
	}

	return b.String()
}
