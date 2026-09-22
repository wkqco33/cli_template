package config

import (
	"os"
	"strconv"
)

// ApplyEnv applies WTEMP overrides without changing the stored config file.
// Provider-specific API key variables take precedence over the generic value.
func ApplyEnv(cfg *Config) {
	if value := os.Getenv("WTEMP_AI_PROVIDER"); value != "" {
		cfg.AI.Provider = value
	}
	if value := os.Getenv("WTEMP_AI_MODEL"); value != "" {
		cfg.AI.Model = value
	}
	if value := os.Getenv("WTEMP_AI_BASE_URL"); value != "" {
		cfg.AI.BaseURL = value
	}
	if value := os.Getenv("WTEMP_AI_TIMEOUT"); value != "" {
		cfg.AI.Timeout = value
	}
	if value := os.Getenv("WTEMP_AI_MAX_RETRIES"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			cfg.AI.MaxRetries = parsed
		}
	}
	if value := os.Getenv("WTEMP_AI_ENDPOINT"); value != "" {
		cfg.AI.Endpoint = value
	}
	if value := os.Getenv("WTEMP_AI_API_VERSION"); value != "" {
		cfg.AI.APIVersion = value
	}
	if value := os.Getenv("WTEMP_AI_API_KEY"); value != "" {
		cfg.AI.APIKey = value
	}
	if value := os.Getenv("OPENAI_API_KEY"); value != "" && (cfg.AI.Provider == "openai" || cfg.AI.Provider == "openai-compatible") {
		cfg.AI.APIKey = value
	}
	if value := os.Getenv("AZURE_OPENAI_API_KEY"); value != "" && cfg.AI.Provider == "azure" {
		cfg.AI.APIKey = value
	}
}
