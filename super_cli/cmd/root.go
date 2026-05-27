package cmd

import (
	"github.com/seoyc/wcli"
	"super_cli/config"
)

var cfg config.Config

// Cfg 현재 설정값을 반환한다
func Cfg() *config.Config {
	return &cfg
}

// InitConfig wcli로 설정을 로드한다. root의 PersistentPreRun에서 호출.
func InitConfig() error {
	return wcli.Load(&cfg,
		wcli.WithFiles("config.yaml"),
		wcli.WithEnv(),
		wcli.WithPrefix("SUPER_CLI"),
	)
}
