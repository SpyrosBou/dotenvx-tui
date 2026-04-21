package validate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var validKeyName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// KeyName validates that an environment variable key name is well-formed.
// Must start with a letter or underscore, contain only alphanumeric and underscores.
// Rejects the reserved DOTENV_ prefix.
func KeyName(name string) error {
	if name == "" {
		return fmt.Errorf("key name cannot be empty")
	}
	if !validKeyName.MatchString(name) {
		return fmt.Errorf("invalid key name %q: must contain only letters, digits, and underscores, starting with a letter or underscore", name)
	}
	if strings.HasPrefix(name, "DOTENV_") {
		return fmt.Errorf("key name %q uses reserved prefix DOTENV_", name)
	}
	return nil
}

// FilePath validates that path resolves within targetDir.
// Prevents path traversal attacks and symlink escapes.
func FilePath(targetDir, path string) error {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("cannot resolve target directory: %w", err)
	}

	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(absTarget, candidate)
	}

	abs, err := filepath.Abs(candidate)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Ensure the resolved path is within the target directory
	if !strings.HasPrefix(abs, absTarget+string(filepath.Separator)) && abs != absTarget {
		return fmt.Errorf("path %q is outside the target directory", path)
	}

	// Check for symlink escape
	real, err := filepath.EvalSymlinks(abs)
	if err == nil {
		realTarget, err2 := filepath.EvalSymlinks(absTarget)
		if err2 == nil {
			if !strings.HasPrefix(real, realTarget+string(filepath.Separator)) && real != realTarget {
				return fmt.Errorf("path %q resolves outside target directory via symlink", path)
			}
		}
	}

	return nil
}
