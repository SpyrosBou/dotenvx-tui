package dotenvx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SpyrosBou/dotenvx-tui/internal/validate"
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
	if err := validate.KeyName(key); err != nil {
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
	if err := validate.KeyName(key); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	return r.rewriteEncryptedFile(ctx, file, func(plain []byte) ([]byte, error) {
		return setPlaintextValue(plain, key, value)
	})
}

// Unset removes one or more keys from an encrypted env file.
// It rewrites the encrypted file through a private staging file so the target
// file is never replaced with plaintext.
func (r *Runner) Unset(ctx context.Context, file string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := validate.FilePath(r.workDir, file); err != nil {
		return err
	}
	for _, key := range keys {
		if err := validate.KeyName(key); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}

	return r.rewriteEncryptedFile(ctx, file, func(plain []byte) ([]byte, error) {
		return unsetPlaintextValues(plain, keySet), nil
	})
}

func (r *Runner) rewriteEncryptedFile(ctx context.Context, file string, transform func([]byte) ([]byte, error)) error {
	filePath := file
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(r.workDir, filePath)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat env file failed: %w", err)
	}
	originalMode := stat.Mode().Perm()

	plain, err := r.decryptToStdout(ctx, file)
	if err != nil {
		return err
	}
	defer zeroBytes(plain)

	nextPlain, err := transform(plain)
	if err != nil {
		return err
	}
	defer zeroBytes(nextPlain)

	encrypted, err := r.encryptPlaintextToStdout(ctx, filePath, nextPlain)
	if err != nil {
		return err
	}
	defer zeroBytes(encrypted)

	if err := atomicReplaceFile(filePath, encrypted, originalMode); err != nil {
		return err
	}

	return nil
}

func (r *Runner) decryptToStdout(ctx context.Context, file string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.binary, "decrypt", "--stdout", "-f", file)
	cmd.Dir = r.workDir
	cmd.Env = minimalEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("dotenvx decrypt failed: %s", strings.TrimSpace(stderr.String()))
	}

	out := make([]byte, stdout.Len())
	copy(out, stdout.Bytes())
	zeroBytes(stdout.Bytes())
	return out, nil
}

func (r *Runner) encryptPlaintextToStdout(ctx context.Context, targetPath string, plain []byte) ([]byte, error) {
	stageDir, err := os.MkdirTemp("", "dotenvx-tui-*")
	if err != nil {
		return nil, fmt.Errorf("create staging directory failed: %w", err)
	}
	defer os.RemoveAll(stageDir)

	plainPath := filepath.Join(stageDir, filepath.Base(targetPath))
	if err := os.WriteFile(plainPath, plain, 0o600); err != nil {
		return nil, fmt.Errorf("write staged plaintext failed: %w", err)
	}

	cmd := exec.CommandContext(ctx, r.binary, "encrypt", "--stdout", "-f", plainPath, "-fk", envKeysPath(targetPath))
	cmd.Dir = r.workDir
	cmd.Env = minimalEnv()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("dotenvx encrypt failed: %s", strings.TrimSpace(stderr.String()))
	}

	out := make([]byte, stdout.Len())
	copy(out, stdout.Bytes())
	zeroBytes(stdout.Bytes())
	return out, nil
}

func unsetPlaintextValues(plain []byte, keySet map[string]struct{}) []byte {
	lines := strings.Split(string(plain), "\n")
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

	return []byte(strings.Join(kept, "\n"))
}

func setPlaintextValue(plain []byte, key string, value []byte) ([]byte, error) {
	formatted, err := formatEnvValue(value)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(plain), "\n")
	replacement := key + "=" + formatted
	replaced := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		eqIdx := strings.Index(trimmed, "=")
		if eqIdx < 0 {
			continue
		}
		lineKey := strings.TrimSpace(trimmed[:eqIdx])
		if lineKey == key {
			lines[i] = replacement
			replaced = true
		}
	}

	if !replaced {
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		lines = append(lines, replacement, "")
	}

	return []byte(strings.Join(lines, "\n")), nil
}

func formatEnvValue(value []byte) (string, error) {
	if bytes.ContainsAny(value, "\n\r") {
		return "", fmt.Errorf("values containing newlines are not supported by the staged dotenvx rewrite")
	}

	if !bytes.ContainsRune(value, '`') {
		return quoteWith(value, '`', true), nil
	}
	if !bytes.ContainsRune(value, '\'') {
		return quoteWith(value, '\'', false), nil
	}
	if !bytes.ContainsRune(value, '"') {
		return quoteWith(value, '"', true), nil
	}

	return "", fmt.Errorf("values containing backticks, single quotes, and double quotes together are not supported by the staged dotenvx rewrite")
}

func quoteWith(value []byte, quote byte, escapeDollar bool) string {
	var b strings.Builder
	b.Grow(len(value) + 2)
	b.WriteByte(quote)
	for _, c := range value {
		if escapeDollar && c == '$' {
			b.WriteString(`\$`)
			continue
		}
		b.WriteByte(c)
	}
	b.WriteByte(quote)
	return b.String()
}

func envKeysPath(envPath string) string {
	return filepath.Join(filepath.Dir(envPath), ".env.keys")
}

func atomicReplaceFile(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create encrypted temp file failed: %w", err)
	}
	tmpPath := tmp.Name()
	closed := false
	defer func() {
		if !closed {
			_ = tmp.Close()
		}
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write encrypted temp file failed: %w", err)
	}
	if err := tmp.Chmod(mode); err != nil {
		return fmt.Errorf("set encrypted temp file mode failed: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync encrypted temp file failed: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close encrypted temp file failed: %w", err)
	}
	closed = true

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace env file failed: %w", err)
	}
	if err := syncDir(dir); err != nil {
		return fmt.Errorf("sync env directory failed: %w", err)
	}
	return nil
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
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
