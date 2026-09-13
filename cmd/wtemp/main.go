package main

import (
	"errors"
	"io"
	"os"
	"runtime/debug"

	"github.com/wkqco33/cli_template/generator"
	"github.com/wkqco33/cli_template/internal/cli"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

// version 릴리스 시 -ldflags "-X main.version=v1.2.3"로 주입한다.
var version = "dev"

// resolveVersion -X로 주입된 버전이 없으면 모듈 버전을 시도한다.
// `go install .../cmd/wtemp@v1.2.3`처럼 소스에서 빌드한 바이너리도 실제 버전을 보고한다.
func resolveVersion() string {
	moduleVersion := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		moduleVersion = info.Main.Version
	}
	return pickVersion(version, moduleVersion)
}

// pickVersion 주입된 버전 → 모듈 버전 → 기본값 순으로 선택한다.
func pickVersion(injected, moduleVersion string) string {
	if injected != "dev" {
		return injected
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}
	return injected
}

// 종료 코드. README의 "종료 코드" 표와 함께 유지한다.
const (
	exitOK          = 0
	exitError       = 1
	exitUsage       = 2
	exitConflict    = 3
	exitExternal    = 4
	exitInputNeeded = 5
)

const rootLong = `wcli 기반 Go CLI 프로젝트 템플릿 생성기.

사용법:
  wtemp <command> [flags]
  wtemp new <project-name> [flags]
  wtemp list [flags]

예시:
  wtemp list
  wtemp new my-app
  wtemp new my-app -t minimal --module github.com/user/my-app
  wtemp new my-app --dry-run --format json

종료 코드:
  0  성공
  1  일반 오류
  2  사용법·입력 검증 오류
  3  대상 경로 충돌
  4  외부 도구(git) 실행 실패
  5  비대화형 환경에서 확인 입력 필요

문서: https://github.com/wkqco33/cli_template
이슈: https://github.com/wkqco33/cli_template/issues`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run CLI를 실행하고 프로세스 종료 코드를 반환한다.
// 결과는 stdout, 진행·오류 메시지는 stderr로 보낸다.
func run(args []string, stdout, stderr io.Writer) int {
	env := cli.NewEnv(stdout, stderr, os.Stdin)
	root := newRootCmd(env)

	if err := root.Execute(args); err != nil {
		rich.Fprintln(env.Stderr, "[red][bold]Error:[/bold] %s[/red]", rich.EscapeMarkup(err.Error()))
		return exitCode(err)
	}
	return exitOK
}

// newRootCmd 하위 커맨드와 전역 플래그를 포함한 루트 커맨드를 구성한다.
func newRootCmd(env *cli.Env) *wcli.Command {
	root := &wcli.Command{
		Use:           "wtemp",
		Short:         "wcli 기반 Go CLI 프로젝트 템플릿 생성기",
		Long:          rootLong,
		Version:       resolveVersion(),
		OutWriter:     env.Stdout,
		ErrWriter:     env.Stderr,
		SilenceErrors: true,
	}

	flags := root.PersistentFlags()
	flags.BoolVar(&env.NoColor, "no-color", "", false, "색상 출력을 사용하지 않음")
	flags.BoolVar(&env.NoInput, "no-input", "", false, "대화형 입력을 받지 않음")
	flags.BoolVar(&env.Quiet, "quiet", "q", false, "진행 메시지 억제 (결과·오류는 유지)")
	flags.BoolVar(&env.Debug, "debug", "d", false, "추가 진단 메시지 출력")
	flags.BoolVar(&env.Yes, "yes", "y", false, "확인 프롬프트에 자동으로 yes")

	root.PersistentPreRun = func(ctx *wcli.Context) error {
		if env.NoColor {
			rich.NoColor = true
		}
		return nil
	}

	// 인자 없이 실행하면 도움말, 알 수 없는 명령이면 사용법 오류를 반환한다.
	root.Run = func(ctx *wcli.Context) error {
		if len(ctx.Args) > 0 {
			return generator.NewUsageError(
				"알 수 없는 명령: %s\n해결 방법: wtemp --help로 사용 가능한 명령을 확인하세요",
				ctx.Args[0],
			)
		}
		root.Help()
		return nil
	}

	root.AddCommand(
		cli.NewCmd(env),
		cli.ListCmd(env),
		wcli.NewCompletionCommand(root),
	)

	return root
}

// exitCode 오류를 문서화된 종료 코드로 매핑한다.
func exitCode(err error) int {
	var usageErr *generator.UsageError
	if errors.As(err, &usageErr) {
		return exitUsage
	}
	var flagErr *wcli.FlagError
	if errors.As(err, &flagErr) {
		return exitUsage
	}
	var validationErr *wcli.ValidationError
	if errors.As(err, &validationErr) {
		return exitUsage
	}
	var conflictErr *generator.ConflictError
	if errors.As(err, &conflictErr) {
		return exitConflict
	}
	var externalErr *generator.ExternalError
	if errors.As(err, &externalErr) {
		return exitExternal
	}
	var inputErr *generator.InputRequiredError
	if errors.As(err, &inputErr) {
		return exitInputNeeded
	}
	return exitError
}
