package cmd

import (
	"fmt"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/rich"
)

var availableTemplates = []struct {
	Name string
	Desc string
}{
	{"minimal", "루트 커맨드만 있는 최소 구조"},
	{"full", "서브커맨드 + wconf 설정이 포함된 전체 구조"},
}

func ListCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "list",
		Short: "사용 가능한 템플릿 목록을 출력합니다",
		Run: func(ctx *wcli.Context) error {
			rich.Println("[bold][cyan]사용 가능한 템플릿:[/cyan][/bold]")
			for _, t := range availableTemplates {
				fmt.Printf("  %-12s %s\n", t.Name, t.Desc)
			}
			return nil
		},
	}
}
