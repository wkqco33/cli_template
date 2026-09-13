package cli

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/wkqco33/cli_template/generator"
)

// TestResolveFormat_DefaultsByTerminal 형식을 지정하지 않으면 TTY는 table,
// 파이프·CI는 plain으로 자동 선택한다.
func TestResolveFormat_DefaultsByTerminal(t *testing.T) {
	tests := []struct {
		name     string
		explicit string
		tty      bool
		want     string
	}{
		{name: "tty-default", tty: true, want: FormatTable},
		{name: "pipe-default", tty: false, want: FormatPlain},
		{name: "explicit-table", explicit: FormatTable, tty: false, want: FormatTable},
		{name: "explicit-plain", explicit: FormatPlain, tty: true, want: FormatPlain},
		{name: "explicit-json", explicit: FormatJSON, tty: true, want: FormatJSON},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveFormat(tc.explicit, tc.tty)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected format %q, got %q", tc.want, got)
			}
		})
	}
}

// TestResolveFormat_UnknownIsUsageError 알 수 없는 형식은 종료 코드 2로 매핑되는 오류다.
func TestResolveFormat_UnknownIsUsageError(t *testing.T) {
	_, err := resolveFormat("yaml", false)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	var usageErr *generator.UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("expected *UsageError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "table") || !strings.Contains(err.Error(), "json") {
		t.Fatalf("expected available formats in message, got %q", err.Error())
	}
}

// TestNewEnv_DefaultsToStandardStreams nil을 넘기면 표준 스트림으로 채운다.
func TestNewEnv_DefaultsToStandardStreams(t *testing.T) {
	env := NewEnv(nil, nil, nil)

	if env.Stdout == nil || env.Stderr == nil || env.Stdin == nil {
		t.Fatalf("expected non-nil streams, got %+v", env)
	}
	if env.Stdout != os.Stdout || env.Stderr != os.Stderr || env.Stdin != os.Stdin {
		t.Fatalf("expected standard streams, got stdout=%T stderr=%T stdin=%T", env.Stdout, env.Stderr, env.Stdin)
	}
}

// TestProgress_And_DebugGate --quiet/--debug 게이트가 출력을 통제한다.
func TestProgress_And_DebugGate(t *testing.T) {
	var buf strings.Builder
	env := &Env{Stdout: &buf, Stderr: &buf, Stdin: strings.NewReader(""), Quiet: true}

	env.progress("quiet-suppressed")
	if buf.Len() != 0 {
		t.Fatalf("expected suppressed progress, got %q", buf.String())
	}

	env.Quiet = false
	env.debug("hidden")
	if buf.Len() != 0 {
		t.Fatalf("expected no debug output without --debug, got %q", buf.String())
	}

	env.Debug = true
	env.debug("shown")
	if !strings.Contains(buf.String(), "shown") {
		t.Fatalf("expected debug output, got %q", buf.String())
	}
}
