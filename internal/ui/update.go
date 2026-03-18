package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	clip "github.com/warui1/dotenvx-tui/internal/clipboard"
	"github.com/warui1/dotenvx-tui/internal/dotenvx"
	"github.com/warui1/dotenvx-tui/internal/theme"
	"github.com/warui1/dotenvx-tui/internal/ui/overlays"
)

// Update handles all messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ctrl+c always quits, even with overlays open
	if kmsg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(kmsg, m.keyMap.Quit) {
			m.cleanup()
			return m, tea.Quit
		}
	}

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
		m, cmd := setStatus(m,"Failed to load keys: "+msg.Err.Error(), StatusError)
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
		m, cmd := setStatus(m,"Failed to decrypt: "+msg.Err.Error(), StatusError)
		return m, cmd

	case SetErrorMsg:
		m, cmd := setStatus(m,"Set failed: "+msg.Err.Error(), StatusError)
		return m, cmd

	case CopyCompleteMsg:
		m, cmd := setStatus(m,"Copied "+msg.Key+" to clipboard", StatusSuccess)
		return m, cmd

	case CopyMultiCompleteMsg:
		m, cmd := setStatus(m,formatCount(msg.Count, "value")+" copied to clipboard", StatusSuccess)
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
		m.cleanup()
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

	case key.Matches(msg, km.Back):
		return m.handleBack()

	case key.Matches(msg, km.Get):
		return m.handleGet()

	case key.Matches(msg, km.Help):
		m.activeOverlay = OverlayHelp
		return m, nil

	case key.Matches(msg, km.Copy):
		return m.handleCopy()

	case key.Matches(msg, km.Set):
		return m.openSetOverlay()

	case key.Matches(msg, km.Diff):
		return m.openDiffOverlay()

	case key.Matches(msg, km.Import):
		return m.openImportOverlay()

	case key.Matches(msg, km.Export):
		return m.openExportOverlay()
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

func (m Model) handleBack() (tea.Model, tea.Cmd) {
	// Esc in browse mode: move focus back one panel, or clear selection
	if m.keyPanel.SelectionCount() > 0 {
		m.keyPanel.ToggleSelectAll() // clears all
		return m, nil
	}
	if m.focusedPanel > PanelScopes {
		m.focusedPanel--
		if m.layout.HideScopes && m.focusedPanel == PanelScopes {
			m.focusedPanel = PanelEnvs
		}
	}
	return m, nil
}

func (m Model) handleGet() (tea.Model, tea.Cmd) {
	if m.focusedPanel != PanelKeys {
		return m, nil
	}
	// Reveal the current key's value in the preview pane
	curKey := m.keyPanel.CursorItem()
	if curKey == "" {
		return m, nil
	}
	m.previewShown = true
	if curKey != m.previewKey {
		return m, m.loadValue(m.currentFile(), curKey)
	}
	return m, nil
}

