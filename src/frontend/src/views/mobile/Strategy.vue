<template>
  <div class="mobile-strategy">
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh" :disabled="!isOnline">
      <van-list v-model:loading="loading" :finished="true" finished-text="">
        <!-- 策略卡片列表 -->
        <div v-for="s in strategies" :key="s.id" class="strategy-card">
          <div class="strategy-header">
            <span class="strategy-name">{{ s.name }}</span>
            <van-tag :type="s.status === 'active' ? 'success' : 'default'" size="medium">
              {{ s.status === 'active' ? '运行中' : '已停止' }}
            </van-tag>
          </div>
          <div class="strategy-desc">{{ s.description || '暂无描述' }}</div>
          <div class="strategy-footer">
            <span class="strategy-time">{{ s.updated_at || s.created_at || '' }}</span>
            <van-button
              size="mini"
              type="primary"
              plain
              :disabled="!isOnline"
              @click="$router.push('/mobile/strategy/editor?id=' + s.id)"
            >
              编辑
            </van-button>
          </div>
        </div>

        <!-- 新建策略按钮 -->
        <div class="new-strategy">
          <van-button
            type="primary"
            icon="plus"
            block
            round
            plain
            :disabled="!isOnline"
            @click="$router.push('/mobile/strategy/editor')"
          >
            新建策略
          </van-button>
        </div>

        <van-empty v-if="!loading && strategies.length === 0" description="暂无策略" />
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import backend from '@/api/backend'
import { useNetworkStatus } from '@/composables/useNetworkStatus'

const { isOnline } = useNetworkStatus()

const loading = ref(false)
const refreshing = ref(false)
const strategies = ref<any[]>([])

async function fetchStrategies() {
  loading.value = true
  try {
    const res: any = await backend.get('/strategy/list')
    strategies.value = res ?? res?.data ?? []
  } catch {
    /* ignore */
  } finally {
    loading.value = false
  }
}

async function onRefresh() {
  if (!isOnline.value) {
    refreshing.value = false
    return
  }
  await fetchStrategies()
  refreshing.value = false
}

// 初始化加载
fetchStrategies()
</script>

<style scoped>
.mobile-strategy {
  padding: 0 0 16px 0;
}

.strategy-card {
  margin: 8px 12px;
  padding: 14px 16px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.strategy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.strategy-name {
  font-size: 15px;
  font-weight: 600;
  color: #323233;
}

.strategy-desc {
  font-size: 13px;
  color: #646566;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.strategy-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.strategy-time {
  font-size: 12px;
  color: #969799;
}

.new-strategy {
  margin: 16px 12px;
}
</style>
