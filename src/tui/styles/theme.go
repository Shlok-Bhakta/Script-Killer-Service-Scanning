package styles

import (
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name   string
	IsDark bool

	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor
	Tertiary  lipgloss.TerminalColor
	Accent    lipgloss.TerminalColor

	BgBase        lipgloss.TerminalColor
	BgBaseLighter lipgloss.TerminalColor
	BgSubtle      lipgloss.TerminalColor
	BgOverlay     lipgloss.TerminalColor

	FgBase      lipgloss.TerminalColor
	FgMuted     lipgloss.TerminalColor
	FgHalfMuted lipgloss.TerminalColor
	FgSubtle    lipgloss.TerminalColor
	FgSelected  lipgloss.TerminalColor

	Border      lipgloss.TerminalColor
	BorderFocus lipgloss.TerminalColor

	Success lipgloss.TerminalColor
	Error   lipgloss.TerminalColor
	Warning lipgloss.TerminalColor
	Info    lipgloss.TerminalColor

	ItemOfflineIcon lipgloss.Style
	ItemBusyIcon    lipgloss.Style
	ItemErrorIcon   lipgloss.Style
	ItemOnlineIcon  lipgloss.Style

	styles *Styles
}

type Styles struct {
	Base         lipgloss.Style
	SelectedBase lipgloss.Style

	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Text         lipgloss.Style
	TextSelected lipgloss.Style
	Muted        lipgloss.Style
	Subtle       lipgloss.Style

	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
}

func (t *Theme) S() *Styles {
	if t.styles == nil {
		t.styles = t.buildStyles()
	}
	return t.styles
}

func (t *Theme) buildStyles() *Styles {
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0"))
	return &Styles{
		Base: base,

		SelectedBase: base.Background(lipgloss.Color("#7d56f4")),

		Title: base.
			Foreground(lipgloss.Color("#f5c542")).
			Bold(true),

		Subtitle: base.
			Foreground(lipgloss.Color("#99e09c")).
			Bold(true),

		Text:         base,
		TextSelected: base.Background(lipgloss.Color("#7d56f4")).Foreground(lipgloss.Color("#ffffff")),

		Muted: base.Foreground(lipgloss.Color("#999999")),

		Subtle: base.Foreground(lipgloss.Color("#666666")),

		Success: base.Foreground(lipgloss.Color("#99e09c")),

		Error: base.Foreground(lipgloss.Color("#f87070")),

		Warning: base.Foreground(lipgloss.Color("#f5c542")),

		Info: base.Foreground(lipgloss.Color("#7aa2f7")),
	}
}

var defaultTheme *Theme

func CurrentTheme() *Theme {
	if defaultTheme == nil {
		defaultTheme = &Theme{
			Name:   "default",
			IsDark: true,

			Primary:   lipgloss.Color("#7d56f4"),
			Secondary: lipgloss.Color("#f5c542"),
			Tertiary:  lipgloss.Color("#99e09c"),
			Accent:    lipgloss.Color("#f5c542"),

			BgBase:        lipgloss.Color("#1a1a1a"),
			BgBaseLighter: lipgloss.Color("#252525"),
			BgSubtle:      lipgloss.Color("#2a2a2a"),
			BgOverlay:     lipgloss.Color("#333333"),

			FgBase:      lipgloss.Color("#e0e0e0"),
			FgMuted:     lipgloss.Color("#999999"),
			FgHalfMuted: lipgloss.Color("#b3b3b3"),
			FgSubtle:    lipgloss.Color("#666666"),
			FgSelected:  lipgloss.Color("#ffffff"),

			Border:      lipgloss.Color("#3a3a3a"),
			BorderFocus: lipgloss.Color("#7d56f4"),

			Success: lipgloss.Color("#99e09c"),
			Error:   lipgloss.Color("#f87070"),
			Warning: lipgloss.Color("#f5c542"),
			Info:    lipgloss.Color("#7aa2f7"),
		}

		defaultTheme.ItemOfflineIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")).SetString("●")
		defaultTheme.ItemBusyIcon = defaultTheme.ItemOfflineIcon.Foreground(lipgloss.Color("#f5c542"))
		defaultTheme.ItemErrorIcon = defaultTheme.ItemOfflineIcon.Foreground(lipgloss.Color("#f87070"))
		defaultTheme.ItemOnlineIcon = defaultTheme.ItemOfflineIcon.Foreground(lipgloss.Color("#99e09c"))
	}
	return defaultTheme
}
