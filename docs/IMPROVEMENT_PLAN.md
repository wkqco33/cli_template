# wtemp 개선 계획 (체크리스트: `ncli view 24`, 2026-09-12)

기준 문서: `ncli view 24`의 **프로젝트 체크리스트**.
점검 대상: `wtemp` (Go 1.26, wcli v0.2.0 기반 CLI 템플릿 생성기).

모든 항목은 **현재 상태를 실제로 실행/측정해서 확인한 증거**를 포함한다.
구현은 `AGENTS.md`의 TDD 규칙(Red → Green → Refactor)을 따른다. 각 항목의 `Red:`에 적은 테스트를 먼저 작성한다.

---

## 0. 점검 방법 (재현 명령)

```bash
go build -o /tmp/wtemp .
go test -count=1 -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out | tail -3
go list -m -u all
git rev-list --objects --all | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' \
  | awk '/^blob/ {print $3, $4}' | sort -rn | head
```

측정된 실제 결과:

| 항목 | 측정값 |
| --- | --- |
| 저장소 크기 / 최대 blob | `.git` 280KB, 최대 blob 12.8KB (`generator/generator.go`) |
| 시크릿/대용량 바이너리 | 추적 파일·전체 git 히스토리에서 미검출 |
| 커버리지 | total 82.0%, `cmd` 87.3%, `generator` 84.7%, `main.go` **0%** |
| 생성 성능 (`library`) | total 234~594µs / render 193~520µs (5회) |
| 오래된 의존성 | `wcli v0.2.0 [v0.2.3]` |

---

## 1. 종합 진단

| 체크리스트 섹션 | 상태 | 요약 |
| --- | --- | --- |
| 개발 품질 — 히스토리/시크릿 | ✅ 대체로 양호 | 히스토리 청정. 자동 탐지 게이트만 없음 |
| 개발 품질 — TDD/AGENTS.md | ✅ 양호 | `AGENTS.md`, 테스트 7종, `-race` CI 존재. 게이트/`main.go` 미커버 |
| 개발 품질 — 문서/주석 | ⚠️ 보통 | README 최신. `task help`에 `--profile` 누락 등 드리프트 |
| 개발 품질 — 스타일/포맷 | ⚠️ 보통 | gofmt/vet CI 있음. lint·`.gitattributes`·`govulncheck` 없음 |
| 개발 품질 — 성능 | ✅ 양호 | 병목 없음(서브밀리초). 회귀 방지 벤치마크만 부재 |
| CLI 가이드라인 | ❌ 미준수 다수 | stdout/stderr 미분리, 종료 코드 단일, `--json`/표준 플래그/확인 절차 없음 |
| API 가이드라인 (AIP) | ➖ 해당 없음 | REST 서버 없음 (템플릿은 스켈레톤만 생성) |
| 라이브러리 가이드라인 | ⚠️ 부분 | 모듈 경로 `cli_template`이 `go install`을 차단. CHANGELOG 없음 |
| 프론트엔드 가이드라인 | ➖ 해당 없음 | `fyne` 템플릿만 존재하며 스모크 매트릭스 제외(의도됨) |
| 공개·배포 | ⚠️ 보통 | LICENSE/CONTRIBUTING/SECURITY 있음. 서명·provenance·CHANGELOG 없음 |
| `library` 템플릿 정확성 | ❌ **버그** | `--module` 사용 시 `github.com/github.com/...` 생성 |

---

## 2. P0 — 즉시 수정 (정확성/배포 차단)

### P0-1. `library` 템플릿이 `--module` 경로를 중복 접두

현재 증거:

```text
$ wtemp new libx -t library --module github.com/user/libx
$ head -1 libx/go.mod
module github.com/github.com/user/libx      # ← 깨진 모듈 경로
$ grep 'go get' libx/README.md
go get github.com/github.com/user/libx
```

