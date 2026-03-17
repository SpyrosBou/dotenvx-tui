package validate

import (
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
