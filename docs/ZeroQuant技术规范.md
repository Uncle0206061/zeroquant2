# ZeroQuant 2.0 技术规范

> 版本：v2.2
> 日期：2026-05-03
> 维护者：零号（PM）
> 状态：生效

---

## 一、技术栈

| 层级 | 技术选型 |
|------|----------|
| **后端（妞妞）** | Go + Gin + PostgreSQL + Redis + WebSocket |
| **前端（大聪明）** | Vue 3 + Element Plus + Vant（移动端）+ ECharts |
| **数据服务（数据官）** | Python + FastAPI + akshare / Tushare Pro + Level 2 数据源 |
| **部署** | Docker（Hola OS 待定） |

---

## 二、系统架构

### 系统架构

```
用户浏览器 / 移动端
    ↓ HTTP/WebSocket
Vue3 + Element Plus + Vant（双端适配）
    ↓ API
Go/Gin 业务服务（8080）← 妞妞
    ↓                 ↓
PostgreSQL          Redis
(biz_ 表)          (biz: 缓存)
    ↓
Python FastAPI（8081）← 数据官
    ↓                 ↓
PostgreSQL          Redis
(data_ 表)         (data: 缓存)
    ↓
akshare + Tushare Pro + Level 2 数据源
```

### 数据流

**实时数据流：** 外部数据源 → Python 采集器 → Redis → WebSocket → 前端

**策略执行流：** 用户编辑 → Go 引擎 → Python 筛选 → Go 执行 → 结果反馈

**回测数据流：** 发起回测 → Go 引擎 → Python 历史数据 → 模拟执行 → 结果返回

---

## 三、项目骨架

```
src/
├── backend/                    ← 妞妞负责（Go）
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/             ← 配置加载
│   │   ├── handler/            ← HTTP handler（Gin 路由处理）
│   │   ├── middleware/         ← 中间件（JWT/CORS/日志）
│   │   ├── model/              ← 数据模型
│   │   ├── repository/        ← 数据访问层
│   │   ├── service/           ← 业务逻辑层
│   │   ├── router/            ← 路由注册
│   │   ├── websocket/         ← WebSocket 实时通信
│   │   └── strategy/          ← 策略执行引擎
│   ├── pkg/
│   │   ├── response/          ← 统一响应
│   │   ├── jwt/               ← JWT 工具
│   │   └── logger/            ← 日志工具
│   ├── migrations/            ← 数据库迁移脚本
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/                   ← 大聪明负责（Vue3）
│   ├── public/
│   ├── src/
│   │   ├── api/                ← API 请求封装
│   │   ├── assets/             ← 静态资源
│   │   ├── components/         ← 公共组件（PC + 移动端）
│   │   ├── composables/        ← 组合式函数
│   │   ├── layouts/            ← 布局组件（PC + 移动端）
│   │   ├── router/             ← 路由配置
│   │   ├── stores/             ← Pinia 状态
│   │   ├── views/              ← 页面组件（PC + 移动端）
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
    │   │       ├── quote.py    ← 行情接口（Level 2）
    │   │       ├── kline.py    ← K 线接口
    │   │       ├── orderbook.py # Level 2 盘口接口
    │   │       └── sector.py   ← 板块接口
    │   ├── collectors/         ← 数据采集器
    │   │   ├── akshare_collector.py
    │   │   ├── tushare_collector.py
    │   │   └── level2_collector.py  # Level 2 采集
    │   ├── models/             ← SQLAlchemy 模型
    │   ├── schemas/            ← Pydantic 模型
    │   ├── services/           ← 业务逻辑
    │   ├── schedulers/         ← 定时任务
    │   ├── config.py
    │   └── main.py
    ├── alembic/
    ├── requirements.txt
    └── Dockerfile
```

---

## 四、API 规范

### 路由前缀

| 服务 | 前缀 | 说明 |
|------|------|------|
| 妞妞业务服务 | `/api/` | 用户操作、交易、策略 |
| 数据官数据服务 | `/data/` | 行情、盘口 K 线、板块 |

### 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误码规范

| 范围 | 含义 |
|------|------|
| 0 | 成功 |
| 1xxxx | 系统级错误（DB/Redis/网络） |
| 2xxxx | 认证授权错误 |
| 3xxxx | 参数校验错误 |
| 4xxxx | 业务逻辑错误（余额不足、风控拦截等） |
| 5xxxx | 外部服务错误（数据源超时等） |

// ⚠️ ZERO.CC 修改：从 40001-50099 格式改为 1xxxx-5xxxx 格式，与整体计划 V1.2 统一

### HTTP 方法

| 操作 | 方法 |
|------|------|
| 查询 | GET |
| 创建 | POST |
| 更新 | PUT |
| 删除 | DELETE |

### 接口性能要求（v2.2 新增）

| 指标 | 要求 |
|------|------|
| 接口超时 | 5 秒 |
| 单 IP 限流 | 60 次 / 分钟 |
| 接口响应（95%） | ≤300ms |
| 最大同时在线 | 30 人 |
| 下单接口 | 串行执行，防止并发重复下单 |

### WebSocket 规则（v2.2 新增）

| 规则 | 值 |
|------|-----|
| 心跳间隔 | 10 秒 |
| 断开自动重连 | 最多 10 次，间隔 2 秒 |
| 消息压缩 | 开启，减少带宽 |
| 推送延迟 | ≤200ms |

