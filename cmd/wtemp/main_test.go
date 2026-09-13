package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wkqco33/cli_template/internal/cli"
	"github.com/wkqco33/wcli/rich"
)

// runCLI run()을 실행하고 stdout/stderr/종료 코드를 반환한다.
func runCLI(args ...string) (stdout, stderr string, code int) {
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return out.String(), errOut.String(), code
}

// setVersion 테스트 동안 version 값을 바꾸고 원복한다.
func setVersion(t *testing.T, v string) {
	t.Helper()

	orig := version
	version = v
	t.Cleanup(func() { version = orig })
}

func TestRun_VersionPrintsToStdout(t *testing.T) {
	setVersion(t, "9.9.9-test")

	stdout, stderr, code := runCLI("--version")

	if code != exitOK {
		t.Fatalf("expected exit code %d, got %d", exitOK, code)
	}
	if !strings.Contains(stdout, "9.9.9-test") {
		t.Fatalf("expected stamped version on stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestRun_ErrorReportedOnceToStderr(t *testing.T) {
	stdout, stderr, code := runCLI("new", "My App")

	if code != exitUsage {
		t.Fatalf("expected exit code %d, got %d (stderr=%s)", exitUsage, code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout on error, got %q", stdout)
	}

	const want = "해결 방법:"
	if !strings.Contains(stderr, want) {
		t.Fatalf("expected stderr to contain %q, got %q", want, stderr)
	}
	if n := strings.Count(stderr, want); n != 1 {
		t.Fatalf("expected error to be reported once, got %d occurrences:\n%s", n, stderr)
	}
}

func TestRun_DryRunListsFilesOnStdout(t *testing.T) {
	chdirTemp(t)

	stdout, stderr, code := runCLI("new", "myapp", "-t", "minimal", "--dry-run")

	if code != exitOK {
		t.Fatalf("expected exit code %d, got %d (stderr=%s)", exitOK, code, stderr)
	}
	if !strings.Contains(stdout, "main.go") {
		t.Fatalf("expected file list on stdout, got %q", stdout)
	}
	if strings.Contains(stdout, "생성될 파일:") {
		t.Fatalf("expected header on stderr, got stdout %q", stdout)
	}
}

// TestRun_ExitCodes 문서화된 종료 코드가 실제로 반환되는지 검증한다.
func TestRun_ExitCodes(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T)
		args  []string
		want  int
	}{
		{
			name: "ok",
			args: []string{"list"},
			want: exitOK,
		},
		{
			name: "usage-missing-args",
			args: []string{"new"},
			want: exitUsage,
		},
		{
			name: "usage-invalid-name",
			args: []string{"new", "My App"},
			want: exitUsage,
		},
		{
			name: "usage-unknown-template",
			args: []string{"new", "myapp", "-t", "nope"},
			want: exitUsage,
		},
		{
			name: "usage-unknown-flag",
			args: []string{"new", "myapp", "--nope"},
			want: exitUsage,
		},
		{
			name: "usage-unknown-command",
			args: []string{"bogus"},
			want: exitUsage,
		},
		{
			name: "usage-root-unknown-flag",
			args: []string{"--nope"},
			want: exitUsage,
		},
		{
			name: "input-required",
			setup: func(t *testing.T) {
				if err := os.MkdirAll("myapp", 0o755); err != nil {
					t.Fatalf("mkdir failed: %v", err)
				}
			},
			args: []string{"new", "myapp", "-t", "minimal"},
			want: exitInputNeeded,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chdirTemp(t)
			if tc.setup != nil {
				tc.setup(t)
			}

			stdout, stderr, code := runCLI(tc.args...)
			if code != tc.want {
				t.Fatalf("args=%v: expected exit code %d, got %d (stdout=%q stderr=%q)",
					tc.args, tc.want, code, stdout, stderr)
			}
		})
	}
}

