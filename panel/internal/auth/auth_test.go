package auth

import (
	"testing"
)

func TestPasswordHash(t *testing.T) {
	password := "test-password-123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash should not be empty")
	}
	if hash == password {
		t.Fatal("hash should not equal plaintext password")
	}

	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword should return true for correct password")
	}
	if VerifyPassword("wrong-password", hash) {
		t.Error("VerifyPassword should return false for wrong password")
	}
}

func TestPasswordHashDifferent(t *testing.T) {
	h1, _ := HashPassword("test")
	h2, _ := HashPassword("test")
	if h1 == h2 {
		t.Error("same password should produce different hashes due to salt")
	}
}
