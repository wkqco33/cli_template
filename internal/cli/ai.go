package cli

import (
	"fmt"
	"strings"

	"github.com/wkqco33/cli_template/ai"
	"github.com/wkqco33/cli_template/config"
	"github.com/wkqco33/cli_template/generator"
	"github.com/wkqco33/wcli"
)

const aiLong = `자연어 요청을 기존 wtemp 템플릿 생성 계획으로 변환합니다.

예시:
  wtemp ai "Gin 기반의 SQLite TODO REST API"
  wtemp ai "작은 Go CLI 도구" --name my-tool --dry-run

AI는 계획만 만들고 실제 파일 생성은 검증된 wtemp generator가 수행합니다.`

type aiResult struct {
	Request string   `json:"request"`
	Plan    ai.Plan  `json:"plan"`
	Files   []string `json:"files,omitempty"`
}

func AICmd(env *Env) *wcli.Command {
	var name, module, provider, model, output, format string
	var sqlite, dryRun, git bool
	cmd := &wcli.Command{
		Use: "ai <request>", Short: "AI로 프로젝트 생성 계획을 만듭니다", Long: aiLong,
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) != 1 || strings.TrimSpace(ctx.Args[0]) == "" {
				return generator.NewUsageError("AI 요청을 입력하세요.\n사용법: wtemp ai \"<request>\"")
			}
			path, err := config.ResolvePath(env.ConfigPath)
			if err != nil {
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}
			config.ApplyEnv(&cfg)
			if provider != "" {
				cfg.AI.Provider = provider
			}
			if model != "" {
				cfg.AI.Model = model
			}
			config.Normalize(&cfg)
			if err := config.Validate(cfg); err != nil {
				return err
			}
			client, err := ai.NewClient(cfg)
			if err != nil {
				return &generator.ExternalError{Tool: "AI provider", Err: err}
			}
			planner := ai.NewPlanner(client, cfg.AI.Model)
			env.progress("[cyan]AI 요청 중...[/cyan]")
			plan, err := planner.PlanWithOverrides(ctx, ctx.Args[0], name, module)
			if err != nil {
				return &generator.ExternalError{Tool: "AI provider", Err: err}
			}
			if sqlite {
				plan.SQLite = true
			}
			if err := ai.ValidatePlan(plan); err != nil {
				return err
			}

			outputDir := output
			if outputDir == "" {
				outputDir = cfg.Generation.DefaultOutput
			}
			opts := generator.Options{Template: plan.Template, SQLite: plan.SQLite, ModuleName: plan.ModuleName, OutputDir: outputDir, DryRun: dryRun}
			var files []string
			if dryRun {
				files, err = generator.DryRun(plan.ProjectName, opts)
				if err != nil {
					return err
				}
			}
			if format == "json" {
				return writeJSON(env.Stdout, aiResult{Request: ctx.Args[0], Plan: plan, Files: files})
			}
			env.progress("[cyan]AI 해석 결과:[/cyan] %s", plan.Summary)
			env.progress("  프로젝트: %s", plan.ProjectName)
			env.progress("  템플릿: %s", plan.Template)
			env.progress("  모듈: %s", plan.ModuleName)
			if plan.SQLite {
				env.progress("  SQLite: 사용")
			}
			if dryRun {
				for _, file := range files {
					fmt.Fprintln(env.Stdout, file)
				}
				return nil
			}
			resolved, err := generator.Resolve(plan.ProjectName, opts)
			if err != nil {
				return err
			}
			proceed, err := confirmOverwrite(env, resolved.TargetPath, false)
			if err != nil {
				return err
			}
			if !proceed {
				return nil
			}
			opts.Force = true
			env.progress("[cyan]생성 중:[/cyan] %s (템플릿: %s)", plan.ProjectName, plan.Template)
			if err := generator.Generate(plan.ProjectName, opts); err != nil {
				return err
			}
			if git {
				if err := generator.InitGit(resolved.TargetPath); err != nil {
					return err
				}
			}
			env.progress("[green][bold]완료![/bold][/green] %s 프로젝트가 생성되었습니다.", plan.ProjectName)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "", "프로젝트 이름을 직접 지정합니다")
	cmd.Flags().StringVar(&module, "module", "", "", "Go 모듈 경로를 직접 지정합니다")
	cmd.Flags().StringVar(&provider, "provider", "", "", "AI provider를 덮어씁니다")
	cmd.Flags().StringVar(&model, "model", "", "", "AI 모델을 덮어씁니다")
	cmd.Flags().StringVar(&output, "output", "o", "", "생성 위치")
	cmd.Flags().BoolVar(&sqlite, "sqlite", "", false, "SQLite 옵션을 추가합니다")
	cmd.Flags().BoolVar(&dryRun, "dry-run", "n", false, "생성 계획만 출력합니다")
	cmd.Flags().BoolVar(&git, "git", "", false, "생성 후 git 저장소를 초기화합니다")
	cmd.Flags().StringVar(&format, "format", "", "", "출력 형식 json")
	return cmd
}
