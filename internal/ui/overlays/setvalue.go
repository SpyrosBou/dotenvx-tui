package overlays

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/SpyrosBou/dotenvx-tui/internal/dotenvx"
	"github.com/SpyrosBou/dotenvx-tui/internal/secret"
	"github.com/SpyrosBou/dotenvx-tui/internal/theme"
	"github.com/SpyrosBou/dotenvx-tui/internal/validate"
)

// SetStep tracks the current step in the set-value flow.
type SetStep int

const (
	SetStepKeyName SetStep = iota
	SetStepValue
)

// SetValueOverlay handles creating/updating encrypted secrets.
type SetValueOverlay struct {
	Active   bool
	Step     SetStep
	KeyInput textinput.Model
	ValInput textinput.Model

	// Batch mode
	Keys         []string // keys to set (from multi-select or single)
	CurrentIndex int
	File         string
	Runner       *dotenvx.Runner
	Styles       theme.Styles

	// Pre-existing key info
	ExistingKey   bool
	CurrentValue  *secret.SecureBytes
	KeyValidError string
}

// NewSetValueOverlay creates a new set-value overlay.
func NewSetValueOverlay(styles theme.Styles) SetValueOverlay {
	ki := textinput.New()
	ki.Placeholder = "KEY_NAME"
	ki.SetWidth(40)
	ki.CharLimit = 128
	ki.Focus()

	vi := textinput.New()
	vi.Placeholder = "value"
	vi.SetWidth(40)
	vi.CharLimit = 4096

	return SetValueOverlay{
		KeyInput: ki,
		ValInput: vi,
		Styles:   styles,
	}
}

// Open activates the overlay for setting a key value.
// If keys is non-empty, enters batch mode with those keys.
// If keys is empty and existingKey is set, pre-fills the key name.
func (o *SetValueOverlay) Open(file string, keys []string, existingKey string, runner *dotenvx.Runner) tea.Cmd {
	o.Active = true
	o.File = file
	o.Runner = runner
	o.CurrentIndex = 0

	if len(keys) > 0 {
		// Batch mode: skip key name input
		o.Keys = keys
		o.Step = SetStepValue
		o.ExistingKey = true
		o.ValInput.SetValue("")
		o.ValInput.Focus()
		o.KeyInput.Blur()
		return o.loadCurrentValue()
	}

	// Single key mode
	o.Keys = nil
	if existingKey != "" {
		o.Keys = []string{existingKey}
		o.Step = SetStepValue
		o.ExistingKey = true
		o.KeyInput.SetValue(existingKey)
		o.ValInput.SetValue("")
		o.ValInput.Focus()
		o.KeyInput.Blur()
		return o.loadCurrentValue()
	}

	// New key
	o.Step = SetStepKeyName
	o.ExistingKey = false
	o.KeyInput.SetValue("")
	o.KeyInput.Focus()
	o.ValInput.Blur()
	return nil
}

// Close deactivates the overlay and clears sensitive data.
func (o *SetValueOverlay) Close() {
	o.Active = false
	if o.CurrentValue != nil {
		o.CurrentValue.Clear()
		o.CurrentValue = nil
	}
	o.Keys = nil
	o.KeyValidError = ""
}

// CurrentKeyName returns the key name being set.
func (o *SetValueOverlay) CurrentKeyName() string {
	if len(o.Keys) > 0 && o.CurrentIndex < len(o.Keys) {
		return o.Keys[o.CurrentIndex]
	}
	return o.KeyInput.Value()
}

// Update handles input for the set-value overlay.
func (o *SetValueOverlay) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !o.Active {
		return nil, false
	}

	switch msg := msg.(type) {
	case setValueCurrentValueMsg:
		if msg.File != o.File || msg.Key != o.CurrentKeyName() {
			if msg.Value != nil {
				msg.Value.Clear()
			}
			return nil, true
		}
		if o.CurrentValue != nil {
			o.CurrentValue.Clear()
		}
		o.CurrentValue = msg.Value
		return nil, true

	case setValueStepDoneMsg:
		return nil, true

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("escape"))):
			o.Close()
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return o.handleEnter()

		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			if o.Step == SetStepKeyName {
				// Validate and move to value step
				name := o.KeyInput.Value()
				if err := validate.KeyName(name); err != nil {
					o.KeyValidError = err.Error()
					return nil, true
				}
				o.KeyValidError = ""
				o.Keys = []string{name}
				o.Step = SetStepValue
				o.KeyInput.Blur()
				o.ValInput.Focus()
				return nil, true
			}
		}

		// Forward to active input
		var cmd tea.Cmd
		if o.Step == SetStepKeyName {
			o.KeyInput, cmd = o.KeyInput.Update(msg)
			// Live validation
			name := o.KeyInput.Value()
			if name != "" {
				if err := validate.KeyName(name); err != nil {
					o.KeyValidError = err.Error()
				} else {
					o.KeyValidError = ""
				}
			} else {
				o.KeyValidError = ""
			}
		} else {
			o.ValInput, cmd = o.ValInput.Update(msg)
		}
		return cmd, true
	}

	return nil, false
}

