// Package config loads and manages application configuration with env-override support.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ProviderConfig holds the LLM endpoint, model, and optional API key.
type ProviderConfig struct {
	BaseURL string `json:"baseURL"`
	Model   string `json:"model"`
	APIKey  string `json:"apiKey"`
}

// Config is the top-level application configuration.
type Config struct {
	Provider     ProviderConfig `json:"provider"`
	SystemPrompt string         `json:"systemPrompt"`
	DataDir      string         `json:"dataDir"`
}

// Default returns a Config populated with sensible defaults for local development.
func Default() Config {
	return Config{
		Provider: ProviderConfig{
			BaseURL: "http://localhost:11434/v1",
			Model:   "qwen3.6:27b",
			APIKey:  "no-key",
		},
		SystemPrompt: "You are a helpful assistant in a terminal chat interface. Respond concisely. Format code with markdown fences.",
		DataDir:      ".sessions",
	}
}

// Load reads .config/config.json and applies environment variable overrides.
func Load() (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(".config/config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return applyEnvOverrides(cfg), nil
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing .config/config.json: %w", err)
	}

	return applyEnvOverrides(cfg), nil
}

func applyEnvOverrides(cfg Config) Config {
	if v := os.Getenv("OLLAMA_BASE_URL"); v != "" {
		cfg.Provider.BaseURL = v
	}
	if v := os.Getenv("OLLAMA_MODEL"); v != "" {
		cfg.Provider.Model = v
	}
	return cfg
}

// EnsureDataDirs creates required directories (data dir and .config) if missing.
func EnsureDataDirs(cfg Config) error {
	dirs := []string{
		cfg.DataDir,
		".config",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}
