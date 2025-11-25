#!/usr/bin/env bash
 
# 生成压缩包 xx.tar.gz或xx.zip
# 使用 ./package.sh -a amd664 -p linux -v v2.0.0
 
# 任何命令返回非0值退出
set -o errexit
# 使用未定义的变量退出
set -o nounset
# 管道中任一命令执行失败退出
set -o pipefail

# 获取 Go 环境变量
GOHOSTOS=$(go env GOHOSTOS)
GOHOSTARCH=$(go env GOHOSTARCH)

# 二进制文件名
BINARY_NAME=''
# main函数所在文件
MAIN_FILE=""
 
# 提取git最新tag作为应用版本
VERSION=''
# 最新git commit id
GIT_COMMIT_ID=''
 
# 外部输入的系统
INPUT_OS=()
# 外部输入的架构
INPUT_ARCH=()
# 未指定OS，默认值
DEFAULT_OS=${GOHOSTOS}
# 未指定ARCH,默认值
DEFAULT_ARCH=${GOHOSTARCH}
# 支持的系统
SUPPORT_OS=(linux darwin windows)
# 支持的架构
SUPPORT_ARCH=(386 amd64 arm64)
 
# 编译参数
LDFLAGS=''
# 需要打包的文件
INCLUDE_FILE=()
# 打包文件生成目录
PACKAGE_DIR=''
# 编译文件生成目录
BUILD_DIR=''
 
# 获取git 最新tag name
git_latest_tag() {
    local COMMIT_ID=""
    local TAG_NAME=""
    COMMIT_ID=`git rev-list --tags --max-count=1`
    TAG_NAME=`git describe --tags "${COMMIT_ID}"`
 
    echo ${TAG_NAME}
}
 
# 获取git 最新commit id
git_latest_commit() {
    echo "$(git rev-parse --short HEAD)"
}
 
# 打印信息
print_message() {
    echo "$1"
}
 
# 打印信息后推出
print_message_and_exit() {
    if [[ -n $1 ]]; then
        print_message "$1"
    fi
    exit 1
}
 
