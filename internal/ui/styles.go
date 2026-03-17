package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds computed colors that may vary by terminal background.
type Theme struct {
	Primary   color.Color
	Dim       color.Color
	Subtle    color.Color
	Text      color.Color
	DimText   color.Color
	Highlight color.Color

	Success color.Color
	Error   color.Color
	Warning color.Color
	Info    color.Color

	DiffAdd     color.Color
	DiffRemove  color.Color
	DiffChanged color.Color
	DiffSame    color.Color
}

// NewTheme creates a theme adapted to the terminal background.
func NewTheme(hasDarkBG bool) Theme {
	ld := lipgloss.LightDark(hasDarkBG)
	return Theme{
		Primary:   ld(lipgloss.Color("#874BFD"), lipgloss.Color("#7D56F4")),
		Dim:       ld(lipgloss.Color("#DDDADA"), lipgloss.Color("#3C3C3C")),
		Subtle:    ld(lipgloss.Color("#9B9B9B"), lipgloss.Color("#5C5C5C")),
		Text:      ld(lipgloss.Color("#1A1A1A"), lipgloss.Color("#FAFAFA")),
		DimText:   ld(lipgloss.Color("#A49FA5"), lipgloss.Color("#777777")),
		Highlight: ld(lipgloss.Color("#F0E6FF"), lipgloss.Color("#2A1F3D")),

		Success: lipgloss.ANSIColor(2),
		Error:   lipgloss.ANSIColor(1),
		Warning: lipgloss.ANSIColor(3),
		Info:    lipgloss.ANSIColor(4),

		DiffAdd:     lipgloss.ANSIColor(2),
		DiffRemove:  lipgloss.ANSIColor(1),
		DiffChanged: lipgloss.ANSIColor(3),
		DiffSame:    lipgloss.ANSIColor(8),
	}
}

// Styles holds all pre-computed lipgloss styles for the UI.
type Styles struct {
	FocusedBorder lipgloss.Style
	BlurredBorder lipgloss.Style
	FocusedTitle  lipgloss.Style
	BlurredTitle  lipgloss.Style

	ActiveItem   lipgloss.Style
	InactiveItem lipgloss.Style
	Cursor       lipgloss.Style

	PreviewKey    lipgloss.Style
	PreviewValue  lipgloss.Style
	PreviewMasked lipgloss.Style

	StatusSuccess lipgloss.Style
	StatusError   lipgloss.Style
	StatusWarning lipgloss.Style
	StatusInfo    lipgloss.Style
	HelpBar       lipgloss.Style

	Overlay      lipgloss.Style
	OverlayTitle lipgloss.Style

	DiffAdd     lipgloss.Style
	DiffRemove  lipgloss.Style
	DiffChanged lipgloss.Style
	DiffSame    lipgloss.Style
}

// NewStyles creates all styles from a theme.
func NewStyles(t Theme) Styles {
	return Styles{
		FocusedBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary),
		BlurredBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Dim),
		FocusedTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Primary).
			Padding(0, 1),
		BlurredTitle: lipgloss.NewStyle().
			Foreground(t.Dim).
			Padding(0, 1),

		ActiveItem: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Text),
		InactiveItem: lipgloss.NewStyle().
			Foreground(t.DimText),
		Cursor: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Highlight).
			Background(t.Primary),

		PreviewKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.ANSIColor(3)),
		PreviewValue: lipgloss.NewStyle().
			Foreground(t.DimText),
		PreviewMasked: lipgloss.NewStyle().
			Foreground(t.Subtle),

		StatusSuccess: lipgloss.NewStyle().Foreground(t.Success),
		StatusError:   lipgloss.NewStyle().Foreground(t.Error),
		StatusWarning: lipgloss.NewStyle().Foreground(t.Warning),
		StatusInfo:    lipgloss.NewStyle().Foreground(t.Info),
		HelpBar:       lipgloss.NewStyle().Foreground(t.Subtle),

		Overlay: lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(t.Primary).
			Padding(1, 2),
		OverlayTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Primary).
			MarginBottom(1),

		DiffAdd:     lipgloss.NewStyle().Foreground(t.DiffAdd),
		DiffRemove:  lipgloss.NewStyle().Foreground(t.DiffRemove),
		DiffChanged: lipgloss.NewStyle().Foreground(t.DiffChanged),
		DiffSame:    lipgloss.NewStyle().Foreground(t.DiffSame),
	}
}

// PanelStyle returns the appropriate border style for a panel based on focus state.
func PanelStyle(s Styles, focused bool, width, height int) lipgloss.Style {
	base := lipgloss.NewStyle().
		Width(width).
		Height(height)

	if focused {
		return base.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(s.FocusedBorder.GetBorderBottomForeground())
	}
	return base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.BlurredBorder.GetBorderBottomForeground())
}
