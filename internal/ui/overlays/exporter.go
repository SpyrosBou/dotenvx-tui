package overlays

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	clip "github.com/warui1/dotenvx-tui/internal/clipboard"
	"github.com/warui1/dotenvx-tui/internal/dotenvx"
	"github.com/warui1/dotenvx-tui/internal/theme"
)

// ExportOverlay handles exporting decrypted keys to clipboard.
type ExportOverlay struct {
	Active bool
	Keys   []string // keys to export
	File   string
	Runner *dotenvx.Runner
	Styles theme.Styles

	// Confirmation
	Cursor int // 0 = clipboard, 1 = cancel
}

// NewExportOverlay creates a new export overlay.
func NewExportOverlay(styles theme.Styles) ExportOverlay {
	return ExportOverlay{Styles: styles}
}

// Open activates the export overlay.
func (o *ExportOverlay) Open(file string, keys []string, runner *dotenvx.Runner) {
	o.Active = true
	o.File = file
	o.Keys = keys
	o.Runner = runner
	o.Cursor = 0
}

// Close deactivates the overlay.
func (o *ExportOverlay) Close() {
	o.Active = false
	o.Keys = nil
}

// Update handles input for the export overlay.
func (o *ExportOverlay) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !o.Active {
		return nil, false
	}

	switch msg := msg.(type) {
	case ExportDoneMsg:
		o.Close()
		return nil, true

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("escape"))):
			o.Close()
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			o.Cursor = (o.Cursor - 1 + 2) % 2
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			o.Cursor = (o.Cursor + 1) % 2
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if o.Cursor == 0 {
				return o.exportToClipboard(), true
			}
			o.Close()
			return nil, true
		}
	}

	return nil, false
}

func (o *ExportOverlay) exportToClipboard() tea.Cmd {
	runner := o.Runner
	file := o.File
	keys := make([]string, len(o.Keys))
	copy(keys, o.Keys)

	if runner == nil {
		return nil
	}

	return func() tea.Msg {
		kv, err := runner.GetAll(context.Background(), file)
		if err != nil {
			return ExportErrorMsg{Err: fmt.Errorf("failed to decrypt: %w", err)}
		}

		// Build KEY=VALUE output for selected keys
		var lines []string
		for _, k := range keys {
			if v, ok := kv[k]; ok {
				lines = append(lines, fmt.Sprintf("%s=%s", k, string(v)))
			}
		}
		// If no specific keys, export all
		if len(keys) == 0 {
			sorted := make([]string, 0, len(kv))
			for k := range kv {
				sorted = append(sorted, k)
			}
			sort.Strings(sorted)
			for _, k := range sorted {
				lines = append(lines, fmt.Sprintf("%s=%s", k, string(kv[k])))
			}
		}

		// Zero all values
		for _, v := range kv {
			for i := range v {
				v[i] = 0
			}
		}

		text := strings.Join(lines, "\n")
		if err := clip.Write(text); err != nil {
			return ExportErrorMsg{Err: err}
		}

		return ExportDoneMsg{Count: len(lines)}
	}
}

// View renders the export overlay.
func (o *ExportOverlay) View(width int) string {
	var b strings.Builder

	keyDesc := fmt.Sprintf("%d keys", len(o.Keys))
	if len(o.Keys) == 1 {
		keyDesc = o.Keys[0]
	} else if len(o.Keys) == 0 {
		keyDesc = "all keys"
	}

	b.WriteString(o.Styles.OverlayTitle.Render(fmt.Sprintf("Export %s", keyDesc)))
	b.WriteString("\n\n")

	options := []string{"Copy to clipboard", "Cancel"}
	for i, opt := range options {
		if i == o.Cursor {
			b.WriteString("  " + o.Styles.Cursor.Render(" "+opt+" ") + "\n")
		} else {
			b.WriteString("  " + o.Styles.InactiveItem.Render(opt) + "\n")
		}
	}

	b.WriteString("\n" + o.Styles.HelpBar.Render("enter: confirm  esc: cancel"))

	return o.Styles.Overlay.
		Width(min(50, width-4)).
		Render(b.String())
}

// Messages.
type ExportDoneMsg struct{ Count int }
type ExportErrorMsg struct{ Err error }
