package overlays

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/warui1/dotenvx-tui/internal/dotenvx"
	"github.com/warui1/dotenvx-tui/internal/theme"
)

// DiffStep tracks the current step in the diff flow.
type DiffStep int

const (
	DiffStepPickEnv DiffStep = iota
	DiffStepShowResult
)

// DiffStatus categorizes a key in the diff.
type DiffStatus int

const (
	DiffOnlyLeft DiffStatus = iota
	DiffOnlyRight
	DiffDifferent
	DiffIdentical
)

// DiffRow represents one row in the diff output.
type DiffRow struct {
	Key    string
	Status DiffStatus
}

// DiffOverlay shows a comparison between two environment files.
type DiffOverlay struct {
	Active bool
	Step   DiffStep

	// Env picker
	Envs     []string
	Cursor   int
	LeftEnv  string
	RightEnv string
	Scope    string

	// Results
	Rows      []DiffRow
	SameCount int
	ScrollY   int
	Error     string

	Runner *dotenvx.Runner
	Files  []dotenvx.EnvFile
	Styles theme.Styles
}

// NewDiffOverlay creates a new diff overlay.
func NewDiffOverlay(styles theme.Styles) DiffOverlay {
	return DiffOverlay{Styles: styles}
}

// Open activates the overlay for diffing.
func (o *DiffOverlay) Open(scope, currentEnv string, envs []string, files []dotenvx.EnvFile, runner *dotenvx.Runner) {
	o.Active = true
	o.Step = DiffStepPickEnv
	o.Scope = scope
	o.LeftEnv = currentEnv
	o.Files = files
	o.Runner = runner
	o.Cursor = 0
	o.Rows = nil
	o.ScrollY = 0
	o.Error = ""

	// Filter out the current env from choices
	var choices []string
	for _, e := range envs {
		if e != currentEnv {
			choices = append(choices, e)
		}
	}
	o.Envs = choices
}

// Close deactivates the overlay.
func (o *DiffOverlay) Close() {
	o.Active = false
	o.Rows = nil
}

// Update handles input for the diff overlay.
func (o *DiffOverlay) Update(msg tea.Msg) (tea.Cmd, bool) {
	if !o.Active {
		return nil, false
	}

	switch msg := msg.(type) {
	case diffResultMsg:
		if msg.Err != nil {
			o.Rows = nil
			o.SameCount = 0
			o.Error = msg.Err.Error()
			o.Step = DiffStepShowResult
			o.ScrollY = 0
			return nil, true
		}
		o.Rows = msg.Rows
		o.SameCount = msg.SameCount
		o.Error = ""
		o.Step = DiffStepShowResult
		o.ScrollY = 0
		return nil, true

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("escape"))):
			if o.Step == DiffStepShowResult {
				o.Step = DiffStepPickEnv
				return nil, true
			}
			o.Close()
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if o.Step == DiffStepPickEnv {
				o.Cursor--
				if o.Cursor < 0 {
					o.Cursor = len(o.Envs) - 1
				}
			} else {
				if o.ScrollY > 0 {
					o.ScrollY--
				}
			}
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if o.Step == DiffStepPickEnv {
				o.Cursor++
				if o.Cursor >= len(o.Envs) {
					o.Cursor = 0
				}
			} else {
				maxScroll := max(0, len(o.Rows)-10)
				if o.ScrollY < maxScroll {
					o.ScrollY++
				}
			}
			return nil, true

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if o.Step == DiffStepPickEnv && len(o.Envs) > 0 {
				o.RightEnv = o.Envs[o.Cursor]
				return o.computeDiff(), true
			}
		}
	}

	return nil, false
}

