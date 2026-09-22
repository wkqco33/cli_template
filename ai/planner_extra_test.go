package ai

import (
	"context"
	"errors"
	"testing"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/cli_template/config"
)

func TestPlanner_Errors(t *testing.T) {
	for name, client := range map[string]CompletionClient{
		"client error": &fakeClient{err: errors.New("network")},
		"invalid json": &fakeClient{response: &llm.ChatResponse{Choices: []llm.Choice{{Message: llm.Message{Content: "not-json"}}}}},
		"invalid plan": &fakeClient{response: &llm.ChatResponse{Choices: []llm.Choice{{Message: llm.Message{Content: `{"project_name":"bad name"}`}}}}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewPlanner(client, "model").Plan(context.Background(), "request"); err == nil {
				t.Fatal("expected planner error")
			}
		})
	}
	if _, err := NewPlanner(&fakeClient{}, "model").Plan(context.Background(), ""); err == nil {
		t.Fatal("expected empty request error")
	}
}

func TestNewClient_Providers(t *testing.T) {
	for _, provider := range []string{"ollama", "openai", "openai-compatible", "azure"} {
		cfg := config.Defaults()
		cfg.AI.Provider = provider
		cfg.AI.Timeout = "1s"
		if provider == "azure" {
			cfg.AI.Endpoint = "https://example.com"
		}
		client, err := NewClient(cfg)
		if err != nil || client == nil {
			t.Fatalf("provider %s failed: client=%v err=%v", provider, client, err)
		}
	}
	cfg := config.Defaults()
	cfg.AI.Provider = "unknown"
	if _, err := NewClient(cfg); err == nil {
		t.Fatal("expected unknown provider error")
	}
	cfg = config.Defaults()
	cfg.AI.Timeout = "bad"
	if _, err := NewClient(cfg); err == nil {
		t.Fatal("expected timeout error")
	}
}
