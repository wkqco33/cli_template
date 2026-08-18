# AGENTS.md — 프로젝트 개발 가이드

이 문서는 이 저장소에서 작업하는 **모든 에이전트(인간/LLM)를 위한 공통 가이드**다.
코드를 수정하기 전에 반드시 읽고, 아래 규칙을 따른다.

---

## 1. 프로젝트 개요

- **wtemp**: `wcli` 기반 Go CLI 프로젝트 템플릿 생성기.
- **언어/도구**: Go 1.26+, [Task](https://taskfile.dev) 빌드 도구.
- **핵심 패키지**:
  - `cmd/` — CLI 커맨드 정의 (`new`, `list`).
  - `generator/` — 템플릿 렌더링, 이름 검증, 원자적 생성 로직.
  - `templates/` — `embed.FS`로 내장된 템플릿 파일들.
- **외부 의존성**: `github.com/wkqco33/wcli` (CLI 라이브러리).

---

## 2. 개발 원칙: TDD (Test-Driven Development)

이 프로젝트는 **TDD로 개발**한다. 코드를 먼저 쓰고 테스트를 나중에 붙이지 않는다.

### TDD 사이클 (Red → Green → Refactor)

1. **Red**: 먼저 실패하는 테스트를 작성한다. (기능/버그를 테스트로 표현)
2. **Green**: 테스트가 통과할 만큼 **최소한의 코드**를 작성한다.
3. **Refactor**: 중복 제거, 이름 개선, 구조 정리. 이때도 테스트는 계속 통과해야 한다.

### 필수 규칙

- **기능 추가/버그 수정은 반드시 테스트와 함께** 커밋한다. 테스트 없는 코드 변경은 금지.
- **테스트를 먼저 작성**하고, 그 테스트가 실패하는 것을 확인한 뒤 구현한다.
- 커밋 전에 반드시 `task test`(전체 단위 테스트)가 통과해야 한다.
- `generator/`의 핵심 로직(검증, 생성, 정리)은 **반드시 테스트로 보호**한다.

---

## 3. 테스트 실행 방법

```bash
task test            # 단위 테스트 (캐시 무시, -count=1)
task test-race       # 데이터 레이스 탐지 포함
task test-watch      # 파일 변경 시 자동 재실행 (watchexec 필요)
task coverage        # 커버리지 측정 + 요약 (coverage.out)
task coverage-html   # HTML 리포트 생성 (coverage.html)
task smoke           # 생성된 템플릿의 go build 검증 (-tags=smoke)
```

- **개발 중**: `task test-watch`를 띄워 두고 수정할 때마다 자동 검증.
- **커밋 직전**: `task test` + `task smoke` 통과 확인.
- **커버리지 확인**: `task coverage`로 패키지별 커버리지를 확인한다.

---

## 4. 테스트 작성 컨벤션

### 파일/함수 명명

- 테스트 파일은 `*_test.go`, 대상 소스와 **같은 패키지**에 둔다.
  - 예: `generator/generator.go` → `generator/generator_test.go`
- 테스트 함수는 `Test<대상>_<시나리오>` 형태.
  - 예: `TestValidateProjectAndModuleName_InvalidProjectName`
- **테이블 드리븐 테스트**를 기본으로 사용한다. (입력/기대값을 구조체 슬라이스로)
- 하위 시나리오는 `t.Run(name, ...)`으로 분리한다.

### 검증 스타일

- `t.Fatalf`/`t.Errorf`에 **실제 값과 기대값을 모두** 출력한다.
  - 예: `t.Fatalf("expected %q, got %q", want, got)`
- 에러 메시지 검증은 `strings.Contains`로 **핵심 문구**만 확인한다.
  - 전체 메시지와의 brittle한 동등 비교는 피한다.

### 격리 (Isolation)

- 파일시스템을 건드리는 테스트는 반드시 `t.TempDir()`을 사용하고,
  `os.Chdir`로 이동했다면 `t.Cleanup`으로 원복한다.
- 전역 상태(예: `renderTemplatesFunc`)를 바꾸는 테스트는 반드시 `t.Cleanup`으로 복원한다.
- 네트워크/외부 도구가 필요한 테스트는 `t.Skip`으로 명시적으로 건너뛴다.

### 예시

```go
func TestSomething_Scenario(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {name: "valid", input: "my-cli", want: ""},
        {name: "invalid", input: "My App", want: "공백 문자가 포함"},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := DoSomething(tc.input)
            if tc.want == "" {
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
                return
            }
            if err == nil || !strings.Contains(err.Error(), tc.want) {
                t.Fatalf("expected error containing %q, got %v", tc.want, err)
            }
        })
    }
}
```

---

## 5. 패키지별 테스트 가이드

### `generator/`

핵심 로직이므로 **가장 높은 우선순위**로 테스트한다.

- `ValidateProjectAndModuleName` — 유효/무효 이름 케이스, 에러 문구.
- `Generate` — 성공 시 산출물 구조, 실패 시 임시 디렉토리 정리(원자성).
- `Templates` / `TemplateNamesCSV` — 카탈로그 무결성.
- `smoke_build_test.go` — `//go:build smoke` 태그. 생성된 템플릿이 실제로 컴파일되는지 검증.

### `cmd/`

- 커맨드 구조(`Use`, `Short`, 등록된 플래그) 검증.
- 실행 경로: 인자 누락/검증 실패 시 에러 반환, `list` 출력 내용.
- `rich.Println`/`fmt.Printf`는 `os.Stdout`에 직접 쓰므로, 출력 캡처가 필요하면
  `os.Pipe`로 `os.Stdout`을 교체하는 헬퍼를 사용한다. (`cmd/cmd_test.go` 참고)

### `templates/`

- 템플릿 파일 자체는 `embed.FS`로 내장된다. 템플릿 문법 오류는
  `generator`의 렌더링 테스트/스모크 테스트가 잡아준다.

---

## 6. 커버리지 기준

- **목표**: `generator/`와 `cmd/` 패키지의 핵심 로직 커버리지를 높게 유지한다.
- **최소 기준**: 새로 추가한 코드는 **테스트로 커버**되어야 한다.
  (커버리지가 낮은 새 코드는 리뷰에서 지적 대상)
- `task coverage`로 확인하고, 핵심 분기(에러 경로 포함)가 빠지지 않았는지 확인한다.

---

## 7. 커밋 규칙

- **하나의 논리적 변경 = 하나의 커밋**. 테스트와 구현을 함께 커밋한다.
- 커밋 메시지는 `type: 요약` 형식 (예: `feat:`, `fix:`, `test:`, `refactor:`, `docs:`).
- 커밋 전 체크리스트:
  - [ ] `task test` 통과
  - [ ] `task smoke` 통과 (템플릿 로직 변경 시)
  - [ ] `go vet ./...` 경고 없음
  - [ ] `gofmt` 적용됨

---

## 8. 주의사항

- **`smoke` 테스트는 느리다**: 네트워크로 의존성을 받고 `go build`를 수행한다.
  템플릿/생성 로직을 바꿀 때만 실행하고, 매 커밋마다 돌리지 않는다.
- **sqlite 케이스**: `CGO_ENABLED=0`, `gcc` 미설치, windows 환경에서는
  스모크 테스트가 `t.Skip`으로 건너뛴다. (의도된 동작)
- **`fyne` 템플릿**: GUI 빌드에 시스템 헤더(X11/OpenGL)가 필요해 스모크 매트릭스에서 제외.
- **`super_cli/`**: 생성된 예제 산출물이므로 직접 수정하지 않는다.
- **`templates/`의 `.tmpl` 파일**: Go `text/template` 문법. 수정 시
  `task smoke`로 생성 결과가 컴파일되는지 반드시 확인한다.

---

## 9. 작업 시작 시 체크리스트

1. `git status`로 작업 트리 상태 확인.
2. `task test`로 현재 테스트가 모두 통과하는지 확인 (Green 기준점).
3. 변경할 기능/버그를 테스트로 먼저 표현 (Red).
4. 최소 구현 (Green) → 리팩터링 (Refactor).
5. `task test` + `task smoke`(필요 시) 통과 후 커밋.
