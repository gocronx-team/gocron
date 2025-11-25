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

.PHONY: run-with-packages
run-with-packages: build-web package-all kill
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
	CGO_ENABLED=0 go build $(RACE) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/gocron-node ./cmd/node

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

.PHONY: package-linux-nosqlite
package-linux-nosqlite: build-web
	@echo "Building Linux packages without SQLite..."
	CGO_ENABLED=0 bash ./package.sh -p linux -a "amd64,arm64"

.PHONY: package-darwin
package-darwin: build-web
	@echo "Building packages for macOS..."
	bash ./package.sh -p darwin -a "amd64,arm64"

.PHONY: package-windows
package-windows: build-web
	@echo "Building packages for Windows..."
	bash ./package.sh -p windows -a "amd64"

.PHONY: package-all
package-all: build-web
	@echo "Building packages for all platforms..."
	bash ./package.sh -p "linux,darwin" -a "amd64,arm64"
	bash ./package.sh -p "windows" -a "amd64"

# 前端构建
.PHONY: build-vue
build-vue:
	@echo "Installing Vue dependencies..."
	cd web/vue && yarn install
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

# 代码质量检查
.PHONY: check
check: fmt vet test
	@echo "✅ All checks passed!"

.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run || echo "⚠️  Linter found issues (non-blocking)"; \
	else \
		echo "⚠️  golangci-lint not installed, skipping..."; \
		echo "   Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@find . -name '*.go' -not -path './internal/statik/statik.go' -exec gofmt -w {} \;

.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@unformatted=$$(gofmt -l . | grep -v 'internal/statik/statik.go'); \
	if [ -n "$$unformatted" ]; then \
		echo "❌ Code not formatted:" && echo "$$unformatted" && echo "Run 'make fmt' to fix" && exit 1; \
	fi
	@echo "✅ Code formatting OK"

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✅ go vet passed"

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

.PHONY: security
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./... || echo "⚠️  Security issues found (non-blocking)"; \
	else \
		echo "⚠️  gosec not installed, skipping..."; \
		echo "   Install: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# 预发布检查（完整检查）
.PHONY: pre-release
pre-release: clean
	@echo "=========================================="
	@echo "Running pre-release checks..."
	@echo "=========================================="
	@$(MAKE) fmt-check
	@$(MAKE) vet
	@$(MAKE) test
	@echo ""
	@echo "=========================================="
	@echo "✅ All pre-release checks passed!"
	@echo "=========================================="
	@echo ""
	@echo "Optional checks (run manually if needed):"
	@echo "  make lint      - Code quality linter"
	@echo "  make security  - Security vulnerability scan"

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
	-rm -f coverage.out coverage.html

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
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# 版本管理
.PHONY: version
version:
	@echo "Current version: $(VERSION)"
	@echo "Recent releases:"
	@git tag -l "v*.*.*" | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort -V | tail -5

.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release VERSION=v1.3.18"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@if git rev-parse $(VERSION) >/dev/null 2>&1; then \
		echo "Error: Tag $(VERSION) already exists!"; \
		exit 1; \
	fi
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "✅ Release $(VERSION) created and pushed successfully!"

.PHONY: release-patch
release-patch:
	@echo "Creating patch release..."
	@$(MAKE) release VERSION=$$($(MAKE) next-patch)

.PHONY: release-minor
release-minor:
	@echo "Creating minor release..."
	@$(MAKE) release VERSION=$$($(MAKE) next-minor)

.PHONY: release-major
release-major:
	@echo "Creating major release..."
	@$(MAKE) release VERSION=$$($(MAKE) next-major)

.PHONY: next-patch
next-patch:
	@latest=$$(git tag -l "v*.*.*" | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort -V | tail -1); \
	if [ -z "$$latest" ]; then \
		echo "v1.0.0"; \
	else \
		echo $$latest | sed 's/^v//' | awk -F. '{printf "v%d.%d.%d", $$1, $$2, $$3+1}'; \
	fi

.PHONY: next-minor
next-minor:
	@latest=$$(git tag -l "v*.*.*" | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort -V | tail -1); \
	if [ -z "$$latest" ]; then \
		echo "v1.0.0"; \
	else \
		echo $$latest | sed 's/^v//' | awk -F. '{printf "v%d.%d.0", $$1, $$2+1}'; \
	fi

.PHONY: next-major
next-major:
	@latest=$$(git tag -l "v*.*.*" | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$$' | sort -V | tail -1); \
	if [ -z "$$latest" ]; then \
		echo "v1.0.0"; \
	else \
		echo $$latest | sed 's/^v//' | awk -F. '{printf "v%d.0.0", $$1+1}'; \
	fi

.PHONY: delete-tag
delete-tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make delete-tag VERSION=v1.3.18"; \
		exit 1; \
	fi
	@echo "Deleting tag $(VERSION)..."
	@git tag -d $(VERSION)
	@git push origin :refs/tags/$(VERSION)
	@echo "✅ Tag $(VERSION) deleted locally and remotely"

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build:"
	@echo "  build          - Build gocron and gocron-node for current platform"
	@echo "  run            - Build and run in development mode"
	@echo "  test           - Run tests"
	@echo "  package-all    - Build packages for all platforms"
	@echo ""
	@echo "Code Quality:"
	@echo "  check          - Run fmt + vet + test"
	@echo "  pre-release    - Run all checks before release"
	@echo "  fmt            - Format code"
	@echo "  fmt-check      - Check code formatting"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run linter (golangci-lint)"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  security       - Run security checks (gosec)"
	@echo ""
	@echo "Version Management:"
	@echo "  version        - Show current version"
	@echo "  release        - Create release tag (VERSION=v1.3.18)"
	@echo "  release-patch  - Auto increment patch version"
	@echo ""
	@echo "Development:"
	@echo "  dev-deps       - Install development dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"
