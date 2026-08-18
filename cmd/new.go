package cmd

import (
	"fmt"

	"cli_template/generator"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

func NewCmd() *wcli.Command {
	var tmplName string
	var sqlite bool
	var profile bool

	cmd := &wcli.Command{
		Use:   "new <project-name>",
		Short: "새 CLI 프로젝트를 생성합니다",
		Long:  "지정한 이름으로 wcli 기반 Go CLI 프로젝트를 생성합니다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("프로젝트 이름을 입력하세요: new <project-name>")
			}
			projectName := ctx.Args[0]

			if err := generator.ValidateProjectAndModuleName(projectName, projectName); err != nil {
				return err
			}

			rich.Println("[cyan]생성 중:[/cyan] %s (템플릿: %s)", projectName, tmplName)

			opts := generator.Options{Template: tmplName, SQLite: sqlite, Profile: profile}
			if err := generator.Generate(projectName, opts); err != nil {
				return err
			}

			rich.Println("[green][bold]완료![/bold][/green] %s 프로젝트가 생성되었습니다.", projectName)
			fmt.Printf("\n  cd %s\n  go mod tidy\n  go build .\n\n", projectName)
			rich.Println("[dim]wcli 라이브러리가 의존성으로 추가되었습니다.[/dim]")
			if sqlite {
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
	return cmd
}
