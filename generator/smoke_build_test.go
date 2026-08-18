//go:build smoke

package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSmokeGeneratedTemplatesBuild(t *testing.T) {
	cases := []struct {
		name       string
		template   string
		sqlite     bool
		requireCGO bool
		requireGCC bool
	}{
		{name: "minimal-default", template: "minimal"},
		{name: "minimal-sqlite", template: "minimal", sqlite: true, requireCGO: true, requireGCC: true},
		{name: "full-default", template: "full"},
		{name: "full-sqlite", template: "full", sqlite: true, requireCGO: true, requireGCC: true},
		{name: "gin-default", template: "gin"},
		{name: "fiber-default", template: "fiber"},
		{name: "echo-default", template: "echo"},
		{name: "library-default", template: "library"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.requireCGO {
				if os.Getenv("CGO_ENABLED") == "0" {
					t.Skip("CGO_ENABLED=0 환경에서는 sqlite 템플릿 빌드를 건너뜁니다")
				}
				if runtime.GOOS == "windows" {
					t.Skip("windows 환경 sqlite 빌드는 CI 도구체인 편차로 건너뜁니다")
				}
			}
			if tc.requireGCC {
				if _, err := exec.LookPath("gcc"); err != nil {
					t.Skip("gcc가 없어 sqlite 템플릿 빌드를 건너뜁니다")
				}
			}

			tmpRoot := t.TempDir()
			projectName := tc.name

			prevDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("현재 경로 확인 실패: %v", err)
			}
			if err := os.Chdir(tmpRoot); err != nil {
				t.Fatalf("임시 디렉토리 이동 실패: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Chdir(prevDir)
			})

			if err := Generate(projectName, Options{Template: tc.template, SQLite: tc.sqlite}); err != nil {
				t.Fatalf("생성 실패 (template=%s sqlite=%v): %v", tc.template, tc.sqlite, err)
			}

			targetPath := filepath.Join(tmpRoot, projectName)
			env := append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = targetPath
			tidyCmd.Env = env
			tidyOut, err := tidyCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go mod tidy 실패 (template=%s sqlite=%v): %v\n%s", tc.template, tc.sqlite, err, string(tidyOut))
			}

			buildCmd := exec.Command("go", "build", "./...")
			buildCmd.Dir = targetPath
			buildCmd.Env = env
			buildOut, err := buildCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go build 실패 (template=%s sqlite=%v): %v\n%s", tc.template, tc.sqlite, err, string(buildOut))
			}
		})
	}
}
