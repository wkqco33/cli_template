package main

import (
	"fmt"
	"os"

	"github.com/seoyc/wcli"
	"cli_template/cmd"
)

const version = "0.1.0"

func main() {
	root := &wcli.Command{
		Use:     "wtemp",
		Short:   "wcli + wconf 기반 Go CLI 프로젝트 템플릿 생성기",
		Version: version,
	}

	root.AddCommand(
		cmd.NewCmd(),
		cmd.ListCmd(),
	)

	if err := root.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