---

## 五、数据库规范

### 表命名

| 前缀 | 负责人 | 示例 |
|------|--------|------|
| `biz_` | 妞妞 | biz_user, biz_strategy, biz_order, biz_position |
| `data_` | 数据官 | data_quote, data_orderbook, data_kline, data_sector |

### 必备字段

每张表必须包含：`id`（主键）、`created_at`、`updated_at`

### 业务表设计（妞妞，biz_ 前缀）

| 表名 | 说明 |
|------|------|
| biz_user | 用户信息（含设备绑定字段） |
| biz_user_profile | 用户画像（板块喜好、买卖习惯等） |
| biz_strategy | 用户策略（含版本号、执行模式开关） |
| biz_strategy_rule | 策略规则（AND/OR/NOT 嵌套条件） |
| biz_watchlist | 自选股 |
| biz_portfolio | 模拟组合（初始资金 100 万） |
| biz_order | 交易记录（委托） |
| biz_position | 持仓（含盈亏计算） |
| biz_backtest | 回测任务 |
| biz_backtest_result | 回测结果（含收益、胜率、回撤） |
| biz_alert | 事件触发提醒 |

### 数据表设计（数据官，data_ 前缀）

| 表名 | 说明 |
|------|------|
| data_quote | 实时行情快照（永久存储） |
| data_kline | K 线数据（永久存储，默认前复权） |
| data_orderbook | Level 2 五档买卖盘（Redis 缓存 3 日） |
| data_tick | 逐笔成交（保存 90 天） |
| data_sector | 板块信息 |
| data_indicator | 计算指标缓存 |

### 安全存储（v2.2 新增）

- 交易密码：AES 加密存储
- 券商 Token：AES 加密存储
- 所有业务接口必须携带有效 JWT Token

---

## 六、用词统一表

### 核心业务术语

| 统一用词 | ❌ 禁用 | 英文 |
|----------|---------|------|
| **盘口** | 行情/报价 | orderbook / level2 |
| **逐笔委托** | 挂单/委托队列 | tick / orderflow |
| **五档买卖盘** | 五档/盘口 | bid/ask levels |
| **策略** | 规则/条件/模型 | strategy |
| **因子** | 指标/条件项 | factor |
| **自动交易** | 自动下单/全自动 | auto_trade |
| **事件触发** | 提醒/通知 | trigger / alert |
| **撮合** | 下单/成交 | match |
| **滑点** | 滑价 | slippage |
| **前复权** | 复权处理 | forward adjustment |
| **风控** | 风控规则/风控阈值 | risk control |

### 技术术语

| 统一用词 | ❌ 禁用 |
|----------|---------|
| 处理器 | 处理函数/控制器 |
| 服务层 | 业务层/逻辑层 |
| 仓库层 | 数据层/DAO |
| 中间件 | 拦截器/过滤器 |

---

## 七、Go 编码规范（妞妞）

| 规则 | 正确 | 错误 |
|------|------|------|
| 文件名 | snake_case | PascalCase |
| 导出函数 | PascalCase | camelCase |
| 私有函数 | camelCase | — |
| 层级调用 | handler → service → repository | 禁止跨层 |
| 错误处理 | 显式处理，无 panic | 吞错误 |
| 下单接口 | 必须串行，不并发 | — |

---

## 八、Vue3 编码规范（大聪明）

- 组件命名：PascalCase
- 响应式优先使用 `<script setup>`
- 移动端适配：Vant UI 组件库
- API 函数：动词 + 名词，如 `getQuote`、`submitOrder`
- 策略编辑器：表单勾选 + 填写形态（v2.2 确定）

---

## 九、Python 编码规范（数据官）

- 文件名：snake_case
- 类名：PascalCase
- 函数名：snake_case + 类型注解
- Level 2 数据处理：高速循环写入

### 数据源容错（v2.2 新增）

| 场景 | 处理方式 |
|------|---------|
| 主数据源断开 | 1 秒内自动切换备用源 |
| 全部源失效 | 返回错误，不返回脏数据 |
| 缺失数据 | 自动补全，无法补全则标记并触发告警 |

---

## 十、Git 规范

- 分支：`main` / `dev` / `feature/{模块}`
- Commit：`type(scope): subject`（feat/fix/docs/refactor）
- 禁止直接 push 到 main

---

## 十一、代码审查清单

- [ ] 文件名、命名符合规范
- [ ] 错误全部显式处理
- [ ] 统一响应格式
- [ ] WebSocket 心跳机制（10 秒）
- [ ] Level 2 数据存储方案（Redis 3 天、K 线永久、逐笔 90 天）
- [ ] 用词与本文档统一
- [ ] Commit message 符合规范
- [ ] 下单接口串行执行
- [ ] JWT Token 鉴权
- [ ] 风控规则前端可配置、后端强制执行

---

## 十二、端口配置

| 服务 | 端口 |
|------|------|
| 妞妞后端 | 8080 |
| 数据官数据服务 | 8081 |
| 前端开发 | 5173 |
| PostgreSQL | 5432 |
| Redis | 6379 |

---

*本规范是 ZeroQuant 2.0 的唯一技术标准。v2.2 更新：接口性能、WebSocket、数据源容错、用词补充。*