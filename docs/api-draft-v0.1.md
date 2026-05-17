# ZeroQuant 2.0 API 草案 v0.1

> 版本：v0.1 | 日期：2026-05-08 | 状态：草案（待实现）
> 后端服务：妞妞（Go + Gin，端口 8080）
> 数据服务：数据官（Python + FastAPI，端口 8081）

---

## 1. 统一响应格式

所有接口均返回以下 JSON 结构：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 0=成功，其他=错误（见错误码表） |
| message | string | 状态描述 |
| data | object | 响应数据（无数据时省略） |

---

## 2. 错误码规范

| 错误码范围 | 含义 | HTTP 状态码 |
|-----------|------|------------|
| 0 | 成功 | 200 |
| 40001~40099 | 参数错误 | 400 |
| 40101~40199 | 认证错误（未登录/Token无效） | 401 |
| 40301~40399 | 权限不足 | 403 |
| 40401~40499 | 资源不存在 | 404 |
| 50001~50099 | 服务器内部错误 | 500 |

### 常用错误码

| 错误码 | 含义 |
|--------|------|
| 40001 | 参数缺失 |
| 40002 | 参数格式错误 |
| 40101 | Token 为空 |
| 40102 | Token 已过期 |
| 40103 | Token 无效 |
| 40401 | 用户不存在 |
| 40402 | 策略不存在 |
| 50001 | 服务器内部错误 |

---

## 3. 路由清单

### 3.1 系统接口（无需认证）

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| GET | /api/v1/health | 健康检查 | 返回服务状态 |
| WS | /api/v1/ws | WebSocket 连接 | 实时数据推送 |

### 3.2 用户认证

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| POST | /api/v1/auth/register | 用户注册 | 手机号+密码 |
| POST | /api/v1/auth/login | 用户登录 | 返回 JWT Token |
| POST | /api/v1/auth/logout | 登出 | 使 Token 失效 |
| POST | /api/v1/auth/refresh | 刷新 Token | 刷新过期 Token |

### 3.3 用户管理

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| GET | /api/v1/user/profile | 获取用户资料 | |
| PUT | /api/v1/user/profile | 更新用户资料 | |
| PUT | /api/v1/user/password | 修改密码 | |

### 3.4 自选股

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| GET | /api/v1/watchlist | 获取自选股列表 | |
| POST | /api/v1/watchlist | 添加自选股 | { stock_code } |
| DELETE | /api/v1/watchlist/{stock_code} | 删除自选股 | |

### 3.5 模拟交易

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| GET | /api/v1/portfolio | 获取持仓组合 | 初始资金 100 万 |
| GET | /api/v1/portfolio/orders | 获取委托记录 | |
| POST | /api/v1/orders | 下单（模拟） | 含风控校验 |
| DELETE | /api/v1/orders/{order_id} | 撤单 | |
| GET | /api/v1/portfolio/positions | 获取持仓明细 | 含盈亏计算 |

### 3.6 策略管理

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| GET | /api/v1/strategies | 获取策略列表 | |
| POST | /api/v1/strategies | 创建策略 | |
| GET | /api/v1/strategies/{id} | 获取策略详情 | |
| PUT | /api/v1/strategies/{id} | 更新策略 | |
| DELETE | /api/v1/strategies/{id} | 删除策略 | |
| POST | /api/v1/strategies/{id}/start | 启动策略 | |
| POST | /api/v1/strategies/{id}/stop | 停止策略 | |

### 3.7 回测

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| POST | /api/v1/backtest | 发起回测 | { strategy_id, start_date, end_date } |
| GET | /api/v1/backtest/{task_id} | 查询回测结果 | 含收益、胜率、回撤 |

### 3.8 数据接口（数据服务 /data/ 前缀）

> 以下由数据官（Python FastAPI 8081）提供，Go 后端代理转发。

| 方法 | 路径 | 说明 | 备注 |
|------|------|------|------|
| GET | /data/v1/quote | 实时行情 | { stock_codes[] } |
| GET | /data/v1/kline | K 线数据 | { stock_code, period, count } |
| GET | /data/v1/orderbook | 五档盘口 | { stock_code } |

---

## 4. 接口性能要求

| 指标 | 要求 |
|------|------|
| 接口超时 | 5 秒 |
| 单 IP 限流 | 60 次/分钟 |
| 接口响应（95%） | ≤300ms |
| 最大同时在线 | 30 人 |
| 下单接口 | 串行执行（防止并发重复下单） |

---

## 5. WebSocket 推送事件

| 事件 | 说明 | 数据内容 |
|------|------|---------|
| `quote_update` | 行情推送 | { stock_code, price, change, change_pct } |
| `order_matched` | 成交推送 | { order_id, stock_code, price, volume, time } |
| `alert_trigger` | 事件触发 | { alert_id, stock_code, type, message } |

### WebSocket 心跳规则

| 规则 | 值 |
|------|---|
| 心跳间隔 | 10 秒 |
| 断开重连 | 最多 10 次，间隔 2 秒 |
| 推送延迟 | ≤200ms |
| 消息压缩 | 开启 |

---

## 6. 认证说明

除 `/api/v1/health` 和 `/api/v1/ws` 外，所有接口需要在 Header 中携带 JWT Token：

```
Authorization: Bearer <token>
```

JWT Payload 示例：
```json
{
  "user_id": 1,
  "username": "testuser",
  "exp": 1735689600,
  "iat": 1735603200,
  "iss": "zeroquant"
}
```

---

## 7. 后续计划

- v0.1 → v0.2：完善策略规则编辑器接口
- v0.2 → v1.0：对接数据服务 /data/ 接口，风控模块
