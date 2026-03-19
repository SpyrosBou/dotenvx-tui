package ui

import (
	"github.com/warui1/dotenvx-tui/internal/dotenvx"
	"github.com/warui1/dotenvx-tui/internal/secret"
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
type KeysLoadedMsg struct{ Keys []string }
type KeysLoadErrorMsg struct{ Err error }

// Preview messages.
type ValueLoadedMsg struct {
	Key   string
	Value *secret.SecureBytes
}
type ValueLoadErrorMsg struct{ Err error }

// Action messages.
type SetErrorMsg struct{ Err error }
type CopyCompleteMsg struct{ Key string }
type CopyMultiCompleteMsg struct{ Count int }

// Status.
type ClearStatusMsg struct{ ID int }

// Auto-mask timer.
type AutoMaskMsg struct{ ID int }
