package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderTemplates_FailsWhenDestRootIsAFile 대상 루트가 파일이면(잘못된 --output)
// 디렉토리 생성 실패로 중단해야 한다.
func TestRenderTemplates_FailsWhenDestRootIsAFile(t *testing.T) {
	tmp := chdirTemp(t)
	filePath := filepath.Join(tmp, "afile")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	data := TemplateData{ProjectName: "app", ModuleName: "app"}
	err := renderTemplates(filePath, "minimal", data)
	if err == nil {
		t.Fatal("expected error when destination root is a file")
	}
	if !strings.Contains(err.Error(), "출력 디렉토리 생성 실패") {
		t.Fatalf("expected directory-creation error, got %q", err.Error())
	}
}

// TestRenderFile_FailsWhenDestinationIsADirectory 파일 자리에 디렉토리가 있으면 실패한다.
func TestRenderFile_FailsWhenDestinationIsADirectory(t *testing.T) {
	tmp := chdirTemp(t)
	dir := filepath.Join(tmp, "adir")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	data := TemplateData{ProjectName: "app", ModuleName: "app"}
	if err := renderFile("minimal/main.go.tmpl", dir, data); err == nil {
		t.Fatal("expected error when destination is a directory")
	}
}

// TestRenderFileTo_MissingNonTemplateSource .tmpl이 아닌 원본이 없으면 오류를 반환하고
// 아무것도 쓰지 않아야 한다.
func TestRenderFileTo_MissingNonTemplateSource(t *testing.T) {
	var buf bytes.Buffer

	err := renderFileTo(&buf, "minimal/README.md", TemplateData{ProjectName: "app"})
	if err == nil {
		t.Fatal("expected error for missing non-template source")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output on failure, got %q", buf.String())
	}
}

// TestCachedTemplate_MissingSource 임베드 FS에 없는 템플릿은 오류로 드러난다.
func TestCachedTemplate_MissingSource(t *testing.T) {
	if _, err := cachedTemplate("minimal/does-not-exist.tmpl"); err == nil {
		t.Fatal("expected error for missing template source")
	}
}

// TestCollectFiles_MissingRoot 없는 디렉토리는 오류로 드러난다.
func TestCollectFiles_MissingRoot(t *testing.T) {
	if _, err := collectFiles(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing root")
	}
}
