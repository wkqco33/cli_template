package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestTemplates_Catalog(t *testing.T) {
	got := Templates()
	if len(got) == 0 {
		t.Fatal("expected non-empty template catalog")
	}

	// 이름은 중복 없이 카탈로그 순서를 유지해야 한다.
	seen := make(map[string]bool)
	for _, meta := range got {
		if meta.Name == "" {
			t.Fatal("expected non-empty template name")
		}
		if meta.Desc == "" {
			t.Fatalf("expected non-empty description for %q", meta.Name)
		}
		if seen[meta.Name] {
			t.Fatalf("duplicate template name %q", meta.Name)
		}
		seen[meta.Name] = true
	}

	// 핵심 템플릿이 반드시 포함되어야 한다.
	for _, want := range []string{"minimal", "full", "gin", "fiber", "echo", "fyne", "library"} {
		if !seen[want] {
			t.Fatalf("expected template %q in catalog, got %v", want, got)
		}
	}
}

func TestTemplateNamesCSV_MatchesCatalog(t *testing.T) {
	names := TemplateNamesCSV()
	if names == "" {
		t.Fatal("expected non-empty CSV")
	}
	parts := strings.Split(names, ", ")
	if len(parts) != len(Templates()) {
		t.Fatalf("expected %d names in CSV, got %d: %q", len(Templates()), len(parts), names)
	}
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			t.Fatalf("expected no empty name in CSV, got %q", names)
		}
	}
}

func TestGenerate_UnknownTemplate(t *testing.T) {
	chdirTemp(t)
	err := Generate("sample", Options{Template: "does-not-exist"})
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "알 수 없는 템플릿") {
		t.Fatalf("expected unknown-template error, got %q", err.Error())
	}
}

func TestGenerate_InvalidName(t *testing.T) {
	chdirTemp(t)
	err := Generate("My App", Options{Template: "minimal"})
	if err == nil {
		t.Fatal("expected validation error for invalid name")
	}
	if !strings.Contains(err.Error(), "해결 방법:") {
		t.Fatalf("expected actionable validation guide, got %q", err.Error())
	}
}

func TestGenerate_ExistingDirectory(t *testing.T) {
	chdirTemp(t)
	if err := os.MkdirAll("sample", 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	err := Generate("sample", Options{Template: "minimal"})
	if err == nil {
		t.Fatal("expected error when target directory already exists")
	}
	if !strings.Contains(err.Error(), "이미 존재합니다") {
		t.Fatalf("expected existing-dir error, got %q", err.Error())
	}
}

func TestGenerate_SQLiteRendersDatabase(t *testing.T) {
	chdirTemp(t)
	projectName := "sqlite-app"
	if err := Generate(projectName, Options{Template: "minimal", SQLite: true}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	for _, p := range []string{
		filepath.Join(projectName, "database", "db.go"),
		filepath.Join(projectName, "database", "models", "example.go"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected sqlite path %s: %v", p, err)
		}
	}
}

func TestGenerate_NoSQLiteSkipsDatabase(t *testing.T) {
	chdirTemp(t)
	projectName := "plain-app"
	if err := Generate(projectName, Options{Template: "minimal", SQLite: false}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectName, "database")); !os.IsNotExist(err) {
		t.Fatalf("expected database/ to be skipped without sqlite, stat err: %v", err)
	}
}

// TestGenerate_ProfileWritesToWriter --profile은 지정한 writer로 단계 시간을 출력한다.
func TestGenerate_ProfileWritesToWriter(t *testing.T) {
	chdirTemp(t)
	var buf bytes.Buffer

	opts := Options{Template: "minimal", Profile: true, ProfileWriter: &buf}
	if err := Generate("profiled-app", opts); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	got := buf.String()
	for _, want := range []string{"[profile]", "total=", "render=", "postprocess="} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected profile output to contain %q, got %q", want, got)
		}
	}
}

// TestGenerate_OutputDirIsAFileFails --output이 파일을 가리키면 디렉토리 생성 실패로 중단한다.
func TestGenerate_OutputDirIsAFileFails(t *testing.T) {
	tmp := chdirTemp(t)
	filePath := filepath.Join(tmp, "afile")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	err := Generate("app", Options{Template: "minimal", OutputDir: filePath})
	if err == nil {
		t.Fatal("expected error when output path is a file")
	}
	// 파일을 출력 경로로 쓰면 대상 경로 확인 단계에서 ENOTDIR로 실패한다.
	if !strings.Contains(err.Error(), "afile") {
		t.Fatalf("expected error to name the offending path, got %q", err.Error())
	}
}
