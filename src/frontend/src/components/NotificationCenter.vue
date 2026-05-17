<template>
  <div class="notification-center">
    <!-- 铃铛图标 + 角标 -->
    <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
      <el-popover
        placement="bottom-end"
        :width="380"
        trigger="click"
        v-model:visible="showPanel"
        popper-class="notif-popover"
      >
        <template #reference>
          <div class="bell-trigger" :class="{ ringing: unreadCount > 0 }">
            <el-icon :size="20"><Bell /></el-icon>
          </div>
        </template>

        <!-- 通知面板 -->
        <div class="notif-panel">
          <div class="notif-header">
            <span class="notif-title">消息通知</span>
            <div class="notif-actions">
              <el-switch
                v-model="soundEnabled"
                size="small"
                inline-prompt
                active-text="🔊"
                inactive-text="🔇"
                @change="onSoundChange"
              />
              <el-button link type="primary" size="small" @click="markAllRead">
                全部已读
              </el-button>
              <el-button link type="info" size="small" @click="clearAll">
                清除
              </el-button>
            </div>
          </div>
          <el-divider style="margin: 8px 0" />

          <!-- 通知列表 -->
          <div class="notif-list" v-if="notifications.length > 0">
            <div
              v-for="n in notifications"
              :key="n.id"
              class="notif-item"
              :class="{ unread: !n.read }"
              @click="markRead(n)"
            >
              <div class="notif-item-header">
                <el-tag :type="typeTagMap[n.type] || 'info'" size="small" effect="plain">
                  {{ typeLabelMap[n.type] || n.type }}
                </el-tag>
                <span class="notif-time">{{ formatTime(n.time) }}</span>
              </div>
              <div class="notif-content">{{ n.content }}</div>
            </div>
          </div>
          <el-empty v-else description="暂无消息" :image-size="60" />
        </div>
      </el-popover>
    </el-badge>

    <!-- 右下角浮窗提醒 -->
    <transition-group name="toast">
      <div
        v-for="t in activeToasts"
        :key="t.id"
        class="toast-popup"
        :class="t.type"
      >
        <el-icon v-if="t.type === 'order_filled'" color="#67C23A"><CircleCheck /></el-icon>
        <el-icon v-if="t.type === 'order_rejected'" color="#F56C6C"><CircleClose /></el-icon>
        <el-icon v-if="t.type === 'backtest_complete'" color="#409EFF"><Finished /></el-icon>
        <el-icon v-if="t.type === 'position_alert'" color="#E6A23C"><WarningFilled /></el-icon>
        <el-icon v-if="t.type === 'system'" color="#909399"><InfoFilled /></el-icon>
        <span class="toast-text">{{ t.content }}</span>
      </div>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  Bell, CircleCheck, CircleClose, Finished,
  WarningFilled, InfoFilled,
} from '@element-plus/icons-vue'
import { safeGetWebSocket } from '@/utils/websocket'

// ========== 类型定义 ==========
interface Notification {
  id: string
  type: string
  content: string
  time: string
  read: boolean
}

interface Toast extends Notification {
  timer: number
}

// ========== 事件映射 ==========
const typeTagMap: Record<string, string> = {
  order_filled: 'success',
  order_rejected: 'danger',
  backtest_complete: '',
  position_alert: 'warning',
  system: 'info',
}

const typeLabelMap: Record<string, string> = {
  order_filled: '成交',
  order_rejected: '拒绝',
  backtest_complete: '回测完成',
  position_alert: '持仓预警',
  system: '系统',
}

// ========== 状态 ==========
const showPanel = ref(false)
const soundEnabled = ref(true)
const notifications = ref<Notification[]>([])
const activeToasts = ref<Toast[]>([])

// 未读数量（从 localStorage 恢复）
const unreadCount = computed(() => notifications.value.filter((n) => !n.read).length)

// ========== localStorage ==========
const STORAGE_KEY = 'zq_notifications'
const SOUND_KEY = 'zq_sound_enabled'

function loadHistory() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const all: Notification[] = JSON.parse(raw)
    // 只保留 24 小时内的
    const cutoff = Date.now() - 24 * 60 * 60 * 1000
    notifications.value = all.filter((n) => new Date(n.time).getTime() > cutoff)
  } catch {
    /* ignore */
  }
}

function saveNotifications() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(notifications.value))
  } catch {
    /* ignore */
  }
}

// ========== 通知操作 ==========
function addNotification(type: string, content: string) {
  const item: Notification = {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    type,
    content,
    time: new Date().toISOString(),
    read: false,
  }
  notifications.value.unshift(item)
  saveNotifications()
  showToast(item)
  playSound()
}

function markRead(n: Notification) {
  n.read = true
  saveNotifications()
}

function markAllRead() {
  notifications.value.forEach((n) => { n.read = true })
  saveNotifications()
}

