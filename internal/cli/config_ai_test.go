package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wkqco33/cli_template/config"
)

func TestConfigCmd_InitPathSetShowValidate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	te := newTestEnv()
	te.Env.ConfigPath = path
	if err := te.run(t, ConfigCmd, "config", "init"); err != nil {
		t.Fatalf("config init failed: %v", err)
	}
	if err := te.run(t, ConfigCmd, "config", "set", "ai.model", "test-model"); err != nil {
		t.Fatalf("config set failed: %v", err)
	}
	if err := te.run(t, ConfigCmd, "config", "validate"); err != nil {
		t.Fatalf("config validate failed: %v", err)
	}

	show := newTestEnv()
	show.Env.ConfigPath = path
	if err := show.run(t, ConfigCmd, "config", "show", "--format", "json"); err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	var got config.Config
	if err := json.Unmarshal(show.Stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode config show: %v", err)
	}
	if got.AI.Model != "test-model" {
		t.Fatalf("expected configured model, got %q", got.AI.Model)
	}

	pathOut := newTestEnv()
	pathOut.Env.ConfigPath = path
	if err := pathOut.run(t, ConfigCmd, "config", "path"); err != nil {
		t.Fatalf("config path failed: %v", err)
	}
	if strings.TrimSpace(pathOut.Stdout.String()) != path {
		t.Fatalf("expected path %q, got %q", path, pathOut.Stdout.String())
	}
}

func TestConfigCmd_ShowMasksAPIKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := config.Defaults()
	cfg.AI.APIKey = "secret-value"
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}
	te := newTestEnv()
	te.Env.ConfigPath = path
	if err := te.run(t, ConfigCmd, "config", "show"); err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if strings.Contains(te.Stdout.String(), "secret-value") || !strings.Contains(te.Stdout.String(), "********") {
		t.Fatalf("expected masked API key, got %q", te.Stdout.String())
	}
}

func TestAICmd_DryRunUsesConfiguredProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"project_name\":\"todo-api\",\"module_name\":\"github.com/example/todo-api\",\"template\":\"gin\",\"sqlite\":true,\"summary\":\"Gin API\"}"}}]}`))
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := config.Defaults()
	cfg.AI.Provider = "openai-compatible"
	cfg.AI.BaseURL = server.URL
	cfg.AI.Model = "test-model"
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}
	te := newTestEnv()
	te.Env.ConfigPath = path
	if err := te.run(t, AICmd, "ai", "Gin API", "--dry-run", "--format", "json"); err != nil {
		t.Fatalf("ai dry-run failed: %v\nstderr=%s", err, te.Stderr.String())
	}
	if !strings.Contains(te.Stdout.String(), `"template":"gin"`) || !strings.Contains(te.Stdout.String(), "main.go") {
		t.Fatalf("unexpected AI result: %s", te.Stdout.String())
	}
}

func TestAICmd_MissingRequest(t *testing.T) {
	te := newTestEnv()
	if err := te.run(t, AICmd, "ai"); err == nil {
		t.Fatal("expected missing request error")
	}
}

func TestConfigCmd_InitRejectsExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("version: 1\n"), 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	te := newTestEnv()
	te.Env.ConfigPath = path
	if err := te.run(t, ConfigCmd, "config", "init"); err == nil {
		t.Fatal("expected existing config error")
	}
}
