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

func TestLoad_ModeAgent(t *testing.T) {
	dir := t.TempDir()
	content := "mode: agent\ninstructions: test\n"
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Mode != ModeAgent {
		t.Errorf("got mode %q, want %q", cfg.Mode, ModeAgent)
	}
	if !cfg.IsAgentMode() {
		t.Error("expected IsAgentMode() to return true")
	}
}

func TestLoad_ModeOrchestrator(t *testing.T) {
	dir := t.TempDir()
	content := "mode: orchestrator\n"
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Mode != ModeOrchestrator {
		t.Errorf("got mode %q, want %q", cfg.Mode, ModeOrchestrator)
	}
	if cfg.IsAgentMode() {
		t.Error("expected IsAgentMode() to return false")
	}
}

func TestLoad_ModeDefault(t *testing.T) {
	dir := t.TempDir()
	content := "instructions: test\n"
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Mode != ModeOrchestrator {
		t.Errorf("expected default mode %q, got %q", ModeOrchestrator, cfg.Mode)
	}
}

func TestLoad_ModeInvalid(t *testing.T) {
	dir := t.TempDir()
	content := "mode: turbo\n"
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}
