# AGENTS.md — 프로젝트 개발 가이드

이 문서는 이 저장소에서 작업하는 **모든 에이전트(인간/LLM)를 위한 공통 가이드**다.
코드를 수정하기 전에 반드시 읽고, 아래 규칙을 따른다.

---

## 1. 프로젝트 개요

- **wtemp**: `wcli` 기반 Go CLI 프로젝트 템플릿 생성기.
- **언어/도구**: Go 1.26+, [Task](https://taskfile.dev) 빌드 도구.
- **핵심 패키지**:
  - `internal/cli/` — CLI 커맨드 정의 (`new`, `list`)와 실행 환경(`Env`).
- `cmd/wtemp/` — 메인 패키지(`run`/종료 코드 매핑).
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
task coverage-check  # 커버리지 임계값(85%) 검증
task coverage-html   # HTML 리포트 생성 (coverage.html)
task lint            # gofmt + go vet
task vuln            # govulncheck (별도 설치 필요)
task bench           # 생성 성능 벤치마크
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

- `ValidateProjectAndModuleName` / `Resolve` — 유효/무효 이름 케이스, 에러 문구, 타입화된 오류(`UsageError`).
- `Generate` — 성공 시 산출물 구조, 실패 시 임시 디렉토리 정리, `--force` 실패 시 기존 디렉토리 보존(원자성).
- `DryRun` — 파일시스템을 변경하지 않고 실제 생성과 같은 파일 목록을 내놓는다.
- `planFiles` — dry-run과 실제 생성이 공유하는 단일 계획 소스. SQLite 스킵 규칙 포함.
- `Templates` / `TemplateNamesCSV` — 카탈로그 무결성.
- `template_contract_test.go` — 템플릿 계약(버전 주입 가능, `templates/*/go.mod.tmpl`의 Go·wcli 버전이 루트 `go.mod`와 일치).
- `benchmark_test.go` — 생성 성능 회귀 감지.
- `smoke_build_test.go` — `//go:build smoke` 태그. 생성된 템플릿이 실제로 컴파일되는지 검증.

### `internal/cli/`

- 커맨드 구조(`Use`, `Short`, 등록된 플래그·단축키) 검증.
- 실행 경로: 인자 누락/검증 실패 시 에러 반환, `list` 출력 내용.
- 출력은 `Env`(`internal/cli/env.go`)로 주입한다. `Env.Stdout`은 결과, `Env.Stderr`는 진행·경고·오류다.
  테스트는 `newTestEnv()`로 버퍼와 `strings.Reader`를 주입하고, 대화형 경로는 `interactive("y\n")`로 만든다.
  터미널 판정은 `isTerminalFD`(ioctl)를 쓰므로 `/dev/null`을 TTY로 오인하지 않는다.
- 오류 타입(`generator.UsageError` 등)은 `errors.As`로 검증한다. 문구만 비교하지 않는다.

### 메인 (`cmd/wtemp/main.go`)

- `run(args, stdout, stderr)`가 종료 코드를 반환한다. 종료 코드 표를 바꾸면 테스트를 함께 갱신한다.
- `cmd/wtemp/docs_drift_test.go` — README의 플래그 문서와 `task help`가 실제 플래그 목록과 어긋나면 실패한다.

### `templates/`

- 템플릿 파일 자체는 `embed.FS`로 내장된다. 템플릿 문법 오류는
  `generator`의 렌더링 테스트/스모크 테스트가 잡아준다.
- 의존성·Go 버전을 올릴 때는 루트 `go.mod`와 `templates/*/go.mod.tmpl`을 **함께** 수정한다
  (계약 테스트가 강제한다).

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
  - [ ] `task coverage-check` 통과 (총 커버리지 85% 이상)
  - [ ] `task smoke` 통과 (템플릿 로직 변경 시)
  - [ ] `task lint` 경고 없음 (gofmt + go vet)
  - [ ] 플래그·종료 코드·출력 형식을 바꿨다면 `README.md`와 `CHANGELOG.md` 갱신

---

## 8. 보안 및 저장소 위생

- API 키, 토큰, 비밀번호, 개인정보, 개인키를 코드·템플릿·테스트·커밋에 넣지 않는다.
- `.env`, 키 파일, 바이너리, 릴리스 아카이브는 커밋하지 않는다. 생성된 예제 바이너리는 소스와 함께 배포하지 않는다.
- 파일 삭제·덮어쓰기 기능은 경로 검증과 원자성을 유지하고, 관련 실패 경로를 테스트한다.
- 의존성 변경 후 `go mod verify`, `go vet ./...`, `task test`를 실행한다.
- 보안 취약점은 공개 이슈 대신 `SECURITY.md` 절차로 신고한다.

## 9. 문서 및 품질 게이트

- 사용자 동작, 플래그, 템플릿 목록을 바꾸면 `README.md`와 관련 템플릿 README를 함께 갱신한다.
- 주석은 코드에서 드러나지 않는 이유와 제약만 설명한다. 구현을 그대로 번역하거나 작업 과정을 기록하는 장황한 주석은 추가하지 않는다.
- Go 코드는 `gofmt`를 적용하고, 커밋 전 다음 명령을 실행한다.

```bash
gofmt -w .
go mod verify
go vet ./...
task test
task test-race
task coverage-check
```

- 템플릿 또는 생성 로직을 변경하면 `task smoke`도 통과해야 한다.
- 릴리스 전에 `CHANGELOG.md`의 `Unreleased`를 버전 섹션으로 확정한다.
  릴리스 노트는 그 섹션에서 생성되고, 없으면 태그 메시지로 폴백한다(`CONTRIBUTING.md` 릴리스 절차).
- 릴리스는 `ppm.json`의 `bin_name`과 일치하는 플랫폼별 아카이브 및 SHA-256 체크섬을 제공하고,
  빌드 provenance를 함께 게시한다. 릴리스 파이프라인은 태그 커밋에서 테스트를 먼저 통과시킨다.

## 10. CLI 계약 (바꾸기 전에 확인)

- **출력 스트림**: 결과는 stdout, 진행·경고·오류·프로파일은 stderr. 새 메시지를 추가할 때 이 규칙을 지킨다.
- **종료 코드**: 0 성공 / 1 일반 / 2 사용법·검증 / 3 충돌 / 4 외부 도구 / 5 입력 필요.
  새 오류는 `generator`의 타입화된 오류(`UsageError`, `ConflictError`, `ExternalError`)로 표현한다.
- **파괴적 작업**: 덮어쓰기는 TTY에서 확인을 받고, 비TTY에서는 `--force`/`--yes`가 없으면 실패한다.
  프롬프트는 `Env.canPrompt()`(TTY + `--no-input` 아님)일 때만 띄우며 `rich.FConfirm`으로 stderr/stdin을 쓴다.
- **기계 판독 출력**: `--format plain|json`은 항상 stdout에만 쓴다. JSON 스키마를 바꾸면 README를 갱신한다.
- **호환성**: 플래그와 서브커맨드는 additive하게 유지한다. 제거·개명 시 deprecation 경고를 먼저 넣는다.

## 11. 주의사항

- **`smoke` 테스트는 느리다**: 네트워크로 의존성을 받고 `go build`를 수행한다.
  템플릿/생성 로직을 바꿀 때만 실행하고, 매 커밋마다 돌리지 않는다.
- **sqlite 케이스**: `CGO_ENABLED=0`, `gcc` 미설치, windows 환경에서는
  스모크 테스트가 `t.Skip`으로 건너뛴다. (의도된 동작)
- **`fyne` 템플릿**: GUI 빌드에 시스템 헤더(X11/OpenGL)가 필요해 스모크 매트릭스에서 제외.
- **`templates/`의 `.tmpl` 파일**: Go `text/template` 문법. 수정 시
  `task smoke`로 생성 결과가 컴파일되는지 반드시 확인한다.
- **전역 상태**: `rich.NoColor`는 루트의 `--no-color`에서 설정된다. 테스트에서 바꿨다면 `t.Cleanup`으로 복원한다.

---

## 12. 작업 시작 시 체크리스트

1. `git status`로 작업 트리 상태 확인.
2. `task test`로 현재 테스트가 모두 통과하는지 확인 (Green 기준점).
3. 변경할 기능/버그를 테스트로 먼저 표현 (Red).
4. 최소 구현 (Green) → 리팩터링 (Refactor).
5. `task test` + `task coverage-check`(+ 필요 시 `task smoke`) 통과 후 커밋.
