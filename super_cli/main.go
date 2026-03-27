package main

import (
	"fmt"
	"os"

	"github.com/seoyc/wcli"
	"super_cli/cmd"
)

const version = "0.1.0"

func main() {
	root := &wcli.Command{
		Use:     "super_cli",
		Short:   "super_cli CLI 도구",
		Version: version,
	}

	root.AddCommand(
		cmd.VersionCmd(version),
		cmd.ExampleCmd(),
	)

	if err := root.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
