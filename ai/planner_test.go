package ai

import (
	"context"
	"strings"
	"testing"

	llm "github.com/wkqco33/LLM_client_go"
)

type fakeClient struct {
	request  llm.ChatRequest
	response *llm.ChatResponse
	err      error
}

func (f *fakeClient) Complete(_ context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	f.request = req
	return f.response, f.err
}

func (f *fakeClient) Stream(context.Context, llm.ChatRequest) (llm.Stream, error) { return nil, nil }
func (f *fakeClient) CreateEmbeddings(context.Context, llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, nil
}
func (f *fakeClient) TokenCounter(string) any { return nil }

func TestValidatePlan_Valid(t *testing.T) {
	plan := Plan{
		ProjectName: "todo-api",
		ModuleName:  "github.com/example/todo-api",
		Template:    "gin",
		SQLite:      true,
	}
	if err := ValidatePlan(plan); err != nil {
		t.Fatalf("ValidatePlan failed: %v", err)
	}
}

func TestValidatePlan_RejectsUnknownTemplate(t *testing.T) {
	plan := Plan{ProjectName: "todo-api", ModuleName: "todo-api", Template: "unknown"}
	if err := ValidatePlan(plan); err == nil || !strings.Contains(err.Error(), "템플릿") {
		t.Fatalf("expected template validation error, got %v", err)
	}
}

func TestPlanner_ParsesStructuredResponse(t *testing.T) {
	client := &fakeClient{response: &llm.ChatResponse{
		Choices: []llm.Choice{{Message: llm.Message{
			Role:    llm.RoleAssistant,
			Content: `{"project_name":"todo-api","module_name":"github.com/example/todo-api","template":"gin","sqlite":true,"summary":"Gin API"}`,
		}}},
	}}
	planner := NewPlanner(client, "test-model")

	plan, err := planner.Plan(context.Background(), "Gin 기반 SQLite TODO API")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}
	if plan.Template != "gin" || !plan.SQLite || plan.ProjectName != "todo-api" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if len(client.request.Messages) != 2 {
		t.Fatalf("expected system and user messages, got %d", len(client.request.Messages))
	}
	if client.request.ResponseFormat == nil {
		t.Fatal("expected structured response format")
	}
}

func TestPlanner_RejectsEmptyChoices(t *testing.T) {
	planner := NewPlanner(&fakeClient{response: &llm.ChatResponse{}}, "test-model")
	if _, err := planner.Plan(context.Background(), "something"); err == nil {
		t.Fatal("expected empty choices error")
	}
}
