# ZeroQuant2 前端代码审查报告

> **项目**：ZeroQuant 2.0 量化交易系统
> **审查范围**：D:\ZeroQuant2\src\frontend 全部代码
> **审查人**：Claude Code
> **审查日期**：2026-05-14
> **代码规模**：约 5,000+ 行 Vue/TypeScript，27 新建 + 14 修改 = 41 文件

---

## 一、总体评价

| 维度 | 评分 | 说明 |
|------|------|------|
| 功能完整性 | ⭐⭐⭐⭐ | 完成 7/7 任务卡，PC+移动端双端交付 |
| 代码质量 | ⭐⭐⭐ | 架构清晰，TypeScript 使用得当 |
| 安全性 | ⭐⭐ | 存在 token key 不一致等严重问题 |
| 最佳实践 | ⭐⭐⭐ | 整体符合 Vue 3 规范，少量重复代码 |

---

## 二、严重问题（必须修复）

### 2.1 🔴 Token Key 不一致 - 登录完全无法工作

**问题描述**：
多处 localStorage key 不一致，导致 WebSocket 连接无法携带有效 token，登录流程存在阻塞性 bug。

| 文件 | 行号 | localStorage key | 问题 |
|------|------|------------------|------|
| `api/backend.ts` | 31 | `zq_token` | ✅ 正确 |
| `utils/websocket.ts` | 34 | `token` | ❌ 错误 |
| `stores/auth.ts` | 15 | `zq_token` | ✅ 正确 |
| `router/index.ts` | 96 | `zq_token` | ✅ 正确 |

**影响**：用户登录后，WebSocket 连接使用的 key 与存储的 key 不匹配，导致 WS 无法认证。

**修复方案**：
```typescript
// utils/websocket.ts:34
// 修改前
const token = localStorage.getItem('token')

// 修改后
const token = localStorage.getItem('zq_token')
```

---

### 2.2 🔴 WebSocket 监听器未清理 - 内存泄漏

**问题描述**：
组件卸载时未调用 `ws.off()`，导致监听器累积和内存泄漏。

**受影响文件**：

| 文件 | 问题 | 行号 |
|------|------|------|
| `composables/useOrders.ts` | 无 off 调用 | 118-134 |
| `composables/usePosition.ts` | 无 off 调用 | 101-128 |
| `components/NotificationCenter.vue` | 清理 toast 但未清理 ws | 270-272 |
| `components/NotificationMobile.vue` | 无 onUnmounted | - |

**修复方案**：
```typescript
// composables/useOrders.ts 示例
import { ref, onMounted, onUnmounted } from 'vue'
import ws from '@/utils/websocket'

export function useOrders() {
  // ... 其他代码 ...

  // 保存回调引用
  const orderUpdateHandler = (data: any) => {
    const updated = normalizeOrder(data)
    const idx = allOrders.value.findIndex((o) => o.order_id === updated.order_id)
    if (idx >= 0) {
      allOrders.value[idx] = { ...allOrders.value[idx], ...updated }
    } else {
      allOrders.value.unshift(updated)
    }
  }

  onMounted(() => {
    fetchOrders()
    ws.on('order_update', orderUpdateHandler)
  })

  onUnmounted(() => {
    ws.off('order_update', orderUpdateHandler)  // 关键修复
  })

  // ... 其他代码 ...
}
```

---

### 2.3 🔴 mobile/Backtest.vue 重复添加监听器

**位置**：`src/views/mobile/Backtest.vue:215-221`

**问题描述**：
`runBacktest()` 函数每次调用都添加新的监听器，不清理旧监听器。连续点击"发起回测"会导致事件被触发 N 次。

**当前代码**：
```typescript
async function runBacktest() {
  // ...
  ws.on('backtest_progress', (data: any) => {
    if (data?.task_id === taskId) progress.value = data.progress || 0
  })

  ws.on('backtest_result', (data: any) => {
    if (data?.task_id === taskId) handleResult(data)
  })
  // ...
}
```

**修复方案**：
```typescript
// 保存回调引用
const progressHandler = (data: any) => {
  if (data?.task_id === taskId) progress.value = data.progress || 0
}
const resultHandler = (data: any) => {
  if (data?.task_id === taskId) handleResult(data)
}

ws.on('backtest_progress', progressHandler)
ws.on('backtest_result', resultHandler)
```

或在函数开始时先移除：
```typescript
ws.off('backtest_progress', progressHandler)
ws.off('backtest_result', resultHandler)
```

---

### 2.4 🔴 路由未定义但被引用

**问题描述**：
`MobileLayout.vue:78` 引用 `/mobile/strategy/editor` 路由，但 `router/index.ts` 中未定义此路由。

**当前引用**：
```typescript
// MobileLayout.vue
const map = {
  '/mobile/strategy/editor': '策略编辑',  // 路由不存在
  // ...
}
```

**修复方案**：
在 `router/index.ts` 中添加路由：
```typescript
{
  path: 'strategy/editor',
  name: 'MobileStrategyEditor',
  component: () => import('@/views/mobile/StrategyEditor.vue'),
  meta: { title: '策略编辑' },
}
```

