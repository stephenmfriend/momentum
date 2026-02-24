// Package config provides repo-specific configuration for Momentum.
//
// Momentum looks for a .momentum.yaml file in the working directory.
// If found, its settings override built-in defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const filename = ".momentum.yaml"

// Mode controls how momentum manages task lifecycle.
type Mode string

const (
	// ModeOrchestrator is the default — momentum manages status transitions
	// (in_progress before spawn, done on exit 0, planning on user stop).
	ModeOrchestrator Mode = "orchestrator"

	// ModeAgent delegates status management to the agent. Momentum only
	// resets to planning on user-initiated stops (safety net).
	ModeAgent Mode = "agent"
)

// RepoConfig holds repo-specific Momentum configuration.
type RepoConfig struct {
	// Mode controls lifecycle management: "orchestrator" (default) or "agent".
	Mode Mode `yaml:"mode"`

	// Instructions replaces the default agent prompt preamble.
	// Task context (ID, title, AC, guardrails) is always appended.
	Instructions string `yaml:"instructions"`
}

// IsAgentMode returns true when the agent owns the task lifecycle.
func (c RepoConfig) IsAgentMode() bool {
	return c.Mode == ModeAgent
}

// Load reads .momentum.yaml from dir. Returns a zero-value RepoConfig
// (not an error) if the file doesn't exist.
func Load(dir string) (RepoConfig, error) {
	path := filepath.Join(dir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return RepoConfig{}, nil
		}
		return RepoConfig{}, err
	}

	var cfg RepoConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return RepoConfig{}, err
	}

	// Validate mode
	switch cfg.Mode {
	case "", ModeOrchestrator:
		cfg.Mode = ModeOrchestrator
	case ModeAgent:
		// valid
	default:
		return RepoConfig{}, fmt.Errorf("invalid mode %q (use \"orchestrator\" or \"agent\")", cfg.Mode)
	}

	return cfg, nil
}
