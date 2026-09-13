package generator

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/wkqco33/cli_template/templates"
)

// TestTemplates_VersionIsStampable CLI 템플릿은 버전 값을 상수로 굳히지 않고
// -ldflags "-X main.version=..."로 주입할 수 있어야 한다.
func TestTemplates_VersionIsStampable(t *testing.T) {
	tests := []struct {
		template string
		wire     *regexp.Regexp
	}{
		{template: "minimal", wire: regexp.MustCompile(`(?m)\.Version\s*=\s*version`)},
		{template: "full", wire: regexp.MustCompile(`(?m)Version:\s*version`)},
		{template: "gin", wire: regexp.MustCompile(`(?m)Version:\s*version`)},
		{template: "fiber", wire: regexp.MustCompile(`(?m)Version:\s*version`)},
		{template: "echo", wire: regexp.MustCompile(`(?m)Version:\s*version`)},
	}

	for _, tc := range tests {
		t.Run(tc.template, func(t *testing.T) {
			content := readTemplateFile(t, tc.template+"/main.go.tmpl")

			if !strings.Contains(content, `var version = "dev"`) {
				t.Fatalf("expected stampable %q declaration, got:\n%s", `var version = "dev"`, content)
			}
			if !tc.wire.MatchString(content) {
				t.Fatalf("expected version wiring matching %q, got:\n%s", tc.wire, content)
			}
		})
	}
}

// TestTemplates_NoHardcodedVersion 모든 템플릿은 릴리스 버전을 상수로 박아두지 않는다.
func TestTemplates_NoHardcodedVersion(t *testing.T) {
	entries, err := templates.FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read embed FS failed: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			for _, path := range templateFiles(t, entry.Name()) {
				content := readTemplateFile(t, path)
				if strings.Contains(content, "const version") {
					t.Errorf("hardcoded version constant in %s:\n%s", path, content)
				}
			}
		})
	}
}

// TestTemplates_GoModMatchesRootModule 템플릿의 go.mod는 이 저장소의
// go.mod와 Go 버전·wcli 버전이 어긋나지 않아야 한다.
func TestTemplates_GoModMatchesRootModule(t *testing.T) {
	rootGo, rootWCLI := readRootGoMod(t)

	entries, err := templates.FS.ReadDir(".")
	if err != nil {
		t.Fatalf("read embed FS failed: %v", err)
	}

	checked := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := entry.Name() + "/go.mod.tmpl"
		content, err := templates.FS.ReadFile(path)
		if err != nil {
			continue
		}
		checked++

		if !strings.Contains(string(content), "go "+rootGo) {
			t.Errorf("%s: expected go directive %q to match root go.mod, got:\n%s", path, "go "+rootGo, content)
		}
		if strings.Contains(string(content), "github.com/wkqco33/wcli") &&
			!strings.Contains(string(content), "github.com/wkqco33/wcli "+rootWCLI) {
			t.Errorf("%s: expected wcli %s to match root go.mod, got:\n%s", path, rootWCLI, content)
		}
	}

	if checked == 0 {
		t.Fatal("no template go.mod files found")
	}
}

// readRootGoMod 저장소 루트 go.mod의 Go 버전과 wcli 버전을 읽는다.
func readRootGoMod(t *testing.T) (goVersion, wcliVersion string) {
	t.Helper()

	content := readRepoFile(t, "../go.mod")
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(trimmed, "go "); ok {
			goVersion = strings.TrimSpace(after)
		}
		if after, ok := strings.CutPrefix(trimmed, "require github.com/wkqco33/wcli "); ok {
			wcliVersion = strings.TrimSpace(after)
		}
	}
	if goVersion == "" || wcliVersion == "" {
		t.Fatalf("failed to parse root go.mod (go=%q wcli=%q)", goVersion, wcliVersion)
	}
	return goVersion, wcliVersion
}

// readRepoFile 저장소 파일을 읽는다(테스트는 패키지 디렉토리에서 실행된다).
func readRepoFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	return string(content)
}

// readTemplateFile 임베드된 템플릿 파일을 읽는다.
func readTemplateFile(t *testing.T, path string) string {
	t.Helper()

	content, err := templates.FS.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	return string(content)
}

// templateFiles 템플릿 디렉토리 아래의 모든 파일 경로를 반환한다.
func templateFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := templates.FS.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s failed: %v", dir, err)
	}

	var paths []string
	for _, entry := range entries {
		path := dir + "/" + entry.Name()
		if entry.IsDir() {
			paths = append(paths, templateFiles(t, path)...)
			continue
		}
		paths = append(paths, path)
	}
	return paths
}