원인: 다른 템플릿은 `module {{.ModuleName}}`인데 `library`만 `github.com/`를 하드코딩한다
(`templates/library/go.mod.tmpl:1`, `README.md.tmpl:8,14`, `lib_test.go.tmpl:6`, `examples/basic/main.go.tmpl:6`).

변경: 5개 파일에서 `github.com/{{.ModuleName}}` → `{{.ModuleName}}`.
README의 `go get`/`import` 예시는 모듈 경로 그대로 노출하고, 기본값 사용 시 `github.com/<name>`을 쓰도록 README에 안내.

- `Red:` `generator/options_test.go`에 `TestGenerate_LibraryModuleName` — `--module github.com/user/libx`로 생성 후 `go.mod` 첫 줄이 정확히 `module github.com/user/libx`이고 `github.com/github.com`이 없음을 검증. 기존 `TestGenerate_ModuleName`이 `minimal`만 검사해 이 버그를 놓쳤다.
- `Red:` `generator/smoke_build_test.go` 매트릭스에 `--module`을 지정한 케이스(`library-module-path`, `full-module-path`) 추가.
- 완료 기준: 위 테스트 통과 + `task smoke` 통과.

### P0-2. 모듈 경로가 `go install`을 차단

현재 증거:

```text
$ go install github.com/wkqco33/cli_template@latest
go: github.com/wkqco33/cli_template@v0.2.1: parsing go.mod:
        module declares its path as: cli_template
                but was required as: github.com/wkqco33/cli_template
```

`go.mod:1`이 `module cli_template`이고 `ppm.json`의 `homepage`는 `github.com/wkqco33/cli_template`이다.
표준 설치 경로(`go install`)가 막혀 있고, 원격에 `v0.2.1` 태그가 이미 게시된 상태다.

변경: `go.mod`의 모듈 경로를 `github.com/wkqco33/cli_template`으로 바꾸고
`main.go`, `cmd/*.go`, `generator/*.go`, `generator/*_test.go`의 임포트를 일괄 수정.

- `Red:` `TestModulePath_MatchesRepo` — `go.mod`의 module 경로가 `ppm.json`의 `homepage`(에서 스킴 제거)와 일치하고 `/vN` 접미사가 v2 미만에서 없음을 검증.
- 검증: `go build ./... && go vet ./... && task test`, 이후 `go install github.com/wkqco33/cli_template@latest` 1회 확인.
- 완료 기준: `go install` 성공, README 설치 절차에 `go install` 경로 추가.

### P0-3. 릴리스 바이너리가 항상 `0.1.0`을 보고

현재 증거: `main.go:11` `const version = "0.1.0"`, `.github/workflows/release.yml:30`은 `task release`(ldflags 없음)로 빌드.
원격 태그는 `v0.2.1`까지 존재하며 `wtemp --version`은 `0.1.0`을 출력한다.

변경:

- `main.go`: `const version` → `var version = "dev"`.
- `Taskfile.yml` `release`: `go build -trimpath -ldflags "-s -w -X main.version=$(git describe --tags --always --dirty)"`.
- `.github/workflows/release.yml`: 태그 검증(`v*` 강제) + 릴리스 노트 자동 생성.
- `templates/*/main.go.tmpl`: 생성 프로젝트도 `-X main.version`으로 스탬프 가능하도록 동일 패턴 적용 + 생성된 README에 방법 1줄 문서화.

- `Red:` `main_test.go`의 `TestVersionFlag_UsesLdflagsValue` — `version`을 바꾼 뒤 `--version` 출력이 그 값을 반환.
- `Red:` `templates` 계약 테스트 — 모든 템플릿 `main.go.tmpl`이 `const version =` 대신 `var version`을 쓸 것.
- 완료 기준: `git tag v0.3.0 && ...` 없이도 `-ldflags` 주입 시 올바른 값 출력.

### P0-4. `--dry-run`이 디렉터리를 실제로 생성 (부작용)

현재 증거:

```text
$ wtemp new myapp -t minimal --dry-run -o made/up
생성될 파일: myapp (템플릿: minimal)
$ ls                      # made/ 디렉터리가 생성됨
made
```

