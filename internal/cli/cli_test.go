package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wkqco33/cli_template/generator"
	"github.com/wkqco33/wcli"
)

// testEnv 실행 환경과 출력 버퍼를 묶어 커맨드 실행을 돕는다.
type testEnv struct {
	Env    *Env
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
}

// newTestEnv 비TTY(파이프) 환경을 기본으로 만든다.
func newTestEnv() *testEnv {
	var stdout, stderr bytes.Buffer
	return &testEnv{
		Env:    NewEnv(&stdout, &stderr, strings.NewReader("")),
		Stdout: &stdout,
		Stderr: &stderr,
	}
}

// interactive 테스트 환경을 대화형(TTY)으로 바꾸고 stdin을 주입한다.
func (te *testEnv) interactive(input string) *testEnv {
	te.Env.Interactive = true
	te.Env.Stdin = strings.NewReader(input)
	return te
}

// run 커맨드를 격리된 루트에 붙여 실행한다.
func (te *testEnv) run(t *testing.T, newSub func(*Env) *wcli.Command, args ...string) error {
	t.Helper()

	root := &wcli.Command{
		Use:           "wtemp",
		OutWriter:     te.Stdout,
		ErrWriter:     te.Stderr,
		SilenceErrors: true,
	}
	root.AddCommand(newSub(te.Env))
	return root.Execute(args)
}

// listSub ListCmd를 run 팩토리 시그니처에 맞춘다.
func listSub(env *Env) *wcli.Command { return ListCmd(env) }

// chdirTemp 테스트를 임시 디렉토리로 이동하고 종료 시 원복한다.
func chdirTemp(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	return tmp
}

func TestNewCmd_Structure(t *testing.T) {
	cmd := NewCmd(NewEnv(io.Discard, io.Discard, nil))

	if cmd.Use != "new <project-name>" {
		t.Fatalf("expected Use %q, got %q", "new <project-name>", cmd.Use)
	}
	if cmd.Short == "" {
		t.Fatal("expected non-empty Short description")
	}

	names := make(map[string]string)
	for _, f := range cmd.Flags().All() {
		names[f.Name] = f.Shorthand
	}
	for _, want := range []string{"template", "sqlite", "profile", "module", "force", "output", "dry-run", "git", "format"} {
		if _, ok := names[want]; !ok {
			t.Fatalf("expected flag %q to be registered, got %v", want, names)
		}
	}
	for flag, short := range map[string]string{
		"template": "t", "output": "o", "force": "f", "dry-run": "n",
	} {
		if names[flag] != short {
			t.Fatalf("expected -%s shorthand for --%s, got %q", short, flag, names[flag])
		}
	}
}

func TestNewCmd_NoArgsReturnsError(t *testing.T) {
	te := newTestEnv()

	err := te.run(t, NewCmd, "new")
	if err == nil {
		t.Fatal("expected error when no project name given")
	}
	if !strings.Contains(err.Error(), "프로젝트 이름을 입력하세요") {
		t.Fatalf("expected missing-name error, got %q", err.Error())
	}
	var usageErr *generator.UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected *UsageError, got %T", err)
	}
}

func TestNewCmd_InvalidNameReturnsValidationError(t *testing.T) {
	te := newTestEnv()

	err := te.run(t, NewCmd, "new", "My App")
	if err == nil {
		t.Fatal("expected validation error for invalid project name")
	}
	if !strings.Contains(err.Error(), "해결 방법:") {
		t.Fatalf("expected actionable validation guide, got %q", err.Error())
	}
	var usageErr *generator.UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected *UsageError, got %T", err)
	}
}

func TestNewCmd_RejectsExtraArgs(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	err := te.run(t, NewCmd, "new", "myapp", "extra", "-t", "minimal")
	if err == nil {
		t.Fatal("expected error for unexpected positional argument")
	}
	if !strings.Contains(err.Error(), "예상하지 못한 인자") {
		t.Fatalf("expected unexpected-arg error, got %q", err.Error())
	}
	if _, statErr := os.Stat("myapp"); !os.IsNotExist(statErr) {
		t.Fatalf("nothing should be generated, stat err: %v", statErr)
	}
}

