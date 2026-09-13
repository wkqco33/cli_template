# wtemp

wcli 기반 Go CLI 프로젝트 템플릿 생성기. 여러 Go CLI, 웹 서버, GUI, 라이브러리 프로젝트를 일관된 구조로 생성합니다.

- 패키지 메타데이터: [`ppm.json`](ppm.json)
- 기여 방법: [`CONTRIBUTING.md`](CONTRIBUTING.md)
- 보안 신고: [`SECURITY.md`](SECURITY.md)
- 라이선스: [`LICENSE`](LICENSE)

## 설치

### ppm

```bash
ppm install wkqco33/cli_template
```

릴리스 아카이브마다 `.sha256` 체크섬 파일이 함께 제공됩니다. 운영 환경에서는 패키지 관리자의 무결성 검증을 활성화하세요.

### go install

```bash
go install github.com/wkqco33/cli_template/cmd/wtemp@latest
```

메인 패키지는 `cmd/wtemp`에 있으므로 위 경로로 설치하면 바이너리 이름이 `wtemp`가 된다.

> `v0.2.1` 이하 릴리스는 모듈 경로가 저장소 URL과 달라 `go install`이 실패하고, 메인 패키지가 루트에 있었다. 다음 릴리스부터 위 명령이 동작한다.

### 직접 빌드

```bash
git clone https://github.com/wkqco33/cli_template.git
cd cli_template
task install
```

## 출력 규칙

결과는 stdout, 진행·경고·오류·프로파일은 stderr로 출력한다.
파이프·CI에서 결과만 받아 쓰려면 stdout을 그대로 사용한다.

```bash
# 생성될 파일 목록만 받기 (헤더/안내는 stderr)
wtemp new myapp -t minimal --dry-run | while read -r f; do echo "check $f"; done
```

`wtemp --version`은 릴리스 빌드 시 `-ldflags "-X main.version=<tag>"`로 주입된 값을 출력한다.

## 기본 확인 명령

```bash
go run ./cmd/wtemp --help
go run ./cmd/wtemp list
task help
```

## 사용법

```bash
# 템플릿 목록
wtemp list

# 프로젝트 생성 (기본 템플릿: full)
wtemp new <project-name>

# 템플릿 지정
wtemp new <project-name> -t minimal
wtemp new <project-name> --template gin

# Go 모듈 경로 지정 (기본값: 프로젝트 이름)
wtemp new <project-name> --module github.com/user/<project-name>

# 대상 디렉토리가 이미 존재해도 덮어쓰기
wtemp new <project-name> --force

# 생성 위치 지정
wtemp new <project-name> -o ./sub/dir

# 실제 생성 없이 생성될 파일 목록만 미리 보기
wtemp new <project-name> --dry-run
wtemp new <project-name> --dry-run --format json

# 생성 후 git 저장소 초기화
wtemp new <project-name> --git

# SQLite + GORM 포함 생성
wtemp new <project-name> --sqlite
wtemp new <project-name> --sqlite -t gin

# 생성 단계별 성능 측정
wtemp new <project-name> -t library --profile
```

`--profile`을 사용하면 생성 단계 시간(`render`, `postprocess`, `total`)이 stderr로 출력된다.

### 플래그

전역 플래그는 서브커맨드 앞뒤 어디에나 쓸 수 있다. (`wtemp --quiet new app`)

