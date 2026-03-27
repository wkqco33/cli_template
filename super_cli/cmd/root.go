package cmd

import (
	"super_cli/config"
	"wconf"
)

var cfg config.Config

// initConfig wconf로 설정을 로드한다. PersistentPreRun에서 호출.
func initConfig() error {
	return wconf.Load(&cfg,
		wconf.WithFiles("config.yaml"),
		wconf.WithEnv(),
		wconf.WithPrefix("SUPER_CLI"),
	)
}
