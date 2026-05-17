<template>
  <div class="mobile-layout">
    <!-- 横屏提示 -->
    <transition name="fade">
      <div v-if="isLandscape" class="landscape-mask">
        <div class="rotate-icon">📱</div>
        <div class="rotate-text">请使用竖屏浏览</div>
      </div>
    </transition>

    <!-- 竖屏主体 -->
    <template v-if="!isLandscape">
      <!-- 顶部导航 -->
      <van-nav-bar
        :title="pageTitle"
        left-arrow
        @click-left="router.back()"
      >
        <template #right>
          <NotificationMobile />
        </template>
      </van-nav-bar>

      <!-- 离线提示 -->
      <van-notice-bar
        v-if="!isOnline"
        left-icon="info-o"
        background="#fff3cd"
        color="#856404"
      >
        当前离线，仅展示缓存数据
      </van-notice-bar>

      <!-- 内容区 -->
      <div class="mobile-content">
        <router-view />
      </div>

      <!-- 底部 Tab 栏 -->
      <van-tabbar v-model="activeTab" route>
        <van-tabbar-item to="/mobile/position" icon="cash-back-record">
          持仓
        </van-tabbar-item>
        <van-tabbar-item to="/mobile/orders" icon="orders-o">
          订单
        </van-tabbar-item>
        <van-tabbar-item to="/mobile/backtest" icon="chart-trending-o">
          回测
        </van-tabbar-item>
        <van-tabbar-item to="/mobile/strategy" icon="setting-o">
          策略
        </van-tabbar-item>
      </van-tabbar>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import NotificationMobile from '@/components/NotificationMobile.vue'
import { useNetworkStatus } from '@/composables/useNetworkStatus'

const route = useRoute()
const router = useRouter()
const { isOnline } = useNetworkStatus()

const activeTab = ref('/mobile/position')
const isLandscape = ref(false)

// 页面标题映射
const pageTitle = computed(() => {
  const map: Record<string, string> = {
    '/mobile/position': '我的持仓',
    '/mobile/orders': '订单列表',
    '/mobile/backtest': '回测分析',
    '/mobile/strategy': '策略管理',
    '/mobile/strategy/editor': '策略编辑',
    '/mobile/login': '登录',
  }
  return map[route.path] || 'ZeroQuant'
})

// 横屏检测
function checkOrientation() {
  isLandscape.value = window.innerWidth > window.innerHeight
}

onMounted(() => {
  checkOrientation()
  window.addEventListener('resize', checkOrientation)
  window.addEventListener('orientationchange', checkOrientation)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkOrientation)
  window.removeEventListener('orientationchange', checkOrientation)
})
</script>

<style scoped>
.mobile-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
}

/* 横屏提示 */
.landscape-mask {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: #f7f8fa;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.rotate-icon {
  font-size: 64px;
  animation: rotateHint 2s ease-in-out infinite;
}

.rotate-text {
  font-size: 16px;
  color: #646566;
  font-weight: 500;
}

@keyframes rotateHint {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(25deg); }
  75% { transform: rotate(-25deg); }
}

/* 内容区 */
.mobile-content {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
