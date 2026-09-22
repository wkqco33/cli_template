package config

import "testing"

func TestApplyEnv_OverridesProviderAndModel(t *testing.T) {
	t.Setenv("WTEMP_AI_PROVIDER", "openai")
	t.Setenv("WTEMP_AI_MODEL", "test-model")
	t.Setenv("WTEMP_AI_BASE_URL", "http://base")
	t.Setenv("WTEMP_AI_TIMEOUT", "2s")
	t.Setenv("WTEMP_AI_MAX_RETRIES", "4")
	t.Setenv("WTEMP_AI_ENDPOINT", "endpoint")
	t.Setenv("WTEMP_AI_API_VERSION", "v2")
	t.Setenv("WTEMP_AI_API_KEY", "generic-secret")
	t.Setenv("OPENAI_API_KEY", "env-secret")
	cfg := Defaults()
	ApplyEnv(&cfg)
	if cfg.AI.Provider != "openai" || cfg.AI.Model != "test-model" || cfg.AI.APIKey != "env-secret" || cfg.AI.BaseURL != "http://base" || cfg.AI.MaxRetries != 4 || cfg.AI.Endpoint != "endpoint" || cfg.AI.APIVersion != "v2" {
		t.Fatalf("unexpected environment config: %+v", cfg.AI)
	}
}

func TestApplyEnv_AzureKey(t *testing.T) {
	t.Setenv("AZURE_OPENAI_API_KEY", "azure-secret")
	cfg := Defaults()
	cfg.AI.Provider = "azure"
	ApplyEnv(&cfg)
	if cfg.AI.APIKey != "azure-secret" {
		t.Fatalf("expected Azure key, got %q", cfg.AI.APIKey)
	}
}
