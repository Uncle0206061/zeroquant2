<template>
  <div class="mobile-notification">
    <van-badge :content="unreadCount" :max="99">
      <van-icon name="bell" size="22" color="#323233" @click="onClick" />
    </van-badge>

    <!-- 通知列表弹窗 -->
    <van-popup
      v-model:show="showPanel"
      position="bottom"
      round
      :style="{ height: '60%' }"
    >
      <div class="notif-panel">
        <div class="notif-header">
          <span class="notif-title">消息通知</span>
          <van-button size="mini" type="primary" plain @click="markAllRead">
            全部已读
          </van-button>
        </div>
        <van-divider :style="{ margin: '4px 0' }" />

        <van-pull-refresh v-model="refreshing" @refresh="refreshing = false">
          <van-cell-group v-if="notifications.length > 0">
            <van-cell
              v-for="n in notifications"
              :key="n.id"
              :title="n.content"
              :label="formatTime(n.time)"
              :border="true"
              :class="{ 'unread-cell': !n.read }"
              clickable
              @click="n.read = true; unreadCount = Math.max(0, unreadCount - 1)"
            />
          </van-cell-group>
          <van-empty v-else description="暂无消息" />
        </van-pull-refresh>
      </div>
    </van-popup>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { safeGetWebSocket } from '@/utils/websocket'

const showPanel = ref(false)
const refreshing = ref(false)
const unreadCount = ref(0)
const notifications = ref<any[]>([])

const typeLabelMap: Record<string, string> = {
  order_filled: '成交',
  order_rejected: '拒绝',
  backtest_complete: '回测完成',
  position_alert: '持仓预警',
  system: '系统',
}

function formatTime(time: string): string {
  const d = new Date(time)
  const now = new Date()
  const h = d.getHours().toString().padStart(2, '0')
  const m = d.getMinutes().toString().padStart(2, '0')
  if (d.toDateString() === now.toDateString()) return `${h}:${m}`
  return `${d.getMonth() + 1}/${d.getDate()} ${h}:${m}`
}

function markAllRead() {
  notifications.value.forEach((n) => { n.read = true })
  unreadCount.value = 0
}

function addNotification(type: string, content: string) {
  notifications.value.unshift({
    id: `${Date.now()}`,
    type,
    content,
    time: new Date().toISOString(),
    read: false,
  })
  unreadCount.value = Math.min(unreadCount.value + 1, 99)
  // 保留最新 50 条
  if (notifications.value.length > 50) {
    notifications.value = notifications.value.slice(0, 50)
  }
}

function onClick() {
  showPanel.value = true
}

// 保存回调引用以便清理
const eventHandlers: Record<string, (data: any) => void> = {}

const events = ['order_update', 'position_update', 'backtest_result', 'system_alert']

onMounted(() => {
  for (const evt of events) {
    eventHandlers[evt] = (data: any) => {
      try {
        const typeMap: Record<string, string> = {
          order_update: data?.status === 'filled' ? 'order_filled' : 'system',
          position_update: data?.alert ? 'position_alert' : 'system',
          backtest_result: 'backtest_complete',
          system_alert: 'system',
        }
        const notifType = typeMap[evt] || 'system'
        const content =
          typeof data === 'string' ? data : data?.message || JSON.stringify(data).substring(0, 80)
        addNotification(notifType, content)
      } catch {
        /* ignore */
      }
    }
    safeGetWebSocket().on(evt, eventHandlers[evt])
  }
})

onUnmounted(() => {
  for (const evt of events) {
    safeGetWebSocket().off(evt, eventHandlers[evt])
  }
})
</script>

<style scoped>
.mobile-notification {
  display: flex;
  align-items: center;
}

.notif-panel {
  padding: 12px 16px;
}

.notif-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.notif-title {
  font-size: 16px;
  font-weight: 600;
}

.unread-cell {
  background: #ecf5ff;
}
</style>
