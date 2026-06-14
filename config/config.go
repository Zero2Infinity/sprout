package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ProviderConfig struct {
	BaseURL string `json:"baseURL"`
	Model   string `json:"model"`
	APIKey  string `json:"apiKey"`
}

type Config struct {
	Provider     ProviderConfig `json:"provider"`
	SystemPrompt string         `json:"systemPrompt"`
	DataDir      string         `json:"dataDir"`
}

func Default() Config {
	return Config{
		Provider: ProviderConfig{
			BaseURL: "http://localhost:11434/v1",
			Model:   "qwen3:27b",
			APIKey:  "ollama",
		},
		SystemPrompt: "You are a helpful assistant in a terminal chat interface. Respond concisely. Format code with markdown fences.",
		DataDir:      "sessions",
	}
}

func Load() (Config, error) {
	cfg := Default()

	data, err := os.ReadFile("config/config.json")
	if err != nil {
		if os.IsNotExist(err) {
			return applyEnvOverrides(cfg), nil
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config/config.json: %w", err)
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

func EnsureDataDirs(cfg Config) error {
	dirs := []string{
		cfg.DataDir,
		filepath.Dir("config/config.json"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return nil
}
