# ZeroQuant 2.0 前端 API 规范文档

> 版本：v0.1 | 最后更新：2026-05-14
> 作者：大聪明（dacongming）

---

## 1. 认证接口

### 1.1 登录
- **POST** `/api/v1/auth/login`
- 请求体：`{ phone: string, password: string }`
- 响应：`{ code: 0, message: "ok", data: { token: "jwt...", username: "..." } }`

### 1.2 注册
- **POST** `/api/v1/auth/register`
- 请求体：`{ phone: string, password: string, nickname?: string }`
- 响应：`{ code: 0, message: "ok" }`

### 1.3 Token 刷新
- **POST** `/api/v1/auth/refresh`
- 请求头：`Authorization: Bearer <token>`

---

## 2. 策略接口

### 2.1 策略列表
- **GET** `/api/v1/strategy/list`
- 查询参数：`?page=1&pageSize=20&status=running`
- 响应：`{ code: 0, data: { list: Strategy[], total: number } }`

### 2.2 策略详情
- **GET** `/api/v1/strategy/:id`

### 2.3 创建策略
- **POST** `/api/v1/strategy/create`
- 请求体：策略 JSON（因子 + 规则组合）

### 2.4 更新策略
- **PUT** `/api/v1/strategy/:id`

### 2.5 删除策略
- **DELETE** `/api/v1/strategy/:id`

### 2.6 启/停策略
- **POST** `/api/v1/strategy/:id/start`
- **POST** `/api/v1/strategy/:id/stop`

---

## 3. 回测接口

### 3.1 启动回测
- **POST** `/api/v1/backtest/run`
- 请求体：`{ strategyId: string, startDate: string, endDate: string }`

### 3.2 回测结果
- **GET** `/api/v1/backtest/:taskId/result`

### 3.3 回测进度
- 通过 WebSocket 推送事件 `backtest_progress`

---

## 4. 持仓接口

### 4.1 当前持仓
- **GET** `/api/v1/position`
- 响应：`{ code: 0, data: Position[] }`

### 4.2 持仓历史
- **GET** `/api/v1/position/history?page=1&pageSize=20`

---

## 5. 订单接口

### 5.1 订单列表
- **GET** `/api/v1/order/list?page=1&pageSize=20&status=all`

### 5.2 手动下单
- **POST** `/api/v1/order/create`
- 请求体：`{ code: string, direction: "buy"|"sell", volume: number, price?: number }`

---

## 6. 数据服务接口（Python）

> Base URL: `/data/v1`（端口 8081，通过 Vite 代理到 8080）

### 6.1 实时行情
- **GET** `/data/v1/quote/:code`
- 响应：`{ code: 0, data: { code, name, price, change, volume, ... } }`

### 6.2 K 线数据
- **GET** `/data/v1/kline/:code?period=1d&count=100`
- period: 1m / 5m / 15m / 30m / 1h / 1d

### 6.3 市场概览
- **GET** `/data/v1/market/:code`

---

## 7. WebSocket 事件

> 连接地址：`ws://localhost:8080/api/v1/ws?token=<jwt>`

### 7.1 心跳
- 客户端 → 服务端：`{ type: "ping" }`，每 10 秒
- 服务端 → 客户端：`{ type: "pong" }`

### 7.2 事件推送
| 事件类型 | type 字段 | 说明 |
|----------|-----------|------|
| 订单更新 | `order_update` | 成交/委托状态变更 |
| 持仓变更 | `position_update` | 仓位变动 |
| 回测进度 | `backtest_progress` | 回测执行进度 |
| 策略触发 | `strategy_trigger` | 策略条件触发通知 |
| 系统公告 | `system_notice` | 系统消息 |

### 7.3 重连策略
- 最大重连：10 次
- 间隔：2 秒
- Token 失效（401）：清除本地 token，跳转 /login

---

## 8. 统一响应格式

```json
{
  "code": 0,       // 0=成功，非0=失败
  "message": "ok",
  "data": { ... }  // 业务数据
}
```

### 错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| 40101 | 未登录 |
| 40102 | Token 过期 |
| 40103 | Token 无效 |
| 40301 | 无权限 |
| 50001 | 服务内部错误 |

---

## 9. 前端 Token 管理

| Key | 说明 |
|-----|------|
| `zq_token` | JWT Token（localStorage） |
| `zq_username` | 用户名（localStorage） |

- 所有请求通过 Axios 拦截器自动附加 `Authorization: Bearer <token>`
- 401 响应自动清除 token 并跳转 /login
- 路由守卫检查 token 存在性

---

## 10. Vite 代理配置

```
/api/*  → http://localhost:8080  (Go 后端)
/data/* → http://localhost:8080  (由 Go 后端转发到 Python 数据服务 8081)
```

开发环境端口：
- 前端：5173
- Go 后端：8080
- Python 数据服务：8081