원인: `generator/generator.go:256` `os.MkdirAll(parentDir, ...)` — dry-run이 임시 디렉터리를 만들어 렌더링한 뒤 목록만 수집한다.
`TestDryRun_ReturnsFilesWithoutCreatingTarget`은 `projectName`만 검사해 `parentDir` 생성은 잡지 못한다.

변경:

- 파일 계획 산출 로직을 `planFiles(tmplName string, data TemplateData) ([]string, error)`로 추출해 **실제 생성과 같은 워커(walker)를 공유** (SQLite 스킵 규칙·`gitignore.tmpl` → `.gitignore` 매핑 단일 소스).
- `DryRun`은 디스크를 전혀 건드리지 않는다. `renderTemplates`는 `planFiles`가 만든 목록을 쓰도록 리팩터.

- `Red:` `TestDryRun_CreatesNoDirectories` — `--output nested/dir`로 dry-run 후 `nested`가 `os.IsNotExist`.
- `Red:` `TestDryRun_SQLiteFileListMatchesGenerate` — dry-run 목록과 실제 생성물 파일 목록이 완전히 일치(SQLite on/off 각각). 두 경로가 드리프트할 수 없게 만든다.
- 완료 기준: 임시 디렉터리도 만들지 않음(`t.TempDir()` 안에 잔여 항목 0).

### P0-5. `--force`가 비원자적으로 기존 디렉터리를 삭제

현재 증거: `generator/generator.go:196` — 렌더링 **전에** `os.RemoveAll(targetPath)`.
렌더링이 실패하면(예: 디스크 부족, 템플릿 오류) 사용자의 기존 디렉터리가 복구 불가로 사라진다.
`TestGenerate_CleansTemporaryArtifactsOnFailure`는 대상이 없는 경우만 검증한다.

변경: 렌더링은 임시 디렉터리에서 완료 → `targetPath`를 `.bak-<ts>`로 `os.Rename` → 임시 디렉터리를 `targetPath`로 `os.Rename` → 성공 시 `.bak` 제거, 실패 시 `.bak`을 원위치로 복원.

- `Red:` `TestGenerate_ForceRenderFailureKeepsExistingDirectory` — 기존 파일이 있는 대상 + `renderTemplatesFunc` 주입 실패 + `Force: true` → 기존 파일 내용이 그대로 보존되고 `.bak-*` 잔여물 0.
- 완료 기준: `TestGenerate_ForceOverwritesExisting`(기존) + 신규 테스트 동시 통과.

### P0-6. 진행/완료 메시지가 stdout으로 출력 (기계 판독 불가)

현재 증거: `cmd/new.go:54,61,67,78,81,82,83,85`의 `rich.Println`/`fmt.Printf`는 모두 `os.Stdout`으로 나간다(`rich.Println` → `rich.Fprint(os.Stdout, ...)`).
`--dry-run`에서 파일 목록과 안내 문구가 같은 스트림에 섞인다. 때문에 `cmd/cmd_test.go`가 `os.Stdout`을 파이프로 교체하는 헬퍼를 써야 한다.

변경:

- `generator`의 `--profile` 출력(stderr, `generator.go:135`)과 동일하게 **결과만 stdout**, 진행/경고/힌트는 stderr.
- `cmd.NewCmd(out, err io.Writer)` / `cmd.ListCmd(out io.Writer)`로 주입(또는 `wcli.Command.OutWriter/ErrWriter`를 main에서 설정하고 클로저로 전달). `rich.Fprintln(w, ...)`, `table.Render(w)`가 이미 공개 API이므로 그대로 사용 가능.
- `main.go`를 `func run(args []string, stdout, stderr io.Writer) int`로 분리 → 테스트 헬퍼 제거 가능.

