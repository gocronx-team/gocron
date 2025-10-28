# gocron - 定时任务管理系统

使用Go语言开发的轻量级定时任务集中调度和管理系统，用于替代Linux-crontab。

## 功能特性

* Web界面管理定时任务
* crontab时间表达式，精确到秒
* 任务执行失败可重试
* 任务执行超时，强制结束
* 任务依赖配置
* 多用户与权限控制
* 双因素认证(2FA)
* 任务类型
    * Shell任务 - 在任务节点上执行shell命令
    * HTTP任务 - 访问指定的URL地址
* 任务执行日志查看
* 任务执行结果通知（邮件、Slack、Webhook）

## 环境要求

* Go 1.23+
* MySQL 或 PostgreSQL
* Node.js 18+ (前端开发)

## 快速开始

### 开发环境

```bash
# 1. 克隆项目
git clone https://github.com/gocronx-team/gocron.git
cd gocron

# 2. 安装依赖
go mod download

# 3. 配置数据库
# 编辑 ~/.gocron/conf/app.ini

# 4. 启动后端（热更新）
air

# 5. 启动前端（另一个终端）
cd web/vue
npm install
npm run dev
```

访问 http://localhost:8080

### 生产部署

```bash
# 1. 编译
make package

# 2. 启动服务
./gocron web

# 3. 启动任务节点
./gocron-node
```

访问 http://localhost:5920

## 命令说明

### gocron

```bash
gocron web              # 启动Web服务
gocron web -p 8080      # 指定端口
gocron web -e dev       # 开发模式
gocron -v               # 查看版本
```

### gocron-node

```bash
gocron-node             # 启动任务节点
gocron-node -s :5921    # 指定监听地址
gocron-node -enable-tls # 启用TLS
```

## 技术栈

* 后端：Gin + GORM + gRPC
* 前端：Vue 3 + Element Plus
* 定时任务：Cron
* 数据库：MySQL / PostgreSQL

## 开发工具

* `make` - 编译项目
* `make run` - 编译并运行
* `air` - 后端热更新
* `npm run dev` - 前端热更新

## 文档

* [2FA使用说明](2FA_USAGE.md)
* [升级指南](UPGRADE_GUIDE.md)

## License

MIT
