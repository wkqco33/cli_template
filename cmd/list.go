package cmd

import (
	"fmt"

	"cli_template/generator"
	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

func ListCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "list",
		Short: "사용 가능한 템플릿 목록을 출력합니다",
		Run: func(ctx *wcli.Context) error {
			rich.Println("[bold][cyan]사용 가능한 템플릿:[/cyan][/bold]")
			for _, t := range generator.Templates() {
				fmt.Printf("  %-12s %s\n", t.Name, t.Desc)
			}
			return nil
		},
	}
}
