package ui

// Layout holds computed panel dimensions.
type Layout struct {
	ScopeWidth int
	EnvWidth   int
	KeysWidth  int
	PanelHeight int
	PreviewHeight int
	TooSmall    bool
	HideScopes  bool
}

const (
	minWidth  = 60
	minHeight = 16
	statusLines = 2 // status bar + help hint
	borderSize  = 2 // top + bottom border per panel
)

// ComputeLayout calculates panel dimensions from terminal size.
func ComputeLayout(width, height int, scopeCount int) Layout {
	if width < minWidth || height < minHeight {
		return Layout{TooSmall: true}
	}

	usableHeight := height - statusLines
	panelHeight := usableHeight - borderSize
	previewHeight := min(5, panelHeight/3)
	panelHeight = panelHeight - previewHeight - borderSize

	if panelHeight < 3 {
		panelHeight = 3
	}

	// Subtract border widths from usable width
	innerWidth := width

	hideScopes := scopeCount <= 1 && innerWidth < 100

	var scopeW, envW, keysW int
	if hideScopes {
		envW = innerWidth * 35 / 100
		keysW = innerWidth - envW
	} else {
		scopeW = max(12, innerWidth*20/100)
		envW = max(12, innerWidth*25/100)
		keysW = innerWidth - scopeW - envW
	}

	return Layout{
		ScopeWidth:    scopeW,
		EnvWidth:      envW,
		KeysWidth:     keysW,
		PanelHeight:   panelHeight,
		PreviewHeight: previewHeight,
		TooSmall:      false,
		HideScopes:    hideScopes,
	}
}
