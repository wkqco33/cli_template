package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wkqco33/cli_template/config"
)

func TestAICmd_GeneratesProjectFromPlan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{\"project_name\":\"generated-app\",\"module_name\":\"github.com/example/generated-app\",\"template\":\"minimal\",\"sqlite\":false,\"summary\":\"minimal app\"}"}}]}`))
	}))
	defer server.Close()

	root := t.TempDir()
	path := filepath.Join(root, "config.yaml")
	cfg := config.Defaults()
	cfg.AI.Provider = "openai-compatible"
	cfg.AI.BaseURL = server.URL
	cfg.AI.Model = "test-model"
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	te := newTestEnv()
	te.Env.ConfigPath = path
	te.Env.Yes = true
	te.Env.ConfigPath = path
	// The command runs in the temporary root so the generated project is isolated.
	cwd := chdirTemp(t)
	_ = cwd
	if err := te.run(t, AICmd, "ai", "minimal CLI", "--name", "generated-app", "--module", "github.com/example/generated-app"); err != nil {
		t.Fatalf("AI generation failed: %v\nstderr=%s", err, te.Stderr.String())
	}
	if _, err := os.Stat("generated-app/main.go"); err != nil {
		t.Fatalf("expected generated main.go: %v", err)
	}
}
