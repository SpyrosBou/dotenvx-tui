package dotenvx

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// excludedNames are filenames to skip entirely.
var excludedNames = map[string]bool{
	".env.keys":  true,
	".env.vault": true,
	".envrc":     true,
}

// Discover finds all dotenvx-encrypted .env.* files under targetDir.
// It filters out non-encrypted files, example files, key files, and vault files.
func Discover(targetDir string) ([]EnvFile, error) {
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, err
	}

	var files []EnvFile

	err = filepath.WalkDir(absTarget, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible directories
		}

		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()

		// Must be a .env.* file
		if !strings.HasPrefix(name, ".env.") {
			return nil
		}

		// Exclude known non-secret files
		if excludedNames[name] {
			return nil
		}

		// Exclude example/template files
		if strings.HasSuffix(name, ".example") || strings.HasSuffix(name, ".sample") {
			return nil
		}

		// Check for DOTENV_PUBLIC_KEY header (indicates dotenvx encryption)
		if !hasPublicKeyHeader(path) {
			return nil
		}

		// Derive scope and env
		relPath, err := filepath.Rel(absTarget, path)
		if err != nil {
			return nil
		}

		scope := filepath.Dir(relPath)
		env := strings.TrimPrefix(name, ".env.")

		files = append(files, EnvFile{
			Path:  relPath,
			Scope: scope,
			Env:   env,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Scope != files[j].Scope {
			return files[i].Scope < files[j].Scope
		}
		return files[i].Env < files[j].Env
	})

	return files, nil
}

// hasPublicKeyHeader checks if a file contains a DOTENV_PUBLIC_KEY line
// in its first 20 lines (dotenvx places it at the top).
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

// Scopes returns unique, sorted scope names from discovered files.
func Scopes(files []EnvFile) []string {
	seen := make(map[string]bool)
	var scopes []string
	for _, f := range files {
		if !seen[f.Scope] {
			seen[f.Scope] = true
			scopes = append(scopes, f.Scope)
		}
	}
	sort.Strings(scopes)
	return scopes
}

// EnvsForScope returns the environment names available in a given scope.
func EnvsForScope(files []EnvFile, scope string) []string {
	var envs []string
	for _, f := range files {
		if f.Scope == scope {
			envs = append(envs, f.Env)
		}
	}
	sort.Strings(envs)
	return envs
}

// FindFile looks up the EnvFile for a given scope and environment.
func FindFile(files []EnvFile, scope, env string) (EnvFile, bool) {
	for _, f := range files {
		if f.Scope == scope && f.Env == env {
			return f, true
		}
	}
	return EnvFile{}, false
}
