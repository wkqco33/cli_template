package cmd

import (
	"fmt"

	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

func ExampleCmd() *wcli.Command {
	var name string
	var verbose bool

	cmd := &wcli.Command{
		Use:   "example",
		Short: "예시 커맨드",
		Run: func(ctx *wcli.Context) error {
			rich.Println("[green]실행 중:[/green] example 커맨드")
			if name != "" {
				fmt.Printf("이름: %s\n", name)
			}
			if verbose {
				fmt.Printf("설정: %+v\n", cfg)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "n", "", "대상 이름")
	cmd.Flags().BoolVar(&verbose, "verbose", "v", false, "상세 출력")
	return cmd
}
