# Ubuntu 裸机部署指南

本文档说明两件事：

1. 如何在本地 macOS 开发机上生成可用于 Ubuntu 的部署产物
2. 如何把产物发布到 Ubuntu 裸机或云服务器，并通过 systemd 常驻运行

本文默认目标环境为 `Ubuntu 22.04/24.04 amd64`。如果你的服务器是 ARM64，请把文中的 `amd64` 改成 `arm64`，并手动使用对应的 `GOARCH` 构建。

***

## 适用场景

适合以下部署方式：

1. 本地在 macOS 上构建产物
2. 通过 `scp` 或 CI 上传到 Ubuntu 服务器
3. 在 Ubuntu 上直接运行二进制，不使用 Docker

这种方式的优点是：

1. 服务器上不需要安装 Go 或 Bun
2. 部署目录清晰，适合做 systemd 托管
3. 可以和现有的 Nginx、MySQL、PostgreSQL、Redis 环境直接集成

***

## 部署前提

本地构建机需要：

1. Go，建议与仓库当前版本一致
2. Bun
3. 能正常执行 `bun install`、`bun run build`、`go build`

Ubuntu 服务器需要：

1. 已创建好系统用户，例如 `ubuntu` 或专门的 `newapi`
2. 已准备好数据库，支持 SQLite、MySQL、PostgreSQL
3. 如需 Redis，已准备好 Redis 连接地址
4. 服务器对外开放业务端口或已配置 Nginx 反向代理

***

## 一、本地生成 Ubuntu 部署产物

### 方式一：使用 Makefile 直接打包

在仓库根目录执行：

```bash
cd /Users/wenqidong/cdinfo/new-api
make package-ubuntu-amd64
```

执行完成后，会生成：

```text
release/new-api-ubuntu-amd64-<VERSION>.tar.gz
```

压缩包中包含：

1. `new-api` Linux 可执行文件
2. `.env.example`
3. `new-api.service`
4. `LICENSE`
5. `NOTICE`
6. `THIRD-PARTY-LICENSES.md`

说明：

1. 该目标会先构建 `web/default` 和 `web/classic`，再把前端静态资源嵌入后端二进制
2. 生成的是 `linux/amd64` 产物，适用于大多数 x86_64 Ubuntu 服务器

如果你只想先生成二进制，不打包压缩文件，也可以执行：

```bash
cd /Users/wenqidong/cdinfo/new-api
make build-ubuntu-amd64
```

产物位置：

```text
release/new-api-linux-amd64
```

### 方式二：手动构建

如果你不想依赖 Makefile，也可以手动执行：

```bash
cd /Users/wenqidong/cdinfo/new-api

cd web/default
bun install
DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat ../../VERSION) bun run build

cd ../classic
bun install
VITE_REACT_APP_VERSION=$(cat ../../VERSION) bun run build

cd /Users/wenqidong/cdinfo/new-api
mkdir -p release/manual
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" \
  -o release/manual/new-api main.go
```

如果要打包：

```bash
cd /Users/wenqidong/cdinfo/new-api
mkdir -p release/manual/package
cp release/manual/new-api release/manual/package/
cp .env.example new-api.service LICENSE NOTICE THIRD-PARTY-LICENSES.md release/manual/package/
cd release/manual
tar -czf new-api-ubuntu-amd64-manual.tar.gz package
```

### ARM64 服务器构建方式

如果你的 Ubuntu 服务器是 ARM64，例如部分云主机或树莓派，请改为：

```bash
cd /Users/wenqidong/cdinfo/new-api
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" \
  -o release/new-api-linux-arm64 main.go
```

***

## 二、上传部署产物到 Ubuntu

以下示例假设：

1. 服务器地址是 `203.0.113.10`
2. 登录用户是 `ubuntu`
3. 程序部署目录是 `/opt/new-api`

### 第一步：上传压缩包

在本地执行：

```bash
cd /Users/wenqidong/cdinfo/new-api
scp release/new-api-ubuntu-amd64-$(cat VERSION).tar.gz ubuntu@203.0.113.10:/tmp/
```

### 第二步：在 Ubuntu 上准备目录

登录服务器：

```bash
ssh ubuntu@203.0.113.10
```

创建目录：

```bash
sudo mkdir -p /opt/new-api/releases
sudo mkdir -p /opt/new-api/shared/logs
sudo mkdir -p /opt/new-api/shared/data
sudo chown -R ubuntu:ubuntu /opt/new-api
```

### 第三步：解压并切换到当前版本

```bash
cd /opt/new-api/releases
tar -xzf /tmp/new-api-ubuntu-amd64-*.tar.gz
ln -sfn /opt/new-api/releases/$(ls -dt /opt/new-api/releases/new-api-ubuntu-amd64-* | head -n 1 | xargs basename) /opt/new-api/current
```

如果你不想用自动选最新目录，也可以手动指定：

```bash
ln -sfn /opt/new-api/releases/new-api-ubuntu-amd64-0.7.0 /opt/new-api/current
```

***

## 三、准备运行配置

程序启动时会自动读取当前工作目录下的 `.env`。

因此推荐把 `.env` 放在：

```text
/opt/new-api/current/.env
```

### MySQL 示例

