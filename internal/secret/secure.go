package secret

import (
	"runtime"
	"strings"
	"sync"
)

// SecureBytes holds sensitive data in a byte slice that is zeroed on cleanup.
// Never convert the internal data to a Go string — strings are immutable and cannot be zeroed.
type SecureBytes struct {
	mu   sync.Mutex
	data []byte
	dead bool
}

// New creates a SecureBytes from src, then zeros the source slice.
// Ownership of the data is transferred — the caller should not use src after this call.
func New(src []byte) *SecureBytes {
	s := &SecureBytes{
		data: make([]byte, len(src)),
	}
	copy(s.data, src)

	// Zero the caller's copy — ownership transferred
	for i := range src {
		src[i] = 0
	}

	runtime.SetFinalizer(s, (*SecureBytes).Clear)
	return s
}

// Clear zeros the internal data. Safe to call multiple times.
func (s *SecureBytes) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dead {
		return
	}
	for i := range s.data {
		s.data[i] = 0
	}
	s.dead = true
	runtime.SetFinalizer(s, nil)
}

// Read calls fn with the raw bytes. The callback MUST NOT retain the slice
// reference beyond the call.
func (s *SecureBytes) Read(fn func([]byte)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dead {
		fn(nil)
		return
	}
	fn(s.data)
}

// Masked returns a display-safe version: first 4 bytes visible, rest as bullets.
// If the value is 4 bytes or fewer, all characters are masked.
func (s *SecureBytes) Masked() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dead || len(s.data) == 0 {
		return "(cleared)"
	}
	const visible = 4
	if len(s.data) <= visible {
		return strings.Repeat("*", len(s.data))
	}
	return string(s.data[:visible]) + strings.Repeat("*", len(s.data)-visible)
}

// String returns the full value as a string. Use sparingly — prefer Read() or Masked().
func (s *SecureBytes) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dead {
		return ""
	}
	return string(s.data)
}

// Len returns the length of the data without exposing it.
func (s *SecureBytes) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}
