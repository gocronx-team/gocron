GO111MODULE=on

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date '+%Y-%m-%d %H:%M:%S')
LDFLAGS = -w -X 'main.AppVersion=$(VERSION)' -X 'main.BuildDate=$(BUILD_DATE)' -X 'main.GitCommit=$(GIT_COMMIT)'

# 构建目录
BIN_DIR = bin
PACKAGE_DIR = packages

# 默认目标
.DEFAULT_GOAL := build

# 本地构建
.PHONY: build
build: gocron node

.PHONY: build-race
build-race: enable-race build

.PHONY: run
run: build kill
	./$(BIN_DIR)/gocron-node &
	./$(BIN_DIR)/gocron web -e dev

.PHONY: run-race
run-race: enable-race run

.PHONY: kill
kill:
	-killall gocron-node

.PHONY: gocron
gocron:
	@mkdir -p $(BIN_DIR)
	go build $(RACE) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/gocron ./cmd/gocron

.PHONY: node
node:
	@mkdir -p $(BIN_DIR)
	go build $(RACE) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/gocron-node ./cmd/node

.PHONY: test
test:
	go test $(RACE) ./...

.PHONY: test-race
test-race: enable-race test

.PHONY: enable-race
enable-race:
	$(eval RACE = -race)

# 多平台打包
.PHONY: package
package: build-web
	@echo "Building packages for current platform..."
	bash ./package.sh

.PHONY: package-linux
package-linux: build-web
	@echo "Building packages for Linux..."
	bash ./package.sh -p linux -a "amd64,arm64"

.PHONY: package-darwin
package-darwin: build-web
	@echo "Building packages for macOS..."
	bash ./package.sh -p darwin -a "amd64,arm64"

.PHONY: package-windows
package-windows: build-web
	@echo "Building packages for Windows..."
	bash ./package.sh -p windows -a "amd64,arm64"

.PHONY: package-all
package-all: build-web
	@echo "Building packages for all platforms..."
	bash ./package.sh -p "linux,darwin,windows" -a "amd64,arm64"

# 前端构建
.PHONY: build-vue
build-vue:
	@echo "Building Vue frontend..."
	cd web/vue && yarn run build
	cp -r web/vue/dist/* web/public/

.PHONY: install-vue
install-vue:
	@echo "Installing Vue dependencies..."
	cd web/vue && yarn install

.PHONY: run-vue
run-vue:
	@echo "Starting Vue dev server..."
	cd web/vue && yarn run dev

.PHONY: statik
statik:
	@echo "Generating static assets..."
	go install github.com/rakyll/statik@latest
	go generate ./...

.PHONY: build-web
build-web: build-vue statik
	@echo "Web build complete!"

# 代码质量
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# 清理
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	-rm -rf $(BIN_DIR)
	-rm -rf $(PACKAGE_DIR)
	-rm -rf gocron-package
	-rm -rf gocron-node-package
	-rm -rf gocron-build
	-rm -rf gocron-node-build

.PHONY: clean-web
clean-web:
	@echo "Cleaning web build artifacts..."
	-rm -rf web/vue/dist
	-rm -rf web/public/static
	-rm -f web/public/index.html

# 开发工具
.PHONY: dev-deps
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/cosmtrek/air@latest
	go install github.com/rakyll/statik@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build gocron and gocron-node for current platform"
	@echo "  run            - Build and run in development mode"
	@echo "  test           - Run tests"
	@echo "  package        - Build packages for current platform"
	@echo "  package-linux  - Build packages for Linux (amd64, arm64)"
	@echo "  package-darwin - Build packages for macOS (amd64, arm64)"
	@echo "  package-windows- Build packages for Windows (amd64, arm64)"
	@echo "  package-all    - Build packages for all platforms"
	@echo "  build-vue      - Build Vue frontend"
	@echo "  install-vue    - Install Vue dependencies"
	@echo "  run-vue        - Start Vue dev server"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev-deps       - Install development dependencies"
	@echo "  help           - Show this help message"
