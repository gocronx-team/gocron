# Quick Start

This guide will help you quickly deploy and run gocron.

## Requirements

- **Go**: 1.23+
- **Database**: MySQL / PostgreSQL / SQLite (see notes below)
- **Node.js**: 20+ (only for frontend development)

## Database Support

| Deployment Method       | MySQL        | PostgreSQL   | SQLite           |
| ----------------------- | ------------ | ------------ | ---------------- |
| Docker Deployment       | ✅ Supported | ✅ Supported | ✅ Supported     |
| Kubernetes Deployment   | ✅ Supported | ✅ Supported | ❌ Not supported |
| Binary Deployment       | ✅ Supported | ✅ Supported | ✅ Supported     |
| Development Environment | ✅ Supported | ✅ Supported | ✅ Supported     |

::: tip Note

- The managed Kubernetes Chart requires MySQL or PostgreSQL so every Pod can remain stateless and horizontally scalable. Other deployment methods retain pure-Go SQLite support.
- **Production Recommendation**: Use MySQL or PostgreSQL for better performance and distributed deployment support
  :::

## Choosing a Deployment Method

| Scenario                          | Recommended            |
| --------------------------------- | ---------------------- |
| Production, single / few machines | **Binary** (preferred) |
| Production, Kubernetes cluster    | Helm Chart             |
| Local evaluation, testing         | Docker Compose         |

gocron compiles into a single, dependency-free static binary (pure Go SQLite, no CGO), so binary deployment is the lightest and best-fitting option for single-machine production.

## Binary Deployment (Production Recommended)

Suitable for production environments, supports all databases (including SQLite).

### Download Package