func TestNewCmd_GeneratesProject(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("new 실행 실패: %v", err)
	}
	for _, p := range []string{"myapp", "myapp/main.go", "myapp/go.mod"} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected generated path %s: %v", p, err)
		}
	}
}

func TestNewCmd_ProgressGoesToStderr(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("new 실행 실패: %v", err)
	}
	if te.Stdout.Len() != 0 {
		t.Fatalf("expected empty stdout on success, got:\n%s", te.Stdout.String())
	}
	for _, want := range []string{"생성 중:", "완료!"} {
		if !strings.Contains(te.Stderr.String(), want) {
			t.Fatalf("expected stderr to contain %q, got:\n%s", want, te.Stderr.String())
		}
	}
}

func TestNewCmd_QuietSuppressesProgress(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()
	te.Env.Quiet = true

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("new 실행 실패: %v", err)
	}
	if te.Stderr.Len() != 0 {
		t.Fatalf("expected quiet stderr, got:\n%s", te.Stderr.String())
	}
}

func TestNewCmd_DebugPrintsDiagnostics(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()
	te.Env.Debug = true

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("new 실행 실패: %v", err)
	}
	if !strings.Contains(te.Stderr.String(), "debug:") {
		t.Fatalf("expected debug diagnostics, got:\n%s", te.Stderr.String())
	}
}

func TestNewCmd_ModuleFlag(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--module", "github.com/user/myapp"); err != nil {
		t.Fatalf("new 실행 실패: %v", err)
	}
	content, err := os.ReadFile(filepath.Join("myapp", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod failed: %v", err)
	}
	if !strings.Contains(string(content), "module github.com/user/myapp") {
		t.Fatalf("expected module path in go.mod, got:\n%s", string(content))
	}
}

func TestNewCmd_ForceFlag(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--force"); err != nil {
		t.Fatalf("new --force 실행 실패: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myapp", "main.go")); err != nil {
		t.Fatalf("expected generated main.go after force: %v", err)
	}
}

func TestNewCmd_OverwriteNoticeShowsAbsolutePath(t *testing.T) {
	tmp := chdirTemp(t)
	te := newTestEnv()
	te.Env.Yes = true

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("new --yes 실행 실패: %v", err)
	}
	if want := filepath.Join(tmp, "myapp"); !strings.Contains(te.Stderr.String(), want) {
		t.Fatalf("expected overwrite notice with absolute path %q, got:\n%s", want, te.Stderr.String())
	}
}

func TestNewCmd_OutputFlag(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--output", "sub/dir"); err != nil {
		t.Fatalf("new --output 실행 실패: %v", err)
	}
	if _, err := os.Stat(filepath.Join("sub", "dir", "myapp", "main.go")); err != nil {
		t.Fatalf("expected generated file in output dir: %v", err)
	}
}

func TestNewCmd_DryRunListsFilesToStdout(t *testing.T) {
	tmp := chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--dry-run", "-o", "nested/dir"); err != nil {
		t.Fatalf("new --dry-run 실행 실패: %v", err)
	}
	if !strings.Contains(te.Stdout.String(), "main.go") {
		t.Fatalf("expected dry-run file list on stdout, got:\n%s", te.Stdout.String())
	}
	if strings.Contains(te.Stdout.String(), "생성될 파일:") {
		t.Fatalf("expected header on stderr, not stdout:\n%s", te.Stdout.String())
	}
	if !strings.Contains(te.Stderr.String(), "생성될 파일:") {
		t.Fatalf("expected header on stderr, got:\n%s", te.Stderr.String())
	}
	if _, err := os.Stat("myapp"); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create target directory, stat err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "nested")); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create output directories, stat err: %v", err)
	}
}

func TestNewCmd_DryRunTableFormatIndents(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--dry-run", "--format", "table"); err != nil {
		t.Fatalf("new --dry-run --format table 실행 실패: %v", err)
	}
	if !strings.Contains(te.Stdout.String(), "\n  main.go\n") {
		t.Fatalf("expected indented table output, got:\n%s", te.Stdout.String())
	}
}