func (m Model) handleSelect() (tea.Model, tea.Cmd) {
	switch m.focusedPanel {
	case PanelScopes:
		m.scopePanel.Select()
		m.focusedPanel = PanelEnvs // advance to envs panel
		return m.selectScope(m.scopePanel.SelectedItem())
	case PanelEnvs:
		m.envPanel.Select()
		m.focusedPanel = PanelKeys // advance to keys panel
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
	file := m.currentFile()
	if file == "" || m.runner == nil {
		return m, nil
	}

	selectedKeys := m.keyPanel.SelectedItems()
	if len(selectedKeys) == 0 {
		return m, nil
	}

	runner := m.runner
	return m, func() tea.Msg {
		if len(selectedKeys) == 1 {
			raw, err := runner.GetValue(context.Background(), file, selectedKeys[0])
			if err != nil {
				return SetErrorMsg{Err: err}
			}
			text := string(raw)
			for i := range raw {
				raw[i] = 0
			}
			if err := clip.Write(text); err != nil {
				return SetErrorMsg{Err: err}
			}
			return CopyCompleteMsg{Key: selectedKeys[0]}
		}

		// Multi-copy
		kv, err := runner.GetAll(context.Background(), file)
		if err != nil {
			return SetErrorMsg{Err: err}
		}
		var lines []string
		for _, k := range selectedKeys {
			if v, ok := kv[k]; ok {
				lines = append(lines, fmt.Sprintf("%s=%s", k, string(v)))
			}
		}
		for _, v := range kv {
			for i := range v {
				v[i] = 0
			}
		}
		text := ""
		for i, l := range lines {
			if i > 0 {
				text += "\n"
			}
			text += l
		}
		if err := clip.Write(text); err != nil {
			return SetErrorMsg{Err: err}
		}
		return CopyMultiCompleteMsg{Count: len(lines)}
	}
}

func (m Model) openSetOverlay() (tea.Model, tea.Cmd) {
	file := m.currentFile()
	if file == "" || m.runner == nil {
		return m, nil
	}
	m.activeOverlay = OverlaySetValue
	keys := m.keyPanel.SelectedItems()
	existingKey := ""
	if len(keys) == 1 && m.keyPanel.SelectionCount() == 0 {
		existingKey = keys[0]
		keys = nil
	}
	cmd := m.setOverlay.Open(file, keys, existingKey, m.runner)
	return m, cmd
}

func (m Model) openDiffOverlay() (tea.Model, tea.Cmd) {
	scope := m.scopePanel.SelectedItem()
	env := m.envPanel.SelectedItem()
	envs := dotenvx.EnvsForScope(m.envFiles, scope)
	if len(envs) < 2 {
		m, cmd := setStatus(m,"Need at least 2 environments to diff", StatusWarning)
		return m, cmd
	}
	m.activeOverlay = OverlayDiff
	m.diffOverlay.Open(scope, env, envs, m.envFiles, m.runner)
	return m, nil
}

func (m Model) openImportOverlay() (tea.Model, tea.Cmd) {
	file := m.currentFile()
	if file == "" || m.runner == nil {
		return m, nil
	}
	m.activeOverlay = OverlayImport
	cmd := m.importOverlay.Open(m.targetDir, file, m.runner)
	return m, cmd
}

func (m Model) openExportOverlay() (tea.Model, tea.Cmd) {
	file := m.currentFile()
	if file == "" || m.runner == nil {
		return m, nil
	}
	m.activeOverlay = OverlayExport
	keys := m.keyPanel.SelectedItems()
	m.exportOverlay.Open(file, keys, m.runner)
	return m, nil
}

func (m Model) updateOverlay(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activeOverlay {
	case OverlaySetValue:
		cmd, handled := m.setOverlay.Update(msg)
		if !handled {
			break
		}
		if !m.setOverlay.Active {
			m.activeOverlay = OverlayNone
		}
		return m, cmd

	case OverlayDiff:
		cmd, handled := m.diffOverlay.Update(msg)
		if !handled {
			break
		}
		if !m.diffOverlay.Active {
			m.activeOverlay = OverlayNone
		}
		return m, cmd

	case OverlayImport:
		cmd, handled := m.importOverlay.Update(msg)
		if !handled {
			break
		}
		if !m.importOverlay.Active {
			m.activeOverlay = OverlayNone
		}
		return m, cmd

	case OverlayExport:
		cmd, handled := m.exportOverlay.Update(msg)
		if !handled {
			break
		}
		if !m.exportOverlay.Active {
			m.activeOverlay = OverlayNone
		}
		return m, cmd

	case OverlayHelp:
		if kmsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(kmsg, m.keyMap.Back) || key.Matches(kmsg, m.keyMap.Help) {
				m.activeOverlay = OverlayNone
				return m, nil
			}
		}
	}

	// Handle overlay result messages globally
	switch msg := msg.(type) {
	case overlays.SetDoneMsg:
		m.activeOverlay = OverlayNone
		m, cmd := setStatus(m,"Set "+msg.Key+" in "+msg.File, StatusSuccess)
		return m, tea.Batch(cmd, m.loadKeys(msg.File))

	case overlays.SetErrorMsg:
		m, cmd := setStatus(m,"Set failed: "+msg.Err.Error(), StatusError)
		return m, cmd

	case overlays.ImportDoneMsg:
		m.activeOverlay = OverlayNone
		m, cmd := setStatus(m,fmt.Sprintf("Imported %d keys into %s", msg.Count, msg.File), StatusSuccess)
		return m, tea.Batch(cmd, m.loadKeys(msg.File))

	case overlays.ImportErrorMsg:
		m, cmd := setStatus(m,"Import failed: "+msg.Err.Error(), StatusError)
		return m, cmd

	case overlays.ExportDoneMsg:
		m.activeOverlay = OverlayNone
		m, cmd := setStatus(m,fmt.Sprintf("Copied %d key-value pairs to clipboard", msg.Count), StatusSuccess)
		return m, cmd

	case overlays.ExportErrorMsg:
		m, cmd := setStatus(m,"Export failed: "+msg.Err.Error(), StatusError)
		return m, cmd
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
	return fmt.Sprintf("%d %ss", n, noun)
}
