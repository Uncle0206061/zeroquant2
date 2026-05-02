# ZeroQuant 2.0 技术规范

> 版本：v1.0
> 日期：2026-05-02
> 维护者：零号（PM）
> 状态：生效

---

## 一、技术栈

| 层级 | 技术选型 |
|------|----------|
| **后端** | Go + Gin + PostgreSQL + Redis + WebSocket |
| **前端** | Vue 3 + Element Plus + TradingView / ECharts |
| **数据服务** | Python + FastAPI + akshare / Tushare Pro |
| **部署** | Docker / K8s |

---

## 二、项目骨架

```
src/
├── backend/                    ← 妞妞负责（Go）
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         ← 入口
│   ├── internal/
│   │   ├── config/             ← 配置加载
│   │   ├── handler/            ← HTTP handler（Gin路由处理）
│   │   ├── middleware/         ← 中间件（JWT/CORS/日志）
│   │   ├── model/              ← 数据模型
│   │   ├── repository/         ← 数据访问层
│   │   ├── service/            ← 业务逻辑层
│   │   └── router/             ← 路由注册
│   ├── pkg/
│   │   ├── response/           ← 统一响应
│   │   ├── jwt/                ← JWT工具
│   │   └── logger/             ← 日志工具
│   ├── migrations/             ← 数据库迁移脚本
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── frontend/                   ← 大聪明负责（Vue3）
│   ├── public/
│   ├── src/
│   │   ├── api/                ← API 请求封装
│   │   ├── assets/             ← 静态资源
│   │   ├── components/         ← 公共组件
│   │   ├── composables/        ← 组合式函数
│   │   ├── layouts/            ← 布局组件
│   │   ├── router/             ← 路由配置
│   │   ├── stores/             ← Pinia 状态
│   │   ├── views/              ← 页面组件
│   │   ├── utils/              ← 工具函数
│   │   ├── App.vue
│   │   └── main.ts
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   └── Dockerfile
│
└── data/                       ← 数据官负责（Python）
    ├── app/
    │   ├── api/                ← FastAPI 路由
    │   │   └── v1/
    │   │       ├── quote.py    ← 行情接口
    │   │       ├── kline.py    ← K线接口
    │   │       └── sector.py   ← 板块接口
    │   ├── collectors/         ← 数据采集器
    │   │   ├── akshare_collector.py
    │   │   └── tushare_collector.py
    │   ├── models/             ← SQLAlchemy 模型
    │   ├── schemas/            ← Pydantic 模型
    │   ├── services/           ← 业务逻辑
    │   ├── schedulers/         ← 定时任务
    │   ├── config.py           ← 配置
    │   └── main.py             ← FastAPI 入口
    ├── alembic/                ← 数据库迁移
    ├── requirements.txt
    └── Dockerfile
```

---

## 三、API 规范

### 3.1 路由前缀

| 服务 | 前缀 | 示例 |
|------|------|------|
| 妞妞业务服务 | `/api/` | `GET /api/health` |
| 数据官数据服务 | `/data/` | `GET /data/quote?code=000001` |

### 3.2 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**错误响应：**
```json
{
  "code": 40001,
  "message": "股票代码不存在",
  "data": null
}
```

### 3.3 错误码规范

| 范围 | 含义 | 示例 |
|------|------|------|
| 0 | 成功 | — |
| 40001-40099 | 参数错误 | 40001=无效股票代码 |
| 40101-40199 | 认证错误 | 40101=Token过期 |
| 40401-40499 | 资源不存在 | 40401=用户不存在 |
| 50001-50099 | 服务器内部错误 | 50001=数据库异常 |

### 3.4 HTTP 方法选择

| 操作 | 方法 | 示例 |
|------|------|------|
| 查询 | GET | `GET /api/health` |
| 创建 | POST | `POST /api/watchlist` |
| 更新 | PUT | `PUT /api/watchlist/:id` |
| 删除 | DELETE | `DELETE /api/watchlist/:id` |

---

## 四、数据库规范

### 4.1 表命名

- 妞妞表前缀：`biz_`（如 `biz_user`, `biz_watchlist`, `biz_order`）
- 数据官表前缀：`data_`（如 `data_stock_daily`, `data_stock_info`）
- 字段命名：snake_case（如 `created_at`, `user_id`）

### 4.2 必备字段

每张表必须包含：
- `id` — 主键（BIGSERIAL 或 UUID）
- `created_at` — 创建时间（TIMESTAMPTZ, DEFAULT NOW()）
- `updated_at` — 更新时间（TIMESTAMPTZ, DEFAULT NOW()）

### 4.3 Redis Key

- 妞妞前缀：`biz:`
- 数据官前缀：`data:`
- 示例：`biz:user:token:{uid}`, `data:quote:{code}`

---

## 五、Go 编码规范（妞妞遵守）

### 5.1 文件与目录

| 规则 | 正确 | 错误 |
|------|------|------|
| 文件名 snake_case | `health_handler.go` | `healthHandler.go` |
| 包名小写单词 | `handler`, `service` | `Handler`, `httpHandler` |

### 5.2 命名