Visit [GitHub Releases](https://github.com/gocronx-team/gocron/releases) to download the latest version.

Choose the package for your platform:

- Linux: `gocron-linux-amd64.tar.gz` or `gocron-linux-arm64.tar.gz`
- macOS: `gocron-darwin-amd64.tar.gz` or `gocron-darwin-arm64.tar.gz`
- Windows: `gocron-windows-amd64.zip` or `gocron-windows-arm64.zip`

### Quick Start

```bash
# 1. Extract the package
tar -xzf gocron-linux-amd64.tar.gz
cd gocron-linux-amd64

# 2. Start service (SQLite by default; config and data dirs are created on first run)
./gocron web

# 3. Open the web interface and set the admin account in the install wizard
# http://localhost:5920
```

::: tip Data location
gocron is rooted at the **binary's own directory**: config at `<binary-dir>/.gocron/conf/app.ini`,
logs at `<binary-dir>/.gocron/log/`, and the SQLite database at `<binary-dir>/data/gocron.db` by default.
To upgrade, just replace the binary — the data directory stays untouched.
:::

### Keeping the process running

For production, run the `gocron web` process under whatever process manager you prefer (systemd, supervisor, pm2, etc.) for auto-start on boot and automatic restart on crash — configure it per that tool's documentation. To upgrade, replace the binary and restart the process; the data directory stays untouched.

### Database Configuration

gocron supports three databases, choose according to your needs:

#### SQLite (Default)

No configuration needed, works out of the box. Suitable for small deployments and testing.

#### MySQL

Edit `.gocron/conf/app.ini`:

```ini
[db]
engine = mysql
host = 127.0.0.1
port = 3306
user = root
password = your_password
database = gocron
charset = utf8mb4
```

#### PostgreSQL

Edit `.gocron/conf/app.ini`:

```ini
[db]
engine = postgres
host = 127.0.0.1
port = 5432
user = postgres
password = your_password
database = gocron
```

## Docker Compose Deployment (Evaluation / Testing)

Suitable for local evaluation and testing.

::: warning Note
The repository's `docker-compose.yml` uses `build:` to **build the image from source on the spot**,
so `docker compose up -d` requires cloning the full repo and compiling the frontend and backend locally (slow).
For production, prefer the "Binary + systemd" method above.
:::

### Steps

```bash
# 1. Clone the project
git clone https://github.com/gocronx-team/gocron.git
cd gocron

# 2. Start services (automatically builds image)
docker compose up -d

# 3. Access web interface
# http://localhost:5920
```

### Default Credentials

- Username: `admin`
- Password: `admin123`

::: tip Tip

- Docker Compose only deploys the gocron management server
- Task nodes (gocron-node) need to be installed separately
- See [Agent Auto-Registration](./agent-registration) for installing task nodes
  :::

## Kubernetes Deployment (Helm)

Deploy to Kubernetes clusters with a single command using Helm Chart.

::: tip Container image
The Helm Chart defaults to the public image `ghcr.io/gocronx-team/gocron`, with
the tag defaulting to the Chart's appVersion. Release tags publish matching
multi-architecture images for Linux AMD64 and ARM64.
:::

### Add Helm Repository

```bash
helm repo add gocron https://gocronx-team.github.io/gocron
helm repo update
```

### Deploy

```bash
# MySQL (the database must already exist)
helm upgrade gocron gocron/gocron --reuse-values \
  --set db.engine=mysql \
  --set db.host=mysql.default \
  --set db.port=3306 \
  --set db.user=gocron \
  --set db.password=your_password \
  --set db.database=gocron \
  --set managed.authSecret=replace-with-at-least-32-random-characters \
  --set managed.encryptionKey=replace-with-another-32-random-characters \
  --set managed.admin.password=replace-with-a-strong-admin-password

# PostgreSQL (the database must already exist)
helm install gocron gocron/gocron \
  --set db.engine=postgres \
  --set db.host=pg.default \
  --set db.port=5432 \
  --set db.user=gocron \
  --set db.password=your_password \
  --set db.database=gocron \
  --set managed.authSecret=replace-with-at-least-32-random-characters \
  --set managed.encryptionKey=replace-with-another-32-random-characters \
  --set managed.admin.password=replace-with-a-strong-admin-password
```

### Configure Ingress

```bash
helm upgrade gocron gocron/gocron --reuse-values \
  --set ingress.enabled=true \
  --set 'ingress.hosts[0].host=gocron.example.com' \
  --set 'ingress.hosts[0].paths[0].path=/' \
  --set 'ingress.hosts[0].paths[0].pathType=Prefix'
```

::: tip Tip
For full Helm configuration options, see [Kubernetes Deployment](./kubernetes).
:::

## Development Environment

Suitable for development and debugging.

### Prerequisites

- Go 1.23+
- Node.js 20+
- Yarn

### Steps

```bash
# 1. Clone the project
git clone https://github.com/gocronx-team/gocron.git
cd gocron

# 2. Install Go dependencies
go mod download

# 3. Configure database
# Edit .gocron/conf/app.ini
# Or copy app.ini.sqlite.example to use SQLite

# 4. Install development tools
go install github.com/air-verse/air@latest

# 5. Start backend (with hot reload)
air

# 6. Start frontend (in another terminal)
cd web/vue
yarn install
yarn run dev
```

Visit http://localhost:8080

## Install Task Nodes

gocron uses a distributed architecture where tasks execute on independent nodes.

### Method 1: Auto-Registration (Recommended)

Use the web interface to generate one-click installation commands. See [Agent Auto-Registration](./agent-registration) for details.

### Method 2: Manual Installation

```bash
# 1. Download gocron-node package
# Visit GitHub Releases to download the package for your platform

# 2. Extract
tar -xzf gocron-node-linux-amd64.tar.gz
cd gocron-node-linux-amd64

# 3. Start node
./gocron-node

# 4. Add node in web interface
# Go to "Task Nodes" page and click "Add Node"
```

## Verify Installation

1. Access web interface: http://localhost:5920
2. Login with default credentials (admin / admin123)
3. Go to "Task Nodes" page and confirm nodes are connected
4. Create a test task and verify execution

## Next Steps

- [Configuration](./configuration) - Learn about detailed configuration options
- [Kubernetes Deployment](./kubernetes) - Full Helm Chart configuration
- [Scheduled Tasks](./scheduled-tasks) - Learn how to create and manage tasks
- [Agent Auto-Registration](./agent-registration) - Quickly deploy task nodes
- [API Documentation](./api) - Manage tasks using API

## FAQ

### Can I use SQLite with Docker deployment?

Yes. Since gocron v1.5.9+, a pure Go SQLite driver is used and Docker images fully support SQLite.

### How to change the default port?

```bash
# Method 1: Command line argument
./gocron web -p 8080

# Method 2: Configuration file
# Edit .gocron/conf/app.ini
[server]
port = 8080
```

### How to enable HTTPS?

See [Security - TLS Mutual Authentication](./security-tls) chapter.

### Task node cannot connect?

1. Check if the node is running properly
2. Check firewall settings
3. Confirm node address is configured correctly
4. Check node logs: `log/cron.log`