---

## 三、高优先级问题

### 3.1 ⚠️ 类型安全问题 - 过多 `any`

**问题描述**：
多处使用 `any` 类型，降低了 TypeScript 的类型安全保障。

**受影响位置**：

| 文件 | 行号 | 代码 |
|------|------|------|
| `useBacktest.ts` | 135 | `const taskId = res?.task_id ... as any` |
| `useOrders.ts` | 63 | `const res: any = await backend.get(...)` |
| `usePosition.ts` | 48, 51 | 同上 |
| `mobile/Strategy.vue` | 63 | 同上 |
| `stores/backtest.ts` | 5-6 | `lastResult: null as any` |

**修复方案**：
定义统一 API 响应类型：

```typescript
// types/api.ts
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export interface BacktestResponse {
  task_id: string
  status: 'pending' | 'completed'
  annual_return?: number
  max_drawdown?: number
  sharpe_ratio?: number
  win_rate?: number
  // ...
}
```

---

### 3.2 ⚠️ useNetworkStatus 未清理监听

**位置**：`composables/useNetworkStatus.ts:10-18`

**问题描述**：
`addEventListener` 后没有对应的 `removeEventListener`。

**当前代码**：
```typescript
onMounted(() => {
  window.addEventListener('online', update)
  window.addEventListener('offline', update)
})

onUnmounted(() => {
  // 缺少 removeEventListener
})
```

**修复方案**：
```typescript
import { ref, onMounted, onUnmounted } from 'vue'

export function useNetworkStatus() {
  const isOnline = ref(navigator.onLine)

  function update() {
    isOnline.value = navigator.onLine
  }

  onMounted(() => {
    window.addEventListener('online', update)
    window.addEventListener('offline', update)
  })

  onUnmounted(() => {
    window.removeEventListener('online', update)
    window.removeEventListener('offline', update)
  })

  return { isOnline }
}
```

---

### 3.3 ⚠️ API 错误处理不一致

| 文件 | 错误处理方式 |
|------|-------------|
| `api/backend.ts` | ✅ 友好错误消息映射 + 自动重试 |
| `api/data.ts` | ❌ 直接 `Promise.reject(error)` |

**修复方案**：
统一 `data.ts` 的错误处理逻辑，与 `backend.ts` 保持一致。

---

## 四、中优先级问题

### 4.1 ⚡ 代码重复 - Backtest 结果处理

**问题描述**：
`mobile/Backtest.vue` 的 `handleResult` 函数与 `useBacktest.ts` 的 `handleResult` 高度重复（约 15 行完全相同）。

**mobile/Backtest.vue:243-259**：
```typescript
function handleResult(data: any) {
  result.value = {
    metrics: [
      { label: '年化收益', value: `${(data.annual_return * 100).toFixed(2)}%`, ... },
      // ... 多行重复
    ],
    equity_curve: data.equity_curve || [],
    // ...
  }
}
```

**修复方案**：
将通用逻辑提取到 `composables/useBacktest.ts` 中导出，移动端复用。

---

### 4.2 ⚡ PC/移动端涨跌颜色逻辑相反

**问题描述**：
A股是红涨绿跌，但移动端颜色配置反了。

| 端 | 涨（红色） | 跌（绿色） |
|---|-----------|-----------|
| PC | `#F56C6C` ✅ | `#67C23A` ✅ |
| Mobile | `#ee0a24` ✅ | `#07c160` ❌ 应该是跌 |

**当前 mobile 错误代码**：
```css
/* mobile/Position.vue:301-302 */
.pnl-up { color: #ee0a24; }   /* 涨 */
.pnl-down { color: #07c160; } /* 跌 */
```

**修复方案**：
A股红涨绿跌，移动端应统一为：
```css
.pnl-up { color: #ee0a24; }   /* 涨 - 红色 */
.pnl-down { color: #07c160; } /* 跌 - 绿色 */
```

---

### 4.3 ⚡ 多处 CSS 样式重复定义

**问题描述**：
相同功能代码在多处重复定义。

| 重复项 | 位置 |
|--------|------|
| `formatMoney()` | `PositionView.vue:107`, `mobile/Position.vue:104` |
| ECharts grid/tooltip | `useBacktest.ts`, `mobile/Backtest.vue` |
| 涨跌颜色 CSS | PC 和 Mobile 各定义一套 |

**修复方案**：
提取为工具函数和共享样式：
```typescript
// utils/format.ts
export function formatMoney(v: number): string {
  return v >= 0 ? `¥${v.toFixed(2)}` : `-¥${Math.abs(v).toFixed(2)}`
}
```

---

### 4.4 ⚡ NotificationMobile.vue 导入位置不规范

**位置**：`components/NotificationMobile.vue:94`

**问题描述**：
`import { onMounted }` 在 `<script>` 中间而非顶部。

**修复方案**：
将所有 import 移到脚本顶部。

---

### 4.5 ⚡ 占位页面无实际功能

**问题描述**：
部分页面仅有 UI 无数据获取逻辑。

