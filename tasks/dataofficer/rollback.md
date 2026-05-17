# TC-M1-03 变更回滚记录
> 执行人: 数据官 | 日期: 2026-05-14 | 任务: 数据服务骨架初始化

## 变更清单

### 新建文件 (28 个)

```
src/data/
├── requirements.txt
├── .env.example
├── Dockerfile
├── app/
│   ├── __init__.py
│   ├── main.py              ← 可运行 FastAPI 入口
│   ├── config.py            ← 配置中心
│   ├── api/
│   │   ├── __init__.py
│   │   └── v1/
│   │       ├── __init__.py
│   │       ├── health.py    ← ✅ 已实现
│   │       ├── market.py    ← 🔄 可调用
│   │       ├── kline.py     ← 🔄 可调用
│   │       ├── orderbook.py ← ⏳ 占位
│   │       ├── sector.py    ← 🔄 可调用
│   │       └── filter.py    ← ⏳ 占位
│   ├── collectors/
│   │   ├── __init__.py
│   │   ├── akshare_collector.py  ← ✅ 完整实现
│   │   ├── tushare_collector.py  ← 🔄 框架
│   │   └── level2_collector.py   ← 🔄 框架
│   ├── models/
│   │   ├── __init__.py
│   │   └── base.py          ← SQLAlchemy 引擎 + data_ 前缀
│   ├── schemas/
│   │   ├── __init__.py
│   │   └── base.py          ← 统一响应模型
│   ├── services/
│   │   ├── __init__.py
│   │   └── placeholder.py
│   └── schedulers/
│       ├── __init__.py
│       └── placeholder.py
├── alembic/
│   └── env.py
└── tests/
    ├── __init__.py
    └── test_health.py       ← 3 个测试用例

docs/
└── data-api-draft-v0.1.md   ← 接口草案文档
```

## 回滚步骤

如需回退到变更前状态（空目录）:

```powershell
# 删除所有新建文件，保留目录结构
Remove-Item 'src/data/*' -Recurse -Force
# 恢复为原始空骨架
```

## 变更前后对比

| 项目 | 变更前 | 变更后 |
|------|--------|--------|
| app/api/v1/ 文件数 | 0 | 6 (health+5 路由) |
| app/collectors/ 文件数 | 0 | 3 (akshare/tushare/level2) |
| 可运行服务 | ❌ | ✅ uvicorn app.main:app |
| 健康检查 | ❌ | ✅ GET /data/v1/health |
| akshare 采集 | ❌ | ✅ K线/行情/板块/资金流向 |
| API 文档 | ❌ | ✅ /data/docs (Swagger) |

## 依赖说明

```bash
pip install -r src/data/requirements.txt
```

核心依赖: fastapi, uvicorn, sqlalchemy, redis, pandas, apscheduler, akshare
## 2026-05-14 本地测试修复记录

### 修复 1: filter.py → filter_routes.py
- 原因: `filter` 是 Python 内置名，导致 `from app.api.v1 import filter` 失败
- 影响: main.py 启动报 ImportError
- 修复: 重命名为 filter_routes.py

### 修复 2: main.py 补充根路径端点
- 原因: GET `/` 返回 404
- 修复: 添加 `@app.get("/")` 返回服务信息

### 修复 3: on_event → lifespan
- 原因: FastAPI `on_event` 已弃用 (DeprecationWarning)
- 修复: 改用 `@asynccontextmanager lifespan`

### 测试结果 (2026-05-14 14:xx)
- test_health.py: 3/3 PASSED
- test_integration.py: 8/8 PASSED
  - health ✓  kline(日K) ✓  kline(周K) ✓  无效代码 ✓
  - sector ✓  orderbook ✓  filter ✓  routes ✓
- akshare 采集器: K线数据可正常拉取 (sandbox 网络限制仅影响实时接口)

### 环境
- Python 3.10.11
- fastapi 0.115.0, uvicorn 0.30.6, akshare 1.18.60
- pytest 9.0.3
- 测试目录: D:\TEST\src\data


## TC-M2-03 变更记录 2026-05-14

### 新增文件 (5)

```
src/data/app/services/
├── factor_registry.py   ← 8种因子注册表
├── indicators.py        ← MA/MACD/RSI 技术指标计算
└── filter_engine.py     ← AND/OR 多因子筛选引擎核心

src/data/app/api/v1/
└── filter_routes.py     ← 重写: POST /filter + /filter/parse

src/data/tests/
└── test_filter.py       ← 16个测试用例
```

