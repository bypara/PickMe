package auth

import (
	"strings"
	"testing"
)

func TestPasswordVerification(t *testing.T) {
	const password = " Мой новый пароль "
	stored := hashPassword(password)
	if stored == hashPassword(password) {
		t.Fatal("password hashes should use different salts")
	}

	if matches, err := verifyPassword(password, stored); err != nil || !matches {
		t.Fatalf("correct password rejected: %v", err)
	}
	if matches, err := verifyPassword(strings.TrimSpace(password), stored); err != nil || matches {
		t.Fatalf("different password: matches = %t, err = %v", matches, err)
	}
	if matches, err := verifyPassword(password, strings.Replace(stored, "m=19456", "m=999999", 1)); err == nil || matches {
		t.Fatal("invalid memory setting in a stored hash was accepted")
	}
}
