package config

const (
	defaultOllamaURL = "http://localhost:11434/v1"
	defaultOpenAIURL = "https://api.openai.com/v1"
)

// Normalize fills provider-specific defaults without overriding explicit values.
func Normalize(cfg *Config) {
	if cfg.AI.BaseURL == "" || cfg.AI.BaseURL == defaultOllamaURL {
		switch cfg.AI.Provider {
		case "openai":
			cfg.AI.BaseURL = defaultOpenAIURL
		case "ollama":
			cfg.AI.BaseURL = defaultOllamaURL
		}
	}
}
