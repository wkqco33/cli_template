package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/cli_template/generator"
)

type CompletionClient interface {
	Complete(context.Context, llm.ChatRequest) (*llm.ChatResponse, error)
}

type Planner struct {
	client CompletionClient
	model  string
}

func NewPlanner(client CompletionClient, model string) *Planner {
	return &Planner{client: client, model: model}
}

func (p *Planner) Plan(ctx context.Context, request string) (Plan, error) {
	return p.plan(ctx, request, "", "")
}

// PlanWithOverrides applies explicit CLI values before validating the model
// response. This allows callers to supply values that the model could not infer.
func (p *Planner) PlanWithOverrides(ctx context.Context, request, projectName, moduleName string) (Plan, error) {
	return p.plan(ctx, request, projectName, moduleName)
}

func (p *Planner) plan(ctx context.Context, request, projectName, moduleName string) (Plan, error) {
	if strings.TrimSpace(request) == "" {
		return Plan{}, fmt.Errorf("AI 요청이 비어 있습니다")
	}
	response, err := p.client.Complete(ctx, llm.ChatRequest{
		Model: p.model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt()},
			{Role: llm.RoleUser, Content: request},
		},
		Temperature: floatPtr(0),
		ResponseFormat: &llm.ResponseFormat{
			Type: "json_schema",
			JSONSchema: &llm.JSONSchemaDef{
				Name:        "wtemp_generation_plan",
				Description: "wtemp project generation plan",
				Strict:      true,
				Schema:      planSchema(),
			},
		},
	})
	if err != nil {
		return Plan{}, fmt.Errorf("AI 요청 실패: %w", err)
	}
	if response == nil || len(response.Choices) == 0 {
		return Plan{}, fmt.Errorf("AI 응답에 선택 결과가 없습니다")
	}
	var plan Plan
	content := normalizeJSONContent(response.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return Plan{}, fmt.Errorf("AI 계획 JSON 파싱 실패: %w", err)
	}
	if projectName != "" {
		plan.ProjectName = projectName
	}
	if moduleName != "" {
		plan.ModuleName = moduleName
	}
	if err := ValidatePlan(plan); err != nil {
		return Plan{}, fmt.Errorf("AI 계획 검증 실패: %w", err)
	}
	return plan, nil
}

func systemPrompt() string {
	var b strings.Builder
	b.WriteString("당신은 wtemp Go 프로젝트 템플릿 선택기입니다.\n")
	b.WriteString("사용자 요청을 기존 템플릿과 옵션으로 변환하고 JSON만 반환하세요.\n")
	b.WriteString("파일 내용, 쉘 명령, 임의 의존성은 생성하지 마세요.\n\n")
	b.WriteString("사용 가능한 템플릿:\n")
	for _, meta := range generator.Templates() {
		fmt.Fprintf(&b, "- %s: %s\n", meta.Name, meta.Desc)
	}
	return b.String()
}

func planSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_name": map[string]any{"type": "string"},
			"module_name":  map[string]any{"type": "string"},
			"template":     map[string]any{"type": "string", "enum": templateNames()},
			"sqlite":       map[string]any{"type": "boolean"},
			"summary":      map[string]any{"type": "string"},
		},
		"required":             []string{"project_name", "module_name", "template", "sqlite", "summary"},
		"additionalProperties": false,
	}
}

func templateNames() []string {
	metas := generator.Templates()
	names := make([]string, 0, len(metas))
	for _, meta := range metas {
		names = append(names, meta.Name)
	}
	return names
}

func floatPtr(v float64) *float64 { return &v }

// normalizeJSONContent accepts the JSON-only response requested by the prompt,
// while tolerating common model formatting such as ```json fences or a short
// explanatory sentence around the object.
func normalizeJSONContent(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		if newline := strings.IndexByte(content, '\n'); newline >= 0 {
			content = content[newline+1:]
		}
		content = strings.TrimSpace(content)
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}
	start := strings.IndexByte(content, '{')
	end := strings.LastIndexByte(content, '}')
	if start >= 0 && end > start {
		return content[start : end+1]
	}
	return content
}
