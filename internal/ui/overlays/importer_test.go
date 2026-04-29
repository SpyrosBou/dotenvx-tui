package overlays

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SpyrosBou/dotenvx-tui/internal/dotenvx"
	"github.com/SpyrosBou/dotenvx-tui/internal/theme"
)

func TestImportOverlayLoadsKeysThroughDotenvxParser(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.import"), []byte("A='  spaced  '\n"), 0o600); err != nil {
		t.Fatalf("write import file: %v", err)
	}

	fakeDotenvx := filepath.Join(dir, "dotenvx")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = get ]; then\n" +
		"  printf '{\"B\":\"two\",\"A\":\"  spaced  \",\"DOTENV_PUBLIC_KEY\":\"x\",\"BAD-NAME\":\"skip\"}'\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 64\n"
	if err := os.WriteFile(fakeDotenvx, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake dotenvx: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	runner, err := dotenvx.NewRunner(dir)
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}

	overlay := NewImportOverlay(theme.NewStyles(theme.NewTheme(true)))
	overlay.Open(dir, ".env.local", runner)
	msg := overlay.loadKeysFromFile(".env.import")()
	loaded, ok := msg.(importKeysLoadedMsg)
	if !ok {
		t.Fatalf("loadKeysFromFile returned %T, want importKeysLoadedMsg", msg)
	}

	if len(loaded.Keys) != 2 {
		t.Fatalf("loaded %d keys, want 2: %#v", len(loaded.Keys), loaded.Keys)
	}
	if loaded.Keys[0].Name != "A" || loaded.Keys[0].Value != "  spaced  " {
		t.Fatalf("first key = %#v, want A with preserved spaces", loaded.Keys[0])
	}
	if loaded.Keys[1].Name != "B" || loaded.Keys[1].Value != "two" {
		t.Fatalf("second key = %#v, want B=two", loaded.Keys[1])
	}
}

func TestImportOverlayRequiresSelectedKey(t *testing.T) {
	overlay := NewImportOverlay(theme.NewStyles(theme.NewTheme(true)))
	overlay.Active = true
	overlay.Step = ImportStepSelectKeys
	overlay.Keys = []ImportKey{{Name: "A", Value: "one", Selected: false}}

	cmd, handled := overlay.handleEnter()
	if !handled {
		t.Fatal("handleEnter was not handled")
	}
	if cmd != nil {
		t.Fatal("handleEnter returned command without selected keys")
	}
	if overlay.Error != "Select at least one key to import" {
		t.Fatalf("Error = %q, want select warning", overlay.Error)
	}
}

func TestImportOverlayShowsNoImportableKeys(t *testing.T) {
	overlay := NewImportOverlay(theme.NewStyles(theme.NewTheme(true)))
	overlay.Active = true
	overlay.Step = ImportStepSelectKeys
	overlay.TargetFile = ".env.local"

	view := overlay.View(100)
	if !strings.Contains(view, "No importable keys found.") {
		t.Fatalf("view did not explain empty key list: %q", view)
	}
}
