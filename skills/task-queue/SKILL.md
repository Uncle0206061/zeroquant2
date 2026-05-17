# Task Queue Skill — ZeroQuant 2.0 任务卡轮询与反馈

> 版本：V1.0 | 维护者：零号（PM） | 适用终端：niuniu（Go 后端）

## 触发场景

每次会话开始时，或每 10 分钟轮询时使用本 skill。

## 核心路径

| 用途 | 路径 |
|------|------|
| 任务收件箱 | `Z:\ZeroQuant2\tasks\niuniu\inbox\` |
| 任务完成箱 | `Z:\ZeroQuant2\tasks\niuniu\done\` |
| 已归档 | `Z:\ZeroQuant2\tasks\archive\` |

## 轮询流程

### Step 1：检查 inbox

```
Get-ChildItem "Z:\ZeroQuant2\tasks\niuniu\inbox\" -File
```

- 有新任务卡 → 读取内容 → 写入 TASKS.md → 开始执行
- 无新任务 → 回复 HEARTBEAT_OK

### Step 2：读取任务卡

每张任务卡字段：
- **任务编号**：TC-XXX-niuniu
- **截止时间**
- **前置依赖**（必须完成后才能开始）
- **任务内容**（具体步骤）
- **交付物清单**（所有 checkbox 项）
- **审核点**（逐项自检）

### Step 3：执行

按任务内容执行，交付物路径或内容写入任务卡。

### Step 4：提交完成

在 `done/` 写入完成反馈卡 DoneCard-XXX.md：

```markdown
# 完成反馈卡 DoneCard-{任务编号}
完成时间：YYYY-MM-DD HH:mm

## 任务编号
TC-XXX-niuniu

## 交付物确认
- [x] 交付物1：路径/说明
- [x] 交付物2：路径/说明

## 自检结果
[对照审核点逐项列出]

## 遗留问题
[如有]

## 耗时
X 小时
```

### Step 5：遇到问题

在 `inbox/` 写入阻塞卡 BlockCard-XXX.md：

```markdown
# 阻塞卡 BlockCard-{任务编号}
发现问题时间：YYYY-MM-DD HH:mm

## 任务编号
TC-XXX-niuniu

## 问题描述
[具体问题]

## 已尝试的解决方式
[列出]

## 需要 PM 支持的内容
[具体说明]
```

## 常用命令速查

| 操作 | 命令 |
|------|------|
| 查看 inbox | `explorer Z:\ZeroQuant2\tasks\niuniu\inbox\` |
| 查看 done | `explorer Z:\ZeroQuant2\tasks\niuniu\done\` |
| 提交完成 | 写入 `Z:\ZeroQuant2\tasks\niuniu\done\TC-XXX_Done.md` |
| 报阻塞 | 写入 `Z:\ZeroQuant2\tasks\niuniu\inbox\TC-XXX_Block.md` |

## 注意事项

- 阻塞卡必须立即写，不要自己憋着，PM SLA 10 分钟内响应
- 完成卡必须包含全部审核点的自检结果
- 交付物路径必须精确（文件名+行号或文件路径）
- 任务完成后从 inbox 移到 done/（即在 done/ 写入即可，inbox 原文件保留给 PM 归档）

## 参考文档

- references/task-card-format.md — 任务卡格式规范
- references/review-checklist.md — 审核检查清单
- references/feedback-template.md — 反馈模板
