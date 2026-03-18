package ui

import "charm.land/bubbles/v2/key"

// KeyMap holds all key bindings for the application.
type KeyMap struct {
	Quit       key.Binding
	Up         key.Binding
	Down       key.Binding
	NextPanel  key.Binding
	PrevPanel  key.Binding
	Select     key.Binding
	ToggleSel  key.Binding
	SelectAll  key.Binding
	Set        key.Binding
	Diff       key.Binding
	Import     key.Binding
	Export     key.Binding
	Copy       key.Binding
	Help       key.Binding
	Back       key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		NextPanel: key.NewBinding(
			key.WithKeys("tab", "l", "right"),
			key.WithHelp("→/tab", "next panel"),
		),
		PrevPanel: key.NewBinding(
			key.WithKeys("shift+tab", "h", "left"),
			key.WithHelp("←/shift+tab", "prev panel"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		ToggleSel: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle select"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		Set: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "set"),
		),
		Diff: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "diff"),
		),
		Import: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "import"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Back: key.NewBinding(
			key.WithKeys("escape"),
			key.WithHelp("esc", "back"),
		),
	}
}
