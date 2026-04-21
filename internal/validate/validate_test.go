package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeyNameValid(t *testing.T) {
	valid := []string{
		"DATABASE_URL",
		"API_KEY",
		"_PRIVATE",
		"a",
		"A",
		"MY_VAR_123",
	}
	for _, name := range valid {
		if err := KeyName(name); err != nil {
			t.Errorf("KeyName(%q) should be valid, got error: %v", name, err)
		}
	}
}

func TestKeyNameInvalid(t *testing.T) {
	invalid := []string{
		"",
		"123_START",
		"has-dash",
		"has space",
		"has.dot",
		"key=value",
		"$VAR",
	}
	for _, name := range invalid {
		if err := KeyName(name); err == nil {
			t.Errorf("KeyName(%q) should be invalid, got nil", name)
		}
	}
}

func TestKeyNameReservedPrefix(t *testing.T) {
	reserved := []string{
		"DOTENV_PUBLIC_KEY",
		"DOTENV_PRIVATE_KEY",
		"DOTENV_ANYTHING",
	}
	for _, name := range reserved {
		err := KeyName(name)
		if err == nil {
			t.Errorf("KeyName(%q) should reject DOTENV_ prefix, got nil", name)
		}
	}
}

func TestFilePathAllowsRelativePathInsideTargetDir(t *testing.T) {
	targetDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(targetDir, ".env.local"), []byte("KEY=value\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	otherDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(otherDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	}()

	if err := FilePath(targetDir, ".env.local"); err != nil {
		t.Fatalf("FilePath rejected relative target path: %v", err)
	}
}

func TestFilePathRejectsRelativeTraversalOutsideTargetDir(t *testing.T) {
	targetDir := t.TempDir()

	if err := FilePath(targetDir, "../outside.env"); err == nil {
		t.Fatal("FilePath accepted traversal outside target dir")
	}
}