| 类型 | 风格 | 示例 |
|------|------|------|
| 导出函数/方法 | PascalCase | `GetHealth()`, `NewRouter()` |
| 私有函数/方法 | camelCase | `parseConfig()`, `buildResponse()` |
| 接口 | PascalCase + er 后缀 | `HealthChecker`, `UserRepository` |
| 结构体 | PascalCase | `HealthResponse`, `RouterConfig` |
| 错误变量 | Err 前缀 | `ErrNotFound`, `ErrUnauthorized` |

### 5.3 层级调用

**规则：** `handler → service → repository`

严禁跨层调用，严禁 handler 直接写 SQL。

### 5.4 错误处理

```go
// ✅ 正确：显式处理每个错误
if err != nil {
    return nil, fmt.Errorf("get health status: %w", err)
}

// ❌ 错误：吞掉错误或 panic
if err != nil {
    log.Println(err)
    return nil, nil
}
```

---

## 六、Vue3 编码规范（大聪明遵守）

### 6.1 文件命名

| 类型 | 风格 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `KlineChart.vue` |
| 组合式函数 | use + PascalCase | `useQuote.ts` |
| Store | use + Name + Store | `useWatchlistStore.ts` |
| API 函数 | 动词 + 名词 | `getQuote`, `createOrder` |

### 6.2 组件结构

```vue
<script setup lang="ts">
// 1. imports
// 2. props/emits
// 3. composables
// 4. reactive state
// 5. computed
// 6. methods
// 7. lifecycle hooks
</script>

<template>
  <!-- 模板 -->
</template>

<style scoped>
  /* 样式 */
</style>
```

---

## 七、Python 编码规范（数据官遵守）

### 7.1 文件与命名

| 类型 | 风格 | 示例 |
|------|------|------|
| 文件名 | snake_case | `akshare_collector.py` |
| 类名 | PascalCase | `QuoteService` |
| 函数名 | snake_case | `get_daily_kline` |
| 常量 | UPPER_SNAKE_CASE | `DEFAULT_LIMIT` |

### 7.2 类型注解

所有函数必须有类型注解：

```python
def get_quote(code: str) -> dict[str, Any]:
    ...
```

---

## 八、用词统一表

### 8.1 核心术语

| 统一用词 | ❌ 禁用 | 代码中的英文 |
|----------|---------|-------------|
| **健康检查** | 健康检测/健康探针 | health |
| **自选股** | 关注股/收藏股 | watchlist |
| **行情** | 报价/行情数据 | quote |
| **K线** | 蜡烛图/OHLC | kline |
| **板块** | 行业/概念 | sector |
| **模拟交易** | 虚拟交易/纸上交易 | paper_trading |
| **策略** | 策略模型/交易策略 | strategy |
| **回测** | 历史回测 | backtest |

### 8.2 技术术语

| 统一用词 | ❌ 禁用 | 代码中的英文 |
|----------|---------|-------------|
| **处理器** | 处理函数/控制器 | handler |
| **服务层** | 业务层/逻辑层 | service |
| **仓库层** | 数据层/DAO层 | repository |
| **中间件** | 拦截器/过滤器 | middleware |
| **路由** | 路由器/路径 | router |
| **模型** | 实体/对象 | model |
| **配置** | 设置/参数 | config |
| **响应** | 返回值/输出 | response |
| **请求** | 输入/参数 | request |
| **认证** | 鉴权/授权 | auth |

### 8.3 业务术语

| 统一用词 | ❌ 禁用 | 代码中的英文 |
|----------|---------|-------------|
| **买入** | 购买/做多 | buy |
| **卖出** | 出售/做空 | sell |
| **持仓** | 头寸/仓位 | position |
| **委托** | 订单/挂单 | order |
| **成交** | 交易/执行 | trade |
| **止损** | 止蚀/割肉 | stop_loss |
| **止盈** | 目标价/获利 | take_profit |
| **收益率** | 回报率/利润率 | return_rate |
| **涨跌幅** | 涨跌/变化率 | change_percent |

---

## 九、Git 规范

### 9.1 分支

| 分支 | 用途 |
|------|------|
| `main` | 稳定版本，禁止直接 push |
| `dev` | 开发分支 |
| `feature/{模块名}` | 功能分支 |
| `hotfix/{描述}` | 紧急修复 |

### 9.2 Commit

格式：`type(scope): subject`

| type | 含义 |
|------|------|
| feat | 新功能 |
| fix | 修复 |
| docs | 文档 |
| refactor | 重构 |
| test | 测试 |
| chore | 杂务 |

示例：
- `feat(backend): 添加健康检查接口`
- `fix(backend): 修复响应格式错误`
- `docs(all): 更新编码规范`

---

## 十、代码审查清单

零号终审时会逐项检查：

- [ ] 文件名符合命名规范
- [ ] 包名/类名符合规范
- [ ] 导出函数 PascalCase，私有 camelCase/snake_case
- [ ] 错误全部显式处理，无 panic/吞错误
- [ ] 统一响应格式（code/message/data）
- [ ] 路由前缀正确（`/api/` 或 `/data/`）
- [ ] Context 传递正确（Go）
- [ ] 类型注解完整（Python）
- [ ] 用词与本文档统一
- [ ] Commit message 符合规范
- [ ] 代码无硬编码配置

---

*本规范是 ZeroQuant 2.0 的唯一技术标准。如有变更，零号负责更新。*
