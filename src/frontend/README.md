# ZeroQuant 2.0 前端

> 量化交易系统前端 — Vue 3 + TypeScript + Vite

## 技术栈

| 类别 | 技术 |
|------|------|
| 框架 | Vue 3 + TypeScript |
| 构建 | Vite |
| PC UI | Element Plus |
| 移动端 UI | Vant 4（自动按需导入） |
| 状态管理 | Pinia + pinia-plugin-persistedstate |
| 路由 | Vue Router |
| 图表 | ECharts（按需引入） |
| HTTP | Axios（自动重试 + JWT） |
| 实时通信 | WebSocket（心跳 + 断线重连） |

## 快速开始

### 安装依赖

```bash
pnpm install
```

### 开发模式

```bash
pnpm dev
# 访问 http://localhost:5173（PC 端）
# 访问 http://localhost:5173/mobile/position（移动端）
# 局域网设备访问：http://<本机IP>:5173
```

### 生产构建

```bash
pnpm build
# 产物在 dist/ 目录
```

## 环境变量

| 文件 | 用途 |
|------|------|
| `.env` | 公共变量 |
| `.env.development` | 开发环境（默认 localhost） |
| `.env.production` | 生产环境（部署前替换域名） |

### 变量说明

```
VITE_API_BASE_URL   # Go 后端 API 前缀
VITE_DATA_BASE_URL  # Python 数据服务前缀
VITE_WS_URL         # WebSocket 地址
```

## 项目结构

```
src/
├── api/                  # API 封装
│   ├── backend.ts        # Go 后端接口（Axios + 重试）
│   └── data.ts           # 数据服务接口
├── components/           # 公共组件
│   ├── NotificationCenter.vue  # PC 端通知中心
│   ├── NotificationMobile.vue  # 移动端通知
│   └── WsStatusBar.vue         # WebSocket 断线提示条
├── composables/          # 组合式函数
│   ├── usePosition.ts    # 持仓数据管理
│   ├── useOrders.ts      # 订单数据管理
│   ├── useBacktest.ts    # 回测逻辑
│   └── useNetworkStatus.ts  # 网络状态检测
├── layouts/              # 布局
│   ├── MainLayout.vue    # PC 端布局（侧边栏 + 顶栏）
│   └── MobileLayout.vue  # 移动端布局（NavBar + Tabbar）
├── router/               # 路由（含权限守卫 + 懒加载）
├── stores/               # Pinia 状态管理
│   ├── auth.ts           # 认证状态
│   └── backtest.ts       # 回测结果持久化
├── utils/                # 工具
│   ├── websocket.ts      # WebSocket 客户端（心跳/重连/状态导出）
│   └── echarts.ts        # ECharts 按需引入
├── views/                # PC 端页面
│   ├── LoginView.vue     # 登录（记住账号）
│   ├── DashboardView.vue # 仪表盘
│   ├── StrategyView.vue  # 策略管理
│   ├── StrategyEditorView.vue  # 策略编辑器
│   ├── BacktestView.vue  # 回测分析（骨架屏）
│   ├── PositionView.vue  # 持仓（A 股红涨绿跌）
│   └── OrdersView.vue    # 订单
└── views/mobile/         # 移动端页面（Vant 4）
    ├── Position.vue      # 持仓（下拉刷新 + 详情弹窗）
    ├── Orders.vue        # 订单（Tabs 筛选 + 撤单）
    ├── Backtest.vue      # 回测（3 ECharts + 成交记录）
    └── Strategy.vue      # 策略列表
```

## 移动端

访问 `/mobile/` 路径进入移动端页面：

- **自动横屏检测**：横屏时提示切换竖屏
- **离线缓存**：断网时显示最后缓存数据，禁用下单操作
- **数据一致**：PC 端回测结果通过 Pinia persist 自动同步到移动端

## WebSocket 事件

| 事件 | 说明 |
|------|------|
| `order_update` | 订单状态更新（成交/拒绝） |
| `position_update` | 持仓数据实时更新 |
| `backtest_progress` | 回测进度（0-100%） |
| `backtest_result` | 回测完成结果 |
| `system_alert` | 系统告警 |

## 部署

### Nginx 部署步骤

1. **构建前端**
   ```bash
   pnpm build
   ```

2. **复制产物**
   ```bash
   # 将 dist/ 内容复制到部署目录
   xcopy /E /Y dist\ D:\ZeroQuant2\deploy\frontend\
   ```

3. **配置 Nginx**
   - 参考项目根目录 `../deploy/nginx.conf`
   - 修改 `server_name` 和 `root` 路径
   - 修改 proxy_pass 地址为实际后端地址

4. **启动 Nginx**
   ```bash
   nginx -c D:\ZeroQuant2\deploy\nginx.conf -t   # 测试配置
   nginx -c D:\ZeroQuant2\deploy\nginx.conf        # 启动
   ```

5. **验证**
   - PC 端：`http://your-domain`
   - 移动端：`http://your-domain/mobile/position`

### 关键配置说明

- **SPA 路由**：`try_files $uri $uri/ /index.html`
- **WebSocket 代理**：`proxy_pass` + `Upgrade` + `Connection "upgrade"`
- **静态缓存**：`/assets/` 目录 1 年缓存（Vite 构建带 hash）
- **gzip**：Nginx 原生 gzip + Vite 预压缩 `.gz` 文件

## 性能指标

| 指标 | 目标 |
|------|------|
| 首屏加载（PC） | ≤ 3 秒 |
| 首屏加载（移动端 4G） | ≤ 4 秒 |
| Lighthouse（PC） | ≥ 80 |
| Lighthouse（移动端） | ≥ 70 |
| 构建产物（gzip） | < 5MB |
| ECharts 包体积 | ~549KB（按需引入，gzip ~180KB） |

## 已知特性

- ✅ 路由懒加载（所有页面组件 `() => import()`）
- ✅ 请求自动重试（5xx / 网络错误，最多 1 次）
- ✅ WebSocket 断线自动重连（最多 10 次，指数退避）
- ✅ 断线全局提示条（WsStatusBar）
- ✅ 登录记住账号（localStorage）
- ✅ 回测骨架屏加载
- ✅ A 股盈亏颜色语义（红涨绿跌）
- ✅ 离线数据缓存（Pinia persist）
