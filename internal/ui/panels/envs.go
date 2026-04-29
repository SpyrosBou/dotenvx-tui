package panels

import (
	"fmt"
	"strings"

	"github.com/SpyrosBou/dotenvx-tui/internal/theme"
)

// EnvPanel displays the list of environments for the selected scope.
type EnvPanel struct {
	Items     []string
	Cursor    int
	Selected  int
	Width     int
	Height    int
	scrollOff int
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
	p.ensureVisible()
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
	p.ensureVisible()
}

// Select marks the current cursor position as selected.
func (p *EnvPanel) Select() {
	if len(p.Items) > 0 {
		p.Selected = p.Cursor
	}
}

// SelectedItem returns the currently selected environment name.
func (p *EnvPanel) SelectedItem() string {
	if len(p.Items) == 0 || p.Selected < 0 || p.Selected >= len(p.Items) {
		return ""
	}
	return p.Items[p.Selected]
}

// Reset clears items and resets cursor.
func (p *EnvPanel) Reset(items []string) {
	p.Items = items
	p.Cursor = 0
	p.Selected = 0
	p.scrollOff = 0
}

func (p *EnvPanel) visibleCount() int {
	v := p.Height - 1
	if v < 1 {
		v = 1
	}
	return v
}

func (p *EnvPanel) ensureVisible() {
	vis := p.visibleCount()
	if p.Cursor < p.scrollOff {
		p.scrollOff = p.Cursor
	} else if p.Cursor >= p.scrollOff+vis {
		p.scrollOff = p.Cursor - vis + 1
	}
}

// Render draws the env panel content (without border).
func (p *EnvPanel) Render(styles theme.Styles, focused bool) string {
	if len(p.Items) == 0 {
		return styles.InactiveItem.Render("  (no envs)")
	}

	vis := p.visibleCount()
	end := min(p.scrollOff+vis, len(p.Items))

	var b strings.Builder
	for i := p.scrollOff; i < end; i++ {
		item := p.Items[i]
		if i == p.Cursor && focused {
			fmt.Fprintf(&b, "  %s", styles.Cursor.Render(" "+item+" "))
		} else if i == p.Selected {
			fmt.Fprintf(&b, "  %s", styles.ActiveItem.Render(item))
		} else {
			fmt.Fprintf(&b, "  %s", styles.InactiveItem.Render(item))
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	if p.scrollOff > 0 || end < len(p.Items) {
		b.WriteString("\n")
		indicator := fmt.Sprintf("  %d/%d", p.Cursor+1, len(p.Items))
		b.WriteString(styles.HelpBar.Render(indicator))
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
