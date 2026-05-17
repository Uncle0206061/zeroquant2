# ZeroQuant 2.0 — Claude Code 速查卡 · 妞妞（Go 后端）

> 版本：V1.0 | 日期：2026-05-05 | 维护者：零号（PM）

---

## 你的角色

**Go 后端开发**，负责 ZeroQuant 2.0 后端骨架、API 设计与实现、数据库建模。

## 环境要求

| 工具 | 版本要求 |
|------|----------|
| Go | ≥ 1.21 |
| PostgreSQL | ≥ 14 |
| Redis | ≥ 6 |
| Git | 最新 |

## 项目结构

```
src/backend/
├── internal/           # 业务逻辑（不要直接暴露）
├── handlers/           # HTTP handlers（Gin）
├── models/             # 数据模型
├── repository/          # 数据访问层
├── config/              # 配置加载
└── main.go
```

## API 路由规范

- 路径前缀：`/api/v1/`
- 响应格式：`{"code": 0, "msg": "success", "data": {...}}`
- 错误码：`400` 参数错误 / `401` 未认证 / `403` 禁止 / `500` 服务器错误

## WebSocket 规范

- 路径：`ws://[host]/api/v1/ws`
- 心跳：ping/pong，每 30s 一次
- 断线重连：指数退避，最大 5 次

## 数据库规范

- 表前缀：`zq_`
- 命名：snake_case（全小写 + 下划线）
- 迁移：用 Goose 或 golang-migrate

## 当前任务（TASK-11）

**输出 Go API 草案**，包含：
1. `/api/v1/` 下各路由设计（行情、用户、交易等）
2. 统一响应结构
3. 五类错误码定义
4. WebSocket 连接方案建议

产出文件：`tasks/niuniu/done/TASK-11_API草案_v1.0.md`

## 质量标准

- 代码自测通过后再交付
- 变更公共接口前先通知 PM 和其他端
- 跨端问题找 PM 协调，不要自己拍板