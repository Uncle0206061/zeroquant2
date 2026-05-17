# ZeroQuant 2.0 — Step 4：三端详细环境配置清单

> 版本：V1.0
> 日期：2026-05-05
> 维护者：零号（PM）

---

## 原则

- 三端各自独立配置，互不干扰
- 所有配置通过 NAS 共享一份 `shared/` 原件，各自改自己那份
- 遇到工具缺失 → 先解决工具，再开工

---

## 一、公共前提

### 1.1 网络与存储

| 项目 | 说明 |
|------|------|
| NAS 地址 | `\\100.65.205.77\homes\zero\ZeroQuant2\` |
| 本地映射盘 | `Z:` |
| 交换目录 | `tasks\{terminal}\inbox\`（通过 NAS 同步） |
| 共享配置 | `Z:\ZeroQuant2\shared\`（配置模板放这里） |

### 1.2 GitHub 仓库（待 TASK-001 完成）

| 角色 | 仓库名 |
|------|--------|
| 零号 | `zero-pm` |
| 妞妞 | `zero-backend` |
| 大聪明 | `zero-frontend` |
| 数据官 | `zero-data` |

---

## 二、妞妞（Go 后端）环境配置

### 2.1 必装工具

| 工具 | 版本要求 | 安装方式 |
|------|----------|----------|
| Go | ≥ 1.21 | go.dev/dl |
| PostgreSQL | ≥ 14 | postgresql.org/download |
| Redis | ≥ 6.2 | redis.io/download |
| Docker Desktop | 最新版 | docker.com |
| VSCode | 最新 | code.visualstudio.com |
| Git | 最新 | git-scm.com |

### 2.2 VSCode 扩展推荐

```
Go (ms-vscode.go)
Thunder Client (rangav.vscode-thunder-client)
PostgreSQL editor (ms-ossdata.vscode-postgresql)
Docker (ms-azuretools.vscode-docker)
GitLens (eamodio.gitlens)
```

### 2.3 项目目录

```
D:\ZeroQuant2\                      ← 本地从 NAS 拉取
  src/
    backend/                        ← 妞妞工作目录
      cmd/server/main.go
      internal/
        config/config.go           ← 配置加载
        handler/                   ← HTTP handlers
        middleware/                ← 中间件
        model/                     ← 数据模型
        repository/               ← 数据访问层
        service/                  ← 业务逻辑
        router/                   ← 路由注册
        websocket/                ← WebSocket
        strategy/                 ← 策略引擎
      pkg/
        response/                 ← 统一响应
        jwt/                     ← JWT 工具
        logger/                  ← 日志
      migrations/                 ← 数据库迁移
      go.mod
      go.sum
      Dockerfile
```

### 2.4 环境变量（.env.local）

```env
# 后端服务
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=ZeroQuant2026
DB_NAME=zeroquant_biz
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=ZeroQuant_JWT_SECRET_CHANGE_ME
JWT_EXPIRE_HOURS=24

# 数据服务（Python）通信
DATA_SERVICE_URL=http://localhost:8081
DATA_SERVICE_WS=ws://localhost:8081/data/v1/ws

# 日志
LOG_LEVEL=debug
LOG_PATH=./logs
```

### 2.5 数据库初始化（PostgreSQL）

```sql
-- 连接 postgres 后执行：
CREATE DATABASE zeroquant_biz;

\c zeroquant_biz

-- 建表由 migrations/ 脚本管理，首次运行：
-- go run ./migrations/main.go up
```

### 2.6 本地启动（开发模式）

```bash
cd D:\ZeroQuant2\src\backend

# 拉取依赖
go mod tidy

# 启动 PostgreSQL（Docker）
docker run -d --name zero-postgres `
  -e POSTGRES_PASSWORD=ZeroQuant2026 `
  -p 5432:5432 `
  postgres:16-alpine

# 启动 Redis（Docker）
docker run -d --name zero-redis `
  -p 6379:6379 `
  redis:7-alpine

# 运行迁移
go run ./migrations/main.go up

# 启动服务
go run ./cmd/server/main.go
```

### 2.7 验证（健康检查）

```bash
curl http://localhost:8080/api/v1/health
# 期望返回：{"code":0,"msg":"success","data":{"status":"ok"}}
```

---

## 三、大聪明（Vue3 前端）环境配置

### 3.1 必装工具

| 工具 | 版本要求 | 安装方式 |
|------|----------|----------|
| Node.js | ≥ 20（LTS） | nodejs.org |
| npm / pnpm | pnpm ≥ 9 | npm i -g pnpm |
| Vue CLI / Vite | 已内置在项目 | — |
| Docker Desktop | 最新 | docker.com |
| VSCode | 最新 | code.visualstudio.com |
| Git | 最新 | git-scm.com |
| Chrome / Edge | 最新 | —（开发调试用） |

