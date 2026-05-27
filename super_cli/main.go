package main

import (
	"fmt"
	"os"

	"github.com/seoyc/wcli"
	"github.com/seoyc/wcli/logging"
	"super_cli/cmd"
)

const version = "0.1.0"

func main() {
	logger := logging.NewDefaultLogger(os.Stderr, logging.LevelInfo, true)
	logging.SetLogger(logger)

	root := &wcli.Command{
		Use:     "super_cli",
		Short:   "super_cli CLI 도구",
		Version: version,
		PersistentPreRun: func(ctx *wcli.Context) error {
			return cmd.InitConfig()
		},
	}

	root.AddCommand(
		cmd.VersionCmd(version),
		cmd.ExampleCmd(),
		wcli.NewCompletionCommand(root),
	)

	if err := root.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
