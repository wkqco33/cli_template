package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath_OverrideAndEnvironment(t *testing.T) {
	if got, err := ResolvePath("custom.yaml"); err != nil || got != "custom.yaml" {
		t.Fatalf("override path: %q, %v", got, err)
	}
	t.Setenv("WTEMP_CONFIG", filepath.Join(t.TempDir(), "env.yaml"))
	got, err := ResolvePath("")
	if err != nil || got != os.Getenv("WTEMP_CONFIG") {
		t.Fatalf("environment path: %q, %v", got, err)
	}
}

func TestDefaultPath_IsUnderUserConfigDir(t *testing.T) {
	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath failed: %v", err)
	}
	if filepath.Base(got) != "config.yaml" || filepath.Base(filepath.Dir(got)) != "wtemp" {
		t.Fatalf("unexpected default path: %q", got)
	}
}
