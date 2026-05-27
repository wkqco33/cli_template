# wtemp

wcli 기반 Go CLI 프로젝트 템플릿 생성기.

## 설치

### ppm

```bash
ppm install wkqco33/cli_template
```

### 직접 빌드

```bash
git clone --recurse-submodules https://github.com/wkqco33/cli_template
cd cli_template
make install
```

## 기본 확인 명령

```bash
go run . --help
go run . list
make help
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

# SQLite + GORM 포함 생성
wtemp new <project-name> --sqlite
wtemp new <project-name> --sqlite -t gin

# 생성 단계별 성능 측정
wtemp new <project-name> -t library --profile
```

`--profile`를 사용하면 생성 단계 시간(`render`, `git`, `postprocess`, `total`)이 stderr로 출력된다.

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
- `fyne`, `library` 템플릿에서는 현재 변경 사항이 없다.
- SQLite 드라이버(`gorm.io/driver/sqlite`)는 CGO가 필요하므로 `gcc`가 설치되어 있어야 한다.

## 생성 후 시작

```bash
wtemp new my-tool
cd my-tool
go mod tidy
go build .
./my-tool --help
```

생성된 프로젝트에는 `wcli`가 git submodule로 자동 추가된다.

## 빌드

```bash
make          # 로컬 빌드
make install  # 빌드 후 ~/.local/bin 에 설치
make release  # 전체 플랫폼 릴리스 빌드 (dist/)
make clean    # 빌드 산출물 삭제
make uninstall
make help     # 사용법 출력
```

## 스모크 검증

핵심 템플릿 생성 후 `go build ./...` 컴파일 가능 여부를 자동 검증한다.

- 커버리지: `minimal/full` 기본 + sqlite on/off, `gin` 기본, `library` 기본
- sqlite 케이스는 `CGO_ENABLED=0`, `gcc` 미설치, windows 환경에서 skip 사유를 명시
- 각 케이스는 `t.TempDir()`를 사용해 생성 산출물을 자동 정리

```bash
go test -tags=smoke ./generator -run TestSmokeGeneratedTemplatesBuild -count=1
```

## 성능 측정 재현(로컬 기준)

네트워크 편차를 줄이기 위해 서브모듈이 없는 `library` 템플릿으로 측정한다.

```bash
go build -o ./wtemp .
for i in $(seq 1 5); do
  ./wtemp new "perf-lib-$i" -t library --profile >/dev/null
done
```
