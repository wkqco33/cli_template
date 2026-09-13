package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wkqco33/cli_template/generator"
	"github.com/wkqco33/wcli"
)

const newLong = `지정한 이름으로 wcli 기반 Go CLI 프로젝트를 생성합니다.

사용법:
  wtemp new <project-name> [flags]

예시:
  wtemp new my-app
  wtemp new my-app -t minimal --module github.com/user/my-app
  wtemp new my-app --dry-run --format json
  wtemp new my-app --sqlite -t gin --git

대상 디렉토리가 이미 있으면 확인을 요청한다. 비대화형(파이프·CI) 환경에서는
--force로 덮어쓰거나 --yes로 확인을 생략해야 한다.

문서: https://github.com/wkqco33/cli_template`

// dryRunResult new --dry-run의 기계 판독 결과
type dryRunResult struct {
	Project  string   `json:"project"`
	Template string   `json:"template"`
	Module   string   `json:"module"`
	SQLite   bool     `json:"sqlite"`
	Files    []string `json:"files"`
}

// NewCmd new 서브커맨드를 만든다. 생성 결과는 stdout,
// 진행·경고·안내 메시지는 stderr로 출력한다.
func NewCmd(env *Env) *wcli.Command {
	var tmplName string
	var sqlite bool
	var profile bool
	var moduleName string
	var force bool
	var outputDir string
	var dryRun bool
	var git bool
	var format string

	cmd := &wcli.Command{
		Use:   "new <project-name>",
		Short: "새 CLI 프로젝트를 생성합니다",
		Long:  newLong,
		Run: func(ctx *wcli.Context) error {
			if len(ctx.Args) == 0 {
				return generator.NewUsageError("프로젝트 이름을 입력하세요.\n사용법: wtemp new <project-name> [flags]")
			}
			if len(ctx.Args) > 1 {
				return generator.NewUsageError(
					"예상하지 못한 인자입니다: %s\n해결 방법: 프로젝트 이름 하나만 넘기고 나머지는 플래그로 지정하세요. (예: --module, --template)",
					strings.Join(ctx.Args[1:], " "),
				)
			}
			projectName := ctx.Args[0]

			opts := generator.Options{
				Template:      tmplName,
				SQLite:        sqlite,
				Profile:       profile,
				ProfileWriter: env.Stderr,
				ModuleName:    moduleName,
				OutputDir:     outputDir,
				DryRun:        dryRun,
			}

			resolved, err := generator.Resolve(projectName, opts)
			if err != nil {
				return err
			}
			env.debug("template=%s module=%s target=%s", tmplName, resolved.ModuleName, resolved.TargetPath)

			if sqlite && !generator.SQLiteSupported(tmplName) {
				env.notice("[yellow]주의:[/yellow] %s 템플릿은 --sqlite를 지원하지 않습니다. 옵션이 무시되었습니다.", tmplName)
			}

			if dryRun {
				files, err := generator.DryRun(projectName, opts)
				if err != nil {
					return err
				}
				resolvedFormat, err := resolveFormat(format, isTerminal(env.Stdout))
				if err != nil {
					return err
				}
				return renderDryRun(env, resolvedFormat, projectName, tmplName, resolved.ModuleName, sqlite, files)
			}

			proceed, err := confirmOverwrite(env, resolved.TargetPath, force)
			if err != nil {
				return err
			}
			if !proceed {
				return nil
			}
			opts.Force = true

			env.progress("[cyan]생성 중:[/cyan] %s (템플릿: %s)", projectName, tmplName)
			if err := generator.Generate(projectName, opts); err != nil {
				return err
			}

			if git {
				if err := generator.InitGit(resolved.TargetPath); err != nil {
					return err
				}
				env.progress("[dim]git 저장소가 초기화되었습니다.[/dim]")
			}

			env.progress("[green][bold]완료![/bold][/green] %s 프로젝트가 생성되었습니다.", projectName)
			env.progress("\n  cd %s\n  go mod tidy\n  go build .\n", resolved.TargetPath)
			env.progress("[dim]wcli 라이브러리가 의존성으로 추가되었습니다.[/dim]")
			if sqlite && generator.SQLiteSupported(tmplName) {
				env.progress("[dim]SQLite + GORM이 포함되었습니다. database/ 디렉토리를 확인하세요.[/dim]")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(
		&tmplName,
		"template",
		"t",
		"full",
		fmt.Sprintf("사용할 템플릿 (%s)", generator.TemplateNamesCSV()),
	)
	cmd.Flags().BoolVar(&sqlite, "sqlite", "", false, "SQLite + GORM 지원 추가")
	cmd.Flags().BoolVar(&profile, "profile", "", false, "생성 단계별 성능 프로파일 출력")
	cmd.Flags().StringVar(&moduleName, "module", "", "", "Go 모듈 경로 (기본값: 프로젝트 이름)")
	cmd.Flags().BoolVar(&force, "force", "f", false, "대상 디렉토리가 이미 존재해도 덮어쓰기")
	cmd.Flags().StringVar(&outputDir, "output", "o", "", "생성 위치 (기본값: 현재 디렉토리)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", "n", false, "실제 생성 없이 생성될 파일 목록만 출력")
	cmd.Flags().BoolVar(&git, "git", "", false, "생성 후 git 저장소 초기화")
	cmd.Flags().StringVar(&format, "format", "", "", formatUsage())
	return cmd
}

// renderDryRun dry-run 결과를 지정한 형식으로 출력한다.
func renderDryRun(env *Env, format, projectName, tmplName, moduleName string, sqlite bool, files []string) error {
	switch format {
	case FormatJSON:
		return writeJSON(env.Stdout, dryRunResult{
			Project:  projectName,
			Template: tmplName,
			Module:   moduleName,
			SQLite:   sqlite,
			Files:    files,
		})
	case FormatPlain:
		env.progress("[cyan]생성될 파일:[/cyan] %s (템플릿: %s)", projectName, tmplName)
		for _, f := range files {
			fmt.Fprintln(env.Stdout, f)
		}
	default:
		env.progress("[cyan]생성될 파일:[/cyan] %s (템플릿: %s)", projectName, tmplName)
		for _, f := range files {
			fmt.Fprintf(env.Stdout, "  %s\n", f)
		}
	}
	return nil
}

// confirmOverwrite 대상이 이미 존재할 때 덮어쓸지 결정한다.
// proceed가 false이고 err가 nil이면 호출자는 아무것도 하지 않고 정상 종료한다.
func confirmOverwrite(env *Env, targetPath string, force bool) (proceed bool, err error) {
	exists, err := pathExists(targetPath)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}

	// 되돌릴 수 없는 교체이므로 대상을 절대 경로로 보여준다.
	env.notice("[yellow]덮어쓰기:[/yellow] 기존 %s 를 새 프로젝트로 교체합니다.", absolutePath(targetPath))

	if force || env.Yes {
		return true, nil
	}

	if !env.canPrompt() {
		return false, &generator.InputRequiredError{Path: targetPath}
	}

	ok, err := env.confirm(fmt.Sprintf("%s 디렉토리가 이미 존재합니다. 덮어쓸까요?", targetPath), false)
	if err != nil {
		return false, err
	}
	if !ok {
		env.notice("[yellow]취소되었습니다.[/yellow] 기존 파일을 변경하지 않았습니다.")
		return false, nil
	}
	return true, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("대상 경로 확인 실패 (%s): %w", path, err)
}

// absolutePath 절대 경로를 구하지 못하면 원래 값을 그대로 돌려준다.
func absolutePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
