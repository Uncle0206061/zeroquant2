# ZeroQuant 2.0 Backend

> A股全自动量化交易系统后端服务 | Go + Gin + GORM + PostgreSQL + Redis + WebSocket

## 项目简介

ZeroQuant 2.0 面向小圈子（≤15人）的 A股全自动量化交易系统，后端采用 B/S 架构，
用户通过图形化界面自主编写量化策略，系统全自动执行模拟/实盘交易。

## 技术栈

| 层级 | 技术选型 |
|------|----------|
| 语言 | Go 1.26+ |
| 框架 | Gin Web Framework |
| ORM | GORM (PostgreSQL) |
| 数据库 | PostgreSQL 16 |
| 缓存 | Redis 7 |
| 实时通信 | WebSocket |
| 鉴权 | JWT (HS256) |
| API文档 | Swagger 2.0 |
| 容器化 | Docker + Docker Compose |
| 日志 | 结构化 JSON (自带 request_id) |

## 目录结构

```
backend/
├── cmd/server/          # 程序入口
│   └── main.go
├── internal/
│   ├── broker/          # 券商接口抽象
│   │   ├── broker_interface.go
│   │   └── mock_broker.go      # Phase 1 Mock实现
│   ├── cache/           # Redis缓存层
│   │   └── cache.go
│   ├── config/          # 配置管理（YAML+环境变量覆盖）
│   │   ├── config.go
│   │   └── config.yaml
│   ├── handler/          # HTTP处理器（路由注解→Swagger）
│   │   ├── auth_handler.go          # 用户注册/登录
│   │   ├── strategy_handler.go      # 策略管理
│   │   ├── trade_handler.go         # 模拟交易
│   │   ├── real_trade_handler.go     # 实盘交易
│   │   ├── health_handler.go        # 健康检查/系统监控
│   │   └── admin_handler.go         # 管理功能
│   ├── middleware/      # 中间件（JWT/CORS/RequestID/限流）
│   │   ├── middleware.go
│   │   └── trade_mode.go           # 实盘模式开关
│   ├── model/           # 数据模型（AutoMigrate自动建表）
│   │   ├── user.go
│   │   ├── strategy.go
│   │   ├── order.go
│   │   ├── position.go
│   │   ├── portfolio.go
│   │   ├── real_trade_model.go
│   │   └── backtest.go
│   ├── repository/      # 数据访问层（禁止跨层调用）
│   │   ├── user_repository.go
│   │   ├── strategy_repository.go
│   │   ├── order_repository.go
│   │   ├── position_repository.go
│   │   ├── portfolio_repository.go
│   │   └── real_*.go
│   ├── router/          # 路由注册
│   │   └── router.go
│   ├── service/         # 业务逻辑层（撮合引擎/风控）
│   │   ├── auth_service.go
│   │   ├── strategy_service.go
│   │   ├── order_service.go        # 模拟交易撮合
│   │   ├── real_order_service.go    # 实盘交易
│   │   └── backtest_service.go
│   └── websocket/      # WebSocket Hub（实时推送）
│       ├── hub.go
│       └── websocket.go
├── pkg/
│   ├── logger/         # 结构化JSON日志
│   │   └── logger.go
│   ├── response/       # 统一响应封装
│   │   └── response.go
│   └── jwt/            # JWT工具
│       └── jwt.go
├── docs/               # Swagger文档（swag init自动生成）
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── scripts/
│   └── backup/
│       └── pg_backup.sh    # PostgreSQL每日备份（保留7天）
├── Dockerfile          # 多阶段构建（目标镜像<200MB）
├── docker-compose.yml # PostgreSQL + Redis + Backend一键部署
├── .env.example       # 环境变量模板
├── .dockerignore
├── .gitignore
├── go.mod
└── go.sum
```

## 快速启动

### 方式一：Docker 一键启动（推荐）

```bash
cd backend
cp .env.example .env
# 编辑 .env，填入真实密钥后：
docker compose up -d
# 访问 http://localhost:8080/swagger/index.html
```

### 方式二：源码运行

#### 前置依赖

- Go 1.26+
- PostgreSQL 16（已建库 `biz`，用户 `postgres`，密码 `zq215007`）
- Redis 7+

#### 启动步骤

```bash
cd backend
go mod tidy
go build -o server ./cmd/server/
./server
# 服务启动于 :8080
# 访问 http://localhost:8080/swagger/index.html
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_HOST` | PostgreSQL地址 | localhost |
| `DB_PORT` | PostgreSQL端口 | 5432 |
| `DB_USER` | 数据库用户 | postgres |
| `DB_PASSWORD` | 数据库密码 | zq215007 |
| `DB_NAME` | 数据库名 | biz |
| `REDIS_HOST` | Redis地址 | localhost |
| `REDIS_PORT` | Redis端口 | 6379 |
| `JWT_SECRET` | JWT签名密钥 | （必须设置，>=32字符） |
| `SERVER_PORT` | 服务端口 | 8080 |
| `REAL_TRADE` | 实盘开关 | false |
| `DATA_SERVICE_URL` | Python数据服务URL | http://localhost:8081 |