- `Red:` `TestNewCmd_ProgressGoesToStderr` — 결과 파일 목록은 stdout, `생성 중:`/`완료!`/`주의:`는 stderr.
- `Red:` `TestListCmd_NoANSIToPipe` — 파이프 출력에 ANSI 이스케이프(`\x1b[`)가 없고 박스 문자도 `--plain`에서는 없음.
- 완료 기준: README에 스트림 규칙(결과=stdout, 상태/오류=stderr) 명시.

---

## 3. P1 — CLI 가이드라인 준수

### P1-1. 종료 코드 체계

현재: `main.go:28`이 모든 오류에 `os.Exit(1)`.

변경(문서화 포함): `0` 성공 / `1` 일반·내부 / `2` 사용법(인자 누락·검증 실패·알 수 없는 템플릿) / `3` 충돌(대상 존재) / `4` 외부 도구 실패(git) / `5` 입력 필요(비대화형 확인 불가).

- 구현: `generator`에 `UsageError`, `ConflictError` 센티넬 타입 + `errors.As`로 `run()`에서 매핑.
- `Red:` `main_test.go` 테이블 테스트 6케이스 — 각 시나리오의 `run()` 반환 코드 검증.
- 완료 기준: README "종료 코드" 표 + 각 서브커맨드 `--help`에 안내.

### P1-2. 기계 판독 출력 `--json` / `--plain`

현재: `wtemp list`는 파이프에서도 박스 테이블을 출력 → 스크립트 파싱 불가.

변경: `list`와 `new --dry-run`에 `--format table|plain|json`(+ 단축 `--json`, `--plain`) 추가.

```json
{"templates":[{"name":"minimal","desc":"루트 커맨드만 있는 최소 구조","sqlite":true}]}
{"project":"myapp","template":"minimal","sqlite":false,"files":["main.go","go.mod"]}
```

- `Red:` `TestListCmd_JSONSchema` — `json.Unmarshal` 후 이름 집합이 `generator.Templates()`와 일치. `TestDryRunCmd_JSONSchema` — 파일 목록이 실제 생성물과 일치.
- 완료 기준: `wtemp list --json | jq -e '.templates | length == 7'` 성공.

### P1-3. 표준 플래그 정합

현재 없음: `--no-color`, `--no-input`, `-q/--quiet`, `-d/--debug`, `-y/--yes`, `-n`(`--dry-run` 단축), `-f`(`--force` 단축).
현재 있음: `-h/--help`, `--version`, `-o/--output`, `-t/--template`, `--dry-run`, `--force`.

변경: 루트에 영속 플래그로 추가하고 일관되게 사용.
`--no-color` → `rich.NoColor = true`; `--quiet` → 진행 메시지 억제(결과·오류는 유지); `--debug` → `logging.LevelDebug`.
`--no-input`은 "프롬프트 사용 금지"로 동작(프롬프트가 필요한데 비TTY면 코드 5로 실패).

- `Red:` `TestRootFlags_StandardSet` — 플래그 이름/단축 등록 검증. `TestNoColorFlag_DisablesMarkup` — `--no-color` 출력에 `\x1b[` 없음.
- 완료 기준: `wtemp --help` 플래그 목록과 README 표가 일치.

### P1-4. 파괴적 작업 확인 절차

현재: 대상이 존재하면 `--force` 없이는 즉시 오류, `--force`가 있으면 무조건 삭제. TTY 확인이 없다.

변경:

- TTY + `--force`/`--yes` 없음 → stderr 프롬프트로 삭제 대상을 보여주고 확인(기본 N).
- 비TTY + `--force`/`--yes` 없음 → 코드 3으로 실패하고 "`--force` 또는 `--yes`를 사용하세요" 안내 (체크리스트의 비대화형 요구).
- 삭제 전 대상 경로(절대경로)·포함 항목 수를 stderr에 출력.

- `Red:` `TestNewCmd_ExistingDirNonTTYFailsWithGuidance`, `TestNewCmd_ExistingDirTTYConfirmDeclined` — 삭제가 일어나지 않음.
- 완료 기준: 확인 프롬프트는 `stdin`이 TTY일 때만 표시.

### P1-5. 도움말 완결성

