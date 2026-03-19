package dotenvx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/warui1/dotenvx-tui/internal/validate"
)

const defaultTimeout = 30 * time.Second

// Runner wraps all dotenvx CLI calls with proper security practices.
type Runner struct {
	binary  string // resolved path to dotenvx binary
	workDir string
}

// NewRunner creates a Runner, verifying that dotenvx is available in PATH.
func NewRunner(workDir string) (*Runner, error) {
	path, err := exec.LookPath("dotenvx")
	if err != nil {
		return nil, fmt.Errorf("dotenvx not found in PATH: install with: brew install dotenvx/brew/dotenvx")
	}
	return &Runner{binary: path, workDir: workDir}, nil
}

// GetKeys returns the sorted key names (excluding DOTENV_PUBLIC_KEY) from an encrypted env file.
func (r *Runner) GetKeys(ctx context.Context, file string) ([]string, error) {
	if err := validate.FilePath(r.workDir, file); err != nil {
		return nil, err
	}

	kv, err := r.GetAll(ctx, file)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// GetValue decrypts and returns the raw bytes of a single key.
// The caller should wrap the result in secret.SecureBytes immediately.
func (r *Runner) GetValue(ctx context.Context, file, key string) ([]byte, error) {
	if err := validate.FilePath(r.workDir, file); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.binary, "get", key, "-f", file)
	cmd.Dir = r.workDir
	cmd.Env = minimalEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("dotenvx get failed: %s", strings.TrimSpace(stderr.String()))
	}

	out := bytes.TrimRight(stdout.Bytes(), "\n")
	result := make([]byte, len(out))
	copy(result, out)

	// Zero the buffer
	for i := range stdout.Bytes() {
		stdout.Bytes()[i] = 0
	}

	return result, nil
}

// GetAll decrypts all key-value pairs from an env file.
// Returns a map of key names to raw byte values. DOTENV_PUBLIC_KEY is excluded.
// The caller should wrap values in secret.SecureBytes and zero the map values after use.
func (r *Runner) GetAll(ctx context.Context, file string) (map[string][]byte, error) {
	if err := validate.FilePath(r.workDir, file); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.binary, "get", "-f", file)
	cmd.Dir = r.workDir
	cmd.Env = minimalEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("dotenvx get failed: %s", strings.TrimSpace(stderr.String()))
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse dotenvx output: %w", err)
	}

	result := make(map[string][]byte, len(raw))
	for k, v := range raw {
		if strings.HasPrefix(k, "DOTENV_") {
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			continue
		}
		result[k] = []byte(s)
	}

	// Zero the original buffer
	for i := range stdout.Bytes() {
		stdout.Bytes()[i] = 0
	}

	return result, nil
}

// Set encrypts and stores a value in an env file.
func (r *Runner) Set(ctx context.Context, file, key string, value []byte) error {
	if err := validate.FilePath(r.workDir, file); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.binary, "set", key, string(value), "-f", file)
	cmd.Dir = r.workDir
	cmd.Env = minimalEnv()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotenvx set failed: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

// Unset removes one or more keys from an encrypted env file.
// It decrypts the file, removes matching lines, then re-encrypts.
func (r *Runner) Unset(ctx context.Context, file string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := validate.FilePath(r.workDir, file); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	// Step 1: Decrypt the file in-place
	if err := r.runCmd(ctx, "decrypt", "-f", file); err != nil {
		return fmt.Errorf("decrypt failed: %w", err)
	}

	// From this point, if anything fails we must try to re-encrypt
	reencrypt := func() {
		reCtx, reCancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer reCancel()
		_ = r.runCmd(reCtx, "encrypt", "-f", file)
	}

	// Step 2: Read the plaintext file
	filePath := file
	if !strings.HasPrefix(filePath, "/") {
		filePath = r.workDir + "/" + filePath
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		reencrypt()
		return fmt.Errorf("read decrypted file failed: %w", err)
	}

	// Build key lookup set
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}

	// Step 3: Remove lines matching any key
	lines := strings.Split(string(data), "\n")
	var kept []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			kept = append(kept, line)
			continue
		}
		eqIdx := strings.Index(trimmed, "=")
		if eqIdx < 0 {
			kept = append(kept, line)
			continue
		}
		lineKey := strings.TrimSpace(trimmed[:eqIdx])
		if _, found := keySet[lineKey]; found {
			continue // skip this line
		}
		kept = append(kept, line)
	}

	// Zero the original buffer
	for i := range data {
		data[i] = 0
	}

	// Step 4: Write modified file
	output := strings.Join(kept, "\n")
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		reencrypt()
		return fmt.Errorf("write modified file failed: %w", err)
	}

	// Step 5: Re-encrypt
	if err := r.runCmd(ctx, "encrypt", "-f", file); err != nil {
		return fmt.Errorf("re-encrypt failed: %w", err)
	}

	return nil
}

// runCmd executes a dotenvx subcommand with the given arguments.
func (r *Runner) runCmd(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, r.binary, args...)
	cmd.Dir = r.workDir
	cmd.Env = minimalEnv()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

// minimalEnv returns a minimal set of environment variables for subprocesses.
// This prevents leaking the parent process's full environment.
func minimalEnv() []string {
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
	}
	if term := os.Getenv("TERM"); term != "" {
		env = append(env, "TERM="+term)
	}
	return env
}
