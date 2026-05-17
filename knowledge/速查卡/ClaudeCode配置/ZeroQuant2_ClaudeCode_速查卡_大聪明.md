# ZeroQuant 2.0 — Claude Code 速查卡 · 大聪明（Vue3 前端）

> 版本：V1.0 | 日期：2026-05-05 | 维护者：零号（PM）

---

## 你的角色

**Vue3 前端开发**，负责 ZeroQuant 2.0 前端界面、组件开发、API 接入与用户体验。

## 环境要求

| 工具 | 版本要求 |
|------|----------|
| Node.js | ≥ 18 |
| pnpm | ≥ 8 |
| Vue3 | 3.x |
| Element Plus | 2.x |
| TypeScript | 5.x |
| Vite | 5.x |

## 项目结构

```
src/frontend/
├── src/
│   ├── api/              # API 封装（统一 Axios 实例）
│   ├── components/       # 通用组件
│   ├── views/            # 页面
│   ├── stores/           # Pinia 状态管理
│   ├── router/           # Vue Router
│   └── styles/            # 全局样式
├── public/
└── package.json
```

## API 封装规范

- 使用 Axios，统一 baseURL
- 请求拦截器：自动附加 token
- 响应拦截器：统一错误处理（code != 0 显示错误信息）
- WebSocket：封装连接管理，断线自动重连

## 前端路由规范

- 路径：`/app/` 前缀（后端代理）
- 路由表放 `router/index.ts`

## 当前任务（TASK-13）

**确认前端 API 封装方案**，包含：
1. Axios 实例封装（拦截器设计）
2. WebSocket 前端接入方案
3. 与后端确认 `/api/v1/` 路由是否匹配

产出文件：`tasks/dacongming/done/TASK-13_前端API封装方案_v1.0.md`

## 质量标准

- 组件自测通过后再交付
- 样式规范参考 Element Plus 官方
- 跨端问题找 PM 协调