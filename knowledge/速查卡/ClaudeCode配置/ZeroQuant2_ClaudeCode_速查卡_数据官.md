# ZeroQuant 2.0 — Claude Code 速查卡 · 数据官（Python 数据服务）

> 版本：V1.0 | 日期：2026-05-05 | 维护者：零号（PM）

---

## 你的角色

**Python 数据服务开发**，负责 ZeroQuant 2.0 数据采集、K线计算、行情推送、WebSocket 数据通道。

## 环境要求

| 工具 | 版本要求 |
|------|----------|
| Python | ≥ 3.10 |
| FastAPI | ≥ 0.104 |
| akshare | 最新 |
| Tushare Pro | 最新（如有账号） |
| Redis | ≥ 6 |
| PostgreSQL | ≥ 14 |

## 项目结构

```
src/data/
├──采集器/                # 行情数据采集（akshare/Tushare）
├──处理器/                # K线计算、指标、信号
├──通道/                  # WebSocket 推送服务
├──存储/                  # Redis + PG 写入
└──main.py
```

## API 路由规范

- 路径前缀：`/data/v1/`
- 健康检查：`/data/v1/health`
- 行情接口：`/data/v1/quote/{symbol}`
- WebSocket：`ws://[host]/data/v1/ws`

## 数据采集规范

- Level 2 盘口数据（五档 + 逐笔）
- 采集频率：tick级，按券商接口能力
- 缓存策略：Redis，TTL 按数据类型（行情 5s、K线 60s）

## 当前任务（TASK-12）

**输出 Python 数据接口草案**，包含：
1. `/data/v1/` 路由设计
2. 行情数据结构定义
3. WebSocket 数据格式
4. 与后端对齐 WebSocket 参数

产出文件：`tasks/dataofficer/done/TASK-12_数据接口草案_v1.0.md`

## 质量标准

- 采集脚本自测通过后再交付
- 跨端问题找 PM 协调