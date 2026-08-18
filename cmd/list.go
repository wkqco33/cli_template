package cmd

import (
	"cli_template/generator"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

func ListCmd() *wcli.Command {
	return &wcli.Command{
		Use:   "list",
		Short: "사용 가능한 템플릿 목록을 출력합니다",
		Run: func(ctx *wcli.Context) error {
			table := rich.NewTable("이름", "설명")
			for _, t := range generator.Templates() {
				table.AddRow(t.Name, t.Desc)
			}
			table.Print()
			return nil
		},
	}
}