func (o *DiffOverlay) computeDiff() tea.Cmd {
	scope := o.Scope
	leftEnv := o.LeftEnv
	rightEnv := o.RightEnv
	files := o.Files
	runner := o.Runner

	leftFile, ok1 := dotenvx.FindFile(files, scope, leftEnv)
	rightFile, ok2 := dotenvx.FindFile(files, scope, rightEnv)
	if !ok1 || !ok2 || runner == nil {
		return nil
	}

	return func() tea.Msg {
		leftKV, err := runner.GetAll(context.Background(), leftFile.Path)
		if err != nil {
			return diffResultMsg{Err: fmt.Errorf("failed to decrypt %s: %w", leftFile.Path, err)}
		}
		rightKV, err := runner.GetAll(context.Background(), rightFile.Path)
		if err != nil {
			return diffResultMsg{Err: fmt.Errorf("failed to decrypt %s: %w", rightFile.Path, err)}
		}

		// Collect all keys
		allKeys := make(map[string]bool)
		for k := range leftKV {
			allKeys[k] = true
		}
		for k := range rightKV {
			allKeys[k] = true
		}

		var sorted []string
		for k := range allKeys {
			sorted = append(sorted, k)
		}
		sort.Strings(sorted)

		var rows []DiffRow
		sameCount := 0

		for _, k := range sorted {
			lv, inLeft := leftKV[k]
			rv, inRight := rightKV[k]

			var status DiffStatus
			switch {
			case inLeft && !inRight:
				status = DiffOnlyLeft
			case !inLeft && inRight:
				status = DiffOnlyRight
			case string(lv) != string(rv):
				status = DiffDifferent
			default:
				status = DiffIdentical
				sameCount++
			}

			rows = append(rows, DiffRow{Key: k, Status: status})
		}

		// Zero all values
		for _, v := range leftKV {
			for i := range v {
				v[i] = 0
			}
		}
		for _, v := range rightKV {
			for i := range v {
				v[i] = 0
			}
		}

		return diffResultMsg{Rows: rows, SameCount: sameCount}
	}
}

// View renders the diff overlay.
func (o *DiffOverlay) View(width int) string {
	var b strings.Builder

	if o.Step == DiffStepPickEnv {
		b.WriteString(o.Styles.OverlayTitle.Render(fmt.Sprintf("Diff: %s vs ?", o.LeftEnv)))
		b.WriteString("\n\nCompare with:\n\n")

		for i, env := range o.Envs {
			if i == o.Cursor {
				b.WriteString("  " + o.Styles.Cursor.Render(" "+env+" ") + "\n")
			} else {
				b.WriteString("  " + o.Styles.InactiveItem.Render(env) + "\n")
			}
		}

		b.WriteString("\n" + o.Styles.HelpBar.Render("enter: select  esc: cancel"))
	} else {
		b.WriteString(o.Styles.OverlayTitle.Render(fmt.Sprintf("Diff: %s vs %s", o.LeftEnv, o.RightEnv)))
		b.WriteString("\n\n")

		if o.Error != "" {
			b.WriteString(o.Styles.StatusError.Render(o.Error))
			b.WriteString("\n\n" + o.Styles.HelpBar.Render("esc: back"))
			return o.Styles.Overlay.
				Width(min(65, width-4)).
				Render(b.String())
		}

		// Show non-identical rows first, then identical count
		visible := o.Rows
		end := min(o.ScrollY+15, len(visible))
		start := o.ScrollY

		for _, row := range visible[start:end] {
			var line string
			switch row.Status {
			case DiffOnlyLeft:
				line = o.Styles.DiffRemove.Render(fmt.Sprintf("  - %-30s  only in %s", row.Key, o.LeftEnv))
			case DiffOnlyRight:
				line = o.Styles.DiffAdd.Render(fmt.Sprintf("  + %-30s  only in %s", row.Key, o.RightEnv))
			case DiffDifferent:
				line = o.Styles.DiffChanged.Render(fmt.Sprintf("  ~ %-30s  values differ", row.Key))
			case DiffIdentical:
				line = o.Styles.DiffSame.Render(fmt.Sprintf("  = %-30s  identical", row.Key))
			}
			b.WriteString(line + "\n")
		}

		if o.SameCount > 0 {
			b.WriteString(fmt.Sprintf("\n  %d identical keys\n", o.SameCount))
		}

		b.WriteString("\n" + o.Styles.HelpBar.Render("j/k: scroll  esc: back"))
	}

	return o.Styles.Overlay.
		Width(min(65, width-4)).
		Render(b.String())
}

type diffResultMsg struct {
	Rows      []DiffRow
	SameCount int
	Err       error
}
