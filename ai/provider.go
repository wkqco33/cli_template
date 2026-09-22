package ai

import (
	"fmt"
	"time"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/LLM_client_go/azure"
	"github.com/wkqco33/LLM_client_go/ollama"
	"github.com/wkqco33/LLM_client_go/openai"
	"github.com/wkqco33/LLM_client_go/retry"
	"github.com/wkqco33/cli_template/config"
)

func NewClient(cfg config.Config) (llm.Client, error) {
	config.Normalize(&cfg)
	timeout, err := time.ParseDuration(cfg.AI.Timeout)
	if err != nil {
		return nil, fmt.Errorf("AI timeout 설정 오류: %w", err)
	}
	policy := &retry.Policy{MaxRetries: cfg.AI.MaxRetries}
	switch cfg.AI.Provider {
	case "ollama":
		return ollama.New(ollama.Config{BaseURL: cfg.AI.BaseURL, Timeout: timeout, RetryPolicy: policy}), nil
	case "openai", "openai-compatible":
		return openai.New(openai.Config{APIKey: cfg.AI.APIKey, BaseURL: cfg.AI.BaseURL, Timeout: timeout, RetryPolicy: policy}), nil
	case "azure":
		return azure.New(azure.Config{Endpoint: cfg.AI.Endpoint, APIKey: cfg.AI.APIKey, APIVersion: cfg.AI.APIVersion, Timeout: timeout, RetryPolicy: policy}), nil
	default:
		return nil, fmt.Errorf("지원하지 않는 AI provider: %s", cfg.AI.Provider)
	}
}
