package cli

import (
	"fmt"

	"github.com/wkqco33/cli_template/generator"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

// TemplateInfo list 출력용 템플릿 정보
type TemplateInfo struct {
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	SQLite bool   `json:"sqlite"`
}

type templateList struct {
	Templates []TemplateInfo `json:"templates"`
}

const listLong = `사용 가능한 템플릿과 설명을 출력합니다.

사용법:
  wtemp list [flags]

예시:
  wtemp list
  wtemp list --format plain | cut -f1
  wtemp list --format json

문서: https://github.com/wkqco33/cli_template`

// ListCmd list 서브커맨드를 만든다. 템플릿 목록은 stdout으로 출력한다.
func ListCmd(env *Env) *wcli.Command {
	var format string

	cmd := &wcli.Command{
		Use:   "list",
		Short: "사용 가능한 템플릿 목록을 출력합니다",
		Long:  listLong,
		Run: func(ctx *wcli.Context) error {
			resolved, err := resolveFormat(format, isTerminal(env.Stdout))
			if err != nil {
				return err
			}

			infos := make([]TemplateInfo, 0, len(generator.Templates()))
			for _, t := range generator.Templates() {
				infos = append(infos, TemplateInfo{
					Name:   t.Name,
					Desc:   t.Desc,
					SQLite: generator.SQLiteSupported(t.Name),
				})
			}

			switch resolved {
			case FormatJSON:
				return writeJSON(env.Stdout, templateList{Templates: infos})
			case FormatPlain:
				for _, t := range infos {
					fmt.Fprintf(env.Stdout, "%s\t%s\n", t.Name, t.Desc)
				}
			default:
				table := rich.NewTable("이름", "설명")
				for _, t := range infos {
					table.AddRow(t.Name, t.Desc)
				}
				table.Render(env.Stdout)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "", "", formatUsage())
	return cmd
}
