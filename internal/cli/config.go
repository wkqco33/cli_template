package cli

import (
	"fmt"
	"os"

	"github.com/wkqco33/cli_template/config"
	"github.com/wkqco33/cli_template/generator"
	"github.com/wkqco33/wcli"
	"gopkg.in/yaml.v3"
)

const configLong = `wtemp 설정 파일을 관리합니다.

사용법:
  wtemp config <command> [flags]

명령:
  init       기본 설정 파일을 생성합니다
  path       설정 파일 경로를 출력합니다
  show       병합된 설정을 출력합니다
  set        설정 값을 변경합니다
  unset      설정 값을 기본값으로 되돌립니다
  validate   설정 파일을 검증합니다`

func ConfigCmd(env *Env) *wcli.Command {
	cmd := &wcli.Command{Use: "config", Short: "wtemp 설정 파일을 관리합니다", Long: configLong}
	cmd.AddCommand(configInitCmd(env), configPathCmd(env), configShowCmd(env), configSetCmd(env), configUnsetCmd(env), configValidateCmd(env))
	return cmd
}

func configPath(env *Env) (string, error) { return config.ResolvePath(env.ConfigPath) }

func configInitCmd(env *Env) *wcli.Command {
	var force bool
	cmd := &wcli.Command{Use: "init", Short: "기본 설정 파일을 생성합니다", Run: func(ctx *wcli.Context) error {
		path, err := configPath(env)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil && !force && !env.Yes {
			return &generator.ConflictError{Path: path}
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := config.Save(path, config.Defaults()); err != nil {
			return err
		}
		env.progress("[green]설정 파일이 생성되었습니다:[/green] %s", path)
		return nil
	}}
	cmd.Flags().BoolVar(&force, "force", "f", false, "기존 설정 파일을 덮어씁니다")
	return cmd
}

func configPathCmd(env *Env) *wcli.Command {
	return &wcli.Command{Use: "path", Short: "설정 파일 경로를 출력합니다", Run: func(ctx *wcli.Context) error {
		path, err := configPath(env)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(env.Stdout, path)
		return err
	}}
}

func configShowCmd(env *Env) *wcli.Command {
	var format string
	cmd := &wcli.Command{Use: "show", Short: "병합된 설정을 출력합니다", Run: func(ctx *wcli.Context) error {
		path, err := configPath(env)
		if err != nil {
			return err
		}
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		config.ApplyEnv(&cfg)
		config.Normalize(&cfg)
		if err := config.Validate(cfg); err != nil {
			return err
		}
		cfg = config.MaskSecrets(cfg)
		if format == "json" {
			return writeJSON(env.Stdout, cfg)
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("설정 출력 실패: %w", err)
		}
		_, err = env.Stdout.Write(data)
		return err
	}}
	cmd.Flags().StringVar(&format, "format", "", "yaml", "출력 형식 yaml 또는 json")
	return cmd
}

func configSetCmd(env *Env) *wcli.Command {
	return &wcli.Command{Use: "set <key> <value>", Short: "설정 값을 변경합니다", Run: func(ctx *wcli.Context) error {
		if len(ctx.Args) != 2 {
			return fmt.Errorf("사용법: wtemp config set <key> <value>")
		}
		path, err := configPath(env)
		if err != nil {
			return err
		}
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		if err := config.Set(&cfg, ctx.Args[0], ctx.Args[1]); err != nil {
			return err
		}
		if err := config.Save(path, cfg); err != nil {
			return err
		}
		env.progress("설정이 변경되었습니다: %s", ctx.Args[0])
		return nil
	}}
}

func configUnsetCmd(env *Env) *wcli.Command {
	return &wcli.Command{Use: "unset <key>", Short: "설정을 기본값으로 되돌립니다", Run: func(ctx *wcli.Context) error {
		if len(ctx.Args) != 1 {
			return fmt.Errorf("사용법: wtemp config unset <key>")
		}
		path, err := configPath(env)
		if err != nil {
			return err
		}
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		if err := config.Unset(&cfg, ctx.Args[0]); err != nil {
			return err
		}
		if err := config.Save(path, cfg); err != nil {
			return err
		}
		env.progress("설정이 초기화되었습니다: %s", ctx.Args[0])
		return nil
	}}
}

func configValidateCmd(env *Env) *wcli.Command {
	return &wcli.Command{Use: "validate", Short: "설정 파일을 검증합니다", Run: func(ctx *wcli.Context) error {
		path, err := configPath(env)
		if err != nil {
			return err
		}
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		config.ApplyEnv(&cfg)
		config.Normalize(&cfg)
		if err := config.Validate(cfg); err != nil {
			return err
		}
		env.progress("설정 파일이 유효합니다: %s", path)
		return nil
	}}
}
