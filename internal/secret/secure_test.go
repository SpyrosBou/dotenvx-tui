package secret

import (
	"testing"
)

func TestNewZerosSource(t *testing.T) {
	src := []byte("secret-value")
	_ = New(src)

	for i, b := range src {
		if b != 0 {
			t.Errorf("source byte %d not zeroed: got %d", i, b)
		}
	}
}

func TestClearZerosData(t *testing.T) {
	s := New([]byte("secret-value"))
	s.Clear()

	s.Read(func(b []byte) {
		if b != nil {
			t.Error("expected nil after Clear()")
		}
	})
}

func TestDoubleClear(t *testing.T) {
	s := New([]byte("secret-value"))
	s.Clear()
	s.Clear() // should not panic
}

func TestMasked(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ab", "**"},
		{"abcd", "****"},
		{"abcde", "abcd*"},
		{"postgresql://user:pass@host/db", "post**************************"},
	}

	for _, tt := range tests {
		s := New([]byte(tt.input))
		got := s.Masked()
		if got != tt.want {
			t.Errorf("Masked(%q) = %q, want %q", tt.input, got, tt.want)
		}
		s.Clear()
	}
}

func TestMaskedAfterClear(t *testing.T) {
	s := New([]byte("secret"))
	s.Clear()
	if got := s.Masked(); got != "(cleared)" {
		t.Errorf("Masked() after Clear() = %q, want %q", got, "(cleared)")
	}
}

func TestReadCallback(t *testing.T) {
	original := "hello-world"
	s := New([]byte(original))

	var captured string
	s.Read(func(b []byte) {
		captured = string(b)
	})

	if captured != original {
		t.Errorf("Read() got %q, want %q", captured, original)
	}

	s.Clear()
}

func TestLen(t *testing.T) {
	s := New([]byte("12345"))
	if s.Len() != 5 {
		t.Errorf("Len() = %d, want 5", s.Len())
	}
}

func TestString(t *testing.T) {
	s := New([]byte("hello"))
	if s.String() != "hello" {
		t.Errorf("String() = %q, want %q", s.String(), "hello")
	}
	s.Clear()
	if s.String() != "" {
		t.Errorf("String() after Clear() = %q, want %q", s.String(), "")
	}
}
