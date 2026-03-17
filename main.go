package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/warui1/dotenvx-tui/internal/ui"
)

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

	p := tea.NewProgram(ui.NewModel(targetDir))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