function clearAll() {
  notifications.value = []
  saveNotifications()
}

// ========== 浮窗 ==========
function showToast(notif: Notification) {
  const toast: Toast = { ...notif, timer: window.setTimeout(() => removeToast(toast.id), 5000) }
  activeToasts.value.push(toast)
}

function removeToast(id: string) {
  const idx = activeToasts.value.findIndex((t) => t.id === id)
  if (idx >= 0) {
    clearTimeout(activeToasts.value[idx].timer)
    activeToasts.value.splice(idx, 1)
  }
}

// ========== 声音 ==========
function loadSoundSetting() {
  soundEnabled.value = localStorage.getItem(SOUND_KEY) !== 'false'
}

function onSoundChange(val: boolean) {
  localStorage.setItem(SOUND_KEY, String(val))
}

function playSound() {
  if (!soundEnabled.value) return
  try {
    // 使用 Web Audio API 生成简短提示音（无需外部文件）
    const ctx = new AudioContext()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain)
    gain.connect(ctx.destination)
    osc.frequency.value = 880
    gain.gain.value = 0.3
    osc.start()
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.3)
    osc.stop(ctx.currentTime + 0.3)
  } catch {
    /* 浏览器可能阻止自动播放 */
  }
}

// ========== 时间格式化 ==========
function formatTime(time: string) {
  const d = new Date(time)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  const h = d.getHours().toString().padStart(2, '0')
  const m = d.getMinutes().toString().padStart(2, '0')
  if (isToday) return `${h}:${m}`
  return `${d.getMonth() + 1}/${d.getDate()} ${h}:${m}`
}

// ========== WebSocket 监听 ==========
const wsEventTypes = [
  { type: 'order_update', condition: (d: any) => d?.status === 'filled' },
  { type: 'order_update', condition: (d: any) => d?.status === 'rejected', mapType: 'order_rejected' },
  { type: 'position_update', condition: (d: any) => d?.alert },
  { type: 'backtest_result', condition: () => true, mapType: 'backtest_complete' },
  { type: 'system_alert', condition: () => true, mapType: 'system' },
]

onMounted(() => {
  loadHistory()
  loadSoundSetting()

  // 注册 WebSocket 监听
  for (const evt of wsEventTypes) {
    safeGetWebSocket().on(evt.type, (data: any) => {
      try {
        if (evt.condition(data)) {
          const notifType = (evt as any).mapType || evt.type
          const content =
            typeof data === 'string'
              ? data
              : data?.message || data?.detail || JSON.stringify(data).substring(0, 100)
          addNotification(notifType, content)
        }
      } catch {
        /* ignore */
      }
    })
  }
})

onUnmounted(() => {
  activeToasts.value.forEach((t) => clearTimeout(t.timer))
})
</script>

<style scoped>
.notification-center {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.bell-trigger {
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  transition: all 0.3s;
  color: #909399;
}

.bell-trigger:hover {
  background: #f0f2f5;
}

.bell-trigger.ringing {
  color: #409eff;
  animation: ring 0.5s ease-in-out;
}

@keyframes ring {
  0%, 100% { transform: rotate(0); }
  20% { transform: rotate(15deg); }
  40% { transform: rotate(-10deg); }
  60% { transform: rotate(5deg); }
  80% { transform: rotate(-3deg); }
}

.notif-panel {
  margin: -12px;
}

.notif-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px;
}

.notif-title {
  font-weight: 600;
  font-size: 15px;
}

.notif-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.notif-list {
  max-height: 400px;
  overflow-y: auto;
}

.notif-item {
  padding: 10px 8px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.2s;
  border-radius: 4px;
}

.notif-item:hover {
  background: #f5f7fa;
}

.notif-item.unread {
  background: #ecf5ff;
  border-left: 3px solid #409eff;
}

.notif-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.notif-time {
  font-size: 12px;
  color: #909399;
}

.notif-content {
  font-size: 13px;
  color: #303133;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 浮窗 */
.toast-popup {
  position: fixed;
  bottom: 60px;
  right: 24px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 14px 18px;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
  display: flex;
  align-items: center;
  gap: 10px;
  z-index: 9999;
  min-width: 260px;
  max-width: 400px;
}

.toast-popup.order_filled { border-left: 4px solid #67C23A; }
.toast-popup.order_rejected { border-left: 4px solid #F56C6C; }
.toast-popup.backtest_complete { border-left: 4px solid #409EFF; }
.toast-popup.position_alert { border-left: 4px solid #E6A23C; }
.toast-popup.system { border-left: 4px solid #909399; }

.toast-text {
  font-size: 14px;
  color: #303133;
  flex: 1;
}

/* 浮窗动画 */
.toast-enter-active {
  transition: all 0.35s ease;
}
.toast-leave-active {
  transition: all 0.25s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(40px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(40px);
}
</style>
