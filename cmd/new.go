package cmd

import (
	"fmt"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
	"cli_template/generator"
)

func NewCmd() *wcli.Command {
	var tmplName string
	var sqlite bool

	cmd := &wcli.Command{
		Use:   "new <project-name>",
		Short: "새 CLI 프로젝트를 생성합니다",
		Long:  "지정한 이름으로 wcli + wconf 기반 Go CLI 프로젝트를 생성합니다.",
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return fmt.Errorf("프로젝트 이름을 입력하세요: new <project-name>")
			}
			projectName := ctx.Args[0]

			rich.Println("[cyan]생성 중:[/cyan] %s (템플릿: %s)", projectName, tmplName)

			opts := generator.Options{Template: tmplName, SQLite: sqlite}
			if err := generator.Generate(projectName, opts); err != nil {
				return err
			}

			rich.Println("[green][bold]완료![/bold][/green] %s 프로젝트가 생성되었습니다.", projectName)
			fmt.Printf("\n  cd %s\n  go mod tidy\n  go build .\n\n", projectName)
			rich.Println("[dim]wcli, wconf 서브모듈이 자동으로 추가되었습니다.[/dim]")
			if sqlite {
				rich.Println("[dim]SQLite + GORM이 포함되었습니다. database/ 디렉토리를 확인하세요.[/dim]")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tmplName, "template", "t", "full", "사용할 템플릿 (minimal, full)")
	cmd.Flags().BoolVar(&sqlite, "sqlite", "", false, "SQLite + GORM 지원 추가")
	return cmd
}
