package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/wkqco33/cli_template/internal/cli"
	"github.com/wkqco33/wcli"
)

// repoFile 저장소 루트의 파일을 읽는다(테스트는 패키지 디렉토리에서 실행된다).
func repoFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s failed: %v", path, err)
	}
	return string(content)
}

// TestReadme_DocumentsAllFlags README의 플래그 표가 실제 등록된 플래그를 모두 문서화한다.
func TestReadme_DocumentsAllFlags(t *testing.T) {
	readme := repoFile(t, "../../README.md")

	env := cli.NewEnv(io.Discard, io.Discard, strings.NewReader(""))
	root := newRootCmd(env)

	commands := []*wcli.Command{root, cli.NewCmd(env), cli.ListCmd(env)}

	checked := 0
	for _, c := range commands {
		flagSets := []*wcli.FlagSet{c.Flags(), c.PersistentFlags()}
		for _, fs := range flagSets {
			for _, f := range fs.All() {
				checked++
				if !strings.Contains(readme, "--"+f.Name) {
					t.Errorf("README does not document --%s (command %q)", f.Name, c.Use)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no flags found")
	}
}

// TestTaskfileHelp_DelegatesToBinaryHelp task help는 플래그 표를 중복 정의하지 않고
// 바이너리 도움말을 단일 소스로 사용해야 한다.
func TestTaskfileHelp_DelegatesToBinaryHelp(t *testing.T) {
	taskfile := repoFile(t, "../../Taskfile.yml")

	for _, want := range []string{"go run ./cmd/wtemp --help", "go run ./cmd/wtemp new --help"} {
		if !strings.Contains(taskfile, want) {
			t.Fatalf("expected Taskfile help task to run %q", want)
		}
	}
	for _, duplicated := range []string{"--sqlite", "--dry-run", "--module", "--force"} {
		if strings.Contains(taskfile, duplicated) {
			t.Fatalf("Taskfile must not duplicate flag documentation (%s); use binary --help instead", duplicated)
		}
	}
}
