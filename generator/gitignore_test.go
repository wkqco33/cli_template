package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate_IncludesGitignore(t *testing.T) {
	chdirTemp(t)
	projectName := "myapp"
	if err := Generate(projectName, Options{Template: "minimal"}); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectName, ".gitignore")); err != nil {
		t.Fatalf("expected generated .gitignore: %v", err)
	}
}
