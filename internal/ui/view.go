package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/warui1/dotenvx-tui/internal/theme"
)

// View renders the entire UI.
func (m Model) View() tea.View {
	var content string

	if m.fatalErr != "" {
		content = m.renderFatalError()
	} else if m.layout.TooSmall {
		content = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("Terminal too small (%dx%d)\nMinimum: %dx%d", m.width, m.height, minWidth, minHeight))
	} else if len(m.envFiles) == 0 && !m.loading {
		content = m.renderEmptyState()
	} else {
		content = m.renderMain()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m Model) renderMain() string {
	// Render panels
	var topPanels []string

	if !m.layout.HideScopes {
		scopeContent := m.scopePanel.Render(m.styles, m.focusedPanel == PanelScopes)
		scopeTitle := m.scopePanel.Title(m.styles, m.focusedPanel == PanelScopes)
		scopeBox := m.renderPanel(scopeContent, scopeTitle, m.focusedPanel == PanelScopes, m.layout.ScopeWidth, m.layout.PanelHeight)
		topPanels = append(topPanels, scopeBox)
	}

	envContent := m.envPanel.Render(m.styles, m.focusedPanel == PanelEnvs)
	envTitle := m.envPanel.Title(m.styles, m.focusedPanel == PanelEnvs)
	envBox := m.renderPanel(envContent, envTitle, m.focusedPanel == PanelEnvs, m.layout.EnvWidth, m.layout.PanelHeight)
	topPanels = append(topPanels, envBox)

	keysContent := m.keyPanel.Render(m.styles, m.focusedPanel == PanelKeys)
	keysTitle := m.keyPanel.Title(m.styles, m.focusedPanel == PanelKeys)
	keysBox := m.renderPanel(keysContent, keysTitle, m.focusedPanel == PanelKeys, m.layout.KeysWidth, m.layout.PanelHeight)
	topPanels = append(topPanels, keysBox)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topPanels...)

	// Preview pane
	preview := m.renderPreview()
	previewBox := m.renderPanel(preview, m.styles.BlurredTitle.Render("Preview"), false, m.width, m.layout.PreviewHeight)

	// Status + help
	statusBar := m.renderStatusBar()
	helpBar := m.renderHelpBar()

	main := lipgloss.JoinVertical(lipgloss.Left, topRow, previewBox, statusBar, helpBar)

	// Overlay
	if m.activeOverlay != OverlayNone {
		main = m.renderOverlayOnTop(main)
	}

	return main
}

func (m Model) renderPanel(content, title string, focused bool, width, height int) string {
	// Hard clip content to panel height so it never overflows the border
	lines := strings.Split(content, "\n")
	maxLines := height - 1 // title takes 1 line
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
		content = strings.Join(lines, "\n")
	}
	style := theme.PanelStyle(m.theme, focused, width-2, height) // -2 for border
	return style.Render(title + "\n" + content)
}

func (m Model) renderPreview() string {
	if m.previewKey == "" {
		return m.styles.InactiveItem.Render("  Select a key to preview its value")
	}

	keyStr := m.styles.PreviewKey.Render(m.previewKey)

	var valStr string
	if m.previewValue == nil {
		valStr = m.styles.PreviewMasked.Render("(loading...)")
	} else if m.previewShown {
		valStr = m.styles.PreviewValue.Render(m.previewValue.String())
	} else {
		valStr = m.styles.PreviewMasked.Render(m.previewValue.Masked())
	}

	hint := ""
	if !m.previewShown && m.previewValue != nil {
		hint = m.styles.HelpBar.Render("  (press enter to reveal)")
	}

	return fmt.Sprintf("  %s = %s%s", keyStr, valStr, hint)
}

func (m Model) renderStatusBar() string {
	if m.statusMsg == "" {
		if m.loading {
			return m.styles.StatusInfo.Render("  " + m.loadingMsg)
		}
		return ""
	}

	var style lipgloss.Style
	var prefix string
	switch m.statusLevel {
	case StatusSuccess:
		style = m.styles.StatusSuccess
		prefix = "  ✓ "
	case StatusError:
		style = m.styles.StatusError
		prefix = "  ✗ "
	case StatusWarning:
		style = m.styles.StatusWarning
		prefix = "  ⚠ "
	case StatusInfo:
		style = m.styles.StatusInfo
		prefix = "  ℹ "
	}

	return style.Render(prefix + m.statusMsg)
}

