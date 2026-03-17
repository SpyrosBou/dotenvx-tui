package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	width     int
	height    int
	targetDir string
}

func newModel(targetDir string) model {
	return model{targetDir: targetDir}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() tea.View {
	content := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "dotenvx-tui")

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func main() {
	targetDir := ""
	if len(os.Args) > 1 {
		targetDir = os.Args[1]
	} else {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		targetDir = dir
	}

	p := tea.NewProgram(newModel(targetDir))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
