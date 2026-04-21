package overlays

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/SpyrosBou/dotenvx-tui/internal/dotenvx"
	"github.com/SpyrosBou/dotenvx-tui/internal/theme"
)

func TestBatchSetKeepsOverlayOpenAfterIntermediateKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("DOTENV_PUBLIC_KEY=x\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	fakeDotenvx := filepath.Join(dir, "dotenvx")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"get\" ]; then\n" +
		"  printf 'current-value'\n" +
		"fi\n"
	if err := os.WriteFile(fakeDotenvx, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake dotenvx: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	runner, err := dotenvx.NewRunner(dir)
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}

	styles := theme.NewStyles(theme.NewTheme(true))
	overlay := NewSetValueOverlay(styles)
	if cmd := overlay.Open(".env.local", []string{"FIRST", "SECOND"}, "", runner); cmd == nil {
		t.Fatal("Open returned nil load command")
	}
	overlay.ValInput.SetValue("first-value")

	cmd, handled := overlay.handleEnter()
	if !handled {
		t.Fatal("handleEnter was not handled")
	}
	if !overlay.Active {
		t.Fatal("overlay closed before batch completed")
	}
	if overlay.CurrentIndex != 1 {
		t.Fatalf("CurrentIndex = %d, want 1", overlay.CurrentIndex)
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("handleEnter command returned %T, want tea.BatchMsg", msg)
	}

	foundIntermediateDone := false
	for _, child := range batch {
		switch child().(type) {
		case setValueStepDoneMsg:
			foundIntermediateDone = true
		case SetDoneMsg:
			t.Fatal("intermediate batch key emitted final SetDoneMsg")
		}
	}
	if !foundIntermediateDone {
		t.Fatal("intermediate batch key did not emit setValueStepDoneMsg")
	}
}
