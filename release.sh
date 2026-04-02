#!/bin/bash

# 本地构建并发布到 GitHub Release

set -e

VERSION=""
PRERELEASE=false
SKIP_CHECKS=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        --prerelease)
            PRERELEASE=true
            shift
            ;;
        --skip-checks)
            SKIP_CHECKS=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 -v <version> [--prerelease] [--skip-checks]"
            echo "Example: $0 -v v1.3.21"
            exit 1
            ;;
    esac
done

if [ -z "$VERSION" ]; then
    echo "Error: Version is required"
    echo "Usage: $0 -v <version> [--prerelease] [--skip-checks]"
    exit 1
fi

echo "=========================================="
echo "Local Build and Release to GitHub"
echo "=========================================="
echo "Version: $VERSION"
echo "Prerelease: $PRERELEASE"
echo "Skip Checks: $SKIP_CHECKS"
echo ""

# 0. 代码质量检查
if [ "$SKIP_CHECKS" = false ]; then
    echo "0. Running code quality checks..."
    echo ""
    
    # 格式检查
    echo "  → Checking code formatting..."
    if ! make fmt-check 2>/dev/null; then
        echo "❌ Code formatting check failed!"
        echo "   Run 'make fmt' to fix formatting issues"
        exit 1
    fi
    
    # go vet 检查
    echo "  → Running go vet..."
    if ! make vet 2>/dev/null; then
        echo "❌ go vet check failed!"
        exit 1
    fi
    
    # 运行测试
    echo "  → Running tests..."
    if ! make test 2>/dev/null; then
        echo "❌ Tests failed!"
        exit 1
    fi
    
    # 可选：linter 检查
    echo "  → Running linter (optional)..."
    make lint 2>/dev/null || echo "⚠️  Linter check skipped"
    
    echo ""
    echo "✅ All code quality checks passed!"
    echo ""
else
    echo "⚠️  Skipping code quality checks (--skip-checks flag)"
    echo ""
fi

# 1. 检查是否需要清理
echo "1. Checking existing builds..."
if [ -d "gocron-package" ] && [ -n "$(ls -A gocron-package 2>/dev/null)" ]; then
    echo "Found existing packages. Clean and rebuild? (y/N): "
    read -r CLEAN_RESPONSE
    if [[ $CLEAN_RESPONSE =~ ^[Yy]$ ]]; then
        rm -rf gocron-package gocron-node-package gocron-build gocron-node-build
        echo "✓ Cleaned"
    else
        echo "✓ Keeping existing packages"
    fi
else
    echo "✓ No existing packages"
fi
echo ""

# 2. 构建前端
echo "2. Building frontend..."
cd web/vue

# 检测使用哪个包管理器
if [ -f "pnpm-lock.yaml" ]; then
    echo "Using pnpm..."
    pnpm install --frozen-lockfile
    pnpm run build
elif [ -f "yarn.lock" ]; then
    echo "Using yarn..."
    yarn install --frozen-lockfile
    yarn run build
else
    echo "Using npm..."
    npm ci
    npm run build
fi

cd ../..
echo "✓ Frontend built (output: web/vue/dist/)"
echo ""



# 3. 构建所有平台的包
echo "3. Building packages for all platforms..."
MISSING_PACKAGES=false

# 检查 Linux/macOS gocron 包
for os in linux darwin; do
    for arch in amd64 arm64; do
        if [ ! -f "gocron-package/gocron-${VERSION}-${os}-${arch}.tar.gz" ] || \
           [ ! -f "gocron-node-package/gocron-node-${os}-${arch}.tar.gz" ]; then
            MISSING_PACKAGES=true
            break 2
        fi
    done
done

if [ "$MISSING_PACKAGES" = true ]; then
    echo "Building Linux and macOS packages..."
    ./package.sh -p "linux,darwin" -a "amd64,arm64" -v "$VERSION"
else
    echo "Linux/macOS packages already exist, skipping..."
fi

# 检查 Windows 包
if [ ! -f "gocron-package/gocron-${VERSION}-windows-amd64.zip" ] || \
   [ ! -f "gocron-node-package/gocron-node-windows-amd64.zip" ]; then
    echo "Building Windows packages..."
    ./package.sh -p "windows" -a "amd64" -v "$VERSION"
else
    echo "Windows packages already exist, skipping..."
fi
echo "✓ All packages built"
echo ""

# 4. 显示构建结果
echo "4. Build summary:"
echo ""
echo "gocron packages:"
ls -lh gocron-package/
echo ""
echo "gocron-node packages:"
ls -lh gocron-node-package/
echo ""

# 5. 验证包内容
echo "5. Verifying package contents..."
SAMPLE_PACKAGE=$(ls gocron-package/*.tar.gz 2>/dev/null | head -1)
if [ -n "$SAMPLE_PACKAGE" ]; then
    echo "Checking: $SAMPLE_PACKAGE"
    tar tzf "$SAMPLE_PACKAGE" | head -5
    echo "✓ Package verified"
else
    SAMPLE_PACKAGE=$(ls gocron-package/*.zip 2>/dev/null | head -1)
    if [ -n "$SAMPLE_PACKAGE" ]; then
        echo "Checking: $SAMPLE_PACKAGE"
        unzip -l "$SAMPLE_PACKAGE" | head -5
        echo "✓ Package verified"
    fi
fi
echo ""

# 6. 创建 Git tag
echo "6. Creating Git tag..."
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "Tag $VERSION already exists"
    read -p "Delete and recreate? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "$VERSION"
        git push origin ":refs/tags/$VERSION" 2>/dev/null || true
    else
        echo "Skipping tag creation"
    fi
fi

if ! git rev-parse "$VERSION" >/dev/null 2>&1; then
    git tag -a "$VERSION" -m "Release $VERSION"
    git push origin "$VERSION"
    echo "✓ Tag created and pushed"
else
    echo "✓ Using existing tag"
fi
echo ""

# 7. 创建 GitHub Release
echo "7. Creating GitHub Release..."
echo ""

PRERELEASE_FLAG=""
if [ "$PRERELEASE" = true ]; then
    PRERELEASE_FLAG="--prerelease"
fi

# 生成 release notes
cat > /tmp/release_notes.md <<EOF

feat: multi-tag support, comma-separated storage, el-select multi-picker, GET /api/task/tags
feat: per-task log cleanup, POST /api/task/log/clear/:id, batch delete, admin-only
feat: task-level log retention, log_retention_days field, excludes custom tasks from global cleanup
fix: go run AppDir detection for dev environment
fix: el-alert \\n literal rendering, use v-html with <br>
chore: bump to v1.5.8 with migration 158

EOF

# 检查 gh CLI 是否安装
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed"
    echo "Install it from: https://cli.github.com/"
    echo ""
    echo "Packages are ready in:"
    echo "  - gocron-package/"
    echo "  - gocron-node-package/"
    echo ""
    echo "You can manually create a release on GitHub and upload these files."
    exit 1
fi

# 创建 release
gh release create "$VERSION" \
    --title "Release $VERSION" \
    --notes-file /tmp/release_notes.md \
    $PRERELEASE_FLAG \
    gocron-package/*.tar.gz \
    gocron-package/*.zip \
    gocron-node-package/*.tar.gz \
    gocron-node-package/*.zip

echo ""
echo "=========================================="
echo "✅ Release $VERSION created successfully!"
echo "=========================================="
echo ""
echo "View release: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/releases/tag/$VERSION"
