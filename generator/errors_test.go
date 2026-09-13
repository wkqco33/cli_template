package generator

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// TestResolve_ValidationErrorsAreUsageErrors 입력 검증 실패는 종료 코드 2로
// 매핑될 수 있도록 UsageError여야 한다.
func TestResolve_ValidationErrorsAreUsageErrors(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		opts        Options
	}{
		{name: "invalid-project-name", projectName: "My App", opts: Options{Template: "minimal"}},
		{name: "empty-project-name", projectName: "", opts: Options{Template: "minimal"}},
		{name: "invalid-module", projectName: "myapp", opts: Options{Template: "minimal", ModuleName: "bad module"}},
		{name: "unknown-template", projectName: "myapp", opts: Options{Template: "nope"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Resolve(tc.projectName, tc.opts)
			if err == nil {
				t.Fatal("expected error")
			}
			var usageErr *UsageError
			if !errors.As(err, &usageErr) {
				t.Fatalf("expected *UsageError, got %T: %v", err, err)
			}
			if !strings.Contains(err.Error(), "해결 방법:") && tc.name != "unknown-template" {
				t.Fatalf("expected actionable guide, got %q", err.Error())
			}
		})
	}
}

// TestResolve_ReturnsTargetPath Resolve는 검증된 모듈 이름과 대상 경로를 돌려준다.
func TestResolve_ReturnsTargetPath(t *testing.T) {
	got, err := Resolve("myapp", Options{Template: "minimal", OutputDir: "out"})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if got.ModuleName != "myapp" {
		t.Fatalf("expected module name %q, got %q", "myapp", got.ModuleName)
	}
	if got.TargetPath != "out/myapp" {
		t.Fatalf("expected target path %q, got %q", "out/myapp", got.TargetPath)
	}
}

// TestGenerate_ExistingDirectoryIsConflictError 충돌은 종료 코드 3으로 매핑된다.
func TestGenerate_ExistingDirectoryIsConflictError(t *testing.T) {
	chdirTemp(t)
	if err := os.MkdirAll("sample", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	err := Generate("sample", Options{Template: "minimal"})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected *ConflictError, got %T: %v", err, err)
	}
	if conflict.Path != "sample" {
		t.Fatalf("expected conflict path %q, got %q", "sample", conflict.Path)
	}
}

// TestInitGit_FailureIsExternalError 외부 도구 실패는 종료 코드 4로 매핑된다.
func TestInitGit_FailureIsExternalError(t *testing.T) {
	t.Setenv("PATH", "")

	err := InitGit(t.TempDir())
	if err == nil {
		t.Fatal("expected git init failure")
	}
	var external *ExternalError
	if !errors.As(err, &external) {
		t.Fatalf("expected *ExternalError, got %T: %v", err, err)
	}
	if external.Tool != "git init" {
		t.Fatalf("expected tool %q, got %q", "git init", external.Tool)
	}
}

// TestInputRequiredError_HasActionableGuide 비대화형 확인 필요 오류는 해결 방법을 안내한다.
func TestInputRequiredError_HasActionableGuide(t *testing.T) {
	err := &InputRequiredError{Path: "myapp"}
	for _, want := range []string{"myapp", "--force", "--yes"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected %q in %q", want, err.Error())
		}
	}
}
