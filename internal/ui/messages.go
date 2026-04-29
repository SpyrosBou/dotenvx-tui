package ui

import (
	"github.com/SpyrosBou/dotenvx-tui/internal/dotenvx"
	"github.com/SpyrosBou/dotenvx-tui/internal/secret"
)

// PanelID identifies which panel is focused.
type PanelID int

const (
	PanelScopes PanelID = iota
	PanelEnvs
	PanelKeys
)

// OverlayKind identifies which overlay is active.
type OverlayKind int

const (
	OverlayNone OverlayKind = iota
	OverlaySetValue
	OverlayDiff
	OverlayImport
	OverlayExport
	OverlayDelete
	OverlayHelp
)

// StatusLevel indicates the severity of a status message.
type StatusLevel int

const (
	StatusSuccess StatusLevel = iota
	StatusError
	StatusWarning
	StatusInfo
)

// Discovery messages.
type FilesDiscoveredMsg struct{ Files []dotenvx.EnvFile }
type DiscoveryErrorMsg struct{ Err error }

// Panel cascade messages.
type KeysLoadedMsg struct {
	File string
	Keys []string
}
type KeysLoadErrorMsg struct {
	File string
	Err  error
}

// Preview messages.
type ValueLoadedMsg struct {
	File  string
	Key   string
	Value *secret.SecureBytes
}
type ValueLoadErrorMsg struct {
	File string
	Key  string
	Err  error
}

// Action messages.
type CopyCompleteMsg struct{ Key string }
type CopyMultiCompleteMsg struct{ Count int }
type CopyErrorMsg struct{ Err error }

// Status.
type ClearStatusMsg struct{ ID int }

// Auto-mask timer.
type AutoMaskMsg struct{ ID int }
