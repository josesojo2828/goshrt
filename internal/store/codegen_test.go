package store

import "testing"

func TestGenerateShortCode_Length(t *testing.T) {
	code, err := GenerateShortCode(6)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Errorf("expected length 6, got %d", len(code))
	}
}

func TestGenerateShortCode_ValidChars(t *testing.T) {
	code, err := GenerateShortCode(100)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range code {
		if !containsChar(charset, byte(c)) {
			t.Errorf("invalid character %c in code %s", c, code)
			break
		}
	}
}

func containsChar(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

func TestGenerateShortCode_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := GenerateShortCode(8)
		if err != nil {
			t.Fatal(err)
		}
		if seen[code] {
			t.Errorf("duplicate code generated: %s", code)
		}
		seen[code] = true
	}
}

func TestGenerateShortCode_ZeroLength(t *testing.T) {
	code, err := GenerateShortCode(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 0 {
		t.Errorf("expected empty string for length 0, got %s", code)
	}
}
