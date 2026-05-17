# ZeroQuant 2.0 — Step 5：Task Queue 机制说明

> 版本：V1.0
> 日期：2026-05-05
> 维护者：零号（PM）

---

## 一、什么是 Task Queue（任务队列）

零号通过文件分发任务卡，三端通过文件提交完成反馈，全程不需要实时聊天软件。

核心是 **NAS 共享目录 + 文件交换**，任何人只要能访问 `Z:\ZeroQuant2\tasks\` 就能参与。

---

## 二、目录结构

```
Z:\ZeroQuant2\tasks\
├── all/                       ← 三端共享的全局任务
│   ├── inbox/
│   └── done/
├── niuniu/                   ← 妞妞的任务目录
│   ├── inbox/                ← 零号发任务放这里
│   └── done/                 ← 妞妞完成的任务放这里
├── dacongming/               ← 大聪明的任务目录
│   ├── inbox/
│   └── done/
├── dataofficer/              ← 数据官的任务目录
│   ├── inbox/
│   └── done/
└── archive/                  ← 已完成任务归档（零号维护）
    └── YYYY-MM/             ← 按月归档
```

---

## 三、三端如何接收任务

**零号发任务 → 放到 `tasks\{终端}\inbox\` → 你来拿**

每次开工前检查一次 inbox：

```powershell
# 查看自己的任务 inbox
Get-ChildItem Z:\ZeroQuant2\tasks\niuniu\inbox\

# 或双击打开目录
explorer Z:\ZeroQuant2\tasks\niuniu\inbox\
```

**任务卡格式（零号发的文件）：**

```markdown
# 任务卡

**编号**：TASK-11
**任务名**：Day 1 接口约定输出
**类型**：开发任务
**截止时间**：2026-05-07 18:00
**前置依赖**：TASK-01~TASK-04 完成
**任务内容**：
1. 输出 Go API 草案（路由、响应格式、错误码）
2. 确认 WebSocket 方案
3. 输出《接口约定 v0.1》供 PM 评审

**交付物**：
- `src/backend/docs/api_draft_v0.1.md`
- 或放入 tasks/niuniu/done/TASK-11.md

**审核点**：格式符合规范，内容完整

**反馈要求**：
- 有问题 → 写 BlockCard 到 inbox，零号 10 分钟内响应
- 完成 → 写 DoneCard 到 done/TASK-11.md
```

---

## 四、三端如何提交完成

**方法 A：完成任务后，在 `done/` 目录放完成卡**

在 `Z:\ZeroQuant2\tasks\niuniu\done\` 新建 `TASK-11_Done.md`：

```markdown
# 完成卡

**编号**：TASK-11
**任务名**：Day 1 接口约定输出
**完成时间**：2026-05-07 17:30
**实际交付物**：
- `src/backend/docs/api_draft_v0.1.md`

**完成摘要**：
1. ✅ /api/v1/ 路由草案完成
2. ✅ WebSocket 方案确认（心跳 30s）
3. ✅ 接口约定文档已输出

**审核要求**：
请零号确认接口约定格式是否符合后端规范
```

**方法 B：把交付物文件路径写到完成卡，一起放进 done/**

---

## 五、遇到问题怎么办

**写阻塞卡，不要沉默**

在 `tasks\{终端}\inbox\` 新建 `TASK-XX_Block.md`：

```markdown
# 阻塞卡

**编号**：TASK-11
**阻塞原因**：
需要确认 WebSocket 认证方案 —— 目前后端和前端方案不一致

**需要的支持**：
请零号协调大聪明确认 WebSocket 认证方案

**预计阻塞时间**：等待 2 小时
```

零号每 10 分钟轮询一次 inbox，看到阻塞卡会立即处理。

---

## 六、验收流程

```
你提交完成卡
    ↓
零号审核（最长 24 小时）
    ↓
通过 → 归档到 archive/
不通过 → 零号写反馈卡到你的 inbox，说明问题
    ↓
你修改后重新提交
```

---

## 七、常用操作速查

| 操作 | 操作方法 |
|------|----------|
| 查看自己的任务 | `explorer Z:\ZeroQuant2\tasks\niuniu\inbox\` |
| 提交完成 | 在 `done/` 创建 `TASK-XX_Done.md` |
| 报阻塞 | 在 `inbox/` 创建 `TASK-XX_Block.md` |
| 查看已完成记录 | `explorer Z:\ZeroQuant2\tasks\niuniu\done\` |
| 查看归档 | `explorer Z:\ZeroQuant2\tasks\archive\` |

---

## 八、三端专属 inbox 路径

| 角色 | 任务 inbox 路径 |
|------|----------------|
| 妞妞 | `Z:\ZeroQuant2\tasks\niuniu\inbox\` |
| 大聪明 | `Z:\ZeroQuant2\tasks\dacongming\inbox\` |
| 数据官 | `Z:\ZeroQuant2\tasks\dataofficer\inbox\` |

> 建议：把对应 inbox 目录**添加到资源管理器左侧快速访问栏**，方便每天开工时一眼看到。

---

## 九、零号轮询规则

- **时间窗口**：每天 12:01 — 次日 00:01
- **轮询间隔**：每 10 分钟检查一次 inbox
- **响应 SLA**：
  - 阻塞卡：10 分钟内响应
  - 完成卡：24 小时内审核并反馈

---

> 无需安装任何软件。只要能访问 NAS，就能参与任务队列。
> 第一次开工前记住这 3 件事：
> 1. `explorer Z:\ZeroQuant2\tasks\{你的终端}\inbox\` — 看有没有新任务
> 2. 完成任务 → 写 DoneCard 到 `done/`
> 3. 遇到问题 → 写 BlockCard 到 `inbox/`，不要自己憋着