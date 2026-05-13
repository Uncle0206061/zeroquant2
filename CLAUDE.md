# ZeroQuant2 使用指南

## 启动流程（必须按顺序）

1. 加载记忆 → MEMORY.md → 当前任务卡
2. 检查任务 → `tasks/dataofficer/inbox/` 有无未处理任务卡
3. 搜索信息 → 先搜 `D:\Obsidian\零号记忆\` → 本地 `docs/`

## 项目定位

- **定位:** A股量化辅助决策平台，B/S架构
- **技术栈:** Go+Gin / Vue3+ElementPlus / Python+FastAPI+akshare
- **存储:** NAS `\\100.65.205.77\home`

## 团队角色

| 角色 | 职责 |
|---|---|
| 用户 | 总负责 + Claude Code 编码 |
| 零号 | PM / 架构师 / 需求 / 终审 |
| 零号(暂代) | Python数据服务（数据官未上岗） |
| 妞妞 | Go 后端 + 初审 |
| 大聪明 | Vue3 前端 + 初审 |

## 任务卡工作机制

**流程:** 读任务卡 → 写详细计划 → 用户审查 → 执行 → 汇报 → 备份 → 下一张

**目录:**
- `tasks/dataofficer/inbox/` — 收任务卡
- `tasks/dataofficer/done/` — 反馈卡

**工作时间:** 12:01—00:01，轮询每10分钟

## Skills 调用原则

### 三层架构

| 层级 | 类型 | 数量限制 |
|------|------|----------|
| 日常层 | 环境、工具调用类 | ≤10 |
| 项目层 | 具体代码工作 | 按需 |
| 按需层 | 不常用 | MCP 方式 |

### 日常层（10个）

1. git-workflow — Git 操作
2. coding-standards — 代码规范
3. code-tour — 代码导览
4. codebase-onboarding — 项目上手
5. build-tool-factory — 工具构建
6. creating-skills — 技能创建
7. error-handling — 错误处理
8. observability — 日志监控
9. update-config — 配置管理
10. writing-plans — 任务规划

### 项目层（代码审查用）

本项目（Go 后端）代码审查时加载：

| Skill | 用途 | 优先级 |
|-------|------|--------|
| golang-patterns | Go 代码规范审查 | ⭐⭐⭐ |
| golang-testing | Go 测试检查 | ⭐⭐⭐ |
| security-review | JWT/认证/中间件安全审查 | ⭐⭐⭐ |
| backend-patterns | API 设计/中间件/响应格式审查 | ⭐⭐ |
| code-tour | 代码导览辅助 | ⭐⭐ |

**调用方式：** 本地直接调用（非 MCP）

## 核心工作原则

1. **先审后做** — 生成内容先让用户看，确认后再执行
2. **文件存放** — 生成文件 → `桌面\给用户\`
3. **验证后再汇报** — 做完多验证几遍
4. **不确定时先反馈** — 如实说明，不莽撞
5. **删除前确认** — 信息已迁移 Obsidian
6. **代码质量** — Go 用 golang-patterns，Python 用 python-patterns

## MCP 配置

| 状态 | MCP |
|------|-----|
| 已有 | github, memory, sequential-thinking, obsidian, scrapling |

## 当前开发阶段

- Phase 1：骨架 + DB + 登录 + 健康检查（进行中）
- Phase 2：数据采集 + 行情 + 自选股
- Phase 3：模拟交易 + 回测 + 收益

## 相关记忆

- 项目记忆: `C:\Users\user\.claude\projects\D--ZeroQuant2\MEMORY.md`
- 任务卡: `tasks/dataofficer/inbox/`
- 文档: `docs/ZeroQuant工作手册.md`