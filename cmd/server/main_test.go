package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFileSetsMissingValues(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	if err := os.WriteFile(envPath, []byte("DEMO_ACCESS_PASSWORD=secret\nPORT=9090\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("DEMO_ACCESS_PASSWORD", "")
	t.Setenv("PORT", "")

	if err := os.Unsetenv("DEMO_ACCESS_PASSWORD"); err != nil {
		t.Fatalf("unset demo access password: %v", err)
	}

	if err := os.Unsetenv("PORT"); err != nil {
		t.Fatalf("unset port: %v", err)
	}

	if err := loadEnvFile(envPath); err != nil {
		t.Fatalf("load env file: %v", err)
	}

	if got := os.Getenv("DEMO_ACCESS_PASSWORD"); got != "secret" {
		t.Fatalf("expected demo access password %q, got %q", "secret", got)
	}

	if got := os.Getenv("PORT"); got != "9090" {
		t.Fatalf("expected port %q, got %q", "9090", got)
	}
}

func TestLoadEnvFileDoesNotOverrideExistingValues(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	if err := os.WriteFile(envPath, []byte("DEMO_ACCESS_PASSWORD=secret\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	t.Setenv("DEMO_ACCESS_PASSWORD", "already-set")

	if err := loadEnvFile(envPath); err != nil {
		t.Fatalf("load env file: %v", err)
	}

	if got := os.Getenv("DEMO_ACCESS_PASSWORD"); got != "already-set" {
		t.Fatalf("expected demo access password %q, got %q", "already-set", got)
	}
}
