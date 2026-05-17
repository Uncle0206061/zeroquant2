<template>
  <el-container class="main-layout">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapsed ? '64px' : '200px'" class="aside">
      <div class="logo">
        <span v-if="!isCollapsed">ZeroQuant 2.0</span>
        <span v-else>ZQ</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapsed"
        router
        background-color="#001529"
        text-color="#ffffffa6"
        active-text-color="#ffffff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><Odometer /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>
        <el-menu-item index="/market">
          <el-icon><DataLine /></el-icon>
          <template #title>行情</template>
        </el-menu-item>
        <el-menu-item index="/strategy">
          <el-icon><Setting /></el-icon>
          <template #title>策略管理</template>
        </el-menu-item>
        <el-menu-item index="/position">
          <el-icon><PieChart /></el-icon>
          <template #title>持仓</template>
        </el-menu-item>
        <el-menu-item index="/orders">
          <el-icon><List /></el-icon>
          <template #title>订单</template>
        </el-menu-item>
        <el-menu-item index="/backtest">
          <el-icon><TrendCharts /></el-icon>
          <template #title>回测</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapsed = !isCollapsed">
            <Fold v-if="!isCollapsed" />
            <Expand v-else />
          </el-icon>
        </div>
        <div class="header-right">
          <!-- 消息通知组件 -->
          <NotificationCenter />
          <!-- 用户下拉菜单 -->
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              {{ username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import NotificationCenter from '@/components/NotificationCenter.vue'
import {
  Odometer, Setting, PieChart, List, TrendCharts,
  ArrowDown, Fold, Expand, DataLine,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const isCollapsed = ref(false)

const activeMenu = computed(() => route.path)
const username = computed(() => auth.username || '用户')

function handleCommand(command: string) {
  if (command === 'logout') {
    auth.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.main-layout {
  height: 100vh;
}
.aside {
  background-color: #001529;
  transition: width 0.3s;
  overflow: hidden;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  border-bottom: 1px solid #ffffff1a;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e8e8e8;
  padding: 0 20px;
  background: #fff;
}
.collapse-btn {
  cursor: pointer;
  font-size: 20px;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}
.user-info {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
}
.main-content {
  background: #f0f2f5;
  overflow-y: auto;
}
</style>
