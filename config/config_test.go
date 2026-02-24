package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileExists(t *testing.T) {
	dir := t.TempDir()
	content := `instructions: "Use the flux-task skill."
`
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Instructions != "Use the flux-task skill." {
		t.Errorf("got instructions %q, want %q", cfg.Instructions, "Use the flux-task skill.")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Instructions != "" {
		t.Errorf("expected empty instructions, got %q", cfg.Instructions)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(":\n\t: bad"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
