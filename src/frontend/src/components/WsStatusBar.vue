<template>
  <transition name="slide-down">
    <div v-if="showStatus && status === 'disconnected'" class="ws-status-bar">
      <span class="dot disconnected" />
      <span>连接已断开，正在重连...</span>
    </div>
    <div v-else-if="showStatus && status === 'connecting'" class="ws-status-bar connecting">
      <van-loading size="14" color="#e6a23c" />
      <span>正在连接...</span>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { wsStatus } from '@/utils/websocket'

const status = wsStatus

// 未登录时不显示状态栏
const showStatus = computed(() => {
  const token = localStorage.getItem('zq_token')
  return !!token
})
</script>

<style scoped>
.ws-status-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: #f56c6c;
  color: #fff;
  text-align: center;
  padding: 6px 12px;
  font-size: 13px;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.ws-status-bar.connecting {
  background: #e6a23c;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #fff;
  flex-shrink: 0;
}

.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
}

.slide-down-enter-from,
.slide-down-leave-to {
  transform: translateY(-100%);
  opacity: 0;
}
</style>