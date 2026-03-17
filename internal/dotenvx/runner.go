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
