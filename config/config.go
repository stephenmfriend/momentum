// Package config provides repo-specific configuration for Momentum.
//
// Momentum looks for a .momentum.yaml file in the working directory.
// If found, its settings override built-in defaults.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const filename = ".momentum.yaml"

// RepoConfig holds repo-specific Momentum configuration.
type RepoConfig struct {
	// Instructions replaces the default agent prompt preamble.
	// Task context (ID, title, AC, guardrails) is always appended.
	Instructions string `yaml:"instructions"`
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

	return cfg, nil
}
