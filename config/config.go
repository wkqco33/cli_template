package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const CurrentVersion = 1

type Config struct {
	Version    int              `yaml:"version" json:"version"`
	AI         AIConfig         `yaml:"ai" json:"ai"`
	Generation GenerationConfig `yaml:"generation" json:"generation"`
}

type AIConfig struct {
	Provider   string `yaml:"provider" json:"provider"`
	Model      string `yaml:"model" json:"model"`
	BaseURL    string `yaml:"base_url" json:"base_url"`
	APIKey     string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	Endpoint   string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	APIVersion string `yaml:"api_version,omitempty" json:"api_version,omitempty"`
	Timeout    string `yaml:"timeout" json:"timeout"`
	MaxRetries int    `yaml:"max_retries" json:"max_retries"`
}

type GenerationConfig struct {
	DefaultTemplate string `yaml:"default_template" json:"default_template"`
	DefaultOutput   string `yaml:"default_output" json:"default_output"`
	Confirm         bool   `yaml:"confirm" json:"confirm"`
	Git             bool   `yaml:"git" json:"git"`
}

func Defaults() Config {
	return Config{
		Version: CurrentVersion,
		AI: AIConfig{
			Provider:   "ollama",
			Model:      "llama3.1",
			BaseURL:    "http://localhost:11434/v1",
			Timeout:    "60s",
			MaxRetries: 2,
		},
		Generation: GenerationConfig{
			DefaultTemplate: "full",
			DefaultOutput:   ".",
			Confirm:         true,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("설정 파일 읽기 실패 (%s): %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("설정 파일 형식 오류 (%s): %w", path, err)
	}
	if cfg.Version != CurrentVersion {
		return Config{}, fmt.Errorf("지원하지 않는 config 버전입니다: %d (지원 버전: %d)", cfg.Version, CurrentVersion)
	}
	Normalize(&cfg)
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := Validate(cfg); err != nil {
		return err
	}
	cfg.Version = CurrentVersion
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("설정 디렉터리 생성 실패 (%s): %w", parent, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("임시 설정 파일 쓰기 실패: %w", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("설정 파일 권한 설정 실패: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("설정 파일 적용 실패: %w", err)
	}
	return nil
}

func Validate(cfg Config) error {
	if cfg.Version != 0 && cfg.Version != CurrentVersion {
		return fmt.Errorf("지원하지 않는 config 버전입니다: %d", cfg.Version)
	}
	switch cfg.AI.Provider {
	case "ollama", "openai", "openai-compatible", "azure":
	default:
		return fmt.Errorf("ai.provider가 올바르지 않습니다: %s", cfg.AI.Provider)
	}
	if strings.TrimSpace(cfg.AI.Model) == "" {
		return fmt.Errorf("ai.model은 비어 있을 수 없습니다")
	}
	if cfg.AI.Timeout != "" {
		if _, err := time.ParseDuration(cfg.AI.Timeout); err != nil {
			return fmt.Errorf("ai.timeout이 올바르지 않습니다: %w", err)
		}
	}
	if cfg.AI.MaxRetries < 0 {
		return fmt.Errorf("ai.max_retries는 0 이상이어야 합니다")
	}
	if cfg.Generation.DefaultTemplate == "" {
		return fmt.Errorf("generation.default_template은 비어 있을 수 없습니다")
	}
	return nil
}

func Set(cfg *Config, key, value string) error {
	switch key {
	case "ai.provider":
		cfg.AI.Provider = value
	case "ai.model":
		cfg.AI.Model = value
	case "ai.base_url":
		cfg.AI.BaseURL = value
	case "ai.api_key":
		cfg.AI.APIKey = value
	case "ai.endpoint":
		cfg.AI.Endpoint = value
	case "ai.api_version":
		cfg.AI.APIVersion = value
	case "ai.timeout":
		cfg.AI.Timeout = value
	case "ai.max_retries":
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%s에는 정수를 입력해야 합니다", key)
		}
		cfg.AI.MaxRetries = parsed
	case "generation.default_template":
		cfg.Generation.DefaultTemplate = value
	case "generation.default_output":
		cfg.Generation.DefaultOutput = value
	case "generation.confirm":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%s에는 true 또는 false를 입력해야 합니다", key)
		}
		cfg.Generation.Confirm = parsed
	case "generation.git":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%s에는 true 또는 false를 입력해야 합니다", key)
		}
		cfg.Generation.Git = parsed
	default:
		return fmt.Errorf("알 수 없는 설정 키입니다: %s", key)
	}
	return Validate(*cfg)
}

func Unset(cfg *Config, key string) error {
	switch key {
	case "ai.api_key":
		cfg.AI.APIKey = ""
	case "ai.endpoint":
		cfg.AI.Endpoint = ""
	case "ai.api_version":
		cfg.AI.APIVersion = ""
	case "ai.base_url":
		cfg.AI.BaseURL = Defaults().AI.BaseURL
	case "ai.model":
		cfg.AI.Model = Defaults().AI.Model
	default:
		return fmt.Errorf("unset을 지원하지 않는 설정 키입니다: %s", key)
	}
	return Validate(*cfg)
}

func MaskSecrets(cfg Config) Config {
	if cfg.AI.APIKey != "" {
		cfg.AI.APIKey = "********"
	}
	return cfg
}