func (m Model) renderHelpBar() string {
	var parts []string
	if m.focusedPanel == PanelKeys {
		parts = []string{"enter:reveal", "n:new", "s:set", "x:delete", "d:diff", "i:import", "e:export", "c:copy", "space:select", "?:help", "q:quit"}
	} else {
		parts = []string{"tab:panels", "enter:select", "?:help", "q:quit"}
	}
	return m.styles.HelpBar.Render("  " + strings.Join(parts, "  "))
}

func (m Model) renderFatalError() string {
	box := m.styles.Overlay.
		Width(min(60, m.width-4)).
		Render(
			m.styles.OverlayTitle.Render("Error") + "\n\n" +
				m.fatalErr + "\n\n" +
				m.styles.HelpBar.Render("Press q to quit"),
		)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderEmptyState() string {
	box := m.styles.Overlay.
		Width(min(60, m.width-4)).
		Render(
			m.styles.OverlayTitle.Render("No encrypted .env files found") + "\n\n" +
				"dotenvx-tui looks for .env.* files containing\n" +
				"a DOTENV_PUBLIC_KEY header (created by dotenvx).\n\n" +
				"To get started:\n" +
				"  1. Create a .env.local file\n" +
				"  2. Run: dotenvx encrypt\n" +
				"  3. Run: dotenvx-tui\n\n" +
				m.styles.HelpBar.Render("Press q to quit"),
		)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderOverlayOnTop(bg string) string {
	var overlayContent string

	switch m.activeOverlay {
	case OverlayHelp:
		overlayContent = m.renderHelpOverlay()
	case OverlaySetValue:
		overlayContent = m.setOverlay.View(m.width)
	case OverlayDiff:
		overlayContent = m.diffOverlay.View(m.width)
	case OverlayImport:
		overlayContent = m.importOverlay.View(m.width)
	case OverlayExport:
		overlayContent = m.exportOverlay.View(m.width)
	case OverlayDelete:
		overlayContent = m.deleteOverlay.View(m.width)
	default:
		overlayContent = m.styles.Overlay.
			Width(min(50, m.width-4)).
			Render(m.styles.OverlayTitle.Render("Coming soon...") + "\n\n" +
				m.styles.HelpBar.Render("Press esc to close"))
	}

	canvas := lipgloss.NewCanvas(m.width, m.height)
	compositor := lipgloss.NewCompositor(
		lipgloss.NewLayer(bg).X(0).Y(0).Z(0),
		lipgloss.NewLayer(overlayContent).
			X(max(0, (m.width-lipgloss.Width(overlayContent))/2)).
			Y(max(0, (m.height-lipgloss.Height(overlayContent))/2)).
			Z(1),
	)
	canvas.Compose(compositor)
	return canvas.Render()
}

func (m Model) renderHelpOverlay() string {
	help := strings.Join([]string{
		"Navigation",
		"  tab / shift+tab    Switch panels",
		"  j/k or ↑/↓         Move cursor",
		"  enter              Select / Reveal value",
		"  h/l                Prev/Next panel",
		"",
		"Selection",
		"  space              Toggle select",
		"  a                  Select all / none",
		"",
		"Actions",
		"  n                  New variable",
		"  s                  Set value for key",
		"  x                  Delete variable(s)",
		"  d                  Diff environments",
		"  i                  Import from file",
		"  e                  Export keys",
		"  c                  Copy value to clipboard",
		"",
		"General",
		"  ?                  Toggle this help",
		"  esc                Close overlay / Go back",
		"  q / ctrl+c         Quit",
	}, "\n")

	return m.styles.Overlay.
		Width(min(50, m.width-4)).
		Render(m.styles.OverlayTitle.Render("Help") + "\n\n" + help + "\n\n" +
			m.styles.HelpBar.Render("Press ? or esc to close"))
}
