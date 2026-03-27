# super_cli

wcli + wconf 기반 Go CLI 프로젝트.

## 빌드

```bash
go build -o super_cli .
```

## 사용법

```bash
./super_cli --help
./super_cli version
./super_cli example --name foo --verbose
```

## 설정

`config.yaml` 또는 환경변수로 설정합니다.

| 환경변수 | 기본값 | 설명 |
|---------|-------|------|
| `SUPER_CLI_SERVER_HOST` | `0.0.0.0` | 서버 호스트 |
| `SUPER_CLI_SERVER_PORT` | `8080` | 서버 포트 |
| `SUPER_CLI_LOG_LEVEL` | `info` | 로그 레벨 |
