package generator

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerate_ForceRenderFailureKeepsExistingDirectory --force는 렌더링이
// 실패하면 기존 디렉토리를 그대로 남겨야 한다(비원자적 삭제 방지).
func TestGenerate_ForceRenderFailureKeepsExistingDirectory(t *testing.T) {
	tmp := chdirTemp(t)
	projectName := "myapp"
	stalePath := filepath.Join(projectName, "stale.txt")

	if err := os.MkdirAll(projectName, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(stalePath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write stale file failed: %v", err)
	}

	origRender := renderTemplatesFunc
	renderTemplatesFunc = func(destRoot, tmplName string, data TemplateData) error {
		if err := os.MkdirAll(destRoot, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(destRoot, "partial.txt"), []byte("partial"), 0o644); err != nil {
			return err
		}
		return errors.New("injected render failure")
	}
	t.Cleanup(func() { renderTemplatesFunc = origRender })

	err := Generate(projectName, Options{Template: "minimal", Force: true})
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(err.Error(), "템플릿 렌더링 단계 실패") {
		t.Fatalf("expected wrapped context error, got %q", err.Error())
	}

	content, readErr := os.ReadFile(stalePath)
	if readErr != nil {
		t.Fatalf("existing directory must be preserved, read failed: %v", readErr)
	}
	if string(content) != "old" {
		t.Fatalf("expected stale content %q, got %q", "old", string(content))
	}
	if _, statErr := os.Stat(filepath.Join(projectName, "partial.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("partial output must not be installed, stat err: %v", statErr)
	}

	assertNoTemporaryArtifacts(t, tmp, projectName)
}

// TestGenerate_ForceSuccessRemovesBackup --force 성공 후 백업/임시 디렉토리가 남지 않아야 한다.
func TestGenerate_ForceSuccessRemovesBackup(t *testing.T) {
	tmp := chdirTemp(t)
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
	if _, err := os.Stat(filepath.Join(projectName, "main.go")); err != nil {
		t.Fatalf("expected generated main.go: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectName, "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, stat err: %v", err)
	}

	assertNoTemporaryArtifacts(t, tmp, projectName)
}

// assertNoTemporaryArtifacts 대상 프로젝트 외에 임시/백업 잔여물이 없어야 한다.
func assertNoTemporaryArtifacts(t *testing.T, root, keep string) {
	t.Helper()

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == keep {
			continue
		}
		t.Fatalf("unexpected leftover artifact: %s", entry.Name())
	}
}
