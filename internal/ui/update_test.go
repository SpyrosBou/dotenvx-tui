package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/warui1/dotenvx-tui/internal/secret"
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