### 修改文件 (3)
- app/main.py: 改用 lifespan 替代 on_event, 添加 root 端点
- tests/test_integration.py: 适配新 filter 接口, 网络容错
- tests/test_health.py: 更新为 3 个测试用例

### 删除文件 (1)
- app/services/placeholder.py → 被 factor_registry/indicators/filter_engine 替代

### 修复的 Bug
1. filter.py→filter_routes.py (Python 内置名冲突)
2. main.py GET / 返回 404
3. on_event 弃用警告 → lifespan
4. result['data'] None 时 AttributeError

### 测试结果 (2026-05-14)
- 总计: 19 passed, 9 skipped
- 跳过原因: sandbox 网络阻断 akshare 外部 API (非代码问题)
- 因子注册表完整性: 8/8 ✅
- 策略解析: 5/5 ✅
- 技术指标计算 (RSI/MACD): 2/2 ✅
- 边界场景: 4/4 ✅

### 回滚方法
如需回退筛选引擎:
  Remove-Item src/data/app/services/factor_registry.py
  Remove-Item src/data/app/services/indicators.py
  Remove-Item src/data/app/services/filter_engine.py
  Remove-Item src/data/tests/test_filter.py
  # filter_routes.py 恢复为占位版本

### 接口清单
POST /data/v1/filter       — 多因子筛选执行
POST /data/v1/filter/parse — 策略JSON解析

## TC-M2-01 补全记录 2026-05-14

### 新增文件 (5)
- app/models/models.py        ← 4个数据表: data_kline/quote/tick/sector
- app/cache.py                ← Redis 缓存管理器（优雅降级）
- app/api/v1/monitor.py       ← GET /monitor/status 健康监控
- tests/test_tc_m2_01.py      ← 13个测试用例
- app/collectors/level2_collector.py ← 重写为完整 WebSocket 实现

### 修改文件 (5)
- app/collectors/akshare_collector.py  ← 增加缓存读写（双写）
- app/models/base.py                   ← 延迟创建引擎（懒加载）
- app/main.py                          ← 注册 monitor 路由
- tests/test_integration.py            ← 适配新接口
- tests/test_health.py                 ← 3个测试用例

### 修复的 Bug
1. psycopg2 未安装 → 安装 psycopg2-binary
2. engine 在 import 时创建导致模块加载失败 → 改为延迟创建
3. monitor 端点引用 engine 报错 → 改用 get_engine()

### 测试结果
- 32 passed / 9 skipped (网络) / 0 failed
- 缓存读写: 5/5 ✅
- 监控端点: 3/3 ✅
- Level-2健康: 2/2 ✅
- 模型验证: 2/2 ✅
- 路由注册: 1/1 ✅

### 接口变更
+ GET  /data/v1/monitor/status  ← 数据源/Redis/PG/采集器健康监控

### 回滚方法
删除新增的5个文件，恢复5个被修改文件的旧版本。

## TC-M2-02 变更记录 2026-05-16

### 新增文件 (2)
- app/services/backup.py     ← CSV备份 + manifest.json + MD5校验 + 30天清理
- app/services/recovery.py   ← K线缺口检测 + 自动补全
- app/api/v1/admin.py        ← 备份/恢复/验证 API端点
- tests/test_tc_m2_02.py     ← 17个测试用例

### 重写文件 (2)
- app/api/v1/monitor.py      ← AlertManager(分级告警+WS广播) + metrics/alerts端点
- app/main.py                ← 注册 admin 路由

### 新增端点 (8个)
GET  /monitor/status         数据源/Redis/PG/采集器健康
GET  /monitor/metrics        K线完整性/可用率/吞吐量
GET  /monitor/alerts         告警历史
WS   /monitor/ws             告警实时推送
POST /admin/backup           手动备份
GET  /admin/backup/verify    验证备份
GET  /admin/backup/list      备份列表
POST /admin/recover          数据恢复
GET  /admin/recovery/log     恢复日志

### 告警策略
| critical | 全部源失效 >5min | WebSocket |
| warning  | 单源失效 >30s    | 日志+alert |
| info     | 数据源恢复       | 日志 |

### 测试: 57 passed / 9 skipped / 0 failed

## TC-M2-02 2026-05-16
- app/services/backup.py (CSV+MD5+30d cleanup)
- app/services/recovery.py (gap detect+auto fill)
- app/api/v1/admin.py (8 backup/recovery endpoints)
- app/api/v1/monitor.py (AlertManager+WS+metrics+alerts)
- tests/test_tc_m2_02.py (17 tests)
- 57 passed/9 skipped/0 failed

