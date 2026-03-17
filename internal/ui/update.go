package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/warui1/dotenvx-tui/internal/dotenvx"
	"github.com/warui1/dotenvx-tui/internal/theme"
)

// Update handles all messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle overlay input first if an overlay is active
	if m.activeOverlay != OverlayNone {
		return m.updateOverlay(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout = ComputeLayout(m.width, m.height, len(dotenvx.Scopes(m.envFiles)))
		m.updatePanelSizes()
		return m, nil

	case tea.BackgroundColorMsg:
		m.hasDarkBG = msg.IsDark()
		m.theme = theme.NewTheme(m.hasDarkBG)
		m.styles = theme.NewStyles(m.theme)
		return m, nil

	case FilesDiscoveredMsg:
		return m.handleFilesDiscovered(msg)

	case DiscoveryErrorMsg:
		m.fatalErr = msg.Err.Error()
		return m, nil

	case KeysLoadedMsg:
		m.keyPanel.Reset(msg.Keys)
		m.loading = false
		// Load preview for first key
		if len(msg.Keys) > 0 {
			return m, m.loadValue(m.currentFile(), msg.Keys[0])
		}
		return m, nil

	case KeysLoadErrorMsg:
		m.loading = false
		cmd := m.setStatus("Failed to load keys: "+msg.Err.Error(), StatusError)
		return m, cmd

	case ValueLoadedMsg:
		// Clear old preview value
		if m.previewValue != nil {
			m.previewValue.Clear()
		}
		m.previewKey = msg.Key
		m.previewValue = msg.Value
		m.previewShown = false
		return m, nil

	case ValueLoadErrorMsg:
		cmd := m.setStatus("Failed to decrypt: "+msg.Err.Error(), StatusError)
		return m, cmd

	case SetCompleteMsg:
		cmd := m.setStatus("Set "+msg.Key+" in "+msg.File, StatusSuccess)
		// Reload keys
		return m, tea.Batch(cmd, m.loadKeys(msg.File))

	case SetErrorMsg:
		cmd := m.setStatus("Set failed: "+msg.Err.Error(), StatusError)
		return m, cmd

	case CopyCompleteMsg:
		cmd := m.setStatus("Copied "+msg.Key+" to clipboard", StatusSuccess)
		return m, cmd

	case CopyMultiCompleteMsg:
		cmd := m.setStatus(formatCount(msg.Count, "value")+" copied to clipboard", StatusSuccess)
		return m, cmd

	case ClearStatusMsg:
		if msg.ID == m.statusID {
			m.statusMsg = ""
		}
		return m, nil

	case AutoMaskMsg:
		m.previewShown = false
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleFilesDiscovered(msg FilesDiscoveredMsg) (tea.Model, tea.Cmd) {
	m.envFiles = msg.Files

	scopes := dotenvx.Scopes(m.envFiles)
	m.scopePanel.Items = scopes
	m.scopePanel.Cursor = 0
	m.scopePanel.Selected = 0

	// Try to init runner
	runner, err := dotenvx.NewRunner(m.targetDir)
	if err != nil {
		m.fatalErr = err.Error()
		return m, nil
	}
	m.runner = runner

	// Recalculate layout
	m.layout = ComputeLayout(m.width, m.height, len(scopes))
	m.updatePanelSizes()

	// If we have scopes, populate envs for first scope
	if len(scopes) > 0 {
		return m.selectScope(scopes[0])
	}
	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	km := m.keyMap

	switch {
	case key.Matches(msg, km.Quit):
		// Clear any secrets before quitting
		if m.previewValue != nil {
			m.previewValue.Clear()
		}
		return m, tea.Quit

	case key.Matches(msg, km.NextPanel):
		m.focusedPanel = (m.focusedPanel + 1) % 3
		if m.layout.HideScopes && m.focusedPanel == PanelScopes {
			m.focusedPanel = PanelEnvs
		}
		return m, nil

	case key.Matches(msg, km.PrevPanel):
		m.focusedPanel--
		if m.focusedPanel < 0 {
			m.focusedPanel = PanelKeys
		}
		if m.layout.HideScopes && m.focusedPanel == PanelScopes {
			m.focusedPanel = PanelKeys
		}
		return m, nil

	case key.Matches(msg, km.Up):
		return m.handleCursorUp()

	case key.Matches(msg, km.Down):
		return m.handleCursorDown()

	case key.Matches(msg, km.Select):
		return m.handleSelect()

	case key.Matches(msg, km.ToggleSel):
		if m.focusedPanel == PanelKeys {
			m.keyPanel.ToggleSelect()
		}
		return m, nil

	case key.Matches(msg, km.SelectAll):
		if m.focusedPanel == PanelKeys {
			m.keyPanel.ToggleSelectAll()
		}
		return m, nil

	case key.Matches(msg, km.Help):
		m.activeOverlay = OverlayHelp
		return m, nil

	case key.Matches(msg, km.Copy):
		return m.handleCopy()

	case key.Matches(msg, km.Set):
		m.activeOverlay = OverlaySetValue
		return m, nil

	case key.Matches(msg, km.Diff):
		m.activeOverlay = OverlayDiff
		return m, nil

	case key.Matches(msg, km.Import):
		m.activeOverlay = OverlayImport
		return m, nil

	case key.Matches(msg, km.Export):
		m.activeOverlay = OverlayExport
		return m, nil
	}

	return m, nil
}

func (m Model) handleCursorUp() (tea.Model, tea.Cmd) {
	switch m.focusedPanel {
	case PanelScopes:
		m.scopePanel.CursorUp()
	case PanelEnvs:
		m.envPanel.CursorUp()
	case PanelKeys:
		m.keyPanel.CursorUp()
		// Update preview for new cursor position
		curKey := m.keyPanel.CursorItem()
		if curKey != "" && curKey != m.previewKey {
			return m, m.loadValue(m.currentFile(), curKey)
		}
	}
	return m, nil
}

func (m Model) handleCursorDown() (tea.Model, tea.Cmd) {
	switch m.focusedPanel {
	case PanelScopes:
		m.scopePanel.CursorDown()
	case PanelEnvs:
		m.envPanel.CursorDown()
	case PanelKeys:
		m.keyPanel.CursorDown()
		curKey := m.keyPanel.CursorItem()
		if curKey != "" && curKey != m.previewKey {
			return m, m.loadValue(m.currentFile(), curKey)
		}
	}
	return m, nil
}

func (m Model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.focusedPanel {
	case PanelScopes:
		m.scopePanel.Select()
		return m.selectScope(m.scopePanel.SelectedItem())
	case PanelEnvs:
		m.envPanel.Select()
		return m.selectEnv(m.envPanel.SelectedItem())
	case PanelKeys:
		// Toggle preview reveal
		m.previewShown = !m.previewShown
		return m, nil
	}
	return m, nil
}

func (m Model) selectScope(scope string) (tea.Model, tea.Cmd) {
	envs := dotenvx.EnvsForScope(m.envFiles, scope)
	m.envPanel.Reset(envs)
	if len(envs) > 0 {
		return m.selectEnv(envs[0])
	}
	m.keyPanel.Reset(nil)
	return m, nil
}

func (m Model) selectEnv(_ string) (tea.Model, tea.Cmd) {
	m.envPanel.Select()
	file := m.currentFile()
	if file == "" {
		return m, nil
	}
	m.loading = true
	m.loadingMsg = "Loading keys..."
	// Clear old preview
	if m.previewValue != nil {
		m.previewValue.Clear()
		m.previewValue = nil
	}
	m.previewKey = ""
	return m, m.loadKeys(file)
}

func (m Model) handleCopy() (tea.Model, tea.Cmd) {
	// Placeholder — will be implemented in clipboard task
	cmd := m.setStatus("Copy not yet implemented", StatusWarning)
	return m, cmd
}

func (m Model) updateOverlay(msg tea.Msg) (tea.Model, tea.Cmd) {
	// For now, just handle Esc to close overlays
	if kmsg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(kmsg, m.keyMap.Back) || key.Matches(kmsg, m.keyMap.Quit) {
			m.activeOverlay = OverlayNone
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) updatePanelSizes() {
	m.scopePanel.Width = m.layout.ScopeWidth
	m.scopePanel.Height = m.layout.PanelHeight
	m.envPanel.Width = m.layout.EnvWidth
	m.envPanel.Height = m.layout.PanelHeight
	m.keyPanel.Width = m.layout.KeysWidth
	m.keyPanel.Height = m.layout.PanelHeight
}

func formatCount(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return string(rune('0'+n)) + " " + noun + "s"
}
