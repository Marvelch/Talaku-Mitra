package utils

import (
	"strings"
)

var passwordBlacklist = []string{
	"123456789", "12345678", "123456", "password", "password123",
	"qwerty123", "qwertyuiop", "abc12345", "11111111", "111111111",
	"00000000", "000000000", "87654321", "iloveyou", "admin123",
}

// ValidatePasswordStrength returns an Indonesian error message describing the
// first rule violation found, or "" if the password is acceptable.
func ValidatePasswordStrength(password, fullName, email string) string {
	if len(password) < 8 {
		return "Password minimal 8 karakter."
	}

	hasLetter := false
	hasDigit := false
	for _, r := range password {
		switch {
		case r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "Password harus mengandung kombinasi huruf dan angka."
	}

	lower := strings.ToLower(password)
	for _, blocked := range passwordBlacklist {
		if lower == blocked {
			return "Password terlalu mudah ditebak. Gunakan kombinasi yang lebih unik."
		}
	}

	if isAllSameChar(lower) || isSequential(lower) {
		return "Password tidak boleh berupa karakter berulang atau berurutan."
	}

	if containsSubstringOf(lower, fullName) {
		return "Password tidak boleh mengandung bagian dari nama Anda."
	}

	localPart := email
	if at := strings.Index(email, "@"); at > 0 {
		localPart = email[:at]
	}
	if containsSubstringOf(lower, localPart) {
		return "Password tidak boleh mengandung bagian dari email Anda."
	}

	return ""
}

func isAllSameChar(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func isSequential(s string) bool {
	ascending := true
	descending := true
	for i := 1; i < len(s); i++ {
		if s[i] != s[i-1]+1 {
			ascending = false
		}
		if s[i] != s[i-1]-1 {
			descending = false
		}
	}
	return ascending || descending
}

// containsSubstringOf reports whether password contains any 3+ char word from name.
func containsSubstringOf(password, name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, word := range strings.Fields(name) {
		if len(word) >= 3 && strings.Contains(password, word) {
			return true
		}
	}
	return false
}
