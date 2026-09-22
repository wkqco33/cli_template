package cli

import (
	"io"
	"os"

	"github.com/wkqco33/wcli/rich"
)

// Env 커맨드 실행 환경. 루트의 전역 플래그가 여기에 바인딩된다.
type Env struct {
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader
	Interactive bool   // stdin이 터미널이라 확인 프롬프트를 띄울 수 있음
	Quiet       bool   // 진행 메시지 억제
	Debug       bool   // 추가 진단 메시지
	NoInput     bool   // 대화형 입력 금지
	Yes         bool   // 확인 프롬프트에 자동 yes
	NoColor     bool   // 색상 출력 금지
	ConfigPath  string // 설정 파일 경로 오버라이드
}

// NewEnv 기본값을 채운 실행 환경을 만든다.
func NewEnv(stdout, stderr io.Writer, stdin io.Reader) *Env {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	if stdin == nil {
		stdin = os.Stdin
	}
	return &Env{
		Stdout:      stdout,
		Stderr:      stderr,
		Stdin:       stdin,
		Interactive: isTerminal(stdin),
	}
}

// progress 진행·안내 메시지. --quiet이면 출력하지 않는다.
func (e *Env) progress(format string, a ...any) {
	if e.Quiet {
		return
	}
	rich.Fprintln(e.Stderr, format, a...)
}

// notice 경고처럼 사용자가 반드시 봐야 하는 메시지. --quiet에도 출력한다.
func (e *Env) notice(format string, a ...any) {
	rich.Fprintln(e.Stderr, format, a...)
}

// debug 진단 메시지. --debug일 때만 출력한다.
func (e *Env) debug(format string, a ...any) {
	if !e.Debug {
		return
	}
	rich.Fprintln(e.Stderr, "[dim]debug: "+format+"[/dim]", a...)
}

// canPrompt 확인 프롬프트를 띄울 수 있는지 여부.
func (e *Env) canPrompt() bool {
	return !e.NoInput && e.Interactive && e.Stdin != nil
}

// confirm yes/no 확인을 받는다. 호출 전에 canPrompt를 확인해야 한다.
func (e *Env) confirm(label string, defaultYes bool) (bool, error) {
	return rich.FConfirm(e.Stderr, e.Stdin, label, defaultYes)
}

// isTerminal 값이 TTY인지 확인한다. io.Writer/io.Reader 모두 받는다.
func isTerminal(v any) bool {
	f, ok := v.(*os.File)
	if !ok {
		return false
	}
	return isTerminalFD(int(f.Fd()))
}
