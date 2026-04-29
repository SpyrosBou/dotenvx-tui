package ui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/SpyrosBou/dotenvx-tui/internal/dotenvx"
	"github.com/SpyrosBou/dotenvx-tui/internal/secret"
)

func TestUpdateIgnoresStaleValueLoadedMsg(t *testing.T) {
	m := NewModel(t.TempDir())
	current := secret.New([]byte("current"))
	stale := secret.New([]byte("stale"))
	m.previewKey = "CURRENT"
	m.previewValue = current

	updated, cmd := m.Update(ValueLoadedMsg{Key: "OLD", Value: stale})
	if cmd != nil {
		t.Fatalf("Update returned command for stale value")
	}

	got := updated.(Model)
	if got.previewKey != "CURRENT" {
		t.Fatalf("previewKey = %q, want CURRENT", got.previewKey)
	}
	if got.previewValue.String() != "current" {
		t.Fatalf("previewValue = %q, want current", got.previewValue.String())
	}
	if stale.String() != "" {
		t.Fatalf("stale value was not cleared")
	}
}

func TestUpdateIgnoresStaleValueLoadedMsgFromDifferentFile(t *testing.T) {
	m := modelWithCurrentFile(t, ".env.local", ".", "local")
	current := secret.New([]byte("current"))
	stale := secret.New([]byte("stale"))
	m.previewKey = "TOKEN"
	m.previewValue = current

	updated, cmd := m.Update(ValueLoadedMsg{File: ".env.production", Key: "TOKEN", Value: stale})
	if cmd != nil {
		t.Fatalf("Update returned command for stale value")
	}

	got := updated.(Model)
	if got.previewValue.String() != "current" {
		t.Fatalf("previewValue = %q, want current", got.previewValue.String())
	}
	if stale.String() != "" {
		t.Fatalf("stale value was not cleared")
	}
}

func TestUpdateIgnoresStaleValueLoadErrorMsg(t *testing.T) {
	m := NewModel(t.TempDir())
	m.previewKey = "CURRENT"
	m.previewValue = secret.New([]byte("current"))

	updated, cmd := m.Update(ValueLoadErrorMsg{Key: "OLD", Err: errors.New("boom")})
	if cmd != nil {
		t.Fatalf("Update returned command for stale error")
	}

	got := updated.(Model)
	if got.previewKey != "CURRENT" {
		t.Fatalf("previewKey = %q, want CURRENT", got.previewKey)
	}
	if got.statusMsg != "" {
		t.Fatalf("statusMsg = %q, want empty", got.statusMsg)
	}
}

func TestUpdateClearsPreviewOnCurrentValueLoadError(t *testing.T) {
	m := modelWithCurrentFile(t, ".env.local", ".", "local")
	m.previewKey = "CURRENT"
	m.previewValue = secret.New([]byte("current"))

	updated, cmd := m.Update(ValueLoadErrorMsg{File: ".env.local", Key: "CURRENT", Err: errors.New("boom")})
	if cmd == nil {
		t.Fatal("Update returned nil status command")
	}

	got := updated.(Model)
	if got.previewKey != "" {
		t.Fatalf("previewKey = %q, want empty", got.previewKey)
	}
	if got.previewValue != nil {
		t.Fatal("previewValue was not cleared")
	}
	if got.statusMsg == "" {
		t.Fatal("status message was not set")
	}
}

func TestGetKeyTogglesPreviewReveal(t *testing.T) {
	m := NewModel(t.TempDir())
	m.focusedPanel = PanelKeys
	m.previewKey = "SECRET"
	m.previewValue = secret.New([]byte("value"))

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'g', Text: "g"}))
	got := updated.(Model)
	if !got.previewShown {
		t.Fatal("g key did not reveal preview")
	}
}

func TestKeysLoadedArmsFirstPreviewForCurrentFile(t *testing.T) {
	m := modelWithCurrentFile(t, ".env.local", ".", "local")
	m.loading = true

	updated, cmd := m.Update(KeysLoadedMsg{File: ".env.local", Keys: []string{"API_KEY", "DATABASE_URL"}})
	if cmd != nil {
		t.Fatalf("Update returned command without runner")
	}

	got := updated.(Model)
	if got.loading {
		t.Fatal("loading remained true")
	}
	if got.previewKey != "API_KEY" {
		t.Fatalf("previewKey = %q, want API_KEY", got.previewKey)
	}
	if got.previewValue != nil {
		t.Fatal("previewValue was set before value load completed")
	}
}

