BIN     := wtemp
INSTALL := $(HOME)/.local/bin/$(BIN)
DIST    := dist

PLATFORMS := linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64

.PHONY: all build clean install uninstall release help $(PLATFORMS)

all: build

# wtemp 사용법 출력
# --sqlite 옵션은 생성된 프로젝트에 GORM + SQLite를 추가함 (CGO 필요)
help:
	@echo "사용법: wtemp new <project-name> [options]"
	@echo ""
	@echo "옵션:"
	@echo "  -t, --template  minimal | full    템플릿 선택 (기본값: full)"
	@echo "      --sqlite                       SQLite + GORM 지원 추가"
	@echo ""
	@echo "예시:"
	@echo "  wtemp new my-app"
	@echo "  wtemp new my-app -t minimal"
	@echo "  wtemp new my-app --sqlite"
	@echo "  wtemp new my-app --sqlite -t minimal"
	@echo ""
	@echo "  참고: --sqlite 옵션으로 생성된 프로젝트는 CGO(gcc)가 필요합니다."
	@echo ""
	@echo "Makefile 타겟:"
	@echo "  build     현재 플랫폼용 빌드"
	@echo "  install   $(INSTALL) 에 설치"
	@echo "  release   전체 플랫폼 릴리스 빌드 (dist/)"
	@echo "  clean     빌드 산출물 정리"

build:
	go build -o $(BIN) .

clean:
	rm -f $(BIN)
	rm -rf $(DIST)

install: build
	cp $(BIN) $(INSTALL)
	@echo "설치 완료: $(INSTALL)"

uninstall:
	rm -f $(INSTALL)
	@echo "삭제 완료: $(INSTALL)"

release: clean $(PLATFORMS)
	@echo "릴리스 완료: $(DIST)/"
	@ls -lh $(DIST)/

$(PLATFORMS):
	$(eval OS   := $(word 1, $(subst _, ,$@)))
	$(eval ARCH := $(word 2, $(subst _, ,$@)))
	$(eval OUT  := $(DIST)/$(BIN)_$@)
	@mkdir -p $(DIST)
	$(if $(filter windows,$(OS)), \
		GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUT).exe . && \
		cd $(DIST) && zip $(BIN)_$@.zip $(BIN)_$@.exe && rm $(BIN)_$@.exe, \
		GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUT) . && \
		tar -czf $(OUT).tar.gz -C $(DIST) $(BIN)_$@ && rm $(OUT) \
	)
