package clipboard

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// Write copies text to the system clipboard.
func Write(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("clipboard write failed: %w", err)
	}
	return nil
}
