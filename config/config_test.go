package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaults_UsesOllama(t *testing.T) {
	got := Defaults()
	if got.AI.Provider != "ollama" {
		t.Fatalf("expected default provider %q, got %q", "ollama", got.AI.Provider)
	}
	if got.AI.Model == "" {
		t.Fatal("expected default Ollama model")
	}
	if got.AI.BaseURL == "" {
		t.Fatal("expected default Ollama base URL")
	}
}

func TestLoad_MissingFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load missing file failed: %v", err)
	}
	if got.AI.Provider != "ollama" {
		t.Fatalf("expected Ollama defaults, got %q", got.AI.Provider)
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("version: 1\nai:\n  provider: openai\n  model: gpt-test\n  api_key: secret\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got.AI.Provider != "openai" || got.AI.Model != "gpt-test" || got.AI.APIKey != "secret" {
		t.Fatalf("unexpected config: %+v", got.AI)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("ai: [invalid"), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid YAML error")
	}
}

func TestSaveAndLoad_AtomicConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	cfg := Defaults()
	cfg.AI.Model = "custom-model"
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load after Save failed: %v", err)
	}
	if got.AI.Model != "custom-model" {
		t.Fatalf("expected model %q, got %q", "custom-model", got.AI.Model)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temporary config should be removed, stat error: %v", err)
	}
	if runtime.GOOS != "windows" {
		if info, err := os.Stat(path); err != nil {
			t.Fatalf("stat config failed: %v", err)
		} else if info.Mode().Perm() != 0o600 {
			t.Fatalf("expected config permissions 0600, got %o", info.Mode().Perm())
		}
	}
}

func TestValidate_RejectsUnknownProvider(t *testing.T) {
	cfg := Defaults()
	cfg.AI.Provider = "unknown"
	if err := Validate(cfg); err == nil || !strings.Contains(err.Error(), "provider") {
		t.Fatalf("expected provider validation error, got %v", err)
	}
}

func TestSetAndUnset_SupportNestedKeys(t *testing.T) {
	cfg := Defaults()
	if err := Set(&cfg, "ai.model", "qwen2.5"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if cfg.AI.Model != "qwen2.5" {
		t.Fatalf("expected model %q, got %q", "qwen2.5", cfg.AI.Model)
	}
	if err := Unset(&cfg, "ai.model"); err != nil {
		t.Fatalf("Unset failed: %v", err)
	}
	if cfg.AI.Model != Defaults().AI.Model {
		t.Fatalf("expected default model after unset, got %q", cfg.AI.Model)
	}
}

func TestMaskSecrets(t *testing.T) {
	cfg := Defaults()
	cfg.AI.APIKey = "super-secret"
	masked := MaskSecrets(cfg)
	if masked.AI.APIKey == cfg.AI.APIKey || masked.AI.APIKey == "" {
		t.Fatalf("expected masked API key, got %q", masked.AI.APIKey)
	}
}
