package cmd

import (
	"fmt"
	"path/filepath"

	"cli_template/generator"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

func NewCmd() *wcli.Command {
	var tmplName string
	var sqlite bool
	var profile bool
	var moduleName string
	var force bool
	var outputDir string
	var dryRun bool
	var git bool

	cmd := &wcli.Command{
		Use:   "new <project-name>",
		Short: "새 CLI 프로젝트를 생성합니다",
		Long:  "지정한 이름으로 wcli 기반 Go CLI 프로젝트를 생성합니다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("프로젝트 이름을 입력하세요: new <project-name>")
			}
			projectName := ctx.Args[0]

			if moduleName == "" {
				moduleName = projectName
			}
			if err := generator.ValidateProjectAndModuleName(projectName, moduleName); err != nil {
				return err
			}

			opts := generator.Options{
				Template:   tmplName,
				SQLite:     sqlite,
				Profile:    profile,
				ModuleName: moduleName,
				Force:      force,
				OutputDir:  outputDir,
				DryRun:     dryRun,
			}

			if dryRun {
				files, err := generator.DryRun(projectName, opts)
				if err != nil {
					return err
				}
				rich.Println("[cyan]생성될 파일:[/cyan] %s (템플릿: %s)", projectName, tmplName)
				for _, f := range files {
					fmt.Printf("  %s\n", f)
				}
				return nil
			}

			rich.Println("[cyan]생성 중:[/cyan] %s (템플릿: %s)", projectName, tmplName)
			if err := generator.Generate(projectName, opts); err != nil {
				return err
			}

			if sqlite && !generator.SQLiteSupported(tmplName) {
				rich.Println("[yellow]주의:[/yellow] %s 템플릿은 --sqlite를 지원하지 않습니다. 옵션이 무시되었습니다.", tmplName)
			}

			if git {
				targetDir := projectName
				if outputDir != "" {
					targetDir = filepath.Join(outputDir, projectName)
				}
				if err := generator.InitGit(targetDir); err != nil {
					return err
				}
				rich.Println("[dim]git 저장소가 초기화되었습니다.[/dim]")
			}

			rich.Println("[green][bold]완료![/bold][/green] %s 프로젝트가 생성되었습니다.", projectName)
			fmt.Printf("\n  cd %s\n  go mod tidy\n  go build .\n\n", projectName)
			rich.Println("[dim]wcli 라이브러리가 의존성으로 추가되었습니다.[/dim]")
			if sqlite && generator.SQLiteSupported(tmplName) {
				rich.Println("[dim]SQLite + GORM이 포함되었습니다. database/ 디렉토리를 확인하세요.[/dim]")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(
		&tmplName,
		"template",
		"t",
		"full",
		fmt.Sprintf("사용할 템플릿 (%s)", generator.TemplateNamesCSV()),
	)
	cmd.Flags().BoolVar(&sqlite, "sqlite", "", false, "SQLite + GORM 지원 추가")
	cmd.Flags().BoolVar(&profile, "profile", "", false, "생성 단계별 성능 프로파일 출력")
	cmd.Flags().StringVar(&moduleName, "module", "", "", "Go 모듈 경로 (기본값: 프로젝트 이름)")
	cmd.Flags().BoolVar(&force, "force", "", false, "대상 디렉토리가 이미 존재해도 덮어쓰기")
	cmd.Flags().StringVar(&outputDir, "output", "o", "", "생성 위치 (기본값: 현재 디렉토리)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", "", false, "실제 생성 없이 생성될 파일 목록만 출력")
	cmd.Flags().BoolVar(&git, "git", "", false, "생성 후 git 저장소 초기화")
	return cmd
}
