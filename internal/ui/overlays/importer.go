package overlays

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/warui1/dotenvx-tui/internal/dotenvx"
	"github.com/warui1/dotenvx-tui/internal/theme"
)

// ImportStep tracks the current step in the import flow.
type ImportStep int

const (
	ImportStepPickFile ImportStep = iota
	ImportStepSelectKeys
	ImportStepDone
)

// ImportKey represents a key to import with selection state.
type ImportKey struct {
	Name     string
	Value    string
	Selected bool
}

// ImportOverlay handles importing keys from plaintext env files.
type ImportOverlay struct {
	Active bool
	Step   ImportStep

	// File picker
	Files  []string
	Cursor int

	// Key selection
	Keys       []ImportKey
	KeyCursor  int

	// Target
	TargetFile string
	TargetDir  string
	Runner     *dotenvx.Runner
	Styles     theme.Styles
}

// NewImportOverlay creates a new import overlay.
func NewImportOverlay(styles theme.Styles) ImportOverlay {
	return ImportOverlay{Styles: styles}
}

// Open activates the import overlay.
func (o *ImportOverlay) Open(targetDir, targetFile string, runner *dotenvx.Runner) tea.Cmd {
	o.Active = true
	o.Step = ImportStepPickFile
	o.TargetFile = targetFile
	o.TargetDir = targetDir
	o.Runner = runner
	o.Cursor = 0
	o.Keys = nil

	return o.findPlaintextFiles()
}

// Close deactivates the overlay.
func (o *ImportOverlay) Close() {
	o.Active = false
	o.Files = nil
	o.Keys = nil
}

// Update handles input for the import overlay.
func (o *ImportOverlay) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !o.Active {
		return nil, false
	}

	switch msg := msg.(type) {
	case importFilesFoundMsg:
		o.Files = msg.Files
		if len(o.Files) == 0 {
			return nil, true // will show "no files" in view
		}
		return nil, true

	case importKeysLoadedMsg:
		o.Keys = msg.Keys
		o.KeyCursor = 0
		o.Step = ImportStepSelectKeys
		return nil, true

	case ImportDoneMsg:
		o.Close()
		return nil, true

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("escape"))):
			o.Close()
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if o.Step == ImportStepPickFile {
				o.Cursor = (o.Cursor - 1 + len(o.Files)) % max(1, len(o.Files))
			} else if o.Step == ImportStepSelectKeys {
				o.KeyCursor = (o.KeyCursor - 1 + len(o.Keys)) % max(1, len(o.Keys))
			}
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if o.Step == ImportStepPickFile {
				o.Cursor = (o.Cursor + 1) % max(1, len(o.Files))
			} else if o.Step == ImportStepSelectKeys {
				o.KeyCursor = (o.KeyCursor + 1) % max(1, len(o.Keys))
			}
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("space"))):
			if o.Step == ImportStepSelectKeys && len(o.Keys) > 0 {
				o.Keys[o.KeyCursor].Selected = !o.Keys[o.KeyCursor].Selected
			}
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return o.handleEnter()
		}
	}

	return nil, false
}

func (o *ImportOverlay) handleEnter() (tea.Cmd, bool) {
	if o.Step == ImportStepPickFile && len(o.Files) > 0 {
		return o.loadKeysFromFile(o.Files[o.Cursor]), true
	}

	if o.Step == ImportStepSelectKeys {
		return o.executeImport(), true
	}

	return nil, true
}

func (o *ImportOverlay) findPlaintextFiles() tea.Cmd {
	targetDir := o.TargetDir
	return func() tea.Msg {
		var files []string
		err := filepath.WalkDir(targetDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			name := d.Name()
			if !strings.HasPrefix(name, ".env") {
				return nil
			}
			// Skip encrypted files and non-importable files
			if name == ".env.keys" || name == ".env.vault" || name == ".envrc" {
				return nil
			}
			if strings.HasSuffix(name, ".example") || strings.HasSuffix(name, ".sample") {
				return nil
			}
			// Check if it's NOT encrypted (no DOTENV_PUBLIC_KEY)
			if hasPublicKeyHeader(path) {
				return nil
			}
			rel, _ := filepath.Rel(targetDir, path)
			files = append(files, rel)
			return nil
		})
		if err != nil {
			return importFilesFoundMsg{Files: nil}
		}
		return importFilesFoundMsg{Files: files}
	}
}

