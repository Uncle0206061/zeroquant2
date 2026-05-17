<template>
  <div class="mobile-orders">
    <!-- 状态筛选标签 -->
    <van-tabs v-model:active="activeTab" animated swipeable sticky>
      <van-tab title="全部" name="" />
      <van-tab title="待成交" name="pending" />
      <van-tab title="已成交" name="filled" />
      <van-tab title="已撤单" name="cancelled" />
    </van-tabs>

    <!-- 订单列表 -->
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh" :disabled="!isOnline">
      <van-list v-model:loading="loading" :finished="true" finished-text="">
        <div
          v-for="order in displayOrders"
          :key="order.order_id"
          class="order-card"
        >
          <!-- 头部：股票 + 状态 -->
          <div class="order-header">
            <div class="order-stock">
              <span class="order-name">{{ order.stock_name || order.stock_code }}</span>
              <span class="order-code">{{ order.stock_code }}</span>
            </div>
            <van-tag :type="(statusConfig[order.status] as any)?.type || 'default'" size="medium">
              {{ (statusConfig[order.status] as any)?.label || order.status }}
            </van-tag>
          </div>

          <!-- 中间：方向 + 价格 + 数量 -->
          <div class="order-body">
            <div class="order-direction">
              <span :class="order.direction === 1 ? 'buy-text' : 'sell-text'">
                {{ order.direction === 1 ? '买入' : '卖出' }}
              </span>
            </div>
            <div class="order-price">
              <span class="label">价格</span>
              <span class="value">¥{{ order.price.toFixed(2) }}</span>
            </div>
            <div class="order-qty">
              <span class="label">数量</span>
              <span class="value">{{ order.quantity }}股</span>
            </div>
            <div class="order-filled">
              <span class="label">成交</span>
              <span class="value">{{ order.filled_qty }}股</span>
            </div>
          </div>

          <!-- 底部：时间 + 操作 -->
          <div class="order-footer">
            <span class="order-time">{{ formatDateTime(order.created_at) }}</span>
            <van-button
              v-if="order.status === 'pending' && isOnline"
              size="mini"
              type="danger"
              plain
              @click="handleCancel(order.order_id)"
            >
              撤单
            </van-button>
            <van-button
              v-if="order.status === 'pending' && !isOnline"
              size="mini"
              type="default"
              plain
              disabled
            >
              离线
            </van-button>
          </div>
        </div>

        <van-empty v-if="!loading && displayOrders.length === 0" description="暂无订单" />
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { showConfirmDialog, showToast } from 'vant'
import { useOrders, statusConfig } from '@/composables/useOrders'
import { useNetworkStatus } from '@/composables/useNetworkStatus'

const { isOnline } = useNetworkStatus()
const {
  filteredOrders,
  loading,
  fetchOrders,
  cancelOrder,
  filterStatus,
  onFilterChange,
} = useOrders()

const refreshing = ref(false)
const activeTab = ref('')

// 同步 Vant Tab 和 composable 的筛选
watch(activeTab, (val) => {
  filterStatus.value = val
  onFilterChange()
})

const displayOrders = computed(() => filteredOrders.value)

async function onRefresh() {
  if (!isOnline.value) {
    refreshing.value = false
    return
  }
  await fetchOrders()
  refreshing.value = false
}

async function handleCancel(orderId: string) {
  try {
    await showConfirmDialog({
      title: '撤单确认',
      message: '确认撤销该订单？',
    })
    await cancelOrder(orderId)
    showToast({ message: '撤单成功', type: 'success' })
  } catch (e: any) {
    if (e !== 'cancel') {
      showToast({ message: e.message || '撤单失败', type: 'fail' })
    }
  }
}

function formatDateTime(d: string): string {
  if (!d) return '-'
  const cleaned = d.replace('T', ' ').substring(0, 16)
  return cleaned
}
</script>

<style scoped>
.mobile-orders {
  min-height: 100%;
  background: #f7f8fa;
}

/* 订单卡片 */
.order-card {
  margin: 8px 12px;
  padding: 14px 16px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.order-name {
  font-size: 15px;
  font-weight: 600;
  color: #323233;
}

.order-code {
  margin-left: 6px;
  font-size: 12px;
  color: #969799;
}

.order-body {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 10px;
}

.order-direction {
  font-size: 16px;
  font-weight: 700;
  min-width: 32px;
}

.buy-text { color: #ee0a24; }
.sell-text { color: #07c160; }

.order-price,
.order-qty,
.order-filled {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.order-body .label {
  font-size: 11px;
  color: #969799;
}

.order-body .value {
  font-size: 14px;
  color: #323233;
  font-weight: 500;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid #f5f5f5;
}

.order-time {
  font-size: 12px;
  color: #969799;
}
</style>
