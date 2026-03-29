package config

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestEnsureAdminCredentialsRotatesMissingDisplayPassword(t *testing.T) {
	cfg := &Config{
		AdminUser: "admin",
		AdminPass: "$2a$10$existinghashthatwillbereplaced",
	}

	reason, err := EnsureAdminCredentials(cfg)
	if err != nil {
		t.Fatalf("EnsureAdminCredentials returned error: %v", err)
	}
	if reason == "" {
		t.Fatal("expected credentials to be updated when admin_password is missing")
	}
	if cfg.AdminPassword == "" {
		t.Fatal("expected generated admin password to be stored")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cfg.AdminPass), []byte(cfg.AdminPassword)); err != nil {
		t.Fatalf("generated hash does not match generated password: %v", err)
	}
}

func TestEnsureAdminCredentialsRebuildsMissingHash(t *testing.T) {
	cfg := &Config{
		AdminUser:     "admin",
		AdminPassword: "secret123",
	}

	reason, err := EnsureAdminCredentials(cfg)
	if err != nil {
		t.Fatalf("EnsureAdminCredentials returned error: %v", err)
	}
	if reason == "" {
		t.Fatal("expected credentials to be updated when admin_pass is missing")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cfg.AdminPass), []byte("secret123")); err != nil {
		t.Fatalf("rebuilt hash does not match stored plaintext password: %v", err)
	}
}
