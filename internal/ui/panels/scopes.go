package panels

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/warui1/dotenvx-tui/internal/theme"
)

// ScopePanel displays the list of discovered scopes (directories).
type ScopePanel struct {
	Items    []string
	Cursor   int
	Selected int
	Width    int
	Height   int
}

// CursorUp moves the cursor up with circular wrapping.
func (p *ScopePanel) CursorUp() {
	if len(p.Items) == 0 {
		return
	}
	p.Cursor--
	if p.Cursor < 0 {
		p.Cursor = len(p.Items) - 1
	}
}

// CursorDown moves the cursor down with circular wrapping.
func (p *ScopePanel) CursorDown() {
	if len(p.Items) == 0 {
		return
	}
	p.Cursor++
	if p.Cursor >= len(p.Items) {
		p.Cursor = 0
	}
}

// Select marks the current cursor position as selected.
func (p *ScopePanel) Select() {
	if len(p.Items) > 0 {
		p.Selected = p.Cursor
	}
}

// SelectedItem returns the currently selected scope name.
func (p *ScopePanel) SelectedItem() string {
	if len(p.Items) == 0 {
		return ""
	}
	return p.Items[p.Selected]
}

// Render draws the scope panel content (without border).
func (p *ScopePanel) Render(styles theme.Styles, focused bool) string {
	if len(p.Items) == 0 {
		return styles.InactiveItem.Render("  (no scopes)")
	}

	var b strings.Builder
	for i, item := range p.Items {
		label := scopeLabel(item)
		if i == p.Cursor && focused {
			fmt.Fprintf(&b, "  %s", styles.Cursor.Render(" "+label+" "))
		} else if i == p.Selected {
			fmt.Fprintf(&b, "  %s", styles.ActiveItem.Render(label))
		} else {
			fmt.Fprintf(&b, "  %s", styles.InactiveItem.Render(label))
		}
		if i < len(p.Items)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// Title returns the panel title.
func (p *ScopePanel) Title(styles theme.Styles, focused bool) string {
	title := "Scopes"
	if focused {
		return styles.FocusedTitle.Render(title)
	}
	return styles.BlurredTitle.Render(title)
}

// PanelStyle returns the border style for this panel.
func (p *ScopePanel) PanelStyle(styles theme.Styles, focused bool) lipgloss.Style {
	return theme.PanelStyle(styles, focused, p.Width, p.Height)
}

func scopeLabel(scope string) string {
	if scope == "." {
		return "(root)"
	}
	return scope
}