func TestNewCmd_DryRunJSONFormat(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--dry-run", "--format", "json"); err != nil {
		t.Fatalf("new --dry-run --format json 실행 실패: %v", err)
	}

	var got dryRunResult
	if err := json.Unmarshal(te.Stdout.Bytes(), &got); err != nil {
		t.Fatalf("expected valid JSON, got %q: %v", te.Stdout.String(), err)
	}
	if got.Project != "myapp" || got.Template != "minimal" || got.Module != "myapp" {
		t.Fatalf("unexpected JSON payload: %+v", got)
	}
	hasMain := false
	for _, f := range got.Files {
		if f == "main.go" {
			hasMain = true
		}
	}
	if !hasMain {
		t.Fatalf("expected main.go in JSON files, got %v", got.Files)
	}
	if strings.Contains(te.Stdout.String(), "생성될 파일:") {
		t.Fatalf("JSON output must stay machine-readable, got:\n%s", te.Stdout.String())
	}
}

func TestNewCmd_UnknownFormatRejected(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--dry-run", "--format", "xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	var usageErr *generator.UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected *UsageError, got %T: %v", err, err)
	}
}

func TestNewCmd_ProfileGoesToStderr(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "library", "--profile"); err != nil {
		t.Fatalf("new --profile 실행 실패: %v", err)
	}
	if te.Stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got:\n%s", te.Stdout.String())
	}
	if !strings.Contains(te.Stderr.String(), "[profile]") {
		t.Fatalf("expected profile output on stderr, got:\n%s", te.Stderr.String())
	}
}

func TestNewCmd_SQLiteUnsupportedWarning(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	for _, args := range [][]string{
		{"new", "myapp", "-t", "library", "--sqlite"},
		{"new", "myapp2", "-t", "library", "--sqlite", "--dry-run"},
	} {
		te = newTestEnv()
		if err := te.run(t, NewCmd, args...); err != nil {
			t.Fatalf("%v 실행 실패: %v", args, err)
		}
		if strings.Contains(te.Stdout.String(), "지원하지 않습니다") {
			t.Fatalf("expected warning on stderr, not stdout:\n%s", te.Stdout.String())
		}
		if !strings.Contains(te.Stderr.String(), "지원하지 않습니다") {
			t.Fatalf("expected sqlite-unsupported warning for %v, got:\n%s", args, te.Stderr.String())
		}
	}
}

func TestNewCmd_GitFlagInitializesRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git이 설치되어 있지 않아 건너뜁니다")
	}
	chdirTemp(t)
	te := newTestEnv()

	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal", "--git"); err != nil {
		t.Fatalf("new --git 실행 실패: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myapp", ".git")); err != nil {
		t.Fatalf("expected .git directory after --git: %v", err)
	}
}

func TestNewCmd_ExistingDirRequiresConfirmation(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal")
	if err == nil {
		t.Fatal("expected input-required error in non-interactive environment")
	}
	var inputErr *generator.InputRequiredError
	if !errors.As(err, &inputErr) {
		t.Fatalf("expected *InputRequiredError, got %T: %v", err, err)
	}
	if _, statErr := os.Stat(filepath.Join("myapp", "main.go")); !os.IsNotExist(statErr) {
		t.Fatalf("nothing should be generated, stat err: %v", statErr)
	}
}

func TestNewCmd_DeclinedConfirmationCancels(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv().interactive("n\n")

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("expected clean cancellation, got %v", err)
	}
	if !strings.Contains(te.Stderr.String(), "취소") {
		t.Fatalf("expected cancellation notice, got:\n%s", te.Stderr.String())
	}
	if _, err := os.Stat(filepath.Join("myapp", "main.go")); !os.IsNotExist(err) {
		t.Fatalf("nothing should be generated after declining, stat err: %v", err)
	}
}

func TestNewCmd_AcceptedConfirmationOverwrites(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv().interactive("y\n")

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join("myapp", "stale.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write stale file failed: %v", err)
	}
	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("expected overwrite after confirmation, got %v", err)
	}
	if _, err := os.Stat(filepath.Join("myapp", "main.go")); err != nil {
		t.Fatalf("expected generated main.go: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myapp", "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, stat err: %v", err)
	}
}