현재 증거: `wtemp new --help`

```text
Usage:
  new <project-name> [flags]      # 전체 실행 경로(wtemp) 누락, 예시/링크 없음
```

또한 영어 기본 문구(`Usage:`, `Flags:`, `print help`)가 한국어 설명과 혼재한다.

변경: `new`/`list`의 `Long`에 전체 실행 경로·예시 3개·문서/이슈 링크 추가. 루트 `HelpTemplate`(wcli 지원)으로 `Usage:`/`Flags:` 등 고정 문구를 한국어로 통일.

- `Red:` `TestNewCmd_HelpContainsFullPathAndExamples` — `--help` 출력에 `wtemp new`, `예시:`, 문서 URL이 포함.
- 완료 기준: 각 서브커맨드 도움말만으로 복사-실행 가능.

### P1-6. 입력 검증 강화

현재 증거: `wtemp new myapp extra-arg ...` → 여분 인자를 조용히 무시하고 생성 성공.

변경: 여분 위치 인자는 코드 2로 거부. `--output` 절대경로/상위 탈출(`..`)은 허용하되, `--force` 삭제 시 대상이 CWD 밖이거나 생성물 시그니처가 아니면 경고·거부.

- `Red:` `TestNewCmd_RejectsExtraArgs`, `TestGenerate_ForceRefusesOutsideCWDWithoutFlag`.
- 완료 기준: `validateSafeName`과 동일한 수준의 가이드 메시지(`해결 방법:`) 제공.

### P1-7. 문서/설정 드리프트 자동 방지

현재 드리프트(확인됨):

- `Taskfile.yml` `help` 목록에 `--profile` 누락(`cmd/new.go:47`에는 존재).
- `go.mod`는 `go 1.26.1`인데 `templates/*/go.mod.tmpl`은 전부 `go 1.26.0`.
- `templates/*/go.mod.tmpl`은 `wcli v0.2.0`, 루트도 v0.2.0 — 최신은 v0.2.3.
- README·`AGENTS.md`·`Taskfile help`가 플래그 목록을 각각 복제.

변경:

- `task help`는 플래그 표를 직접 출력하지 말고 `go run . --help` / `go run . new --help` / `go run . list`를 호출.
- `go` 버전·`wcli` 버전은 `templates` 계약 테스트로 고정: 루트 `go.mod` 값을 파싱해 모든 템플릿과 비교(값이 다르면 실패 → 한쪽만 올리는 실수 차단).
- 플래그 문서는 `cmd` 패키지 테스트에서 `cmd.Flags().All()`과 README의 표를 대조.

- `Red:` `TestTemplateGoMod_VersionsMatchRoot`, `TestTaskfileHelp_CoversAllFlags`, `TestReadme_DocumentsAllFlags`.
- 완료 기준: 버전/플래그를 한쪽만 바꾸면 테스트가 실패.

### P1-8. 의존성 최신화

현재: `go list -m -u all` → `github.com/wkqco33/wcli v0.2.0 [v0.2.3]`.
v0.2.3에는 서브커맨드 앞 플래그 라우팅 수정과 Windows CRLF/gofmt 문제 해결이 포함된다(upstream CHANGELOG).

변경: 루트 `go.mod`와 `templates/{minimal,full,gin,fiber,echo,fyne}/go.mod.tmpl`을 v0.2.3으로 통일 후 `go mod verify`, `task test`, `task smoke`.
(upstream v0.2.3 CHANGELOG 항목이 `[Unreleased]`에 있는 점은 이슈로 질문으로 남긴다.)

### P1-9. `super_cli/` 드리프트 제거

현재: 추적 중인 `super_cli/`가 `full` 템플릿과 불일치(`parseLogLevel`/설정 기반 로거/`database.Init` 없음, `go.mod`는 `wcli v0.1.1`). `AGENTS.md`는 "직접 수정하지 말 것"이라고만 하고 아무 테스트도 이 디렉터리를 검증하지 않는다.