func hasPublicKeyHeader(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for i := 0; i < 20 && scanner.Scan(); i++ {
		if strings.Contains(scanner.Text(), "DOTENV_PUBLIC_KEY") {
			return true
		}
	}
	return false
}

func (o *ImportOverlay) loadKeysFromFile(file string) tea.Cmd {
	targetDir := o.TargetDir
	return func() tea.Msg {
		path := filepath.Join(targetDir, file)
		f, err := os.Open(path)
		if err != nil {
			return importKeysLoadedMsg{Keys: nil}
		}
		defer f.Close()

		var keys []ImportKey
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes
			value = strings.Trim(value, `"'`)
			keys = append(keys, ImportKey{Name: name, Value: value, Selected: true})
		}
		return importKeysLoadedMsg{Keys: keys}
	}
}

func (o *ImportOverlay) executeImport() tea.Cmd {
	runner := o.Runner
	file := o.TargetFile
	var toImport []ImportKey
	for _, k := range o.Keys {
		if k.Selected {
			toImport = append(toImport, k)
		}
	}

	if len(toImport) == 0 || runner == nil {
		return nil
	}

	return func() tea.Msg {
		count := 0
		for _, k := range toImport {
			err := runner.Set(context.Background(), file, k.Name, []byte(k.Value))
			if err != nil {
				return ImportErrorMsg{Err: fmt.Errorf("failed to import %s: %w", k.Name, err)}
			}
			count++
		}
		return ImportDoneMsg{Count: count, File: file}
	}
}

// View renders the import overlay.
func (o *ImportOverlay) View(width int) string {
	var b strings.Builder

	b.WriteString(o.Styles.OverlayTitle.Render("Import from plaintext file"))
	b.WriteString("\n\n")

	if o.Step == ImportStepPickFile {
		if len(o.Files) == 0 {
			b.WriteString(o.Styles.InactiveItem.Render("No unencrypted .env files found to import."))
			b.WriteString("\n\n" + o.Styles.HelpBar.Render("esc: close"))
		} else {
			b.WriteString("Select file to import from:\n\n")
			for i, f := range o.Files {
				if i == o.Cursor {
					b.WriteString("  " + o.Styles.Cursor.Render(" "+f+" ") + "\n")
				} else {
					b.WriteString("  " + o.Styles.InactiveItem.Render(f) + "\n")
				}
			}
			b.WriteString("\n" + o.Styles.HelpBar.Render("enter: select  esc: cancel"))
		}
	} else if o.Step == ImportStepSelectKeys {
		b.WriteString(fmt.Sprintf("Keys to import into %s:\n\n", o.TargetFile))
		for i, k := range o.Keys {
			marker := "[ ]"
			if k.Selected {
				marker = "[x]"
			}
			if i == o.KeyCursor {
				b.WriteString("  " + o.Styles.Cursor.Render(fmt.Sprintf(" %s %s ", marker, k.Name)) + "\n")
			} else if k.Selected {
				b.WriteString(fmt.Sprintf("  %s %s\n", o.Styles.ActiveItem.Render(marker), o.Styles.ActiveItem.Render(k.Name)))
			} else {
				b.WriteString(fmt.Sprintf("  %s %s\n", o.Styles.InactiveItem.Render(marker), o.Styles.InactiveItem.Render(k.Name)))
			}
		}
		b.WriteString("\n" + o.Styles.HelpBar.Render("space: toggle  enter: import  esc: cancel"))
	}

	return o.Styles.Overlay.
		Width(min(55, width-4)).
		Render(b.String())
}

// Messages.
type importFilesFoundMsg struct{ Files []string }
type importKeysLoadedMsg struct{ Keys []ImportKey }

type ImportDoneMsg struct {
	Count int
	File  string
}

type ImportErrorMsg struct {
	Err error
}
