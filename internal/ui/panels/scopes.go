package panels

import (
	"fmt"
	"strings"

	"github.com/warui1/dotenvx-tui/internal/theme"
)

// ScopePanel displays the list of discovered scopes (directories).
type ScopePanel struct {
	Items      []string
	Cursor     int
	Selected   int
	Width      int
	Height     int
	scrollOff  int
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
	p.ensureVisible()
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
	p.ensureVisible()
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

func (p *ScopePanel) visibleCount() int {
	v := p.Height - 1 // subtract title line
	if v < 1 {
		v = 1
	}
	return v
}

func (p *ScopePanel) ensureVisible() {
	vis := p.visibleCount()
	if p.Cursor < p.scrollOff {
		p.scrollOff = p.Cursor
	} else if p.Cursor >= p.scrollOff+vis {
		p.scrollOff = p.Cursor - vis + 1
	}
}

// Render draws the scope panel content (without border).
func (p *ScopePanel) Render(styles theme.Styles, focused bool) string {
	if len(p.Items) == 0 {
		return styles.InactiveItem.Render("  (no scopes)")
	}

	vis := p.visibleCount()
	end := min(p.scrollOff+vis, len(p.Items))

	var b strings.Builder
	for i := p.scrollOff; i < end; i++ {
		label := scopeLabel(p.Items[i])
		if i == p.Cursor && focused {
			fmt.Fprintf(&b, "  %s", styles.Cursor.Render(" "+label+" "))
		} else if i == p.Selected {
			fmt.Fprintf(&b, "  %s", styles.ActiveItem.Render(label))
		} else {
			fmt.Fprintf(&b, "  %s", styles.InactiveItem.Render(label))
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	// Scroll indicators
	if p.scrollOff > 0 || end < len(p.Items) {
		b.WriteString("\n")
		indicator := fmt.Sprintf("  %d/%d", p.Cursor+1, len(p.Items))
		b.WriteString(styles.HelpBar.Render(indicator))
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

func scopeLabel(scope string) string {
	if scope == "." {
		return "(root)"
	}
	return scope
}
