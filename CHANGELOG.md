# Changelog

이 프로젝트의 주요 변경 사항을 기록한다.
형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.1.0/),
버전 규칙은 [Semantic Versioning](https://semver.org/lang/ko/)을 따른다.

## [Unreleased]

### Added

- 자연어 요청으로 기존 Go 템플릿을 선택·생성하는 `wtemp ai` 명령
- Ollama 기본 provider 및 OpenAI, OpenAI-compatible, Azure provider 지원
- 플랫폼별 YAML 설정 파일과 `config init|path|show|set|unset|validate` 명령
- `LLM_client_go` 기반 구조화된 AI 생성 계획 검증

### Changed

- 프로젝트 및 템플릿의 Go 버전을 1.26.6으로 상향

## [0.3.1] - 2026-09-13

### Fixed

- `go install github.com/wkqco33/cli_template/cmd/wtemp@vX.Y.Z`로 설치한 바이너리의
  `--version`이 `dev`로 표시되던 문제 (빌드 정보의 모듈 버전으로 폴백)

### Changed

- 릴리스 노트를 `CHANGELOG.md`의 해당 버전 섹션에서 만들고, 없으면 태그 메시지로 폴백한다

## [0.3.0] - 2026-09-13

### Added

- 전역 플래그: `--no-color`, `--no-input`, `-q/--quiet`, `-d/--debug`, `-y/--yes`
- `list`/`new --dry-run`의 `--format table|plain|json` 출력 (비TTY 기본값은 `plain`)
- 문서화된 종료 코드: 0/1/2/3/4/5
- 덮어쓰기 확인 프롬프트(TTY)와 `--force`/`--yes`/`--no-input` 처리
- 서브커맨드 도움말의 전체 실행 경로·예시·문서 링크
- 생성 프로젝트의 `--version` 지원과 `-X main.version` 주입
- `task lint`, `task vuln`, `task coverage-check`, `task bench`
- CI: OS 매트릭스(ubuntu/windows/macos), `govulncheck`, gitleaks, 커버리지 임계값(85%)
- 릴리스: 게시 전 테스트, `-trimpath`, SHA-256 체크섬, SLSA provenance, 릴리스 노트 자동 생성
- `CHANGELOG.md`, `.gitattributes`, 이슈/PR 템플릿, Dependabot 설정

### Changed

- 모듈 경로를 `github.com/wkqco33/cli_template`으로 변경하고 메인 패키지를 `cmd/wtemp`로 이동
  (`go install github.com/wkqco33/cli_template/cmd/wtemp@latest`로 `wtemp` 바이너리 설치)
- `wcli`를 v0.2.3으로 업그레이드 (루트 및 모든 템플릿)
- 결과는 stdout, 진행·경고·오류는 stderr로 분리
- 템플릿 `go.mod`의 Go 버전을 저장소와 일치(1.26.1)

### Fixed

- `library` 템플릿이 `--module` 경로에 `github.com/`을 중복으로 붙이던 문제
- `--dry-run`이 출력 디렉터리를 실제로 생성하던 문제
- `--force`가 렌더링 실패 시 기존 디렉터리를 복구할 수 없게 삭제하던 문제 (백업·복원으로 변경)
- 오류 메시지가 중복 출력되던 문제
- 여분 위치 인자를 조용히 무시하던 문제
- `--version`이 하드코딩된 `0.1.0`을 출력하던 문제

### Removed

- 템플릿과 드리프트된 생성 예제 `super_cli/`
