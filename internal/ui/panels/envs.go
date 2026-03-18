package panels

import (
	"fmt"
	"strings"

	"github.com/warui1/dotenvx-tui/internal/theme"
)

// EnvPanel displays the list of environments for the selected scope.
type EnvPanel struct {
	Items    []string
	Cursor   int
	Selected int
	Width    int
	Height   int
}

// CursorUp moves the cursor up with circular wrapping.
func (p *EnvPanel) CursorUp() {
	if len(p.Items) == 0 {
		return
	}
	p.Cursor--
	if p.Cursor < 0 {
		p.Cursor = len(p.Items) - 1
	}
}

// CursorDown moves the cursor down with circular wrapping.
func (p *EnvPanel) CursorDown() {
	if len(p.Items) == 0 {
		return
	}
	p.Cursor++
	if p.Cursor >= len(p.Items) {
		p.Cursor = 0
	}
}

// Select marks the current cursor position as selected.
func (p *EnvPanel) Select() {
	if len(p.Items) > 0 {
		p.Selected = p.Cursor
	}
}

// SelectedItem returns the currently selected environment name.
func (p *EnvPanel) SelectedItem() string {
	if len(p.Items) == 0 {
		return ""
	}
	return p.Items[p.Selected]
}

// Reset clears items and resets cursor.
func (p *EnvPanel) Reset(items []string) {
	p.Items = items
	p.Cursor = 0
	p.Selected = 0
}

// Render draws the env panel content (without border).
func (p *EnvPanel) Render(styles theme.Styles, focused bool) string {
	if len(p.Items) == 0 {
		return styles.InactiveItem.Render("  (no envs)")
	}

	var b strings.Builder
	for i, item := range p.Items {
		if i == p.Cursor && focused {
			fmt.Fprintf(&b, "  %s", styles.Cursor.Render(" "+item+" "))
		} else if i == p.Selected {
			fmt.Fprintf(&b, "  %s", styles.ActiveItem.Render(item))
		} else {
			fmt.Fprintf(&b, "  %s", styles.InactiveItem.Render(item))
		}
		if i < len(p.Items)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// Title returns the panel title.
func (p *EnvPanel) Title(styles theme.Styles, focused bool) string {
	title := "Environments"
	if focused {
		return styles.FocusedTitle.Render(title)
	}
	return styles.BlurredTitle.Render(title)
}

