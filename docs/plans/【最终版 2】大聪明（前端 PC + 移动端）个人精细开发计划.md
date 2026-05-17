# 大聪明（前端 PC+移动端）个人精细开发计划｜最终版
文档版本：V1.3
// ⚠️ ZERO.CC 修改：适用周期和开发周期改为相对时间
姓名：大聪明
适用周期：以开发启动日为 Day 1，总周期约 18 天
参考文档：ZeroQuant 2.0 需求文档 v2.2、整体开发计划 V1.1、技术规范 v2.2
对齐里程碑：M1、M4、M6

---

## 1. 基本信息
- 负责模块：PC端、移动端、策略编辑器、回测、持仓、提醒、WebSocket
- 开发周期：以开发启动日为 Day 1，总周期约 18 天
- 服务端口：5173
- 技术栈：Vue3 + Vite + Element Plus + Vant + ECharts

---

## 2. 环境配置（必须 T1 当日 18:00 前完成）
- Node：v20+
- IDE：VSCode
- 插件：Vue Devtools、ESLint、Prettier
- 依赖：Axios、Pinia、VueRouter、WebSocket、Element Plus、Vant、ECharts

---

## 3. 目录规范（严格执行，与技术规范 v2.2 完全对齐）
frontend/
├── public/
├── src/
│   ├── api/                ← API 请求封装（统一封装 Go 后端 /api/v1/ 和数据服务 /data/v1/）
│   ├── assets/             ← 静态资源（图片、图标、音效）
│   ├── components/         ← 公共组件（PC + 移动端共用）
│   ├── composables/        ← 组合式函数（逻辑复用）
│   ├── layouts/            ← 布局组件（PC 管理后台布局、移动端布局）
│   ├── router/             ← 路由配置（含权限守卫）
│   ├── stores/             ← Pinia 状态管理
│   ├── utils/              ← 工具函数（格式化、时间、加密等）
│   ├── views/              ← 页面组件（PC + 移动端）
│   ├── App.vue
│   └── main.ts
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
└── Dockerfile

---

## 4. 每日任务 + 审核点
### Day 1（T1）｜M1 项目骨架就绪
- 初始化 Vue3 + Vite 项目
- 配置路由（VueRouter）、状态管理（Pinia）
- API 请求统一封装（Axios，拦截器处理 JWT）
- WebSocket 全局封装（含心跳、重连、推送回调）
- 确认前端 API 统一封装方案（与妞妞/数据官对齐 /api/v1/、/data/v1/ 前缀）
- **WebSocket 对齐（Day1 三端共做）：**
  - 心跳间隔：10 秒
  - 断开自动重连：最多 10 次，间隔 2 秒
  - 推送延迟要求：≤200ms
审核点：项目启动正常、路由跳转正常、请求正常、WebSocket 可连接

### Day 2（T1+1）
- 登录（POST /api/v1/auth/login）、注册（POST /api/v1/auth/register）
- 路由守卫（未登录跳转登录页）
审核点：登录跳转正常、JWT 自动附加到所有请求

### Day 3–4（T1+2~T1+3）
- 策略编辑器（多因子、条件组合，表单勾选 + 填写形态）
- 回测页面（发起回测请求 GET /data/v1/kline/、展示结果）
审核点：可编辑、可提交策略、回测结果展示正确

### Day 5–6（T1+4~T1+5）
- 持仓（GET /api/v1/position）、订单（GET /api/v1/order/list）、成交、盈亏展示
- 消息提醒：弹窗 + 声音 + 角标（保留 24 小时）
- WebSocket 实时接收持仓/订单更新
审核点：数据展示正确、推送实时（≤200ms）

### Day 7（T1+6）｜M4 模拟闭环上线
- 移动端 Vant 适配（仅竖屏，离线只看不操作）
- 前后端联调（PC + 移动端功能一致）
审核点：双端功能一致、无布局错乱、无阻塞 BUG

### Day 8–18（T1+7~T1+17）｜M6 正式发布
- 体验优化、ECharts 图表增强、BUG 修复、打包部署
审核点：稳定、流畅、无报错

---

## 5. 核心页面
- 登录 / 注册（/login、/register）
- 策略编辑（/strategy/edit）
- 回测报告（/backtest）
- 持仓（/position）
- 订单（/order）
- 成交（/trade）
- 消息中心（/notifications）
- 移动端适配版（所有页面移动端版本，带 /m/ 前缀路由）

---

## 6. WebSocket 配置（技术规范强制，与三端对齐）
- 心跳间隔：10 秒
- 断开自动重连：最多 10 次，间隔 2 秒
- 推送延迟：≤200ms
- 全局唯一实例，跨页面复用
- 重连后自动订阅之前订阅的频道

---

## 7. 开发规范
- 统一 ESLint 格式
- API 统一放在 src/api/，按后端服务分组（backend.ts、data.ts）
- WebSocket 全局唯一实例（src/utils/websocket.ts）
- 路由守卫控制权限（未登录强制跳转 /login）
- 每日 23:50 提交 dev
- 提交格式：[大聪明] 说明