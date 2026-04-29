package dotenvx

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create encrypted env files (with DOTENV_PUBLIC_KEY header)
	writeEncrypted(t, dir, ".env.local")
	writeEncrypted(t, dir, ".env.staging")
	writeEncrypted(t, dir, ".env.production")

	// Create a subdirectory with env files
	subDir := filepath.Join(dir, "apps", "api")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	writeEncrypted(t, subDir, ".env.local")
	writeEncrypted(t, subDir, ".env.production")

	// Files that should be excluded
	writePlain(t, dir, ".env.keys")
	writePlain(t, dir, ".env.vault")
	writePlain(t, dir, ".env.example")
	writePlain(t, dir, ".envrc")
	writePlain(t, dir, ".env.unencrypted") // no DOTENV_PUBLIC_KEY

	return dir
}

func writeEncrypted(t *testing.T, dir, name string) {
	t.Helper()
	content := `#/-------------------[DOTENV_PUBLIC_KEY]--------------------/
#/            public-key encryption for .env files          /
#/       [how it works](https://dotenvx.com/encryption)     /
#/----------------------------------------------------------/
DOTENV_PUBLIC_KEY="034a..."

# encrypted values
DATABASE_URL="encrypted:abc123"
`
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write encrypted file %s: %v", name, err)
	}
}

func writePlain(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("SOME_KEY=value\n"), 0o644); err != nil {
		t.Fatalf("write plain file %s: %v", name, err)
	}
}

func TestDiscover(t *testing.T) {
	dir := setupTestDir(t)

	files, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	// Should find 5 encrypted files: 3 in root + 2 in apps/api
	if len(files) != 5 {
		t.Errorf("Discover() found %d files, want 5", len(files))
		for _, f := range files {
			t.Logf("  %s (scope=%s, env=%s)", f.Path, f.Scope, f.Env)
		}
	}
}

func TestDiscoverScopes(t *testing.T) {
	dir := setupTestDir(t)

	files, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	scopes := Scopes(files)
	if len(scopes) != 2 {
		t.Errorf("Scopes() = %v, want 2 scopes", scopes)
	}
}

func TestEnvsForScope(t *testing.T) {
	dir := setupTestDir(t)

	files, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	envs := EnvsForScope(files, ".")
	if len(envs) != 3 {
		t.Errorf("EnvsForScope('.') = %v, want 3 envs", envs)
	}

	apiEnvs := EnvsForScope(files, filepath.Join("apps", "api"))
	if len(apiEnvs) != 2 {
		t.Errorf("EnvsForScope('apps/api') = %v, want 2 envs", apiEnvs)
	}
}

func TestFindFile(t *testing.T) {
	dir := setupTestDir(t)

	files, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	f, ok := FindFile(files, ".", "local")
	if !ok {
		t.Fatal("FindFile('.', 'local') not found")
	}
	if f.Env != "local" || f.Scope != "." {
		t.Errorf("FindFile() = %+v, want scope='.', env='local'", f)
	}

	_, ok = FindFile(files, ".", "nonexistent")
	if ok {
		t.Error("FindFile('.', 'nonexistent') should return false")
	}
}

func TestDiscoverEmpty(t *testing.T) {
	dir := t.TempDir()

	files, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Discover() on empty dir found %d files, want 0", len(files))
	}
}
