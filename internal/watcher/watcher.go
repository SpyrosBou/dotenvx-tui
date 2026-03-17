package watcher

import (
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

// StartWatching creates an fsnotify watcher on the given directories
// and returns a tea.Cmd that emits FileChangedMsg when .env files change.
// The watcher is also returned so it can be closed on program exit.
func StartWatching(dirs []string) (*fsnotify.Watcher, tea.Cmd, error) {
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
		var lastPath string

		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					close(events)
					return
				}
				name := filepath.Base(event.Name)
				if !strings.HasPrefix(name, ".env.") {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}

				lastPath = event.Name
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(100*time.Millisecond, func() {
					events <- lastPath
				})

			case _, ok := <-w.Errors:
				if !ok {
					close(events)
					return
				}
			}
		}
	}()

	// tea.Cmd that blocks on the channel
	cmd := watchCmd(events)

	return w, cmd, nil
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
