package config

import "testing"

func TestSet_AllSupportedKeys(t *testing.T) {
	cases := []struct{ key, value string }{
		{"ai.provider", "ollama"}, {"ai.model", "model"}, {"ai.base_url", "http://localhost"},
		{"ai.api_key", "key"}, {"ai.endpoint", "endpoint"}, {"ai.api_version", "v1"},
		{"ai.timeout", "2s"}, {"ai.max_retries", "3"},
		{"generation.default_template", "minimal"}, {"generation.default_output", "out"},
		{"generation.confirm", "false"}, {"generation.git", "true"},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			cfg := Defaults()
			if err := Set(&cfg, tc.key, tc.value); err != nil {
				t.Fatalf("Set failed: %v", err)
			}
		})
	}
}

func TestSet_RejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"unknown", "x"}, {"ai.max_retries", "x"}, {"generation.confirm", "x"}, {"generation.git", "x"},
	} {
		cfg := Defaults()
		if err := Set(&cfg, tc.key, tc.value); err == nil {
			t.Fatalf("expected error for %s", tc.key)
		}
	}
}

func TestUnset_RejectsUnsupportedKey(t *testing.T) {
	cfg := Defaults()
	if err := Unset(&cfg, "ai.provider"); err == nil {
		t.Fatal("expected unsupported unset error")
	}
}

func TestValidate_RejectsInvalidFields(t *testing.T) {
	cases := []func(*Config){
		func(c *Config) { c.Version = 99 },
		func(c *Config) { c.AI.Model = "" },
		func(c *Config) { c.AI.Timeout = "invalid" },
		func(c *Config) { c.AI.MaxRetries = -1 },
		func(c *Config) { c.Generation.DefaultTemplate = "" },
	}
	for i, mutate := range cases {
		cfg := Defaults()
		mutate(&cfg)
		if err := Validate(cfg); err == nil {
			t.Fatalf("case %d: expected validation error", i)
		}
	}
}

func TestNormalize_ProviderDefaults(t *testing.T) {
	cfg := Defaults()
	cfg.AI.Provider = "openai"
	Normalize(&cfg)
	if cfg.AI.BaseURL != defaultOpenAIURL {
		t.Fatalf("expected OpenAI URL, got %q", cfg.AI.BaseURL)
	}
	cfg = Defaults()
	cfg.AI.Provider = "ollama"
	Normalize(&cfg)
	if cfg.AI.BaseURL != defaultOllamaURL {
		t.Fatalf("expected Ollama URL, got %q", cfg.AI.BaseURL)
	}
}
