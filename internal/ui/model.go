package ui

import (
	"context"
	"io/fs"
	"path/filepath"
	"sort"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/SpyrosBou/dotenvx-tui/internal/dotenvx"
	"github.com/SpyrosBou/dotenvx-tui/internal/secret"
	"github.com/SpyrosBou/dotenvx-tui/internal/theme"
	"github.com/SpyrosBou/dotenvx-tui/internal/ui/overlays"
	"github.com/SpyrosBou/dotenvx-tui/internal/ui/panels"
	"github.com/SpyrosBou/dotenvx-tui/internal/watcher"
)

// Model is the root Bubbletea model for the application.
type Model struct {
	// Terminal dimensions
	width  int
	height int

	// Theme detection
	hasDarkBG bool
	theme     theme.Theme
	styles    theme.Styles

	// Layout
	layout Layout

	// Target directory
	targetDir string

	// Discovered env files
	envFiles []dotenvx.EnvFile

	// Panels
	scopePanel panels.ScopePanel
	envPanel   panels.EnvPanel
	keyPanel   panels.KeyListPanel

	// Focus
	focusedPanel PanelID

	// Preview
	previewKey    string
	previewValue  *secret.SecureBytes
	previewShown  bool // true = revealed, false = masked
	previewMaskID int

	// Status
	statusMsg   string
	statusLevel StatusLevel
	statusID    int

	// Overlays
	activeOverlay OverlayKind
	setOverlay    overlays.SetValueOverlay
	diffOverlay   overlays.DiffOverlay
	importOverlay overlays.ImportOverlay
	exportOverlay overlays.ExportOverlay
	deleteOverlay overlays.DeleteOverlay

	// Key bindings
	keyMap KeyMap

	// Runner
	runner *dotenvx.Runner

	// File watching
	fileWatcher *watcher.Watcher

	// Loading state
	loading    bool
	loadingMsg string

	// Error state (fatal, shown on startup)
	fatalErr string
}

// NewModel creates the initial application model.
func NewModel(targetDir string) Model {
	th := theme.NewTheme(true) // default to dark, updated on BackgroundColorMsg
	styles := theme.NewStyles(th)
	return Model{
		targetDir:     targetDir,
		hasDarkBG:     true,
		theme:         th,
		styles:        styles,
		keyMap:        DefaultKeyMap(),
		keyPanel:      panels.NewKeyListPanel(),
		focusedPanel:  PanelScopes,
		setOverlay:    overlays.NewSetValueOverlay(styles),
		diffOverlay:   overlays.NewDiffOverlay(styles),
		importOverlay: overlays.NewImportOverlay(styles),
		exportOverlay: overlays.NewExportOverlay(styles),
		deleteOverlay: overlays.NewDeleteOverlay(styles),
	}
}

// Init initializes the model, starting file discovery.
func (m Model) Init() tea.Cmd {
	return m.discoverFiles()
}

func (m Model) discoverFiles() tea.Cmd {
	targetDir := m.targetDir
	return func() tea.Msg {
		files, err := dotenvx.Discover(targetDir)
		if err != nil {
			return DiscoveryErrorMsg{Err: err}
		}
		return FilesDiscoveredMsg{Files: files}
	}
}

func setStatus(m Model, msg string, level StatusLevel) (Model, tea.Cmd) {
	m.statusID++
	m.statusMsg = msg
	m.statusLevel = level
	id := m.statusID
	return m, tea.Tick(4*time.Second, func(_ time.Time) tea.Msg {
		return ClearStatusMsg{ID: id}
	})
}

func (m Model) loadKeys(file string) tea.Cmd {
	runner := m.runner
	if runner == nil {
		return nil
	}
	return func() tea.Msg {
		keys, err := runner.GetKeys(context.Background(), file)
		if err != nil {
			return KeysLoadErrorMsg{Err: err}
		}
		return KeysLoadedMsg{Keys: keys}
	}
}

func (m Model) loadValue(file, key string) tea.Cmd {
	runner := m.runner
	if runner == nil {
		return nil
	}
	return func() tea.Msg {
		raw, err := runner.GetValue(context.Background(), file, key)
		if err != nil {
			return ValueLoadErrorMsg{Key: key, Err: err}
		}
		sec := secret.New(raw)
		return ValueLoadedMsg{Key: key, Value: sec}
	}
}

// currentFile returns the currently selected env file path, if any.
func (m Model) currentFile() string {
	scope := m.scopePanel.SelectedItem()
	env := m.envPanel.SelectedItem()
	if scope == "" || env == "" {
		return ""
	}
	f, ok := dotenvx.FindFile(m.envFiles, scope, env)
	if !ok {
		return ""
	}
	return f.Path
}

// cleanup zeroes all sensitive data before shutdown.
func (m *Model) cleanup() {
	if m.previewValue != nil {
		m.previewValue.Clear()
		m.previewValue = nil
	}
	if m.fileWatcher != nil {
		_ = m.fileWatcher.Close()
		m.fileWatcher = nil
	}
	m.setOverlay.Close()
}

func (m Model) watchDirs() []string {
	dirs := make(map[string]struct{})
	add := func(dir string) {
		if dir == "" {
			return
		}
		dirs[filepath.Clean(dir)] = struct{}{}
	}

	add(m.targetDir)
	_ = filepath.WalkDir(m.targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" {
				return filepath.SkipDir
			}
			add(path)
		}
		return nil
	})

	out := make([]string, 0, len(dirs))
	for dir := range dirs {
		out = append(out, dir)
	}
	sort.Strings(out)
	return out
}

func (m Model) startFileWatcher() (Model, tea.Cmd, error) {
	dirs := m.watchDirs()
	if len(dirs) == 0 {
		return m, nil, nil
	}

	fileWatcher, cmd, err := watcher.StartWatching(dirs)
	if err != nil {
		return m, nil, err
	}

	if m.fileWatcher != nil {
		_ = m.fileWatcher.Close()
	}
	m.fileWatcher = fileWatcher
	return m, cmd, nil
}
