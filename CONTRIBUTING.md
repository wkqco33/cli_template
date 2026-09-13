# 기여 가이드

## 시작하기

```bash
git clone https://github.com/wkqco33/cli_template.git
cd cli_template
task test
```

Go 1.26.1 이상과 [Task](https://taskfile.dev/)가 필요합니다.

## 개발 절차

1. 이슈를 확인하고 변경 범위를 정합니다.
2. 버그 또는 기능을 재현하는 테스트를 먼저 작성합니다(Red).
3. 테스트를 통과시키는 최소 구현을 작성합니다(Green).
4. 중복과 이름을 정리합니다(Refactor).
5. 템플릿을 변경했다면 `task smoke`를 실행합니다.

테스트와 구현은 같은 변경에 포함해야 합니다. 자세한 규칙은 [`AGENTS.md`](AGENTS.md)를 참고하세요.

## 검증 명령

```bash
gofmt -w .
go vet ./...
go mod verify
task test
task test-race
task coverage-check
task smoke        # 템플릿/생성 로직을 바꿨을 때
```

템플릿의 smoke 테스트는 네트워크로 의존성을 내려받을 수 있어야 합니다.
CI는 ubuntu·windows·macos에서 단위 테스트를 돌리고, Linux에서 race·커버리지·smoke를
추가로 수행합니다. `govulncheck`와 시크릿 스캔(gitleaks)도 함께 돌아간다.

## 커밋 및 풀 리퀘스트

- 커밋 제목은 `type: 요약` 형식을 사용합니다. 예: `fix: validate module path`
- 변경 이유, 테스트 결과, 사용자에게 영향을 주는 변경을 PR 본문에 적습니다.
- 생성된 바이너리, 릴리스 아카이브, 비밀 정보는 커밋하지 않습니다.
- 호환성이 깨지는 변경은 사전에 이슈로 논의합니다.
- 사용자 동작·플래그·출력·종료 코드가 바뀌면 `README.md`와 `CHANGELOG.md`를 갱신합니다.

## 템플릿 변경

`templates/` 파일은 생성 프로젝트의 소스가 됩니다. 템플릿을 수정할 때는 해당 결과물이 `go build ./...`를 통과하는지 확인하고, 필요한 경우 생성 결과의 README도 함께 갱신합니다.

- `templates/*/go.mod.tmpl`의 Go 버전·wcli 버전은 루트 `go.mod`와 **함께** 올려야 합니다.
  (계약 테스트 `TestTemplates_GoModMatchesRootModule`이 어긋나면 실패합니다.)
- CLI 템플릿의 `version`은 상수가 아니라 변수여야 합니다(`-X main.version` 주입).
  (계약 테스트 `TestTemplates_VersionIsStampable`이 검사합니다.)

## 라이선스

기여물은 저장소의 [MIT License](LICENSE)에 따라 제공되는 것으로 간주합니다.
