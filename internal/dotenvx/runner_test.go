package dotenvx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetStagesPlaintextWithoutPassingSecretAsDotenvxArg(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, ".env.local")
	if err := os.WriteFile(targetPath, []byte("ORIGINAL_ENCRYPTED\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.keys"), []byte("DOTENV_PRIVATE_KEY_LOCAL=x\n"), 0o600); err != nil {
		t.Fatalf("write keys file: %v", err)
	}

	argsPath := filepath.Join(dir, "args.txt")
	stagePath := filepath.Join(dir, "stage.txt")
	fakeDotenvx := filepath.Join(dir, "dotenvx")
	script := "#!/bin/sh\n" +
		"printf '%s\\0' \"$@\" >> " + shellQuote(argsPath) + "\n" +
		"printf '\\n' >> " + shellQuote(argsPath) + "\n" +
		"if [ \"$1\" = decrypt ]; then\n" +
		"  printf 'DOTENV_PUBLIC_KEY_LOCAL=\"x\"\\nFOO=\"old\"\\n'\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = encrypt ]; then\n" +
		"  env_file=''\n" +
		"  prev=''\n" +
		"for arg do\n" +
		"    if [ \"$prev\" = -f ]; then env_file=\"$arg\"; fi\n" +
		"    prev=\"$arg\"\n" +
		"  done\n" +
		"  cat \"$env_file\" > " + shellQuote(stagePath) + "\n" +
		"  printf 'ENCRYPTED\\n'\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 64\n"
	if err := os.WriteFile(fakeDotenvx, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake dotenvx: %v", err)
	}

	runner := &Runner{binary: fakeDotenvx, workDir: dir}
	secret := "--help$TOKEN"
	if err := runner.Set(context.Background(), ".env.local", "FOO", []byte(secret)); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read args: %v", err)
	}
	args := string(argsData)
	if strings.Contains(args, "set") {
		t.Fatalf("dotenvx set was called; args log: %q", args)
	}
	if strings.Contains(args, secret) {
		t.Fatalf("secret appeared in dotenvx argv: %q", args)
	}

	stageData, err := os.ReadFile(stagePath)
	if err != nil {
		t.Fatalf("read staged plaintext: %v", err)
	}
	if got, want := string(stageData), "DOTENV_PUBLIC_KEY_LOCAL=\"x\"\nFOO=`--help\\$TOKEN`\n"; got != want {
		t.Fatalf("staged plaintext = %q, want %q", got, want)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if got, want := string(targetData), "ENCRYPTED\n"; got != want {
		t.Fatalf("target = %q, want %q", got, want)
	}
}

func TestUnsetDoesNotReplaceTargetWithPlaintextWhenEncryptFails(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, ".env.local")
	original := []byte("ORIGINAL_ENCRYPTED\n")
	if err := os.WriteFile(targetPath, original, 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.keys"), []byte("DOTENV_PRIVATE_KEY_LOCAL=x\n"), 0o600); err != nil {
		t.Fatalf("write keys file: %v", err)
	}

	fakeDotenvx := filepath.Join(dir, "dotenvx")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = decrypt ]; then\n" +
		"  printf 'DOTENV_PUBLIC_KEY_LOCAL=\"x\"\\nFOO=\"secret\"\\nBAR=\"keep\"\\n'\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [ \"$1\" = encrypt ]; then\n" +
		"  printf 'encrypt failed\\n' >&2\n" +
		"  exit 1\n" +
		"fi\n" +
		"exit 64\n"
	if err := os.WriteFile(fakeDotenvx, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake dotenvx: %v", err)
	}

	runner := &Runner{binary: fakeDotenvx, workDir: dir}
	if err := runner.Unset(context.Background(), ".env.local", []string{"FOO"}); err == nil {
		t.Fatal("Unset returned nil error")
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(targetData) != string(original) {
		t.Fatalf("target changed after failed encrypt: %q", string(targetData))
	}
}

func TestRunnerSetUnsetWithDotenvxCLI(t *testing.T) {
	if _, err := exec.LookPath("dotenvx"); err != nil {
		t.Skip("dotenvx CLI not installed")
	}

	dir := t.TempDir()
	seed := exec.Command("dotenvx", "set", "FOO", "old", "-f", ".env.local")
	seed.Dir = dir
	if out, err := seed.CombinedOutput(); err != nil {
		t.Fatalf("seed dotenvx file: %v\n%s", err, out)
	}

	runner, err := NewRunner(dir)
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}

	want := `--help$TOKEN "quoted" #hash \slash`
	if err := runner.Set(context.Background(), ".env.local", "FOO", []byte(want)); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := runner.GetValue(context.Background(), ".env.local", "FOO")
	if err != nil {
		t.Fatalf("GetValue after Set: %v", err)
	}
	defer zeroBytes(got)
	if string(got) != want {
		t.Fatalf("GetValue = %q, want %q", string(got), want)
	}

	if err := runner.Set(context.Background(), ".env.local", "BAR", []byte("keep")); err != nil {
		t.Fatalf("Set BAR: %v", err)
	}
	if err := runner.Unset(context.Background(), ".env.local", []string{"FOO"}); err != nil {
		t.Fatalf("Unset: %v", err)
	}

	keys, err := runner.GetKeys(context.Background(), ".env.local")
	if err != nil {
		t.Fatalf("GetKeys after Unset: %v", err)
	}
	if strings.Join(keys, ",") != "BAR" {
		t.Fatalf("keys = %#v, want only BAR", keys)
	}
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
