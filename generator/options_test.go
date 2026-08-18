package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateProjectAndModuleName_ModulePath(t *testing.T) {
	valid := []string{
		"github.com/user/my-app",
		"example.com/org/tool",
		"gitlab.com/a/b/c",
	}
	for _, m := range valid {
		t.Run(m, func(t *testing.T) {
			if err := ValidateProjectAndModuleName("myapp", m); err != nil {
				t.Fatalf("expected valid module path %q, got error: %v", m, err)
			}
		})
	}

	invalid := []struct {
		module      string
		wantContain string
	}{
		{module: "github.com//user", wantContain: "빈 경로 세그먼트"},
		{module: "/github.com/user", wantContain: "시작하거나 끝날 수 없습니다"},
		{module: "github.com/user/", wantContain: "시작하거나 끝날 수 없습니다"},
		{module: "github.com/User", wantContain: "형식이 올바르지 않습니다"},
	}
	for _, tc := range invalid {
		t.Run(tc.module, func(t *testing.T) {
			err := ValidateProjectAndModuleName("myapp", tc.module)
			if err == nil {
				t.Fatalf("expected error for module %q", tc.module)
			}
			if !strings.Contains(err.Error(), tc.wantContain) {
				t.Fatalf("expected error to contain %q, got %q", tc.wantContain, err.Error())
			}
		})
	}
}

func TestGenerate_ModuleName(t *testing.T) {
	chdirTemp(t)
	projectName := "myapp"
	moduleName := "github.com/user/myapp"
	if err := Generate(projectName, Options{Template: "minimal", ModuleName: moduleName}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(projectName, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod failed: %v", err)
	}
	if !strings.Contains(string(content), "module "+moduleName) {
		t.Fatalf("expected go.mod to contain %q, got:\n%s", "module "+moduleName, string(content))
	}
}

func TestGenerate_ForceOverwritesExisting(t *testing.T) {
	chdirTemp(t)
	projectName := "myapp"
	if err := os.MkdirAll(projectName, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectName, "stale.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write stale file failed: %v", err)
	}

	if err := Generate(projectName, Options{Template: "minimal", Force: true}); err != nil {
		t.Fatalf("generate with force failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectName, "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, stat err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectName, "main.go")); err != nil {
		t.Fatalf("expected generated main.go: %v", err)
	}
}

func TestGenerate_OutputDir(t *testing.T) {
	chdirTemp(t)
	projectName := "myapp"
	outputDir := filepath.Join("nested", "dir")
	if err := Generate(projectName, Options{Template: "minimal", OutputDir: outputDir}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	for _, p := range []string{
		filepath.Join(outputDir, projectName, "main.go"),
		filepath.Join(outputDir, projectName, "go.mod"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected generated path %s: %v", p, err)
		}
	}
}

func TestDryRun_ReturnsFilesWithoutCreatingTarget(t *testing.T) {
	chdirTemp(t)
	projectName := "myapp"
	files, err := DryRun(projectName, Options{Template: "minimal"})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected non-empty file list from dry run")
	}
	hasMain := false
	for _, f := range files {
		if f == "main.go" {
			hasMain = true
		}
	}
	if !hasMain {
		t.Fatalf("expected main.go in dry-run list, got %v", files)
	}
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		t.Fatalf("dry run must not create target directory, stat err: %v", err)
	}
}

func TestDryRun_UnknownTemplate(t *testing.T) {
	chdirTemp(t)
	_, err := DryRun("myapp", Options{Template: "nope"})
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "알 수 없는 템플릿") {
		t.Fatalf("expected unknown-template error, got %q", err.Error())
	}
}

func TestSQLiteSupported(t *testing.T) {
	tests := []struct {
		template string
		want     bool
	}{
		{template: "minimal", want: true},
		{template: "full", want: true},
		{template: "gin", want: true},
		{template: "fiber", want: true},
		{template: "echo", want: true},
		{template: "fyne", want: false},
		{template: "library", want: false},
		{template: "unknown", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.template, func(t *testing.T) {
			if got := SQLiteSupported(tc.template); got != tc.want {
				t.Fatalf("SQLiteSupported(%q) = %v, want %v", tc.template, got, tc.want)
			}
		})
	}
}

func TestInitGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git이 설치되어 있지 않아 건너뜁니다")
	}
	dir := t.TempDir()
	if err := InitGit(dir); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf("expected .git directory after init: %v", err)
	}
}
