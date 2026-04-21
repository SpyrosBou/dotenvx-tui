package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/warui1/dotenvx-tui/internal/ui"
)

var version = "dev"

const usage = `dotenvx-tui — interactive TUI for managing dotenvx-encrypted environment variables

Install:
  npm install -g dotenvx-tui
  npx dotenvx-tui

Usage:
  dotenvx-tui [directory]

Arguments:
  directory    Project directory to manage (default: current directory)

Keybindings:
  tab/shift+tab   Switch panels (Scopes → Envs → Keys)
  j/k or ↑/↓      Navigate within panel
  enter            Select / Reveal value
  space            Toggle multi-select (Keys panel)
  s                Set key value
  g                Get / decrypt value
  d                Diff two environments
  i                Import from plaintext .env file
  e                Export to clipboard
  c                Copy value to clipboard
  ?                Help overlay
  q / ctrl+c       Quit

Options:
  -h, --help       Show this help
  -v, --version    Show version`

func main() {
	// Parse flags
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			fmt.Println(usage)
			os.Exit(0)
		case "-v", "--version":
			fmt.Println("dotenvx-tui " + version)
			os.Exit(0)
		}
	}

	targetDir := ""
	if len(os.Args) > 1 && os.Args[1] != "" {
		abs, err := filepath.Abs(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		targetDir = abs
	} else {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		targetDir = dir
	}

	// Verify directory exists
	info, err := os.Stat(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: directory not found: %s\n", targetDir)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: not a directory: %s\n", targetDir)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewModel(targetDir))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
