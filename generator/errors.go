package generator

import "fmt"

// UsageError 입력·사용법 오류. 프로세스 종료 코드 2로 매핑된다.
type UsageError struct {
	Message string
}

func (e *UsageError) Error() string { return e.Message }

// NewUsageError 사용법 오류를 만든다.
func NewUsageError(format string, a ...any) error {
	return &UsageError{Message: fmt.Sprintf(format, a...)}
}

// ConflictError 대상 경로가 이미 존재하는 충돌. 종료 코드 3.
type ConflictError struct {
	Path string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("디렉토리가 이미 존재합니다: %s", e.Path)
}

// InputRequiredError 비대화형 환경이라 확인 입력을 받을 수 없을 때 반환된다. 종료 코드 5.
type InputRequiredError struct {
	Path string
}

func (e *InputRequiredError) Error() string {
	return fmt.Sprintf(
		"대상 디렉토리가 이미 존재합니다: %s\n해결 방법: 덮어쓰려면 --force, 확인 프롬프트를 건너뛰려면 --yes를 사용하세요",
		e.Path,
	)
}

// ExternalError 외부 도구 실행 실패. 종료 코드 4.
type ExternalError struct {
	Tool string
	Err  error
}

func (e *ExternalError) Error() string {
	return fmt.Sprintf("%s 실행 실패: %v", e.Tool, e.Err)
}

func (e *ExternalError) Unwrap() error { return e.Err }
