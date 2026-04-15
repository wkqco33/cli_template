# wtemp

`wcli` + `wconf` 기반 Go CLI 프로젝트 템플릿 생성기.

새 프로젝트를 시작할 때 보일러플레이트 설정 없이 바로 기능 구현에 집중할 수 있도록 프로젝트 골격을 자동으로 생성해준다.

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

## 사용법

```bash
# 템플릿 목록 확인
wtemp list

# 새 프로젝트 생성 (기본: full 템플릿)
wtemp new <project-name>

# 템플릿 지정
wtemp new <project-name> --template minimal
wtemp new <project-name> --template full

# SQLite + GORM 추가
wtemp new <project-name> --sqlite
wtemp new <project-name> --sqlite --template minimal
```

## 템플릿 종류

### `minimal`

루트 커맨드 하나만 있는 최소 구조. 단순한 CLI 도구에 적합.

```
my-tool/
├── go.mod
├── main.go
└── cmd/
│   └── root.go
├── wcli/        ← git submodule
```

### `full`

서브커맨드, wconf 설정 관리, 로깅 구조가 포함된 전체 구조. 실제 서비스 수준의 CLI 도구에 적합.

```
my-tool/
├── go.mod
├── main.go
├── config.yaml
├── README.md
├── cmd/
│   ├── root.go
│   ├── version.go
│   └── example.go
├── config/
│   └── config.go
├── wcli/        ← git submodule
└── wconf/       ← git submodule
```

### `--sqlite` 옵션

어느 템플릿에든 `--sqlite` 플래그를 추가하면 GORM + SQLite 설정이 포함된다.

```
my-tool/
├── ...
└── database/
    ├── db.go            ← DB 초기화 (gorm.Open, AutoMigrate)
    └── models/
        └── example.go   ← 예시 모델 (gorm.Model 임베드)
```

> `gorm.io/driver/sqlite`는 CGO가 필요하다. 빌드 환경에 gcc가 있어야 한다.

## 생성 후 시작하기

```bash
wtemp new my-tool
cd my-tool
go mod tidy
go build .
./my-tool --help
```

`wcli`, `wconf`는 프로젝트 생성 시 git submodule로 자동 추가된다.

## 설정 (full 템플릿)

`config.yaml` 또는 환경변수로 설정을 관리한다. 환경변수는 `{프로젝트명 대문자}_` 접두사를 사용한다.

```yaml
# config.yaml
name: my-tool
server:
  host: 0.0.0.0
  port: 8080
log:
  level: info
```

```bash
MY_TOOL_SERVER_PORT=9090 ./my-tool example
```

## 빌드

```bash
make          # 로컬 빌드
make install  # 빌드 후 ~/.local/bin 에 설치
make release  # 전체 플랫폼 릴리스 빌드 (dist/)
make clean    # 빌드 산출물 삭제
make uninstall
make help     # 사용법 출력
```

## 의존 라이브러리

| 라이브러리 | 역할 |
|-----------|------|
| [wcli](https://github.com/wkqco33/wcli) | CLI 프레임워크 (커맨드 트리, 플래그, rich 출력) |
| [wconf](https://github.com/wkqco33/wconf) | 설정 관리 (env, .env, YAML, TOML) |
| [gorm](https://gorm.io) | ORM (`--sqlite` 옵션 시 포함) |
| [gorm/driver/sqlite](https://github.com/go-gorm/sqlite) | SQLite 드라이버 (`--sqlite` 옵션 시 포함, CGO 필요) |
