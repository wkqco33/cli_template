package cmd

import (
	"super_cli/config"
	"github.com/seoyc/wcli"
)

var cfg config.Config

// InitConfig wcli로 설정을 로드한다. root의 PersistentPreRun에서 호출.
func InitConfig() error {
	return wcli.Load(&cfg,
		wcli.WithFiles("config.yaml"),
		wcli.WithEnv(),
		wcli.WithPrefix("SUPER_CLI"),
	)
}