func TestNewCmd_NoInputFlagBlocksPrompt(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv().interactive("y\n")
	te.Env.NoInput = true

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal")
	var inputErr *generator.InputRequiredError
	if !errors.As(err, &inputErr) {
		t.Fatalf("expected *InputRequiredError with --no-input, got %T: %v", err, err)
	}
}

func TestNewCmd_YesFlagSkipsPrompt(t *testing.T) {
	chdirTemp(t)
	te := newTestEnv()
	te.Env.Yes = true

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := te.run(t, NewCmd, "new", "myapp", "-t", "minimal"); err != nil {
		t.Fatalf("new --yes 실행 실패: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myapp", "main.go")); err != nil {
		t.Fatalf("expected generated main.go: %v", err)
	}
}

func TestListCmd_Structure(t *testing.T) {
	cmd := ListCmd(NewEnv(io.Discard, io.Discard, nil))

	if cmd.Use != "list" {
		t.Fatalf("expected Use %q, got %q", "list", cmd.Use)
	}
	if cmd.Short == "" {
		t.Fatal("expected non-empty Short description")
	}
	names := make(map[string]bool)
	for _, f := range cmd.Flags().All() {
		names[f.Name] = true
	}
	if !names["format"] {
		t.Fatalf("expected --format flag, got %v", names)
	}
}

func TestListCmd_DefaultFormatIsPlainWhenNotTerminal(t *testing.T) {
	te := newTestEnv()

	if err := te.run(t, listSub, "list"); err != nil {
		t.Fatalf("list 실행 실패: %v", err)
	}
	for _, name := range []string{"minimal", "full", "gin", "fiber", "echo", "fyne", "library"} {
		if !strings.Contains(te.Stdout.String(), name) {
			t.Fatalf("expected list output to contain %q, got:\n%s", name, te.Stdout.String())
		}
	}
	if strings.Contains(te.Stdout.String(), "\x1b[") {
		t.Fatalf("expected no ANSI escape in non-terminal output, got:\n%q", te.Stdout.String())
	}
	if strings.Contains(te.Stdout.String(), "│") || strings.Contains(te.Stdout.String(), "+--") {
		t.Fatalf("expected plain output without box drawing, got:\n%s", te.Stdout.String())
	}
	if te.Stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got:\n%s", te.Stderr.String())
	}
}

func TestListCmd_JSONFormatSchema(t *testing.T) {
	te := newTestEnv()

	if err := te.run(t, listSub, "list", "--format", "json"); err != nil {
		t.Fatalf("list --format json 실행 실패: %v", err)
	}

	var got templateList
	if err := json.Unmarshal(te.Stdout.Bytes(), &got); err != nil {
		t.Fatalf("expected valid JSON, got %q: %v", te.Stdout.String(), err)
	}
	if len(got.Templates) != len(generator.Templates()) {
		t.Fatalf("expected %d templates, got %d", len(generator.Templates()), len(got.Templates))
	}
	want := make(map[string]bool)
	for _, meta := range generator.Templates() {
		want[meta.Name] = true
	}
	for _, info := range got.Templates {
		if !want[info.Name] {
			t.Fatalf("unexpected template %q in JSON output", info.Name)
		}
		if info.Desc == "" {
			t.Fatalf("expected non-empty description for %q", info.Name)
		}
	}
}

func TestListCmd_TableFormatHasBoxDrawing(t *testing.T) {
	te := newTestEnv()

	if err := te.run(t, listSub, "list", "--format", "table"); err != nil {
		t.Fatalf("list --format table 실행 실패: %v", err)
	}
	if !strings.Contains(te.Stdout.String(), "이름") {
		t.Fatalf("expected table header, got:\n%s", te.Stdout.String())
	}
}

func TestListCmd_UnknownFormatRejected(t *testing.T) {
	te := newTestEnv()

	err := te.run(t, listSub, "list", "--format", "yaml")
	var usageErr *generator.UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected *UsageError, got %T: %v", err, err)
	}
}
