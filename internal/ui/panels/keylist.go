package panels

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/warui1/dotenvx-tui/internal/theme"
)

// KeyListPanel displays environment variable keys with multi-select support.
type KeyListPanel struct {
	Items    []string
	Cursor   int
	Selected map[int]struct{} // multi-select indices
	Width    int
	Height   int
}

// NewKeyListPanel creates a new key list panel.
func NewKeyListPanel() KeyListPanel {
	return KeyListPanel{
		Selected: make(map[int]struct{}),
	}
}

// CursorUp moves the cursor up with circular wrapping.
func (p *KeyListPanel) CursorUp() {
	if len(p.Items) == 0 {
		return
	}
	p.Cursor--
	if p.Cursor < 0 {
		p.Cursor = len(p.Items) - 1
	}
}

// CursorDown moves the cursor down with circular wrapping.
func (p *KeyListPanel) CursorDown() {
	if len(p.Items) == 0 {
		return
	}
	p.Cursor++
	if p.Cursor >= len(p.Items) {
		p.Cursor = 0
	}
}

// ToggleSelect toggles the selection state of the item at the cursor.
func (p *KeyListPanel) ToggleSelect() {
	if len(p.Items) == 0 {
		return
	}
	if _, ok := p.Selected[p.Cursor]; ok {
		delete(p.Selected, p.Cursor)
	} else {
		p.Selected[p.Cursor] = struct{}{}
	}
}

// ToggleSelectAll selects all items if none are selected, otherwise deselects all.
func (p *KeyListPanel) ToggleSelectAll() {
	if len(p.Selected) > 0 {
		p.Selected = make(map[int]struct{})
	} else {
		for i := range p.Items {
			p.Selected[i] = struct{}{}
		}
	}
}

// CursorItem returns the key name at the cursor position.
func (p *KeyListPanel) CursorItem() string {
	if len(p.Items) == 0 || p.Cursor >= len(p.Items) {
		return ""
	}
	return p.Items[p.Cursor]
}

// SelectedItems returns all selected key names, or the cursor item if none selected.
func (p *KeyListPanel) SelectedItems() []string {
	if len(p.Selected) == 0 {
		item := p.CursorItem()
		if item != "" {
			return []string{item}
		}
		return nil
	}
	var items []string
	for i := range p.Items {
		if _, ok := p.Selected[i]; ok {
			items = append(items, p.Items[i])
		}
	}
	return items
}

// SelectionCount returns the number of selected items.
func (p *KeyListPanel) SelectionCount() int {
	return len(p.Selected)
}

// Reset clears items and selection, resets cursor.
func (p *KeyListPanel) Reset(items []string) {
	p.Items = items
	p.Cursor = 0
	p.Selected = make(map[int]struct{})
}

// Render draws the key list panel content (without border).
func (p *KeyListPanel) Render(styles theme.Styles, focused bool) string {
	if len(p.Items) == 0 {
		return styles.InactiveItem.Render("  (no keys)")
	}

	var b strings.Builder
	for i, item := range p.Items {
		marker := "[ ]"
		if _, ok := p.Selected[i]; ok {
			marker = "[x]"
		}

		if i == p.Cursor && focused {
			line := fmt.Sprintf(" %s %s ", marker, item)
			fmt.Fprintf(&b, "%s", styles.Cursor.Render(line))
		} else if _, ok := p.Selected[i]; ok {
			fmt.Fprintf(&b, "  %s %s", styles.ActiveItem.Render(marker), styles.ActiveItem.Render(item))
		} else {
			fmt.Fprintf(&b, "  %s %s", styles.InactiveItem.Render(marker), styles.InactiveItem.Render(item))
		}
		if i < len(p.Items)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// Title returns the panel title with selection count.
func (p *KeyListPanel) Title(styles theme.Styles, focused bool) string {
	title := "Keys"
	if count := p.SelectionCount(); count > 0 {
		title = fmt.Sprintf("Keys (%d selected)", count)
	}
	if focused {
		return styles.FocusedTitle.Render(title)
	}
	return styles.BlurredTitle.Render(title)
}

// PanelStyle returns the border style for this panel.
func (p *KeyListPanel) PanelStyle(t theme.Theme, focused bool) lipgloss.Style {
	return theme.PanelStyle(t, focused, p.Width, p.Height)
}
