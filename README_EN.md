# gocron - Distributed Scheduled Task Management System

[![Release](https://img.shields.io/github/release/gocronx-team/gocron.svg?label=Release)](https://github.com/gocronx-team/gocron/releases) [![Downloads](https://img.shields.io/github/downloads/gocronx-team/gocron/total.svg)](https://github.com/gocronx-team/gocron/releases) [![License](https://img.shields.io/github/license/gocronx-team/gocron.svg)](https://github.com/gocronx-team/gocron/blob/master/LICENSE)

English | [简体中文](README.md)

A lightweight distributed scheduled task management system developed in Go, designed to replace Linux-crontab.

## 📖 Documentation

Full documentation is available at: **[document](https://gocron-docs.pages.dev/en/)**

- 🚀 [Quick Start](https://gocron-docs.pages.dev/en/guide/quick-start) - Installation and deployment guide
- 🤖 [Agent Auto-Registration](https://gocron-docs.pages.dev/en/guide/agent-registration) - One-click task node deployment
- ⚙️ [Configuration](https://gocron-docs.pages.dev/en/guide/configuration) - Detailed configuration guide
- 🔌 [API Documentation](https://gocron-docs.pages.dev/en/guide/api) - API reference

## ✨ Features

* **Web Interface**: Intuitive task management interface
* **Second-level Precision**: Supports Crontab expressions with second precision
* **Distributed Architecture**: Master-Worker architecture, high availability
* **Task Retry**: Configurable retry policies for failed tasks
* **Task Dependency**: Supports task dependency configuration
* **Access Control**: Comprehensive user and permission management
* **2FA Security**: Two-Factor Authentication support
* **Agent Auto-Registration**: One-click installation for Linux/macOS/Windows
* **Multi-Database**: MySQL / PostgreSQL / SQLite support
* **Log Management**: Complete execution logs with auto-cleanup
* **Notifications**: Email, Slack, Webhook support

## 🚀 Quick Start (Docker)

The easiest way to deploy is using Docker Compose:

```bash
# 1. Clone the project
git clone https://github.com/gocronx-team/gocron.git
cd gocron

# 2. Start services
docker-compose up -d

# 3. Access Web Interface
# http://localhost:5920
# Default credentials: admin / admin123
```

For more deployment methods (Binary, Development), please refer to the [Installation Guide](https://gocron-docs.pages.dev/en/guide/quick-start).

## 📸 Screenshots

![Scheduled Tasks](assets/screenshot/scheduler_en.png)

![Agent Auto-Registration](assets/screenshot/agent_en.png)

![Agent自动注册](assets/screenshot/task_en.png)

![Agent自动注册](assets/screenshot/notification_en.png)

## 🤝 Contributing

We warmly welcome community contributions!

- 🐛 **Report Bugs**: Please submit via GitHub Issues
- 💡 **Submit Code**: Please follow the [Contributing Guide](https://gocron-docs.pages.dev/en/guide/contributing) to submit PRs

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=gocronx-team/gocron&type=Date)](https://www.star-history.com/#gocronx-team/gocron&Date)