# 设置系统、CPU架构
set_os_arch() {
    if [[ ${#INPUT_OS[@]} = 0 ]];then
        INPUT_OS=("${DEFAULT_OS}")
    fi
 
    if [[ ${#INPUT_ARCH[@]} = 0 ]];then
        INPUT_ARCH=("${DEFAULT_ARCH}")
    fi
 
    for OS in "${INPUT_OS[@]}"; do
        if [[  ! "${SUPPORT_OS[*]}" =~ ${OS} ]]; then
            print_message_and_exit "不支持的系统${OS}"
        fi
    done
 
    for ARCH in "${INPUT_ARCH[@]}";do
        if [[ ! "${SUPPORT_ARCH[*]}" =~ ${ARCH} ]]; then
            print_message_and_exit "不支持的CPU架构${ARCH}"
        fi
    done
}
 
# 初始化
init() {
    set_os_arch
 
    if [[ -z "${VERSION}" ]];then
        VERSION=`git_latest_tag`
    fi
    GIT_COMMIT_ID=`git_latest_commit`
    LDFLAGS="-w -X 'main.AppVersion=${VERSION}' -X 'main.BuildDate=`date '+%Y-%m-%d %H:%M:%S'`' -X 'main.GitCommit=${GIT_COMMIT_ID}'"
 
    PACKAGE_DIR=${BINARY_NAME}-package
    BUILD_DIR=${BINARY_NAME}-build
 
    # 只清理 BUILD_DIR，保留 PACKAGE_DIR 以支持增量构建
    if [[ -d ${BUILD_DIR} ]];then
        rm -rf ${BUILD_DIR}
    fi
 
    mkdir -p ${BUILD_DIR}
    mkdir -p ${PACKAGE_DIR}
}
 
# 编译
build() {
    local FILENAME=''
    for OS in "${INPUT_OS[@]}";do
        for ARCH in "${INPUT_ARCH[@]}";do
            # gocron-node 不需要数据库，强制禁用 CGO
            # gocron 需要 SQLite，根据平台决定是否启用 CGO
            local CGO_ENABLED_VALUE='1'
            local CC_COMPILER=''
            
            if [[ "${BINARY_NAME}" = "gocron-node" ]]; then
                CGO_ENABLED_VALUE='0'
                print_message "编译 gocron-node ${OS}-${ARCH} 版本（纯静态编译）"
            elif [[ "${OS}" != "${GOHOSTOS}" ]] || [[ "${ARCH}" != "${GOHOSTARCH}" ]]; then
                # 检查是否安装了交叉编译工具链
                if [[ "${OS}" = "windows" ]] && [[ "${ARCH}" = "amd64" ]] && command -v x86_64-w64-mingw32-gcc &> /dev/null; then
                    # macOS/Linux 交叉编译 Windows amd64，使用 MinGW
                    CC_COMPILER='x86_64-w64-mingw32-gcc'
                    print_message "使用 MinGW 交叉编译 Windows amd64 版本（支持 SQLite）"
                elif [[ "${OS}" = "linux" ]] && [[ "${ARCH}" = "amd64" ]] && command -v x86_64-linux-musl-gcc &> /dev/null; then
                    # macOS 交叉编译 Linux amd64，使用 musl-cross，完全静态链接
                    CC_COMPILER='x86_64-linux-musl-gcc'
                    print_message "使用 musl-cross 交叉编译 Linux amd64 版本（支持 SQLite，完全静态）"
                elif [[ "${OS}" = "linux" ]] && [[ "${ARCH}" = "arm64" ]] && command -v aarch64-linux-musl-gcc &> /dev/null; then
                    # macOS 交叉编译 Linux arm64，使用 musl-cross，完全静态链接
                    CC_COMPILER='aarch64-linux-musl-gcc'
                    print_message "使用 musl-cross 交叉编译 Linux arm64 版本（支持 SQLite，完全静态）"
                elif [[ "${OS}" = "darwin" ]]; then
                    # macOS 同平台不同架构编译，保持 CGO 启用
                    print_message "macOS 交叉架构编译 ${OS}-${ARCH} 版本（支持 SQLite）"
                else
                    # 没有交叉编译工具链或不支持的架构，禁用 CGO
                    CGO_ENABLED_VALUE='0'
                    print_message "警告: 跨平台编译 ${OS}-${ARCH} 未找到交叉编译工具链，禁用 CGO（不支持 SQLite）"
                fi
            fi
            
            if [[ "${OS}" = "windows"  ]];then
                FILENAME=${BINARY_NAME}.exe
            else
                FILENAME=${BINARY_NAME}
            fi
            
            if [[ -n "${CC_COMPILER}" ]]; then
                if [[ "${OS}" = "linux" ]]; then
                    # Linux 使用静态链接
                    env CGO_ENABLED=${CGO_ENABLED_VALUE} CC=${CC_COMPILER} GOOS=${OS} GOARCH=${ARCH} go build -ldflags "${LDFLAGS} -extldflags '-static'" -o ${BUILD_DIR}/${BINARY_NAME}-${OS}-${ARCH}/${FILENAME} ${MAIN_FILE}
                else
                    env CGO_ENABLED=${CGO_ENABLED_VALUE} CC=${CC_COMPILER} GOOS=${OS} GOARCH=${ARCH} go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-${OS}-${ARCH}/${FILENAME} ${MAIN_FILE}
                fi
            else
                env CGO_ENABLED=${CGO_ENABLED_VALUE} GOOS=${OS} GOARCH=${ARCH} go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-${OS}-${ARCH}/${FILENAME} ${MAIN_FILE}
            fi
        done
    done
}
 
# 打包
package_binary() {
    cd ${BUILD_DIR}
 
    for OS in "${INPUT_OS[@]}";do
        for ARCH in "${INPUT_ARCH[@]}";do
        package_file ${BINARY_NAME}-${OS}-${ARCH}
        
        # 如果 VERSION 为空，不带版本号
        if [[ -z "${VERSION}" ]]; then
            if [[ "${OS}" = "windows" ]];then
                zip -rq ../${PACKAGE_DIR}/${BINARY_NAME}-${OS}-${ARCH}.zip ${BINARY_NAME}-${OS}-${ARCH}
            else
                tar czf ../${PACKAGE_DIR}/${BINARY_NAME}-${OS}-${ARCH}.tar.gz ${BINARY_NAME}-${OS}-${ARCH}
            fi
        else
            if [[ "${OS}" = "windows" ]];then
                zip -rq ../${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.zip ${BINARY_NAME}-${OS}-${ARCH}
            else
                tar czf ../${PACKAGE_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.tar.gz ${BINARY_NAME}-${OS}-${ARCH}
            fi
        fi
        done
    done
 
    cd ${OLDPWD}
}
 
# 打包文件
package_file() {
    if [[ "${#INCLUDE_FILE[@]}" = "0" ]];then
        return
    fi
    for item in "${INCLUDE_FILE[@]}"; do
            cp -r ../${item} $1
    done
}
 
# 清理
clean() {
    if [[ -d ${BUILD_DIR} ]];then
        rm -rf ${BUILD_DIR}
    fi
}
 
# 运行
run() {
    init
    build
    package_binary
    clean
}

package_gocron() {
    BINARY_NAME='gocron'
    MAIN_FILE="./cmd/gocron/gocron.go"
    # 如果 gocron-node-package 目录存在，将其包含到 gocron 包中
    if [[ -d "gocron-node-package" ]] && [[ -n "$(ls -A gocron-node-package)" ]]; then
        INCLUDE_FILE=("gocron-node-package")
    else
        INCLUDE_FILE=()
    fi

    run
}

package_gocron_node() {
    BINARY_NAME='gocron-node'
    MAIN_FILE="./cmd/node/node.go"
    INCLUDE_FILE=()

    run
}
 
# p 平台 linux darwin windows
# a 架构 386 amd64 arm64
# v 版本号  默认取git最新tag
# t 类型 all(默认), gocron, node
BUILD_TYPE="all"
while getopts "p:a:v:t:" OPT;
do
    case ${OPT} in
    p) IFS=',' read -r -a INPUT_OS <<< "${OPTARG}"
    ;;
    a) IFS=',' read -r -a INPUT_ARCH <<< "${OPTARG}"
    ;;
    v) VERSION=$OPTARG
    ;;
    t) BUILD_TYPE=$OPTARG
    ;;
    *)
    ;;
    esac
done
 
# 默认构建所有
if [[ -z "${BUILD_TYPE}" ]]; then
    BUILD_TYPE="all"
fi

if [[ "${BUILD_TYPE}" = "all" ]] || [[ "${BUILD_TYPE}" = "gocron" ]]; then
    package_gocron
fi

if [[ "${BUILD_TYPE}" = "all" ]] || [[ "${BUILD_TYPE}" = "node" ]]; then
    package_gocron_node
fi

