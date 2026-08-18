package generator

import (
	"testing"

	"cli_template/templates"
)

func TestTemplates_CatalogMatchesEmbedFS(t *testing.T) {
	entries, err := templates.FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read embed FS failed: %v", err)
	}

	fsDirs := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			fsDirs[e.Name()] = true
		}
	}

	catalogNames := make(map[string]bool)
	for _, meta := range Templates() {
		catalogNames[meta.Name] = true
		if !fsDirs[meta.Name] {
			t.Fatalf("catalog template %q has no matching directory in embed.FS", meta.Name)
		}
	}

	for dir := range fsDirs {
		if !catalogNames[dir] {
			t.Fatalf("embed.FS directory %q is missing from the template catalog", dir)
		}
	}
}
