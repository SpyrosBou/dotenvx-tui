package overlays

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/SpyrosBou/dotenvx-tui/internal/dotenvx"
	"github.com/SpyrosBou/dotenvx-tui/internal/theme"
)

// DeleteOverlay handles confirming and executing key deletion.
type DeleteOverlay struct {
	Active bool
	Keys   []string
	File   string
	Runner *dotenvx.Runner
	Styles theme.Styles
}

// NewDeleteOverlay creates a new delete overlay.
func NewDeleteOverlay(styles theme.Styles) DeleteOverlay {
	return DeleteOverlay{Styles: styles}
}

// Open activates the overlay for deleting keys.
func (o *DeleteOverlay) Open(file string, keys []string, runner *dotenvx.Runner) {
	o.Active = true
	o.File = file
	o.Keys = keys
	o.Runner = runner
}

// Close deactivates the overlay.
func (o *DeleteOverlay) Close() {
	o.Active = false
	o.Keys = nil
}

// Update handles input for the delete overlay.
func (o *DeleteOverlay) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !o.Active {
		return nil, false
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("escape"))):
			o.Close()
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			keys := o.Keys
			file := o.File
			runner := o.Runner

			if runner == nil || len(keys) == 0 {
				return nil, true
			}

			o.Close()
			return func() tea.Msg {
				err := runner.Unset(context.Background(), file, keys)
				if err != nil {
					return DeleteErrorMsg{Err: err}
				}
				return DeleteDoneMsg{Keys: keys, File: file}
			}, true
		}

		return nil, true
	}

	return nil, false
}

// View renders the delete confirmation overlay.
func (o *DeleteOverlay) View(width int) string {
	var b strings.Builder

	if len(o.Keys) == 1 {
		b.WriteString(o.Styles.OverlayTitle.Render("Delete " + o.Keys[0] + "?"))
	} else {
		title := fmt.Sprintf("Delete %d keys?", len(o.Keys))
		b.WriteString(o.Styles.OverlayTitle.Render(title))
	}
	b.WriteString("\n\n")

	if len(o.Keys) > 1 {
		maxShow := 10
		for i, k := range o.Keys {
			if i >= maxShow {
				remaining := len(o.Keys) - maxShow
				b.WriteString(o.Styles.InactiveItem.Render(fmt.Sprintf("  ... and %d more", remaining)))
				b.WriteString("\n")
				break
			}
			b.WriteString(o.Styles.InactiveItem.Render("  " + k))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(o.Styles.StatusWarning.Render("This cannot be undone."))
	b.WriteString("\n\n")
	b.WriteString(o.Styles.HelpBar.Render("enter: confirm  esc: cancel"))

	return o.Styles.Overlay.
		Width(min(55, width-4)).
		Render(b.String())
}

// Messages emitted by the delete overlay.
type DeleteDoneMsg struct {
	Keys []string
	File string
}

type DeleteErrorMsg struct {
	Err error
}