| 플래그 | 단축 | 대상 | 설명 |
| --- | --- | --- | --- |
| `--template` | `-t` | `new` | 사용할 템플릿 (기본값: `full`) |
| `--sqlite` | | `new` | SQLite + GORM 추가 |
| `--profile` | | `new` | 생성 단계별 성능 프로파일 출력 |
| `--module` | | `new` | Go 모듈 경로 (기본값: 프로젝트 이름) |
| `--force` | `-f` | `new` | 대상이 이미 존재해도 확인 없이 덮어쓰기 |
| `--output` | `-o` | `new` | 생성 위치 (기본값: 현재 디렉토리) |
| `--dry-run` | `-n` | `new` | 생성 없이 파일 목록만 출력 |
| `--git` | | `new` | 생성 후 git 저장소 초기화 |
| `--format` | | `new`, `list` | 출력 형식 `table` · `plain` · `json` |
| `--no-color` | | 전역 | 색상 출력 사용 안 함 (`NO_COLOR` 환경변수도 동일) |
| `--no-input` | | 전역 | 대화형 입력 금지 (프롬프트가 필요하면 종료 코드 5로 실패) |
| `--quiet` | `-q` | 전역 | 진행 메시지 억제 (결과·오류는 유지) |
| `--debug` | `-d` | 전역 | 추가 진단 메시지 출력 |
| `--yes` | `-y` | 전역 | 확인 프롬프트에 자동으로 yes |
| `--help` | `-h` | 전역 | 도움말 |
| `--version` | | 전역 | 버전 출력 |

### 출력 스트림

- **stdout**: 결과 (템플릿 목록, dry-run 파일 목록, JSON)
- **stderr**: 진행·경고·안내·오류·프로파일

파이프·CI에서는 stdout이 터미널이 아니면 `--format` 기본값이 `plain`이 된다.

```bash
wtemp list --format json | jq -r '.templates[].name'
wtemp new myapp --dry-run | while read -r f; do echo "check $f"; done
```

### 종료 코드

| 코드 | 의미 |
| --- | --- |
| 0 | 성공 (사용자가 프롬프트에서 취소한 경우 포함) |
| 1 | 일반·내부 오류 |
| 2 | 사용법·입력 검증 오류 (인자 누락, 잘못된 이름/템플릿/플래그) |
| 3 | 대상 경로 충돌 |
| 4 | 외부 도구(git) 실행 실패 |
| 5 | 비대화형 환경에서 확인 입력이 필요 |

### 덮어쓰기 확인

대상 디렉토리가 이미 존재하면 덮어쓸지 확인한다.

- TTY: `y/N` 프롬프트로 확인 (기본값 N)
- 비TTY(파이프·CI): `--force` 또는 `--yes` 없이는 종료 코드 5로 실패
- `--no-input`을 주면 프롬프트를 띄우지 않고 같은 방식으로 실패
- 확인 후 교체는 원자적으로 수행되며, 생성이 실패하면 기존 디렉토리가 복원된다

## 템플릿 목록

템플릿 이름/설명의 기준은 `wtemp list` 출력이다. `--format plain`은 `이름<TAB>설명`, `--format json`은 JSON 한 줄이다.

현재 템플릿:

- `minimal` — 루트 커맨드만 있는 최소 구조
- `full` — 서브커맨드 + wcli 설정이 포함된 전체 구조
- `gin` — CLI + Gin 웹서버 (REST API 스켈레톤)
- `fiber` — CLI + Fiber 웹서버 (REST API 스켈레톤)
- `echo` — CLI + Echo 웹서버 (REST API 스켈레톤)
- `fyne` — CLI + Fyne GUI 앱
- `library` — Go 라이브러리 스켈레톤

`--format json`은 이름·설명에 더해 해당 템플릿의 `--sqlite` 지원 여부(`sqlite`)를 함께 출력한다.

## `--sqlite` 옵션

- `--sqlite`는 `minimal`, `full`, `gin`, `fiber`, `echo` 템플릿에서만 실제 코드/의존성에 반영된다.
- `fyne`, `library` 템플릿에서는 지원하지 않으며, 지정 시 경고가 출력되고 옵션이 무시된다.
- SQLite 드라이버(`gorm.io/driver/sqlite`)는 CGO가 필요하므로 `gcc`가 설치되어 있어야 한다.

## 생성 후 시작

```bash
wtemp new my-tool
cd my-tool
go mod tidy
go build .
./my-tool --help
```