### 3.2 VSCode 扩展推荐

```
Vue - Official (Vue.volar)
TypeScript Vue Plugin (Vue.vscode-typescript-vue-plugin)
Tailwind CSS IntelliSense (bradlc.vscode-tailwindcss)
Auto Close Tag (formulahendry.auto-close-tag)
Auto Rename Tag (formulahendry.auto-rename-tag)
Prettier (esbenp.prettier-vscode)
GitLens (eamodio.gitlens)
Docker (ms-azuretools.vscode-docker)
REST Client (humao.rest-client-client)
```

### 3.3 项目目录

```
D:\ZeroQuant2\src\frontend\                ← 大聪明工作目录
  public/
  src/
    api/
      index.ts                            ← Axios 实例 + 统一拦截
      modules/
        auth.ts                           ← 认证相关 API
        quote.ts                          ← 行情 API
        strategy.ts                       ← 策略 API
        trade.ts                          ← 交易 API
    assets/
    components/
      pc/                                 ← PC 端组件
      mobile/                             ← 移动端组件
    composables/
      useAuth.ts
      useWebSocket.ts
      useQuote.ts
    layouts/
      PcLayout.vue
      MobileLayout.vue
    router/
      index.ts
    stores/
      auth.ts
      quote.ts
      strategy.ts
    views/
      pc/
        Dashboard.vue
        Quote.vue
        Strategy.vue
        Backtest.vue
        Trade.vue
      mobile/
        (同上结构)
    utils/
    App.vue
    main.ts
  index.html
  vite.config.ts
  tsconfig.json
  package.json
  Dockerfile
```

### 3.4 环境变量（.env.local）

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_DATA_WS_URL=ws://localhost:8081/data/v1/ws
VITE_APP_TITLE=ZeroQuant 2.0
VITE_APP_VERSION=0.1.0
```

### 3.5 本地启动（开发模式）

```bash
cd D:\ZeroQuant2\src\frontend

# 安装依赖（推荐 pnpm）
npm install -g pnpm
pnpm install

# 启动开发服务器
pnpm dev
# 访问 http://localhost:5173
```

### 3.6 验证

浏览器打开 `http://localhost:5173`，确认页面加载无报错，控制台无跨域错误。

---

## 四、数据官（Python 数据服务）环境配置

### 4.1 必装工具

| 工具 | 版本要求 | 安装方式 |
|------|----------|----------|
| Python | ≥ 3.10 | python.org |
| uv | 最新 | astral.sh/uv |
| Redis | ≥ 6.2 | redis.io/download |
| PostgreSQL | ≥ 14 | postgresql.org/download |
| Docker Desktop | 最新 | docker.com |
| VSCode | 最新 | code.visualstudio.com |
| Git | 最新 | git-scm.com |

### 4.2 VSCode 扩展推荐

```
Python (ms-python.python)
Pylance (ms-python.vscode-pylance)
Jupyter (ms-toolsai.jupyter)
REST Client (humao.rest-client-client)
Docker (ms-azuretools.vscode-docker)
GitLens (eamodio.gitlens)
AutoDocstring (njpwerner.autodocstring)
```

### 4.3 项目目录

```
D:\ZeroQuant2\src\data\                    ← 数据官工作目录
  app/
    api/
      v1/
        quote.py                           ← 行情接口
        kline.py                           ← K 线接口
        orderbook.py                       ← Level 2 盘口
        sector.py                          ← 板块接口
        health.py                          ← 健康检查
    collectors/
      akshare_collector.py
      tushare_collector.py
      level2_collector.py
    models/
      quote.py
      kline.py
    schemas/
      quote.py                             ← Pydantic 模型
    services/
      quote_service.py
      kline_service.py
    schedulers/
      daily_quote_sync.py
    config.py
    main.py
    redis_client.py
  alembic/                                 ← 数据库迁移
  tests/
  requirements.txt
  Dockerfile
```

### 4.4 环境变量（.env.local）

```env
# 数据服务
DATA_HOST=0.0.0.0
DATA_PORT=8081

# 数据库
DATA_DB_HOST=localhost
DATA_DB_PORT=5432
DATA_DB_USER=postgres
DATA_DB_PASSWORD=ZeroQuant2026
DATA_DB_NAME=zeroquant_data
DATA_DB_SSLMODE=disable

# Redis
DATA_REDIS_HOST=localhost
DATA_REDIS_PORT=6379
DATA_REDIS_PASSWORD=
DATA_REDIS_DB=1

# 数据源
AKSHARE_FREE_MODE=true
TUSHARE_TOKEN=                    # 有账号则填写
LEVEL2_PROVIDER=none              # none | futu | tgb（待定）

# 后端服务通信
BACKEND_API_URL=http://localhost:8080/api/v1

# 日志
LOG_LEVEL=debug
```