변경(택1): ① 삭제(생성물이므로 저장소에 둘 이유 없음) ② 살아있는 예제로 유지하려면 `TestSuperCLIMatchesGeneratedOutput`(생성 → 디렉터리 트리·파일 내용 비교) 추가 + `smoke` 매트릭스에 포함.

- 완료 기준: 드리프트가 생기면 CI가 실패.

### P1-10. 테스트/커버리지

현재: `main.go` 0%, 미검증 경로 = `--version`, `--profile` 스트림, `--git` E2E, `list` 출력 스키마, `InitGit` 실패, `os.Rename` 실패.

변경: 위 항목을 P0/P1 작업의 `Red` 테스트로 흡수하고, `main.go`는 `run()` 추출로 테스트 가능하게 만든다.
CI에 커버리지 임계값(예: total ≥ 85%, `main`·`cmd`·`generator` 각각 ≥ 80%)을 추가하고 `coverage.out` 아티팩트 업로드.

- 완료 기준: `task coverage` 요약이 임계값 이상, 신규 코드는 모두 커버.

---

## 4. P2 — 공개·배포 및 공급망

### P2-1. CI 강화 (`.github/workflows/ci.yml`)

현재: gofmt check, `go vet`, build, `go mod verify`, `task test`, `task test-race`, `task smoke` — ubuntu-latest 단일.

추가:

- `golangci-lint`(또는 `staticcheck`) 단계.
- `govulncheck ./...` 단계.
- 시크릿 스캔(`gitleaks`) — 체크리스트 "히스토리/캐시에 시크릿 없음"의 상시 게이트.
- 테스트 OS 매트릭스: `ubuntu-latest` + `windows-latest` + `macos-latest`(단위 테스트만). `validateSafeName`/`filepath` 처리는 OS 의존이므로 실제 위험.
- `go test -shuffle=on -count=1`.
- 워크플로 권한은 이미 `contents: read`로 최소화됨 — 유지.

### P2-2. 릴리스 파이프라인 (`.github/workflows/release.yml`)

추가:

- 태그 푸시 시 **테스트를 선행** 실행(`needs: test` 또는 ci 재사용). 현재는 태그만으로 바로 게시된다.
- `-trimpath`, `-ldflags "-s -w -X main.version=<tag>"`, 재현 가능 빌드(`CGO_ENABLED=0`, `SOURCE_DATE_EPOCH`).
- `actions/attest-build-provenance`(SLSA provenance) + `id-token: write`. 체크섬은 이미 생성 중.
- `generate_release_notes: true` + `CHANGELOG.md` 링크.
- 서드파티 액션은 태그가 아닌 커밋 SHA로 고정.

### P2-3. 공개 저장소 위생

