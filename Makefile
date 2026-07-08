APP_NAME := git-worktree-manager
DIST_DIR := dist
FYNE := $(shell command -v fyne 2>/dev/null)
FYNE_CROSS := $(shell command -v fyne-cross 2>/dev/null)

.PHONY: all run test build package package-local package-all install-tools clean

all: test build

run:
	go run .

test:
	go test ./...

build:
	go build -o $(APP_NAME) .

install-tools:
	go install fyne.io/tools/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest

package-local: install-tools
	@mkdir -p $(DIST_DIR)
	fyne package -release -name "$(APP_NAME)"
	@mv "$(APP_NAME).app" $(DIST_DIR)/ 2>/dev/null || true
	@mv "$(APP_NAME)"* $(DIST_DIR)/ 2>/dev/null || true
	@mv *.exe $(DIST_DIR)/ 2>/dev/null || true
	@mv *.tar.xz $(DIST_DIR)/ 2>/dev/null || true
	@mv *.zip $(DIST_DIR)/ 2>/dev/null || true
	@echo "Packaged for local OS into $(DIST_DIR)/"

package-linux:
	@command -v fyne-cross >/dev/null || (echo "Run: make install-tools" && exit 1)
	fyne-cross linux -arch=amd64,arm64 -name $(APP_NAME) -release
	@mkdir -p $(DIST_DIR)/linux-amd64 $(DIST_DIR)/linux-arm64
	@cp -r fyne-cross/dist/linux-amd64/* $(DIST_DIR)/linux-amd64/ 2>/dev/null || true
	@cp -r fyne-cross/dist/linux-arm64/* $(DIST_DIR)/linux-arm64/ 2>/dev/null || true
	@echo "Linux packages in $(DIST_DIR)/linux-*/"

package-windows:
	@command -v fyne-cross >/dev/null || (echo "Run: make install-tools" && exit 1)
	fyne-cross windows -arch=amd64,arm64 -name $(APP_NAME) -release
	@mkdir -p $(DIST_DIR)/windows-amd64 $(DIST_DIR)/windows-arm64
	@cp -r fyne-cross/dist/windows-amd64/* $(DIST_DIR)/windows-amd64/ 2>/dev/null || true
	@cp -r fyne-cross/dist/windows-arm64/* $(DIST_DIR)/windows-arm64/ 2>/dev/null || true
	@echo "Windows packages in $(DIST_DIR)/windows-*/"

package-darwin:
	@command -v fyne-cross >/dev/null || (echo "Run: make install-tools" && exit 1)
	fyne-cross darwin -arch=amd64,arm64 -name $(APP_NAME) -release
	@mkdir -p $(DIST_DIR)/darwin-amd64 $(DIST_DIR)/darwin-arm64
	@cp -r fyne-cross/dist/darwin-amd64/* $(DIST_DIR)/darwin-amd64/ 2>/dev/null || true
	@cp -r fyne-cross/dist/darwin-arm64/* $(DIST_DIR)/darwin-arm64/ 2>/dev/null || true
	@echo "macOS packages in $(DIST_DIR)/darwin-*/"

package-all: package-linux package-windows package-darwin
	@echo "All platform packages copied under $(DIST_DIR)/"

clean:
	rm -rf $(DIST_DIR) fyne-cross $(APP_NAME) $(APP_NAME).app
