package generator

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_CleansTemporaryArtifactsOnFailure(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	origRender := renderTemplatesFunc
	origInit := initSubmodulesFunc
	renderTemplatesFunc = func(projectName, tmplName string, data TemplateData) error {
		if err := os.MkdirAll(projectName, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(projectName, "marker.txt"), []byte("partial"), 0o644)
	}
	initSubmodulesFunc = func(projectName string, meta TemplateMeta) error {
		return errors.New("injected submodule failure")
	}
	t.Cleanup(func() {
		renderTemplatesFunc = origRender
		initSubmodulesFunc = origInit
	})

	err = Generate("sample", Options{Template: "library"})
	if err == nil {
		t.Fatal("expected failure")
	}
	if !strings.Contains(err.Error(), "서브모듈 초기화 단계 실패") {
		t.Fatalf("expected wrapped context error, got %q", err.Error())
	}

	if _, statErr := os.Stat("sample"); !os.IsNotExist(statErr) {
		t.Fatalf("expected no final output directory, got stat err: %v", statErr)
	}

	entries, err := os.ReadDir(tmpRoot)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".sample.tmp-") {
			t.Fatalf("temporary dir should be removed, found %s", entry.Name())
		}
	}
}

func TestGenerate_SuccessKeepsExpectedOutputStructure(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmpRoot := t.TempDir()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	projectName := "libsample"
	if err := Generate(projectName, Options{Template: "library"}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	expectedPaths := []string{
		filepath.Join(projectName, "README.md"),
		filepath.Join(projectName, "go.mod"),
		filepath.Join(projectName, "lib.go"),
		filepath.Join(projectName, "lib_test.go"),
		filepath.Join(projectName, "examples", "basic", "main.go"),
	}
	for _, p := range expectedPaths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected generated path %s: %v", p, err)
		}
	}
}
