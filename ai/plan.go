package ai

import (
	"fmt"

	"github.com/wkqco33/cli_template/generator"
)

type Plan struct {
	ProjectName string `json:"project_name" yaml:"project_name"`
	ModuleName  string `json:"module_name" yaml:"module_name"`
	Template    string `json:"template" yaml:"template"`
	SQLite      bool   `json:"sqlite" yaml:"sqlite"`
	Summary     string `json:"summary" yaml:"summary"`
}

func ValidatePlan(plan Plan) error {
	if plan.ProjectName == "" {
		return fmt.Errorf("AI 계획의 프로젝트 이름이 비어 있습니다")
	}
	if plan.ModuleName == "" {
		return fmt.Errorf("AI 계획의 모듈 이름이 비어 있습니다")
	}
	if err := generator.ValidateProjectAndModuleName(plan.ProjectName, plan.ModuleName); err != nil {
		return err
	}
	known := false
	for _, meta := range generator.Templates() {
		if meta.Name == plan.Template {
			known = true
			break
		}
	}
	if !known {
		return fmt.Errorf("AI 계획의 알 수 없는 템플릿: %s", plan.Template)
	}
	if plan.SQLite && !generator.SQLiteSupported(plan.Template) {
		return fmt.Errorf("%s 템플릿은 SQLite를 지원하지 않습니다", plan.Template)
	}
	return nil
}