| 文件 | 问题 |
|------|------|
| `views/DashboardView.vue` | 仅显示 "--"，无数据获取 |
| `views/StrategyView.vue` | 表格数据为空 |
| `components/HelloWorld.vue` | Vite 模板遗留代码，应删除 |

**修复方案**：
- 删除 `HelloWorld.vue`
- 补充 Dashboard 和 Strategy 列表的数据获取逻辑

---

## 五、低优先级问题

### 5.1 `.env.production` 未配置实际域名

**问题**：`your-domain.com` 占位符未替换，生产部署会失败。

---

### 5.2 `vite.config.ts:45` - `host: true` 安全隐患

**问题**：
开发环境 `host: true` 允许局域网访问，注释说是"真机测试用"。

**建议**：
通过环境变量控制，仅开发环境开启。

```typescript
server: {
  host: process.env.NODE_ENV === 'development',
  // ...
}
```

---

### 5.3 WebSocket 默认 URL 硬编码

**位置**：`utils/websocket.ts:164`

**问题**：`ws://localhost:8080` 硬编码

**建议**：通过环境变量或配置注入。

---

### 5.4 部分变量命名不清晰

| 文件 | 位置 | 问题 |
|------|------|------|
| `useStrategyEditor.ts` | 79 | `_data` 参数名不清晰 |
| `useBacktest.ts` | 63, 67 | 混用 `res` 和 `res?.data` |

---

## 六、安全问题汇总

| 问题 | 严重性 | 位置 | 建议 |
|------|--------|------|------|
| Token key 不一致 | 🔴 严重 | 多文件 | 统一使用 `zq_token` |
| localStorage 未加密存储 token | ⚠️ 中等 | auth store | 考虑加密或使用 HttpOnly Cookie |
| 生产环境未禁用 API 调试 | ⚠️ 中等 | vite.config.ts | 添加环境检查 |
| API 响应未校验结构 | ⚠️ 中等 | 多处 | 添加运行时类型校验 |

---

## 七、建议改进项

### 7.1 立即可做（阻塞性问题）
- [ ] 统一 localStorage token key 为 `zq_token`
- [ ] 清理所有 WebSocket 监听器
- [ ] 修复 mobile/Backtest 重复监听问题
- [ ] 添加缺失的 `/mobile/strategy/editor` 路由

### 7.2 短期优化
- [ ] 定义统一 API 类型，减少 `any` 使用
- [ ] 提取重复代码到 composable
- [ ] 统一 PC/移动端涨跌颜色为 A 股标准
- [ ] 修复 useNetworkStatus 监听器清理
- [ ] 统一 API 错误处理逻辑

### 7.3 中长期建议
- [ ] 删除 `HelloWorld.vue` 遗留文件
- [ ] 补充 Dashboard 和 Strategy 列表数据获取
- [ ] 添加单元测试（composable 逻辑）
- [ ] 添加错误边界（errorCaptured）
- [ ] 补充 E2E 测试（Playwright）
- [ ] 建立 CI/CD 流程

---

## 八、审查总结

大聪明本次交付的代码**整体质量良好**，架构设计合理，功能覆盖完整。代码风格统一，注释清晰，体现了扎实的前端开发能力。

但存在几个**阻塞性问题**必须立即修复：

1. **Token key 不一致** → 登录后 WebSocket 无法携带 token
2. **监听器泄漏** → 长期使用会导致性能下降
3. **mobile/Backtest 重复监听** → 连续操作会触发多次

修复上述问题后，代码质量可达到生产就绪标准。

---

## 九、文件清单

| 文件路径 | 说明 |
|----------|------|
| `src/api/backend.ts` | Go 后端 API 封装 |
| `src/api/data.ts` | Python 数据服务 API |
| `src/api/index.ts` | API 统一导出 |
| `src/composables/useBacktest.ts` | 回测逻辑 composable |
| `src/composables/usePosition.ts` | 持仓数据 composable |
| `src/composables/useOrders.ts` | 订单数据 composable |
| `src/composables/useStrategyEditor.ts` | 策略编辑器逻辑 |
| `src/composables/useNetworkStatus.ts` | 网络状态检测 |
| `src/stores/auth.ts` | 认证状态 |
| `src/stores/backtest.ts` | 回测结果持久化 |
| `src/utils/websocket.ts` | WebSocket 封装 |
| `src/utils/echarts.ts` | ECharts 按需引入 |
| `src/components/NotificationCenter.vue` | PC 通知中心 |
| `src/components/NotificationMobile.vue` | 移动端通知 |
| `src/components/WsStatusBar.vue` | WS 状态栏 |
| `src/layouts/MainLayout.vue` | PC 主布局 |
| `src/layouts/MobileLayout.vue` | 移动端布局 |
| `src/views/*.vue` | 各页面组件 |
| `src/views/mobile/*.vue` | 移动端页面 |
| `src/data/factors.ts` | 因子库数据 |
| `src/router/index.ts` | 路由配置 |
| `vite.config.ts` | Vite 构建配置 |
| `package.json` | 项目依赖 |
| `main.ts` | 应用入口 |

---

*报告生成时间：2026-05-14*
*生成工具：Claude Code*