func TestRun_QuietSuppressesProgress(t *testing.T) {
	chdirTemp(t)

	stdout, stderr, code := runCLI("new", "myapp", "-t", "minimal", "--quiet")
	if code != exitOK {
		t.Fatalf("expected exit code %d, got %d (stderr=%s)", exitOK, code, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Fatalf("expected no output with --quiet, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestRun_DebugGoesToStderr(t *testing.T) {
	chdirTemp(t)

	_, stderr, code := runCLI("new", "myapp", "-t", "minimal", "--dry-run", "--debug")
	if code != exitOK {
		t.Fatalf("expected exit code %d, got %d (stderr=%s)", exitOK, code, stderr)
	}
	if !strings.Contains(stderr, "debug:") {
		t.Fatalf("expected debug output on stderr, got %q", stderr)
	}
}

func TestRun_NoColorFlagDisablesMarkup(t *testing.T) {
	orig := rich.NoColor
	t.Cleanup(func() { rich.NoColor = orig })
	rich.NoColor = false

	if _, _, code := runCLI("--no-color", "list"); code != exitOK {
		t.Fatalf("expected exit code %d, got %d", exitOK, code)
	}
	if !rich.NoColor {
		t.Fatal("expected --no-color to disable markup")
	}
}

func TestRun_ListJSONFormat(t *testing.T) {
	stdout, stderr, code := runCLI("list", "--format", "json")
	if code != exitOK {
		t.Fatalf("expected exit code %d, got %d (stderr=%s)", exitOK, code, stderr)
	}

	var payload struct {
		Templates []struct {
			Name   string `json:"name"`
			Desc   string `json:"desc"`
			SQLite bool   `json:"sqlite"`
		} `json:"templates"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("expected valid JSON, got %q: %v", stdout, err)
	}
	if len(payload.Templates) == 0 {
		t.Fatalf("expected template list, got %q", stdout)
	}
}

func TestRun_HelpIncludesDocsAndExitCodes(t *testing.T) {
	stdout, _, code := runCLI("--help")
	if code != exitOK {
		t.Fatalf("expected exit code %d, got %d", exitOK, code)
	}
	for _, want := range []string{
		"wtemp new <project-name>",
		"예시:",
		"종료 코드:",
		"https://github.com/wkqco33/cli_template",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected help to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestNewCmdHelp_HasFullPathAndExamples(t *testing.T) {
	stdout, _, code := runCLI("new", "--help")
	if code != exitOK {
		t.Fatalf("expected exit code %d, got %d", exitOK, code)
	}
	for _, want := range []string{
		"wtemp new <project-name> [flags]",
		"예시:",
		"https://github.com/wkqco33/cli_template",
		"--format",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected new help to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestRootFlags_StandardSet(t *testing.T) {
	env := cli.NewEnv(io.Discard, io.Discard, strings.NewReader(""))
	root := newRootCmd(env)

	names := make(map[string]string)
	for _, f := range root.PersistentFlags().All() {
		names[f.Name] = f.Shorthand
	}
	for flag, short := range map[string]string{
		"no-color": "", "no-input": "", "quiet": "q", "debug": "d", "yes": "y",
	} {
		got, ok := names[flag]
		if !ok {
			t.Fatalf("expected persistent flag %q, got %v", flag, names)
		}
		if got != short {
			t.Fatalf("expected -%s shorthand for --%s, got %q", short, flag, got)
		}
	}
}

// TestModulePath_MatchesRepoHomepage 모듈 경로는 저장소 URL과 일치해야
// `go install github.com/<owner>/<repo>@latest`가 동작한다.
func TestModulePath_MatchesRepoHomepage(t *testing.T) {
	modPath := readModulePath(t)

	raw, err := os.ReadFile("../../ppm.json")
	if err != nil {
		t.Fatalf("read ppm.json failed: %v", err)
	}
	var meta struct {
		Homepage string `json:"homepage"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("parse ppm.json failed: %v", err)
	}

	want := strings.TrimPrefix(strings.TrimPrefix(meta.Homepage, "https://"), "http://")
	if modPath != want {
		t.Fatalf("go.mod module = %q, ppm.json homepage = %q", modPath, meta.Homepage)
	}
	if strings.HasSuffix(modPath, "/v2") || strings.HasSuffix(modPath, "/v3") {
		t.Fatalf("module path must not carry a major-version suffix before v2: %q", modPath)
	}
}

func readModulePath(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("read go.mod failed: %v", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(after)
		}
	}
	t.Fatal("module directive not found in go.mod")
	return ""
}

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
	return filepath.Clean(tmp)
}