### 4.5 数据库初始化（PostgreSQL）

```sql
-- 连接 postgres 后执行：
CREATE DATABASE zeroquant_data;

-- 迁移脚本由 alembic 管理：
-- cd D:\ZeroQuant2\src\data
-- alembic upgrade head
```

### 4.6 本地启动（开发模式）

```bash
cd D:\ZeroQuant2\src\data

# 使用 uv 管理环境（推荐）
pip install uv
uv sync

# 或传统方式
pip install -r requirements.txt

# 启动 Redis（Docker）
docker run -d --name zero-redis-data `
  -p 6380:6379 `
  redis:7-alpine

# 启动数据服务
uv run uvicorn app.main:app --reload --port 8081
```

### 4.7 验证

```bash
curl http://localhost:8081/data/v1/health
# 期望返回：{"status":"ok","data_service":"running"}
```

---

## 五、三端共同配置（shared/）

### 5.1 目录结构

```
Z:\ZeroQuant2\shared\
├── config/
│   ├── backend.env.example          ← 妞妞 .env.local 模板
│   ├── frontend.env.example        ← 大聪明 .env.local 模板
│   └── data.env.example            ← 数据官 .env.local 模板
├── docs/
│   ├── 技术规范.md                   ← 当前 v2.2
│   ├── 接口约定_v0.1.md             ← 待 TASK-14 输出
│   └── 数据库设计_v0.1.md           ← 待输出
├── docs/roles/                     ← 各端专属文档
│   ├── 零号/
│   ├── 妞妞/
│   ├── 大聪明/
│   └── 数据官/
└── scripts/
    ├── init_postgres.sh
    ├── init_redis.sh
    └── dev_start.sh                ← 三端一键启动（待写）
```

### 5.2 数据库表前缀规范

| 前缀 | 所属 | 说明 |
|------|------|------|
| `biz_` | 妞妞（zeroquant_biz） | 业务表（用户、策略、订单） |
| `data_` | 数据官（zeroquant_data） | 数据表（行情、K线、盘口） |

### 5.3 端口约定

| 端口 | 服务 |
|------|------|
| 8080 | 妞妞 Go 后端 |
| 8081 | 数据官 Python 服务 |
| 5173 | 大聪明 Vue3 前端（Vite 默认） |
| 5432 | PostgreSQL |
| 6379/6380 | Redis |

---

## 六、三端同步机制

### 6.1 每日开工同步

```bash
# 每次开工前，从 NAS 拉取最新文件
# 三端通用

# 拉取共享内容
robocopy Z:\ZeroQuant2 D:\ZeroQuant2 /MIR /XD .git node_modules __pycache__ /XF .DS_Store

# 或用 Git（推荐，等 TASK-001 完成后）
cd D:\ZeroQuant2
git pull origin main
```

### 6.2 任务卡同步

- 发任务 → `tasks\{terminal}\inbox\{card}.md`
- 完成任务 → `tasks\{terminal}\done\{card}.md`
- 通过 NAS 交换，零号 10 分钟轮询一次

---

## 七、环境检查清单

### 7.1 妞妞检查项

- [ ] `go version` 返回 ≥ 1.21
- [ ] `psql --version` 返回 ≥ 14
- [ ] `redis-cli ping` 返回 PONG
- [ ] `docker ps` 正常
- [ ] `curl http://localhost:8080/api/v1/health` 返回成功

### 7.2 大聪明检查项

- [ ] `node -v` 返回 ≥ 20
- [ ] `pnpm -v` 返回 ≥ 9
- [ ] `pnpm install` 无报错
- [ ] `pnpm dev` 启动成功，页面可访问

### 7.3 数据官检查项

- [ ] `python --version` 返回 ≥ 3.10
- [ ] `uv --version` 或 `pip --version` 正常
- [ ] `redis-cli -p 6380 ping` 返回 PONG
- [ ] `curl http://localhost:8081/data/v1/health` 返回成功

---

## 八、常见问题

| 问题 | 解法 |
|------|------|
| Docker 无权限 | 以管理员身份运行 VSCode / 终端 |
| pnpm install 报错 | 先 `corepack enable` 再 `pnpm install` |
| Go 拉包慢 | 设置 GOPROXY=https://goproxy.cn,direct |
| akshare 取不到数据 | 检查网络，或改用 Tushare |
| Redis 连接失败 | 确认 Docker 容器运行中，检查端口映射 |
| Vue 跨域报错 | 确认 Vite proxy 配置指向 :8080 |
