# 数据官（Python 数据服务）个人精细开发计划｜最终版
文档版本：V1.3
// ⚠️ ZERO.CC 修改：适用周期和开发周期改为相对时间
姓名：数据官
适用周期：以开发启动日为 Day 1，总周期约 18 天
参考文档：ZeroQuant 2.0 需求文档 v2.2、整体开发计划 V1.1、技术规范 v2.2
对齐里程碑：M1、M2、M4、M6

---

## 1. 基本信息
- 负责模块：Level‑2 盘口、行情、K线、采集器、数据接口、筛选引擎、缓存
- 开发周期：以开发启动日为 Day 1，总周期约 18 天
- 服务端口：8081
- 技术栈：Python 3.10+、FastAPI、Redis、PostgreSQL、akshare、Tushare Pro

---

## 2. 环境配置（必须 5.4 18:00 前完成）
- Python：3.10+
- 框架：FastAPI + Uvicorn
- 工具：PyCharm/VSCode、Pandas、APScheduler、Git
- 数据库：PostgreSQL、Redis

---

## 3. 目录规范（严格执行）
data/
├── app/
│   ├── api/
│   │   └── v1/              ← 接口必须带版本 /v1/
│   │       ├── health.py    ← 健康检查（Day1 实现）
│   │       ├── market.py   ← 行情接口
│   │       ├── kline.py    ← K 线接口
│   │       ├── orderbook.py # Level 2 盘口接口
│   │       ├── sector.py   ← 板块接口
│   │       └── filter.py   ← 筛选接口
│   ├── collectors/
│   │   ├── akshare_collector.py
│   │   ├── tushare_collector.py
│   │   └── level2_collector.py
│   ├── models/             ← SQLAlchemy 模型
│   ├── schemas/            ← Pydantic 模型
│   ├── services/           ← 业务逻辑
│   ├── schedulers/         ← 定时任务
│   ├── config.py
│   └── main.py
├── alembic/               ← 数据库迁移
├── requirements.txt
└── Dockerfile

> **数据库表前缀规范**：所有数据表统一加 `data_` 前缀（如 `data_quote`、`data_kline`），由 PostgreSQL `data` 库管理，与 Go 后端的 `biz_` 表（`biz` 库）完全隔离。

---

## 4. 每日任务 + 审核点
### Day 1（T1）｜M1 项目骨架就绪
- 搭建虚拟环境、初始化 FastAPI
- 连接 PostgreSQL + Redis
- 对接 akshare/Tushare/Level-2 数据源
- **实现 /data/v1/health 接口**（返回 {code:0, data:{status:"ok"}}）
- 输出 Python 数据服务接口草案（/data/v1/ 路径，含健康检查）
审核点：服务启动正常、数据源可拉取、健康检查接口返回 200

### Day 2–3（T1+1~T1+2）｜M2 数据服务可用
- Level‑2 五档盘口、逐笔采集
- 行情、K线、分时数据入库（data_quote、data_kline、data_orderbook、data_tick）
- 缓存、降级、补全机制
审核点：数据不丢、不重、实时性达标

### Day 4–5（T1+3~T1+4）｜支撑 M3 业务核心
- 多因子筛选接口（/data/v1/filter/`）
- 策略条件解析服务
- 与 Go 后端联调
审核点：筛选准确、响应 ≤200ms

### Day 6–7（T1+5~T1+6）｜支撑 M4 模拟闭环
- 回测历史数据供给（`/data/v1/kline/`）
- 多源自动切换、稳定性加固
审核点：数据稳定、无堆积

### Day 8–18（T1+7~T1+17）｜M6 正式发布
- 监控、指标计算、数据优化、备份
审核点：数据可靠、服务稳定

---

## 5. 核心接口（统一 /data/v1/ 前缀）
- GET /data/v1/health   ← Day1 必做，三端联调第一枪
- GET /data/v1/market/**    ← 实时行情
- GET /data/v1/kline/**     ← K 线数据
- GET /data/v1/orderbook/** ← Level 2 五档盘口
- GET /data/v1/sector/**    ← 板块数据
- POST /data/v1/filter/     ← 多因子筛选（策略调用）

> **路由版本规范**：所有接口必须带 `/v1/` 版本前缀，如 `/data/v1/health`、`/data/v1/filter/`，与前端/Go 后端保持一致。

---

## 6. 核心数据表（统一 data_ 前缀）
data_quote、data_kline、data_orderbook、data_tick、data_sector、data_indicator

> **前缀说明**：所有表统一加 `data_` 前缀，由 PostgreSQL `data` 库管理，与 Go 后端的 `biz_` 表（`biz` 库）隔离。

---

## 7. WebSocket 实时数据配置
- 心跳间隔：10 秒
- 断开自动重连：最多 10 次，间隔 2 秒
- 推送延迟：≤200ms
- 缓存策略：盘口数据 Redis 缓存 3 日、逐笔保存 90 天、K 线永久存储

---

## 8. 数据源容错规范
| 场景 | 处理方式 |
|------|---------|
| 主数据源断开 | 1 秒内自动切换备用源 |
| 全部源失效 | 返回错误，不返回脏数据 |
| 缺失数据 | 自动补全，无法补全则标记并触发告警 |

---

## 9. 开发规范
- 接口必须带版本 /v1/
- 数据必须落盘 + 缓存双写
- 采集任务异常自动重试
- 每日 23:50 提交 dev
- 提交格式：[数据官] 说明