func (o *SetValueOverlay) handleEnter() (tea.Cmd, bool) {
	if o.Step == SetStepKeyName {
		name := o.KeyInput.Value()
		if err := validate.KeyName(name); err != nil {
			o.KeyValidError = err.Error()
			return nil, true
		}
		o.KeyValidError = ""
		o.Keys = []string{name}
		o.Step = SetStepValue
		o.KeyInput.Blur()
		o.ValInput.Focus()
		return nil, true
	}

	// Submit value
	keyName := o.CurrentKeyName()
	value := o.ValInput.Value()
	file := o.File
	runner := o.Runner

	if runner == nil || keyName == "" {
		return nil, true
	}

	// Clear current value preview
	if o.CurrentValue != nil {
		o.CurrentValue.Clear()
		o.CurrentValue = nil
	}

	isFinal := o.CurrentIndex+1 >= len(o.Keys)
	cmd := func() tea.Msg {
		err := runner.Set(context.Background(), file, keyName, []byte(value))
		if err != nil {
			return SetErrorMsg{Err: err}
		}
		if !isFinal {
			return setValueStepDoneMsg{}
		}
		return SetDoneMsg{Key: keyName, File: file}
	}

	// Move to next key in batch, or close
	o.CurrentIndex++
	if o.CurrentIndex < len(o.Keys) {
		o.ValInput.SetValue("")
		return tea.Batch(cmd, o.loadCurrentValue()), true
	}

	o.Close()
	return cmd, true
}

func (o *SetValueOverlay) loadCurrentValue() tea.Cmd {
	if o.Runner == nil || o.CurrentIndex >= len(o.Keys) {
		return nil
	}
	key := o.Keys[o.CurrentIndex]
	file := o.File
	runner := o.Runner
	return func() tea.Msg {
		raw, err := runner.GetValue(context.Background(), file, key)
		if err != nil {
			return setValueCurrentValueMsg{File: file, Key: key, Value: nil}
		}
		return setValueCurrentValueMsg{File: file, Key: key, Value: secret.New(raw)}
	}
}

// View renders the set-value overlay.
func (o *SetValueOverlay) View(width int) string {
	var b strings.Builder

	// Title
	if len(o.Keys) > 1 {
		title := fmt.Sprintf("Set value (%d/%d): %s", o.CurrentIndex+1, len(o.Keys), o.CurrentKeyName())
		b.WriteString(o.Styles.OverlayTitle.Render(title))
	} else if o.ExistingKey {
		b.WriteString(o.Styles.OverlayTitle.Render("Set value: " + o.CurrentKeyName()))
	} else {
		b.WriteString(o.Styles.OverlayTitle.Render("Set new key"))
	}
	b.WriteString("\n\n")

	// Key name input (if in key name step)
	if o.Step == SetStepKeyName {
		b.WriteString("Key name:\n")
		b.WriteString(o.KeyInput.View())
		if o.KeyValidError != "" {
			b.WriteString("\n")
			b.WriteString(o.Styles.StatusError.Render(o.KeyValidError))
		}
		b.WriteString("\n\n")
		b.WriteString(o.Styles.HelpBar.Render("enter/tab: next  esc: cancel"))
	} else {
		// Show current value (masked)
		if o.CurrentValue != nil {
			b.WriteString("Current: ")
			b.WriteString(o.Styles.PreviewMasked.Render(o.CurrentValue.Masked()))
			b.WriteString("\n\n")
		}

		b.WriteString("New value:\n")
		b.WriteString(o.ValInput.View())
		b.WriteString("\n\n")
		b.WriteString(o.Styles.HelpBar.Render("enter: save  esc: cancel"))
	}

	return o.Styles.Overlay.
		Width(min(55, width-4)).
		Render(b.String())
}

// Messages internal to the set overlay.
type setValueCurrentValueMsg struct {
	File  string
	Key   string
	Value *secret.SecureBytes
}

type setValueStepDoneMsg struct{}

// ClearSensitiveMsg clears secret-bearing overlay messages that were not
// accepted by the currently active overlay.
func ClearSensitiveMsg(msg tea.Msg) {
	switch msg := msg.(type) {
	case setValueCurrentValueMsg:
		if msg.Value != nil {
			msg.Value.Clear()
		}
	}
}

// Messages emitted by the set overlay.
type SetDoneMsg struct {
	Key  string
	File string
}

type SetErrorMsg struct {
	Err error
}
