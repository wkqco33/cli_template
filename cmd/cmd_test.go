package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wkqco33/wcli"
)

// captureStdout os.Stdout를 임시 파이프로 교체해 실행 중 출력을 캡처한다.
// rich.Println / fmt.Printf가 os.Stdout에 직접 쓰기 때문에 필요하다.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe 생성 실패: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	return <-done
}

// newRootCmd 테스트용 루트 커맨드에 하위 커맨드를 붙여 반환한다.
func newRootCmd(sub *wcli.Command) *wcli.Command {
	root := &wcli.Command{Use: "wtemp"}
	root.AddCommand(sub)
	return root
}

func TestNewCmd_Structure(t *testing.T) {
	cmd := NewCmd()

	if cmd.Use != "new <project-name>" {
		t.Fatalf("expected Use %q, got %q", "new <project-name>", cmd.Use)
	}
	if cmd.Short == "" {
		t.Fatal("expected non-empty Short description")
	}

	flags := cmd.Flags().All()
	names := make(map[string]bool)
	for _, f := range flags {
		names[f.Name] = true
	}
	for _, want := range []string{"template", "sqlite", "profile", "module", "force", "output", "dry-run", "git"} {
		if !names[want] {
			t.Fatalf("expected flag %q to be registered, got %v", want, names)
		}
	}
}

func TestNewCmd_NoArgsReturnsError(t *testing.T) {
	root := newRootCmd(NewCmd())
	err := root.Execute([]string{"new"})
	if err == nil {
		t.Fatal("expected error when no project name given")
	}
	if !strings.Contains(err.Error(), "프로젝트 이름을 입력하세요") {
		t.Fatalf("expected missing-name error, got %q", err.Error())
	}
}

func TestNewCmd_InvalidNameReturnsValidationError(t *testing.T) {
	root := newRootCmd(NewCmd())
	err := root.Execute([]string{"new", "My App"})
	if err == nil {
		t.Fatal("expected validation error for invalid project name")
	}
	if !strings.Contains(err.Error(), "해결 방법:") {
		t.Fatalf("expected actionable validation guide, got %q", err.Error())
	}
}

func TestNewCmd_GeneratesProject(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	root := newRootCmd(NewCmd())
	if err := root.Execute([]string{"new", "myapp", "-t", "minimal"}); err != nil {
		t.Fatalf("new 실행 실패: %v", err)
	}

	for _, p := range []string{"myapp", "myapp/main.go", "myapp/go.mod"} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected generated path %s: %v", p, err)
		}
	}
}

func TestNewCmd_ModuleFlag(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	root := newRootCmd(NewCmd())
	if err := root.Execute([]string{"new", "myapp", "-t", "minimal", "--module", "github.com/user/myapp"}); err != nil {
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
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	if err := os.MkdirAll("myapp", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	root := newRootCmd(NewCmd())
	if err := root.Execute([]string{"new", "myapp", "-t", "minimal", "--force"}); err != nil {
		t.Fatalf("new --force 실행 실패: %v", err)
	}
	if _, err := os.Stat(filepath.Join("myapp", "main.go")); err != nil {
		t.Fatalf("expected generated main.go after force: %v", err)
	}
}

func TestNewCmd_OutputFlag(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	root := newRootCmd(NewCmd())
	if err := root.Execute([]string{"new", "myapp", "-t", "minimal", "--output", "sub/dir"}); err != nil {
		t.Fatalf("new --output 실행 실패: %v", err)
	}
	if _, err := os.Stat(filepath.Join("sub", "dir", "myapp", "main.go")); err != nil {
		t.Fatalf("expected generated file in output dir: %v", err)
	}
}

func TestNewCmd_DryRunFlag(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	root := newRootCmd(NewCmd())
	out := captureStdout(t, func() {
		if err := root.Execute([]string{"new", "myapp", "-t", "minimal", "--dry-run"}); err != nil {
			t.Fatalf("new --dry-run 실행 실패: %v", err)
		}
	})
	if !strings.Contains(out, "main.go") {
		t.Fatalf("expected dry-run output to list main.go, got:\n%s", out)
	}
	if _, err := os.Stat("myapp"); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create target directory, stat err: %v", err)
	}
}

func TestNewCmd_SQLiteUnsupportedWarning(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	root := newRootCmd(NewCmd())
	out := captureStdout(t, func() {
		if err := root.Execute([]string{"new", "myapp", "-t", "library", "--sqlite"}); err != nil {
			t.Fatalf("new 실행 실패: %v", err)
		}
	})
	if !strings.Contains(out, "지원하지 않습니다") {
		t.Fatalf("expected sqlite-unsupported warning, got:\n%s", out)
	}
}

func TestListCmd_Structure(t *testing.T) {
	cmd := ListCmd()
	if cmd.Use != "list" {
		t.Fatalf("expected Use %q, got %q", "list", cmd.Use)
	}
	if cmd.Short == "" {
		t.Fatal("expected non-empty Short description")
	}
}

func TestListCmd_PrintsTemplateNames(t *testing.T) {
	root := newRootCmd(ListCmd())

	out := captureStdout(t, func() {
		if err := root.Execute([]string{"list"}); err != nil {
			t.Fatalf("list 실행 실패: %v", err)
		}
	})

	for _, name := range []string{"minimal", "full", "gin", "fiber", "echo", "fyne", "library"} {
		if !strings.Contains(out, name) {
			t.Fatalf("expected list output to contain %q, got:\n%s", name, out)
		}
	}
}