생성된 프로젝트는 공개된 `wcli` 라이브러리를 Go 모듈 의존성으로 사용한다.

## 빌드

[Task](https://taskfile.dev)를 빌드 도구로 사용한다.

```bash
task           # 사용 가능한 태스크 목록 출력
task build     # 로컬 빌드
task install   # 빌드 후 ~/.local/bin 에 설치
task release   # 전체 플랫폼 릴리스 빌드 (dist/)
task test            # 단위 테스트 실행
task test-race       # 데이터 레이스 탐지 포함
task lint            # gofmt + go vet
task coverage-check  # 커버리지 임계값(85%) 검증
task vuln            # govulncheck (별도 설치 필요)
task bench           # 생성 성능 벤치마크
task smoke           # 생성된 템플릿의 go build 검증
task clean           # 빌드 산출물 삭제
task uninstall
task help            # wtemp 사용법 출력 (바이너리 --help 기반)
```

## 개발 (TDD)

이 프로젝트는 **TDD(Test-Driven Development)** 방식으로 개발한다.
에이전트/개발자를 위한 상세 규칙은 [`AGENTS.md`](AGENTS.md)를 참고한다.

```bash
task test          # 단위 테스트 (캐시 무시)
task test-race     # 데이터 레이스 탐지 포함
task test-watch    # 파일 변경 시 자동 재실행 (watchexec 필요)
task coverage      # 커버리지 측정 + 요약 (coverage.out)
task coverage-check # 커버리지 임계값(85%) 검증
task coverage-html # HTML 리포트 생성 (coverage.html)
task lint          # gofmt + go vet
task vuln          # 알려진 취약점 검사
```

- 기능 추가/버그 수정은 **테스트를 먼저 작성**하고(Red), 최소 구현(Green), 리팩터링(Refactor) 순서로 진행한다.
- 커밋 전 `task test`가 통과해야 하며, 템플릿/생성 로직 변경 시 `task smoke`도 확인한다.
- 템플릿 계약 테스트가 `templates/*/go.mod.tmpl`의 Go/wcli 버전을 루트 `go.mod`와 대조하고,
  README·`task help` 드리프트도 테스트로 감지한다.
- 커버리지 임계값은 85%이며, 집계는 Go 툴체인에 따라 다르다(go1.26.1 기준 88.7%).

## 스모크 검증

핵심 템플릿 생성 후 `go build ./...` 컴파일 가능 여부를 자동 검증한다.

- 커버리지: `minimal/full` 기본 + sqlite on/off, `gin`/`fiber`/`echo` 기본, `library` 기본,
  `--module` 경로 지정 케이스(`full`, `library`)
- `fyne`은 GUI 빌드에 X11/OpenGL 등 시스템 헤더가 필요해 스모크 매트릭스에서 제외(수동 빌드로 확인)
- sqlite 케이스는 `CGO_ENABLED=0`, `gcc` 미설치, windows 환경에서 skip 사유를 명시
- 각 케이스는 `t.TempDir()`를 사용해 생성 산출물을 자동 정리

```bash
task smoke
# 또는
go test -tags=smoke ./generator -run TestSmokeGeneratedTemplatesBuild -count=1
```

## 성능 측정

생성 경로는 서브밀리초 수준이며, 회귀 감지를 위해 벤치마크를 둔다.

```bash
task bench

# 단계별 시간을 포함한 수동 측정
go build -o ./wtemp ./cmd/wtemp
./wtemp new perf-lib -t library --profile
```

측정된 기준값(로컬):

- `--profile`(CLI 전체): total 234~594µs / render 193~520µs
- `task bench`(`library`): 약 191µs/op (`full` 약 226µs/op, `minimal` 약 115µs/op)

실제 지연의 대부분은 생성 후 사용자가 실행하는 `go mod tidy`(네트워크)에 있으므로,
생성 경로는 벤치마크로 회귀만 감시하고 렌더링 병렬화 같은 최적화는 하지 않는다.