环境变量会覆盖 `config.yaml` 中的同名配置。

## API 文档

Swagger 交互式文档：服务启动后访问

```
http://localhost:8080/swagger/index.html
```

所有业务接口（认证除外）需携带 JWT Token：

```
Authorization: Bearer <token>
```

### 核心接口一览

**认证**
| 方法 | 路由 | 说明 |
|------|------|------|
| POST | /api/v1/auth/register | 用户注册 |
| POST | /api/v1/auth/login | 用户登录 |
| GET | /api/v1/auth/me | 当前用户信息 |

**策略管理**
| 方法 | 路由 | 说明 |
|------|------|------|
| POST | /api/v1/strategy/create | 创建策略 |
| GET | /api/v1/strategy/list | 策略列表 |
| GET | /api/v1/strategy/:id | 策略详情 |
| PUT | /api/v1/strategy/:id | 更新策略 |
| DELETE | /api/v1/strategy/:id | 删除策略 |
| POST | /api/v1/strategy/:id/submit | 提交策略执行 |
| GET | /api/v1/strategy/:id/backtests | 回测记录 |

**模拟交易**
| 方法 | 路由 | 说明 |
|------|------|------|
| POST | /api/v1/account/simulate/create | 创建模拟账户（初始100万） |
| GET | /api/v1/account | 查询账户 |
| POST | /api/v1/order/submit | 提交订单（实时撮合） |
| GET | /api/v1/order/list | 订单列表 |
| GET | /api/v1/order/:id | 订单详情 |
| DELETE | /api/v1/order/:id | 撤单 |
| GET | /api/v1/position | 持仓列表 |
| GET | /api/v1/position/:stock_code | 持仓详情 |

**实盘交易（需 REAL_TRADE=true）**
| 方法 | 路由 | 说明 |
|------|------|------|
| POST | /api/v1/trade/real/account/create | 创建实盘账户 |
| GET | /api/v1/trade/real/account | 查询实盘账户 |
| POST | /api/v1/trade/real/order/submit | 下单（→pending，需二次确认） |
| POST | /api/v1/trade/real/order/confirm | 二次确认成交 |
| GET | /api/v1/trade/real/order/list | 订单列表 |
| GET | /api/v1/trade/real/order/:id | 订单详情 |
| DELETE | /api/v1/trade/real/order/:id | 撤单 |
| GET | /api/v1/trade/real/position | 持仓 |
| GET | /api/v1/trade/real/log | 操作日志 |

**系统监控**
| 方法 | 路由 | 说明 |
|------|------|------|
| GET | /api/v1/health | 健康检查 |
| GET | /api/v1/ping | 存活探测 |
| GET | /api/v1/stats | 系统指标（内存/连接池/WS/Redis） |

**WebSocket**
```
ws://localhost:8080/ws?token=<jwt_token>
```

推送事件类型：`order_update` | `position_update` | `system_alert` | `strategy_signal`

## 风控规则（强制）

| 规则 | 值 |
|------|-----|
| 单只个股最大仓位 | 总资金 30% |
| 单日最大总亏损 | 总资金 5%（超限冻结当日交易）|
| 单日最大交易次数 | 50次 |
| 禁止买入 | ST股、*ST股、退市整理股、未开板新股 |
| 止盈止损 | 固定5%/10%/自定义 |

## 系统架构

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   前端 (Vue3)    │────▶│  Go后端 (:8080) │────▶│  PostgreSQL     │
│   大聪明开发     │     │  妞妞开发       │     │  :5432          │
│   :5173         │     │                 │     │                 │
└─────────────────┘     │                 │     └─────────────────┘
                        │                 │     ┌─────────────────┐
┌─────────────────┐     │                 │────▶│  Redis          │
│  Python数据服务  │────▶│  撮合引擎       │     │  :6379          │
│  数据官开发     │     │  风控引擎       │     └─────────────────┘
│  :8081           │     │  策略执行       │
└─────────────────┘     │  WebSocket推送   │     ┌─────────────────┐
                        └─────────────────┘────▶│  券商接口        │
                                                  │  (Phase 2接入)   │
                                                  └─────────────────┘
```

## 开发规范

- **文件命名**：snake_case（`*_service.go`）
- **导出规则**：PascalCase 对外接口，camelCase 私有方法
- **层级调用**：handler → service → repository，禁止跨层
- **错误处理**：显式处理，无 panic
- **敏感数据**：AES 加密存储（交易密码/券商Token）
- **接口鉴权**：所有业务接口必须携带有效 JWT Token

## 部署架构

```bash
# 生产环境一键启动
docker compose up -d

# 查看日志
docker compose logs -f backend

# 查看服务状态
docker compose ps
```

默认配置：
- Backend: `:8080`
- PostgreSQL: `:5432`（数据持久化到 `pgdata/`）
- Redis: `:6379`

## License

私有项目，仅供团队内部使用。