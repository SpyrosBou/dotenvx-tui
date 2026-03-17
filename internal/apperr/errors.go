package apperr

import "fmt"

// Kind categorizes application errors for appropriate user-facing handling.
type Kind int

const (
	KindMissingDependency Kind = iota
	KindMissingKeys
	KindDecryptFailed
	KindFileNotFound
	KindInvalidInput
	KindTimeout
)

// AppError represents a user-facing application error.
// Message must NEVER contain decrypted secret values.
type AppError struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates an AppError with the given kind and message.
func New(kind Kind, msg string) *AppError {
	return &AppError{Kind: kind, Message: msg}
}

// Wrap creates an AppError wrapping an underlying error.
func Wrap(kind Kind, msg string, cause error) *AppError {
	return &AppError{Kind: kind, Message: msg, Cause: cause}
}

// MissingDependency creates an error for a missing external tool.
func MissingDependency(tool, installHint string) *AppError {
	return &AppError{
		Kind:    KindMissingDependency,
		Message: fmt.Sprintf("%s is not installed. %s", tool, installHint),
	}
}

// MissingKeys creates an error for a missing .env.keys file.
func MissingKeys(envFile string) *AppError {
	return &AppError{
		Kind:    KindMissingKeys,
		Message: fmt.Sprintf("Cannot decrypt: no .env.keys file found alongside %s. Run 'dotenvx encrypt' first or obtain the keys file from a team member.", envFile),
	}
}
