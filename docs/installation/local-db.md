# 本地数据库环境运行指南

本文档说明如何在本地数据库环境下运行 New API，适用于以下两种场景：

1. 使用 Docker 运行程序，同时连接宿主机上的 MySQL 或 PostgreSQL
2. 直接在本机运行源码，连接本地 SQLite、MySQL 或 PostgreSQL

如果你的目标只是先把项目跑起来，优先使用“方案一”。如果你需要本地调试 Go 或前端代码，使用“方案二”。

***

## 方案选择

| 场景 | 推荐方案 |
| --- | --- |
| 只想尽快启动服务 | 方案一：Docker + 本地数据库 |
| 需要调试 Go 后端代码 | 方案二：源码运行 |
| 需要调试前端页面 | 方案二：源码运行 + 前端开发服务器 |
| 不想先准备 MySQL/PostgreSQL | 方案二：先用 SQLite |

***

## 方案一：Docker 运行程序，连接本机数据库

这种方式最省事。程序运行在容器里，数据库仍然使用你 macOS 本机上的 MySQL 或 PostgreSQL。

### 前提条件

1. 已安装 Docker Desktop
2. 本机 MySQL 或 PostgreSQL 已启动
3. 数据库中已经创建好目标库，例如 `newapi`

### 关键注意事项

如果数据库运行在宿主机上，容器内不能使用 `localhost` 连接数据库。

应使用：

```text
host.docker.internal
```

因为容器里的 `localhost` 指向的是容器自身，不是你的 macOS 主机。

### 使用本机 MySQL

```bash
cd /Users/wenqidong/cdinfo/new-api

docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -e SESSION_SECRET='replace-with-a-random-string' \
  -e SQL_DSN='root:123456@tcp(host.docker.internal:3306)/newapi?charset=utf8mb4&parseTime=true&loc=Local' \
  -v "$PWD/data:/data" \
  calciumion/new-api:latest
```

### 使用本机 PostgreSQL

```bash
cd /Users/wenqidong/cdinfo/new-api

docker run --name new-api -d --restart always \
  -p 3000:3000 \
  -e TZ=Asia/Shanghai \
  -e SESSION_SECRET='replace-with-a-random-string' \
  -e SQL_DSN='postgresql://postgres:123456@host.docker.internal:5432/newapi?sslmode=disable' \
  -v "$PWD/data:/data" \
  calciumion/new-api:latest
```

### 可选：连接本机 Redis

如果你本机也有 Redis，需要额外传入以下环境变量：

```bash
-e REDIS_CONN_STRING='redis://host.docker.internal:6379/0' \
-e CRYPTO_SECRET='replace-with-another-random-string' \
```

说明：

1. 不使用 Redis 时，可以不传 `REDIS_CONN_STRING`
2. 只有在使用 Redis 时，才需要配置 `CRYPTO_SECRET`

### 启动后验证

启动后访问：

```text
http://localhost:3000
```

首次启动时，如果数据库为空，系统会自动建表，并创建初始管理员账号：

```text
用户名: root
密码: 123456
```

首次登录后应立即修改密码。

***

## 方案二：本机直接运行源码

这种方式适合本地开发和调试。

### 前提条件

1. 已安装 Go，建议使用 `1.25.x`
2. 已安装 Bun
3. 已安装本地数据库，或准备先使用 SQLite

### 重要说明

当前仓库中的前端构建产物目录默认不存在，因此不能直接执行：

```bash
go run main.go
```

原因是后端代码使用了 `go:embed` 嵌入前端静态文件，启动前需要先构建前端。

### 第一步：准备 `.env`

项目启动时会自动读取根目录下的 `.env` 文件。

最少要注意以下几点：

1. `SESSION_SECRET` 不能使用示例值 `random_string`
2. 不配置 `SQL_DSN` 时，默认会回退到 SQLite
3. 不使用 Redis 时，可以不配置 `REDIS_CONN_STRING`

### SQLite 示例

适合先把项目快速跑起来：

```env
SESSION_SECRET=replace-with-a-random-string
SQLITE_PATH=./data/new-api.db?_busy_timeout=30000
PORT=3000
GIN_MODE=debug
```

### MySQL 示例

```env
SESSION_SECRET=replace-with-a-random-string
SQL_DSN=root:123456@tcp(127.0.0.1:3306)/newapi?charset=utf8mb4&parseTime=true&loc=Local
PORT=3000
GIN_MODE=debug
```

### PostgreSQL 示例

```env
SESSION_SECRET=replace-with-a-random-string
SQL_DSN=postgresql://postgres:123456@127.0.0.1:5432/newapi?sslmode=disable
PORT=3000
GIN_MODE=debug
```

### 可选：Redis 示例

如果你本机 Redis 也要一起接入，可以额外加入：

```env
REDIS_CONN_STRING=redis://127.0.0.1:6379/0
CRYPTO_SECRET=replace-with-another-random-string
```

### 第二步：构建前端

在仓库根目录执行：

```bash
cd /Users/wenqidong/cdinfo/new-api
mkdir -p data

cd web/default
bun install
bun run build

cd ../classic
bun install
bun run build
```

### 第三步：启动后端

```bash
cd /Users/wenqidong/cdinfo/new-api
go run main.go
```

启动后访问：

```text
http://localhost:3000
```

### 第四步：前端开发模式（可选）

如果你要调试默认前端，可以额外启动前端开发服务器：

```bash
cd /Users/wenqidong/cdinfo/new-api/web/default
bun install
bun run dev
```

此时访问：

```text
http://localhost:3001
```

前端开发服务器会把 API 请求自动代理到：

```text
http://localhost:3000
```

***

## 推荐的最小启动路径

如果你只是想先验证项目能不能跑起来，建议按下面顺序：

1. 先使用 SQLite
2. 先用源码模式或 Docker 任意一种跑通
3. 登录后台确认服务正常
4. 再切换到本机 MySQL 或 PostgreSQL

这样可以把问题拆开，避免同时排查数据库、Redis、前端和程序启动问题。

***

## 常见问题

### 1. 为什么 `go run main.go` 直接报错？

通常是因为前端 `dist` 目录还没有构建。先执行：

```bash
cd web/default && bun install && bun run build
cd ../classic && bun install && bun run build
```

### 2. 为什么 Docker 连不上本机数据库？

最常见原因是把数据库地址写成了 `localhost`。在 Docker 容器里应改为：

```text
host.docker.internal
```

### 3. 为什么登录后会话异常或启动失败？

检查 `SESSION_SECRET` 是否配置，且不要使用示例值 `random_string`。

### 4. Redis 是必须的吗？

不是。没有 Redis 也能启动。

只有在你要启用 Redis 时，才需要同时配置：

1. `REDIS_CONN_STRING`
2. `CRYPTO_SECRET`

### 5. 首次管理员账号是什么？

当数据库为空时，系统会自动创建初始管理员：

```text
用户名: root
密码: 123456
```

建议首次登录后立即修改。

***

## 小结

最省事的方式：

1. 用 Docker 跑程序
2. 用本机 MySQL 或 PostgreSQL 当数据库
3. 数据库地址使用 `host.docker.internal`

本地开发最常用的方式：

1. 写好根目录 `.env`
2. 先构建 `web/default` 和 `web/classic`
3. 再执行 `go run main.go`

如果只是为了先启动成功，先用 SQLite 是阻力最小的方案。