현재 있음: `LICENSE`(MIT), `CONTRIBUTING.md`, `SECURITY.md`, README.
추가: `CHANGELOG.md`(Keep a Changelog + SemVer), `.gitattributes`(`* text=auto eol=lf`), `.github/ISSUE_TEMPLATE/*`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/dependabot.yml`(gomod + github-actions), 선택적으로 `CODE_OF_CONDUCT.md`, `CODEOWNERS`.

`.gitattributes`는 upstream wcli가 겪은 "Windows CRLF로 gofmt 전량 실패" 문제를 예방한다.

### P2-4. 성능 (측정 결과에 따른 결론)

측정: `library` 생성 total 234~594µs, render 193~520µs, postprocess 7~13µs — **병목 없음**.
따라서 파일 렌더링 병렬화는 지금 하지 않는다(에러 처리 복잡도 대비 이득 무시 가능). 대신:

- `BenchmarkGenerate/{minimal,full,library}` + `BenchmarkRenderTemplates` 추가로 회귀만 감시.
- `--profile` 출력에 파일 개수·바이트 수도 추가하고, `--profile`을 `task help`/README 플래그 표에 포함(P1-7과 동일 작업).
- 실제 지연은 생성 후 사용자가 실행하는 `go mod tidy`(네트워크)에 있으므로, README에 그 사실을 명시(개선 대상이 아님을 문서화).

### P2-5. 문서·주석 정리

- README: 설치(`ppm` + `go install` + 직접 빌드), 스트림 규칙, 종료 코드 표, `--json`/`--plain`, 표준 플래그, 버전 스탬프 방법, 서브커맨드 도움말 링크.
- `AGENTS.md`: 새 게이트(`task lint`, `task vuln`, 커버리지 임계값)와 종료 코드 규칙을 반영.
- 주석: `generator/generator.go:312` `renderTemplates(projectName, ...)`은 실제로 대상 루트를 받는 파라미터인데 이름이 프로젝트명과 혼동된다 → `destRoot`로 변경. 작업 과정/번역식 주석은 제거.

---

## 5. 적용 제외 항목 (근거)

- **API 가이드라인(AIP)**: 이 저장소는 REST API 서버가 아니다. `gin`/`fiber`/`echo` 템플릿은 사용자 프로젝트용 스켈레톤이며, 그 스켈레톤의 API 설계를 AIP에 맞추는 일은 별도 템플릿 이슈로 분리한다(원한다면 `templates/*/handler`를 리소스 표준 메서드 형태로 개선하는 후속 작업으로 등록).
- **프론트엔드 앱 가이드라인**: `wtemp` 자체에 프론트엔드 없음. `fyne` 템플릿은 데스크톱 GUI이며 이미 스모크 매트릭스에서 의도적으로 제외되어 있고, 접근성 가이드라인(WCAG)은 템플릿 스켈레톤에 적용할 실익이 낮다.

---

## 6. 실행 순서 (각 단계 = 하나의 PR/커밋 단위)

| 단계 | 작업 | 필수 게이트 |
| --- | --- | --- |
| S1 (P0 버그/배포) | P0-1, P0-4, P0-5, P0-2, P0-3 | `task test`, `task smoke`, `go install` 1회 확인 |
| S2 (CLI 계약) | P0-6 → P1-1 → P1-2 → P1-3 → P1-4 → P1-5 → P1-6 | `task test`, `task test-race`, 스트림/종료 코드 통합 테스트 |
| S3 (품질 게이트) | P1-7, P1-8, P1-9, P1-10, P2-4 | `task coverage` 임계값, `task smoke`, 벤치마크 |
| S4 (배포/공개) | P2-1, P2-2, P2-3, P2-5 | CI 전체 green + 태그 없이 release 워크플로 dry-run |

각 단계 완료 조건(공통 DoD): `gofmt -w .` → `go mod verify` → `go vet ./...` → `task test` → `task test-race` → (템플릿 변경 시) `task smoke`, 그리고 테스트와 구현을 같은 커밋에 포함.

---

## 7. 성과 지표

- `wtemp --version`이 릴리스 태그와 항상 일치.
- 표준 설치 3경로(`ppm`, `go install`, 직접 빌드) 모두 성공.
- `wtemp list --json | jq`가 스크립트에서 동작.
- 종료 코드 6종 문서화 + 테스트로 고정.
- `--dry-run`이 파일시스템을 전혀 변경하지 않음(테스트로 고정).
- `--force` 실패 시 기존 디렉터리 100% 보존(테스트로 고정).
- CI: lint + vet + vuln + secret scan + 3 OS 단위 테스트 + gofmt + 커버리지 임계값 + smoke.
- 템플릿/README/help/`go.mod` 버전 드리프트가 테스트 실패로 탐지됨.
- 커버리지 total ≥ 85% (현재 82.0%, `main.go` 0% → 테스트 대상에 포함).

---

## 8. 진행 상태 (2026-09-12 구현 완료)

### 완료

| 항목 | 결과 |
| --- | --- |
| P0-1 library 모듈 경로 | 5개 템플릿 파일 수정 + `TestGenerate_NoModulePathDuplication`(전체 템플릿) + 스모크 `*-module-path` 2건 |
| P0-2 모듈 경로 | `github.com/wkqco33/cli_template`으로 변경 + 메인 패키지를 `cmd/wtemp`로 이동(설치 바이너리 이름 `wtemp` 보장, `internal/cli` 분리), `TestModulePath_MatchesRepoHomepage`, replace 기반 소비 모듈로 경로 검증 |
| P0-3 버전 스탬프 | `var version`, Taskfile `-trimpath -ldflags -X main.version`, release `fetch-depth: 0` + 버전 검증 단계, 템플릿 5종 `var version` + 계약 테스트 |
| P0-4 dry-run | `planFiles` 추출로 디스크 미변경 + 실제 생성과 동일 목록 (`TestDryRun_*` 4종) |
| P0-5 --force 원자성 | 백업 → 교체 → 실패 시 복원 (`TestGenerate_ForceRenderFailureKeepsExistingDirectory`) |
| P0-6 스트림 분리 | `cli.Env` 주입, 결과=stdout/진행=stderr, `run()` 분리로 오류 중복 출력 제거 |
| P1-1 종료 코드 | 0/1/2/3/4/5 + `errors.As` 매핑 + 테이블 테스트 8케이스 |
| P1-2 기계 판독 출력 | `--format table\|plain\|json`(비TTY 기본 plain), JSON 스키마 테스트 |
| P1-3 표준 플래그 | `--no-color`, `--no-input`, `-q`, `-d`, `-y`, `-f`, `-n` + 플래그 등록 테스트 |
| P1-4 확인 절차 | TTY 확인 프롬프트, 비TTY는 코드 5 + 안내, `--no-input`/`--yes`/`--force` 처리 (테스트 5종) |
| P1-5 도움말 | 전체 실행 경로·예시·문서 링크, 루트 Long에 종료 코드 표 (`TestNewCmdHelp_*`) |
| P1-6 입력 검증 | 여분 인자 거부, 덮어쓰기 대상 절대 경로 안내 |
| P1-7 드리프트 | `task help`가 바이너리 도움말 사용, 템플릿 Go/wcli 버전 계약 테스트, README 플래그 대조 테스트 |
| P1-8 의존성 | wcli v0.2.3 (루트 + 템플릿 6종) |
| P1-9 super_cli | 삭제(템플릿과 드리프트된 생성물, 참조 없음) |
| P1-10 커버리지 | 86.6% (82.0% → 86.6%), `task coverage-check` 임계값 85% + CI 적용 |
| P2-1 CI | OS 매트릭스, `task lint`, shuffle, govulncheck, gitleaks, 커버리지 아티팩트 |
| P2-2 릴리스 | 게시 전 verify 잡, `-trimpath`+버전 주입, SHA-256, SLSA provenance, 릴리스 노트, 액션 SHA 고정, `cache: false` |
| P2-3 위생 | CHANGELOG, `.gitattributes`, 이슈/PR 템플릿, Dependabot(cooldown 포함) |
| P2-4 성능 | 벤치마크(`Generate` 115~226µs/op, `PlanFiles` 2.5~3.7µs) + `task bench`; 병목 없어 최적화는 하지 않음 |
| P2-5 문서 | README(플래그 표·스트림·종료 코드·JSON), AGENTS.md(CLI 계약 섹션), CONTRIBUTING.md |

### 보류 (근거 있음)

| 항목 | 사유 |
| --- | --- |
| 도움말 고정 문구 한국어화 | wcli v0.2.x 기본 템플릿이 `CommandPath`를 제공하지 않아 커스텀 템플릿을 복제·유지해야 함 → upstream 개선 요청으로 분리 (전체 경로·예시는 `Long`으로 이미 제공) |
| `--output` 상위 경로 거부 | 프로젝트 이름 검증으로 대상이 CWD/조상이 될 수 없고, 교체는 원자적이며 절대 경로를 고지한다. 거부 플래그는 사용자 요청 시 추가 |
| CODE_OF_CONDUCT / CODEOWNERS | 선택 항목. 필요해지면 추가 |