func TestKeysLoadedIgnoresStaleFile(t *testing.T) {
	m := modelWithCurrentFile(t, ".env.local", ".", "local")
	m.loading = true
	m.keyPanel.Reset([]string{"CURRENT"})

	updated, cmd := m.Update(KeysLoadedMsg{File: ".env.production", Keys: []string{"STALE"}})
	if cmd != nil {
		t.Fatalf("Update returned command for stale keys")
	}

	got := updated.(Model)
	if !got.loading {
		t.Fatal("stale key load cleared loading state")
	}
	if got.keyPanel.CursorItem() != "CURRENT" {
		t.Fatalf("current key = %q, want CURRENT", got.keyPanel.CursorItem())
	}
}

func TestFilesDiscoveredHandledWhileOverlayOpen(t *testing.T) {
	m := NewModel(t.TempDir())
	m.activeOverlay = OverlayHelp
	m.loading = true

	updated, _ := m.Update(FilesDiscoveredMsg{Files: nil})
	got := updated.(Model)
	defer got.cleanup()

	if got.loading {
		t.Fatal("FilesDiscoveredMsg was dropped while overlay was open")
	}
	if got.activeOverlay != OverlayHelp {
		t.Fatalf("activeOverlay = %v, want OverlayHelp", got.activeOverlay)
	}
}

func TestFilesDiscoveredPreservesSelectionAndNormalizesHiddenScopeFocus(t *testing.T) {
	dir := t.TempDir()
	writeFakeDotenvx(t, dir)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	m := NewModel(dir)
	m.width = 80
	m.height = 30
	m.focusedPanel = PanelScopes
	m.envFiles = []dotenvx.EnvFile{
		{Path: ".env.local", Scope: ".", Env: "local"},
		{Path: filepath.Join("apps", "api", ".env.production"), Scope: filepath.Join("apps", "api"), Env: "production"},
	}
	m.scopePanel.Items = []string{".", filepath.Join("apps", "api")}
	m.scopePanel.Selected = 1
	m.scopePanel.Cursor = 1
	m.envPanel.Reset([]string{"production"})
	m.envPanel.Selected = 0

	files := []dotenvx.EnvFile{
		{Path: filepath.Join("apps", "api", ".env.local"), Scope: filepath.Join("apps", "api"), Env: "local"},
		{Path: filepath.Join("apps", "api", ".env.production"), Scope: filepath.Join("apps", "api"), Env: "production"},
	}
	updated, _ := m.Update(FilesDiscoveredMsg{Files: files})
	got := updated.(Model)
	defer got.cleanup()

	if got.focusedPanel != PanelEnvs {
		t.Fatalf("focusedPanel = %v, want PanelEnvs", got.focusedPanel)
	}
	if got.scopePanel.SelectedItem() != filepath.Join("apps", "api") {
		t.Fatalf("selected scope = %q, want apps/api", got.scopePanel.SelectedItem())
	}
	if got.envPanel.SelectedItem() != "production" {
		t.Fatalf("selected env = %q, want production", got.envPanel.SelectedItem())
	}
	if got.currentFile() != filepath.Join("apps", "api", ".env.production") {
		t.Fatalf("currentFile = %q, want apps/api/.env.production", got.currentFile())
	}
}

func modelWithCurrentFile(t *testing.T, file, scope, env string) Model {
	t.Helper()
	m := NewModel(t.TempDir())
	m.envFiles = []dotenvx.EnvFile{{Path: file, Scope: scope, Env: env}}
	m.scopePanel.Items = []string{scope}
	m.scopePanel.Selected = 0
	m.envPanel.Reset([]string{env})
	m.envPanel.Selected = 0
	return m
}

func writeFakeDotenvx(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "dotenvx")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatalf("write fake dotenvx: %v", err)
	}
}
