# super_cli

wcli 기반 Go CLI 예제 프로젝트.

## 빌드

```bash
go build -o super_cli .
```

## 실행 예시

```bash
./super_cli --help
./super_cli version
./super_cli example --name foo
./super_cli example --name foo --verbose
./super_cli completion bash > super_cli.bash
```

## 설정

`config.yaml`과 환경변수(`SUPER_CLI_` 접두사)를 함께 사용한다.

| 환경변수 | 기본값 | 설명 |
|---------|-------|------|
| `SUPER_CLI_NAME` | `super_cli` | 앱 이름 |
| `SUPER_CLI_SERVER_HOST` | `0.0.0.0` | 서버 호스트 |
| `SUPER_CLI_SERVER_PORT` | `8080` | 서버 포트 |
| `SUPER_CLI_LOG_LEVEL` | `info` | 로그 레벨 |
