package ai

import (
	"context"
	"errors"
	"testing"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/cli_template/config"
)

func TestPlanner_ParsesMarkdownJSON(t *testing.T) {
	content := "```json\n{\"project_name\":\"todo-api\",\"module_name\":\"github.com/example/todo-api\",\"template\":\"full\",\"sqlite\":false,\"summary\":\"CLI app\"}\n```"
	client := &fakeClient{response: &llm.ChatResponse{Choices: []llm.Choice{{Message: llm.Message{Content: content}}}}}
	plan, err := NewPlanner(client, "model").Plan(context.Background(), "CLI app")
	if err != nil {
		t.Fatalf("expected fenced JSON to parse: %v", err)
	}
	if plan.Template != "full" {
		t.Fatalf("expected full template, got %q", plan.Template)
	}
}

func TestPlanner_ExplicitOverridesFillMissingNames(t *testing.T) {
	client := &fakeClient{response: &llm.ChatResponse{Choices: []llm.Choice{{Message: llm.Message{Content: `{"project_name":"","module_name":"","template":"full","sqlite":false,"summary":"CLI app"}`}}}}}
	plan, err := NewPlanner(client, "model").PlanWithOverrides(context.Background(), "CLI app", "testcli", "github.com/example/testcli")
	if err != nil {
		t.Fatalf("expected overrides to complete plan: %v", err)
	}
	if plan.ProjectName != "testcli" || plan.ModuleName != "github.com/example/testcli" {
		t.Fatalf("unexpected plan: %+v", plan)
	}

	client.response.Choices[0].Message.Content = `{"project_name":"","module_name":"","template":"full","sqlite":false,"summary":"CLI app"}`
	plan, err = NewPlanner(client, "model").PlanWithOverrides(context.Background(), "CLI app", "my-tool", "")
	if err != nil {
		t.Fatalf("expected project name to provide module default: %v", err)
	}
	if plan.ModuleName != "my-tool" {
		t.Fatalf("expected module default %q, got %q", "my-tool", plan.ModuleName)
	}
}

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
