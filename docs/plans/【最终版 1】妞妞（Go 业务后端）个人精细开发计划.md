# 妞妞（Go 业务后端）个人精细开发计划｜最终版
文档版本：V1.3
// ⚠️ ZERO.CC 修改：适用周期和开发周期改为相对时间
姓名：妞妞
适用周期：以开发启动日为 Day 1，总周期约 18 天
参考文档：ZeroQuant 2.0 需求文档 v2.2、整体开发计划 V1.1、技术规范 v2.2
对齐里程碑：M1、M3、M4、M5

---

## 1. 基本信息
- 负责模块：用户体系、策略引擎、模拟交易、WebSocket、权限风控、订单/持仓
- 开发周期：以开发启动日为 Day 1，总周期约 18 天
- 服务端口：8080
- 数据库/缓存：PostgreSQL（biz 库）、Redis（biz: 前缀缓存）

---

## 2. 环境配置（必须 T1 当日 18:00 前完成）
- 系统：Windows / Linux
- Go 版本：1.22+
- IDE：GoLand / VSCode
- 必备工具：Postman、ApiFox、GoTest、Git
- 核心依赖：Gin、GORM、Redis、JWT、WebSocket、CORS

---

## 3. 目录规范（严格执行）
backend/
├── cmd/server/main.go
├── internal/
│   ├── config/             ← 配置加载
│   ├── handler/            ← HTTP handler（Gin 路由处理）
│   ├── middleware/         ← 中间件（JWT/CORS/日志）
│   ├── model/              ← 数据模型
│   ├── repository/         ← 数据访问层
│   ├── service/            ← 业务逻辑层
│   ├── router/             ← 路由注册
│   ├── websocket/          ← WebSocket 实时通信
│   └── strategy/           ← 策略执行引擎
├── pkg/
│   ├── response/           ← 统一响应格式
│   ├── jwt/                ← JWT 工具
│   └── logger/             ← 日志工具
├── migrations/             ← 数据库迁移脚本
├── go.mod
└── Dockerfile

---

## 4. 每日任务 + 审核点
### Day 1（T1）｜M1 项目骨架就绪
- 完成环境安装、项目初始化、DB/Redis 连接
- 实现 /api/v1/health 接口（返回 {code:0, data:{status:"ok"}}）
- 输出 API 草案 v0.1（含统一响应格式、错误码规范）
审核点：服务启动正常、数据库连通、健康检查接口可用

### Day 2–3（T1+1~T1+2）｜支撑 M2 数据服务可用
- 用户注册（POST /api/v1/auth/register）、登录（POST /api/v1/auth/login）
- JWT 鉴权中间件
- 权限中间件
- 用户画像接口（GET /api/v1/user/profile）
审核点：登录态正常、权限拦截生效

### Day 4–5（T1+3~T1+4）｜M3 业务核心可用
- 策略增删改查（/api/v1/strategy/**）
- 策略调用 Python 数据服务（/data/v1/filter/）
- WebSocket 心跳、重连、实时推送
- **WebSocket 配置（技术规范强制）：**
  - 心跳间隔：10 秒
  - 断开自动重连：最多 10 次，间隔 2 秒
  - 推送延迟：≤200ms
  - 全局唯一 WebSocket 实例
审核点：策略可下发可回调、推送延迟 ≤200ms、重连机制有效

### Day 6–7（T1+5~T1+6）｜M4 模拟交易闭环
- 模拟账户、订单、持仓、撮合
- 风控规则：仓位/亏损/交易次数/ST 限制
- 与前端联调
- 下单接口串行执行（禁止并发）
审核点：模拟交易全流程通、风控生效

### Day 8–11（T1+7~T1+10）｜M5 实盘能力就绪
- 券商接口对接
- 模拟/实盘完全隔离
- 二次确认、一键暂停、异常冻结
审核点：实盘安全、隔离有效、暂停 <1s

### Day 12–18（T1+11~T1+17）｜配合 M6 正式发布
- 性能优化、监控、日志、备份、Docker 构建
审核点：系统稳定、交付物齐全

---

## 5. 核心接口（统一 /api/v1/ 前缀）
- GET  /api/v1/health              ← 健康检查
- POST /api/v1/auth/register       ← 用户注册
- POST /api/v1/auth/login          ← 用户登录
- GET  /api/v1/auth/me             ← 当前用户
- GET  /api/v1/user/profile        ← 用户画像
- PUT  /api/v1/user/profile        ← 更新画像
- GET  /api/v1/strategy/list       ← 策略列表
- POST /api/v1/strategy/create     ← 创建策略
- GET  /api/v1/strategy/:id        ← 策略详情
- PUT  /api/v1/strategy/:id        ← 更新策略
- DELETE /api/v1/strategy/:id      ← 删除策略
- POST /api/v1/strategy/:id/submit ← 提交执行
- GET  /api/v1/order/list           ← 订单列表
- GET  /api/v1/position            ← 持仓
- GET  /api/v1/portfolio           ← 组合
- WS   /api/v1/ws                  ← WebSocket

> **路由版本规范**：所有接口必须带 `/v1/` 版本前缀，与数据服务（/data/v1/）保持一致。

---

## 6. 核心数据表（统一 biz_ 前缀，PostgreSQL biz 库）
biz_user、biz_user_profile、biz_strategy、biz_strategy_rule、biz_watchlist、biz_portfolio、biz_order、biz_position、biz_backtest、biz_backtest_result、biz_alert

> **前缀说明**：所有表统一加 `biz_` 前缀，由 PostgreSQL `biz` 库管理，与数据官的 `data_` 表（`data` 库）完全隔离。

---

## 7. 开发规范
- 每日 23:50 前提交 dev
- 下单接口必须串行加锁，禁止并发
- 错误不允许 panic
- JWT 标准格式：Bearer Token
- 提交格式：[妞妞] 说明