```bash
cat >/opt/new-api/current/.env <<'EOF'
SESSION_SECRET=replace-with-a-random-string
SQL_DSN=newapi_app:12345678@tcp(127.0.0.1:3306)/newapi?charset=utf8mb4&parseTime=true&loc=Local
PORT=3000
GIN_MODE=release
EOF
```

### PostgreSQL 示例

```bash
cat >/opt/new-api/current/.env <<'EOF'
SESSION_SECRET=replace-with-a-random-string
SQL_DSN=postgresql://postgres:123456@127.0.0.1:5432/newapi?sslmode=disable
PORT=3000
GIN_MODE=release
EOF
```

### SQLite 快速示例

```bash
cat >/opt/new-api/current/.env <<'EOF'
SESSION_SECRET=replace-with-a-random-string
SQLITE_PATH=/opt/new-api/shared/data/new-api.db?_busy_timeout=30000
PORT=3000
GIN_MODE=release
EOF
```

### Redis 可选配置

如果要启用 Redis，再额外加入：

```env
REDIS_CONN_STRING=redis://127.0.0.1:6379/0
CRYPTO_SECRET=replace-with-another-random-string
```

关键注意事项：

1. `SESSION_SECRET` 不能使用默认示例值 `random_string`
2. 开启 Redis 时必须同时配置 `CRYPTO_SECRET`
3. 如果使用 MySQL，`SQL_DSN` 的开头不要写成 `local:`，否则当前项目会把它识别成 SQLite
4. `LOG_SQL_DSN` 不配置时，日志表默认仍写入主数据库

***

## 四、配置 systemd 服务

仓库里已经提供了 `new-api.service` 模板，但里面的用户、路径和端口需要改成你的实际值。

推荐在 Ubuntu 上使用下面这个版本：

```bash
sudo tee /etc/systemd/system/new-api.service >/dev/null <<'EOF'
[Unit]
Description=New API Service
After=network.target

[Service]
User=ubuntu
WorkingDirectory=/opt/new-api/current
ExecStart=/opt/new-api/current/new-api --port 3000 --log-dir /opt/new-api/shared/logs
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
```

然后执行：

```bash
sudo systemctl daemon-reload
sudo systemctl enable new-api
sudo systemctl start new-api
sudo systemctl status new-api
```

查看实时日志：

```bash
journalctl -u new-api -f
```

如果程序正常启动，访问：

```text
http://服务器IP:3000
```

首次初始化时，如果数据库为空，系统会自动建表并进入初始化流程。

***

## 五、升级发布流程

后续升级可以直接复用同一套流程：

1. 本地重新执行 `make package-ubuntu-amd64`
2. 上传新的压缩包到服务器
3. 解压到 `/opt/new-api/releases/<新版本目录>`
4. 更新 `/opt/new-api/current` 软链接
5. 重启服务

示例：

```bash
sudo systemctl restart new-api
sudo systemctl status new-api
```

这种目录结构的好处是：

1. 版本目录和共享数据目录分离
2. 回滚时只需要把 `current` 软链接切回旧版本
3. systemd 配置无需每次修改

***

## 六、可选的 Nginx 反向代理

如果你不希望直接暴露 `3000` 端口，可以在 Nginx 中转发到本地服务：

```bash
sudo tee /etc/nginx/sites-available/new-api >/dev/null <<'EOF'
server {
    listen 80;
    server_name your-domain.example.com;

    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF
```

启用配置：

```bash
sudo ln -sfn /etc/nginx/sites-available/new-api /etc/nginx/sites-enabled/new-api
sudo nginx -t
sudo systemctl reload nginx
```

如果需要 HTTPS，可以再接入 Let's Encrypt。

***

## 七、常见问题

### 1. 为什么 Ubuntu 上不需要 Bun？

因为前端静态文件已经在本地构建完成，并通过 `go:embed` 打进了后端二进制。Ubuntu 服务器只负责运行 `new-api`。

### 2. 为什么服务启动后还是读不到 `.env`？

通常是以下原因：

1. `WorkingDirectory` 不是 `/opt/new-api/current`
2. `.env` 不在工作目录下
3. 启动的是旧版本目录中的二进制

### 3. 为什么明明配置了 MySQL，却表现得像 SQLite？

检查 `SQL_DSN` 是否以 `local` 开头。这个项目里，`SQL_DSN` 只要以 `local` 开头，就会走 SQLite 分支。

### 4. 如何确认当前服务是否真的起来了？

优先看下面三个地方：

1. `systemctl status new-api`
2. `journalctl -u new-api -f`
3. `curl -I http://127.0.0.1:3000`

### 5. 回滚怎么做？

把 `/opt/new-api/current` 软链接切回旧版本目录，然后执行：

```bash
sudo systemctl restart new-api
```

***

## 推荐的目录结构

```text
/opt/new-api/
├── current -> /opt/new-api/releases/new-api-ubuntu-amd64-<VERSION>
├── releases/
│   ├── new-api-ubuntu-amd64-<OLD_VERSION>/
│   └── new-api-ubuntu-amd64-<VERSION>/
└── shared/
    ├── data/
    └── logs/
```

这种结构更适合长期维护、升级和回滚。