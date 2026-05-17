# ZeroQuant2 开发日志

## 2026-05-14

### 任务：大聪明前端代码审查与修复

#### 一、代码审查

审查了大聪明（Vue3+ElementPlus）的前端代码，发现 20 个问题：

| 优先级 | 数量 |
|--------|------|
| 🔴 阻塞性 | 6 |
| ⚠️ 高优先级 | 3 |
| ⚡ 中优先级 | 7 |
| 📝 低优先级 | 4 |

#### 二、已修复问题

##### 1. 阻塞性问题 (5/6)

| 任务 | 说明 |
|------|------|
| T-001 | token key 统一为 `zq_token` |
| T-002 | useOrders 监听器清理 |
| T-003 | usePosition 监听器清理 |
| T-004 | NotificationMobile 修复 |
| T-005 | mobile/Backtest 重复监听修复 |
| T-006 | ⏸️ 搁置（需先创建移动端页面） |

##### 2. 高优先级 (3/3)

| 任务 | 说明 |
|------|------|
| T-007 | useNetworkStatus 已正确实现 |
| T-008 | data.ts 错误处理统一 |
| T-009 | 新建 types/api.ts |

##### 3. 中优先级 (4/7)

| 任务 | 说明 |
|------|------|
| T-010 | 移动端涨跌颜色已正确 |
| T-011 | 新建 utils/format.ts |
| T-012 | ⏸️ 搁置（PC/移动端差异大） |
| T-013 | 导入位置已修复 |

##### 4. 低优先级 (4/4)

| 任务 | 说明 |
|------|------|
| T-014 | HelloWorld.vue 已删除 |
| T-015 | Dashboard 数据已对接持仓 |
| T-016 | Strategy 列表已对接 API |
| T-017 | .env 已正确配置 |
| T-018 | vite host 环境变量控制 |
| T-019 | WS URL 已支持参数传入 |
| T-020 | _data 是合理前缀 |

#### 三、新增功能

1. **注册页面** - `/register`
2. **登录页注册链接** - 跳转到注册页
3. **Dashboard 数据对接** - 对接持仓 composable
4. **Strategy 列表 API** - 对接后端 `/strategy/list`

#### 四、临时修复

1. **WebSocket 暂时禁用** - 后端 WebSocket 服务未就绪，等就绪后开启
2. **数据服务超时保护** - 防止 502 导致页面卡死

#### 五、待处理

1. 搁置 T-006：移动端策略编辑器页面
2. 搁置 T-012：Backtest 结果处理逻辑（PC/移动端差异大）
3. 后端 WebSocket 服务就绪后，开启 `WS_ENABLED`

#### 六、文件变更

**新增文件：**
- `src/types/api.ts` - 统一 API 类型定义
- `src/utils/format.ts` - 格式化工具函数
- `src/views/RegisterView.vue` - 注册页面

**修改文件：**
- `src/api/backend.ts` - 添加 register 函数
- `src/api/data.ts` - 错误处理统一
- `src/stores/auth.ts` - 修复 username 字段
- `src/router/index.ts` - 添加 /register 路由
- `src/views/LoginView.vue` - 添加注册链接
- `src/views/DashboardView.vue` - 数据对接
- `src/views/StrategyView.vue` - API 对接
- `src/composables/usePosition.ts` - 超时保护
- `src/composables/useOrders.ts` - 监听器清理
- `src/components/WsStatusBar.vue` - 登录检查
- `src/components/NotificationMobile.vue` - 监听器清理
- `src/utils/websocket.ts` - 安全包装 + 暂时禁用

**删除文件：**
- `src/components/HelloWorld.vue` - Vite 模板遗留

---

*End of Day*
