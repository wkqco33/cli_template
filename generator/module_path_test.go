package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerate_NoModulePathDuplication --module에 전체 모듈 경로를 넘겼을 때
// 템플릿이 도메인 접두사를 다시 붙여 github.com/github.com/... 같은 경로를
// 만들지 않아야 한다. (library 템플릿이 github.com/을 하드코딩했던 회귀 방지)
func TestGenerate_NoModulePathDuplication(t *testing.T) {
	const moduleName = "github.com/user/sample"

	for _, meta := range Templates() {
		t.Run(meta.Name, func(t *testing.T) {
			tmp := chdirTemp(t)
			projectName := "sample-" + meta.Name

			if err := Generate(projectName, Options{Template: meta.Name, ModuleName: moduleName}); err != nil {
				t.Fatalf("generate failed (template=%s): %v", meta.Name, err)
			}

			err := filepath.WalkDir(filepath.Join(tmp, projectName), func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if strings.Contains(string(content), "github.com/github.com") {
					t.Errorf("module path duplicated in %s:\n%s", path, string(content))
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walk failed: %v", err)
			}
		})
	}
}

// TestGenerate_LibraryModuleName library 템플릿의 go.mod는 지정한 모듈 경로를
// 그대로 사용해야 한다. (기존 TestGenerate_ModuleName은 minimal만 검사했다)
func TestGenerate_LibraryModuleName(t *testing.T) {
	chdirTemp(t)
	const (
		projectName = "libx"
		moduleName  = "github.com/user/libx"
	)

	if err := Generate(projectName, Options{Template: "library", ModuleName: moduleName}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(projectName, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if lines[0] != "module "+moduleName {
		t.Fatalf("expected go.mod first line %q, got %q", "module "+moduleName, lines[0])
	}

	for _, rel := range []string{
		filepath.Join("README.md"),
		filepath.Join("lib_test.go"),
		filepath.Join("examples", "basic", "main.go"),
	} {
		b, err := os.ReadFile(filepath.Join(projectName, rel))
		if err != nil {
			t.Fatalf("read %s failed: %v", rel, err)
		}
		if !strings.Contains(string(b), moduleName) {
			t.Fatalf("expected %s to reference module path %q, got:\n%s", rel, moduleName, string(b))
		}
	}
}
