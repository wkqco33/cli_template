## 변경 요약

<!-- 무엇을, 왜 바꿨는지 적어 주세요. -->

## 관련 이슈

Closes #

## 검증

```bash
gofmt -w .
go mod verify
go vet ./...
task test
task test-race
task coverage-check
task smoke     # 템플릿/생성 로직을 바꾼 경우
```

- [ ] 선행 실패 테스트(Red)를 추가한 뒤 구현했다
- [ ] 테스트와 구현을 같은 커밋에 포함했다
- [ ] 템플릿을 바꿨다면 `task smoke`를 통과했다
- [ ] 사용자 동작·플래그·출력이 바뀌었다면 `README.md`/`CHANGELOG.md`를 갱신했다
- [ ] `templates/*/go.mod.tmpl`의 Go·wcli 버전을 루트 `go.mod`와 맞췄다

## 사용자 영향

<!-- 종료 코드, 출력 스트림, 플래그 변경 등 -->
