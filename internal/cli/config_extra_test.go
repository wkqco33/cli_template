package cli

import (
	"path/filepath"
	"testing"
)

func TestConfigCmd_UnsetAndErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	te := newTestEnv()
	te.Env.ConfigPath = path
	if err := te.run(t, ConfigCmd, "config", "init"); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if err := te.run(t, ConfigCmd, "config", "set", "ai.api_key", "secret"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if err := te.run(t, ConfigCmd, "config", "unset", "ai.api_key"); err != nil {
		t.Fatalf("unset failed: %v", err)
	}
	if err := te.run(t, ConfigCmd, "config", "set", "unknown", "value"); err == nil {
		t.Fatal("expected unknown key error")
	}
	if err := te.run(t, ConfigCmd, "config", "unset", "ai.provider"); err == nil {
		t.Fatal("expected unsupported unset error")
	}
}
