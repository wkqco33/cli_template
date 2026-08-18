# wtemp

wcli 기반 Go CLI 프로젝트 템플릿 생성기.

## 설치

### ppm

```bash
ppm install wkqco33/cli_template
```

### 직접 빌드

```bash
git clone https://github.com/wkqco33/cli_template
cd cli_template
task install
```

## 기본 확인 명령

```bash
go run . --help
go run . list
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

# 생성 후 git 저장소 초기화
wtemp new <project-name> --git

# SQLite + GORM 포함 생성
wtemp new <project-name> --sqlite
wtemp new <project-name> --sqlite -t gin

# 생성 단계별 성능 측정
wtemp new <project-name> -t library --profile
```

`--profile`를 사용하면 생성 단계 시간(`render`, `postprocess`, `total`)이 stderr로 출력된다.

## 템플릿 목록

템플릿 이름/설명의 기준은 `wtemp list` 출력이다.

현재 출력:

```text
minimal      루트 커맨드만 있는 최소 구조
full         서브커맨드 + wcli 설정이 포함된 전체 구조
gin          CLI + Gin 웹서버 (REST API 스켈레톤)
fiber        CLI + Fiber 웹서버 (REST API 스켈레톤)
echo         CLI + Echo 웹서버 (REST API 스켈레톤)
fyne         CLI + Fyne GUI 앱
library      Go 라이브러리 스켈레톤
```

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
task test      # 단위 테스트 실행
task smoke     # 생성된 템플릿의 go build 검증
task clean     # 빌드 산출물 삭제
task uninstall
task help      # wtemp 사용법 출력
```

## 개발 (TDD)

이 프로젝트는 **TDD(Test-Driven Development)** 방식으로 개발한다.
에이전트/개발자를 위한 상세 규칙은 [`AGENTS.md`](AGENTS.md)를 참고한다.

```bash
task test          # 단위 테스트 (캐시 무시)
task test-race     # 데이터 레이스 탐지 포함
task test-watch    # 파일 변경 시 자동 재실행 (watchexec 필요)
task coverage      # 커버리지 측정 + 요약 (coverage.out)
task coverage-html # HTML 리포트 생성 (coverage.html)
```

- 기능 추가/버그 수정은 **테스트를 먼저 작성**하고(Red), 최소 구현(Green), 리팩터링(Refactor) 순서로 진행한다.
- 커밋 전 `task test`가 통과해야 하며, 템플릿/생성 로직 변경 시 `task smoke`도 확인한다.

## 스모크 검증

핵심 템플릿 생성 후 `go build ./...` 컴파일 가능 여부를 자동 검증한다.

- 커버리지: `minimal/full` 기본 + sqlite on/off, `gin`/`fiber`/`echo` 기본, `library` 기본
- `fyne`은 GUI 빌드에 X11/OpenGL 등 시스템 헤더가 필요해 스모크 매트릭스에서 제외(수동 빌드로 확인)
- sqlite 케이스는 `CGO_ENABLED=0`, `gcc` 미설치, windows 환경에서 skip 사유를 명시
- 각 케이스는 `t.TempDir()`를 사용해 생성 산출물을 자동 정리

```bash
task smoke
# 또는
go test -tags=smoke ./generator -run TestSmokeGeneratedTemplatesBuild -count=1
```

## 성능 측정 재현(로컬 기준)

네트워크 편차를 줄이기 위해 의존성이 적은 `library` 템플릿으로 측정한다.

```bash
go build -o ./wtemp .
for i in $(seq 1 5); do
  ./wtemp new "perf-lib-$i" -t library --profile >/dev/null
done
```
