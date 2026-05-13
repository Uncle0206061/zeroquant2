# ZeroQuant 2.0

> A股量化辅助决策平台
> 启动日期：2026-05-02

---

## 项目定位

面向小圈子的 A 股量化辅助决策平台，支持行情查看、自选管理、模拟交易、简单回测。

## 技术栈

- **后端**：Go + Gin + PostgreSQL + Redis
- **前端**：Vue 3 + Element Plus + TradingView
- **数据服务**：Python + FastAPI + akshare

## 团队分工

| 角色 | 职责 |
|------|------|
| 用户 | 项目总负责人，编码 |
| 零号 | PM / 架构师 / 需求 / 终审 / 协调 |
| 妞妞 | Go 后端开发 |
| 大聪明 | Vue3 前端开发 |
| 数据官 | Python 数据服务 |

## 文档

- [工作手册](docs/ZeroQuant工作手册.md) — 开发流程、通信机制、轮询规则
- [技术规范](docs/ZeroQuant技术规范.md) — 编码规范、API 规范、用词统一
- [项目记忆](docs/ZeroQuant项目记忆.md) — 技术栈、MVP 功能、历史记录

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/Uncle0206061/zeroquant2.git

# 后端启动
cd src/backend
go run cmd/server/main.go

# 前端启动
cd src/frontend
npm install && npm run dev

# 数据服务启动
cd src/data
pip install -r requirements.txt
python app/main.py
```

## 目录结构

```
ZeroQuant2/
├── docs/           # 文档
├── src/            # 源代码
│   ├── backend/    # Go 后端
│   ├── frontend/   # Vue3 前端
│   └── data/       # Python 数据服务
├── tasks/          # 任务卡
├── knowledge/      # 知识库
├── skills/         # 技能库
└── backups/        # 备份
```

## 许可证

私有项目，仅供授权成员使用。
