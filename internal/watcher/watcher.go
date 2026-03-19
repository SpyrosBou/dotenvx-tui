package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/fsnotify/fsnotify"
)

// FileChangedMsg is sent when a .env file changes on disk.
type FileChangedMsg struct {
	Path string
}

// Watcher wraps an fsnotify watcher and the debounced event channel used by Bubble Tea.
type Watcher struct {
	w      *fsnotify.Watcher
	events <-chan string
}

// StartWatching creates an fsnotify watcher on the given directories
// and returns a tea.Cmd that emits FileChangedMsg when .env files change.
// The watcher is also returned so it can be closed on program exit.
func StartWatching(dirs []string) (*Watcher, tea.Cmd, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	for _, dir := range dirs {
		if err := w.Add(dir); err != nil {
			w.Close()
			return nil, nil, err
		}
	}

	events := make(chan string, 10)

	// Goroutine that debounces fsnotify events and sends to channel
	go func() {
		var debounceTimer *time.Timer
		var debounceC <-chan time.Time
		var lastPath string

		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					if debounceTimer != nil {
						if !debounceTimer.Stop() {
							select {
							case <-debounceTimer.C:
							default:
							}
						}
					}
					close(events)
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
					continue
				}

				name := filepath.Base(event.Name)
				if !strings.HasPrefix(name, ".env.") {
					if info, err := os.Stat(event.Name); err != nil || !info.IsDir() {
						continue
					}
				}

				if debounceTimer != nil {
					if !debounceTimer.Stop() {
						select {
						case <-debounceTimer.C:
						default:
						}
					}
				}
				lastPath = event.Name
				debounceTimer = time.NewTimer(100 * time.Millisecond)
				debounceC = debounceTimer.C

			case <-debounceC:
				debounceTimer = nil
				debounceC = nil
				events <- lastPath

			case _, ok := <-w.Errors:
				if !ok {
					if debounceTimer != nil {
						if !debounceTimer.Stop() {
							select {
							case <-debounceTimer.C:
							default:
							}
						}
					}
					close(events)
					return
				}
			}
		}
	}()

	return &Watcher{w: w, events: events}, watchCmd(events), nil
}

// Cmd returns a tea.Cmd that waits for the next file change event.
func (w *Watcher) Cmd() tea.Cmd {
	if w == nil || w.events == nil {
		return nil
	}
	return watchCmd(w.events)
}

// Close stops the underlying fsnotify watcher.
func (w *Watcher) Close() error {
	if w == nil || w.w == nil {
		return nil
	}
	return w.w.Close()
}

// watchCmd returns a tea.Cmd that waits for a file change event.
func watchCmd(events <-chan string) tea.Cmd {
	return func() tea.Msg {
		path, ok := <-events
		if !ok {
			return nil
		}
		return FileChangedMsg{Path: path}
	}
}

// ContinueWatching returns a new tea.Cmd that waits for the next event.
// Call this after processing a FileChangedMsg to keep listening.
func ContinueWatching(events <-chan string) tea.Cmd {
	return watchCmd(events)
}
