package generator

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"text/template"
)

// TestDryRun_CreatesNoDirectories --dry-run은 파일 목록만 계산하고
// 중간 디렉터리를 포함해 파일시스템에 아무것도 만들지 않아야 한다.
func TestDryRun_CreatesNoDirectories(t *testing.T) {
	tmp := chdirTemp(t)

	files, err := DryRun("myapp", Options{Template: "minimal", OutputDir: filepath.Join("nested", "dir")})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected non-empty file list from dry run")
	}

	if _, err := os.Stat(filepath.Join(tmp, "nested")); !os.IsNotExist(err) {
		t.Fatalf("dry run must not create output directories, stat err: %v", err)
	}

	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	for _, entry := range entries {
		t.Fatalf("dry run must not create any entry, found %s", entry.Name())
	}
}

// TestDryRun_FileListMatchesGenerate dry-run 목록과 실제 생성 결과가
// 드리프트하지 않아야 한다. (sqlite on/off 포함)
func TestDryRun_FileListMatchesGenerate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		sqlite   bool
	}{
		{name: "minimal", template: "minimal"},
		{name: "minimal-sqlite", template: "minimal", sqlite: true},
		{name: "library", template: "library"},
		{name: "full-sqlite", template: "full", sqlite: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp := chdirTemp(t)
			opts := Options{Template: tc.template, SQLite: tc.sqlite}

			planned, err := DryRun("planned", opts)
			if err != nil {
				t.Fatalf("dry run failed: %v", err)
			}

			if err := Generate("actual", opts); err != nil {
				t.Fatalf("generate failed: %v", err)
			}
			actual, err := collectFiles(filepath.Join(tmp, "actual"))
			if err != nil {
				t.Fatalf("collect files failed: %v", err)
			}

			sort.Strings(planned)
			sort.Strings(actual)
			if !reflect.DeepEqual(planned, actual) {
				t.Fatalf("dry-run list != generated files\n dry-run: %v\n actual:  %v", planned, actual)
			}
		})
	}
}

// TestDryRun_SurfacesTemplateExecutionErrors dry-run은 파일 목록만 세는 것이
// 아니라 템플릿을 실제로 실행해 검증해야 한다(디스크에는 쓰지 않음).
func TestDryRun_SurfacesTemplateExecutionErrors(t *testing.T) {
	const srcPath = "minimal/main.go.tmpl"
	bad, err := template.New("bad").Parse("{{.NoSuchField}}")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	cachedTemplateForTest(t, srcPath, bad)

	if _, err := DryRun("myapp", Options{Template: "minimal"}); err == nil {
		t.Fatal("expected template execution error from dry run")
	}
}

// cachedTemplateForTest 지정한 임베드 경로의 파싱 캐시를 일시적으로 교체한다.
func cachedTemplateForTest(t *testing.T, srcPath string, tmpl *template.Template) {
	t.Helper()

	prev, had := parsedTemplateCache.Load(srcPath)
	parsedTemplateCache.Store(srcPath, tmpl)
	t.Cleanup(func() {
		if had {
			parsedTemplateCache.Store(srcPath, prev)
			return
		}
		parsedTemplateCache.Delete(srcPath)
	})
}
