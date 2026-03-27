BIN     := wtemp
INSTALL := $(HOME)/.local/bin/$(BIN)
DIST    := dist

PLATFORMS := linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64

.PHONY: all build clean install uninstall release $(PLATFORMS)

all: build

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
