<template>
  <div class="orders-page">
    <div class="page-header">
      <h2>订单列表</h2>
      <el-button :icon="Refresh" @click="fetchOrders" :loading="loading" text>
        刷新
      </el-button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-select
        v-model="filterDirection"
        placeholder="方向"
        clearable
        style="width: 100px"
        @change="onFilterChange"
      >
        <el-option label="买入" value="1" />
        <el-option label="卖出" value="2" />
      </el-select>
      <el-select
        v-model="filterStatus"
        placeholder="状态"
        clearable
        style="width: 120px"
        @change="onFilterChange"
      >
        <el-option label="待成交" value="pending" />
        <el-option label="部分成交" value="partial_filled" />
        <el-option label="已成交" value="filled" />
        <el-option label="已撤单" value="cancelled" />
        <el-option label="已拒绝" value="rejected" />
      </el-select>
      <el-button link @click="resetFilter">重置</el-button>
    </div>

    <!-- 订单表格 -->
    <el-table
      :data="paginatedOrders"
      v-loading="loading"
      stripe
      class="orders-table"
      @row-click="showDetail"
    >
      <el-table-column label="订单号" min-width="200">
        <template #default="{ row }">
          <span class="order-id">{{ row.order_id }}</span>
        </template>
      </el-table-column>
      <el-table-column label="时间" width="160">
        <template #default="{ row }">
          {{ formatDateTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="股票" width="150">
        <template #default="{ row }">
          <span class="stock-code">{{ row.stock_code }}</span>
          <span class="stock-name">{{ row.stock_name }}</span>
        </template>
      </el-table-column>
      <el-table-column label="方向" width="70" align="center">
        <template #default="{ row }">
          <el-tag
            :type="row.direction === 1 ? 'danger' : 'success'"
            size="small"
            effect="plain"
          >
            {{ row.direction === 1 ? '买入' : '卖出' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="委托价" width="90" align="right">
        <template #default="{ row }">¥{{ row.price.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="委托量" width="80" align="right">
        <template #default="{ row }">{{ row.quantity }}</template>
      </el-table-column>
      <el-table-column label="成交量" width="80" align="right">
        <template #default="{ row }">{{ row.filled_qty }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag
            :type="(statusConfig[row.status] as any)?.type || 'info'"
            size="small"
          >
            {{ (statusConfig[row.status] as any)?.label || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80" align="center">
        <template #default="{ row }">
          <el-button
            v-if="row.status === 'pending'"
            size="small"
            type="danger"
            link
            @click.stop="handleCancel(row.order_id)"
          >
            撤单
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-wrap" v-if="total > pageSize">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        small
      />
    </div>

    <!-- 空状态 -->
    <el-empty v-if="!loading && allOrders.length === 0" description="暂无订单" />

    <!-- 成交明细弹窗 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="560px" destroy-on-close>
      <template v-if="detail">
        <el-descriptions :column="2" border size="default">
          <el-descriptions-item label="订单号" :span="2">
            {{ detail.order_id }}
          </el-descriptions-item>
          <el-descriptions-item label="股票">
            {{ detail.stock_code }} {{ detail.stock_name }}
          </el-descriptions-item>
          <el-descriptions-item label="方向">
            <el-tag
              :type="detail.direction === 1 ? 'danger' : 'success'"
              size="small"
            >
              {{ detail.direction === 1 ? '买入' : '卖出' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag
              :type="(statusConfig[detail.status] as any)?.type || 'info'"
              size="small"
            >
              {{ (statusConfig[detail.status] as any)?.label || detail.status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDateTime(detail.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="委托价格">
            ¥{{ detail.price.toFixed(2) }}
          </el-descriptions-item>
          <el-descriptions-item label="委托数量">
            {{ detail.quantity }}
          </el-descriptions-item>
          <el-descriptions-item label="成交数量">
            {{ detail.filled_qty }}
          </el-descriptions-item>
          <el-descriptions-item label="成交均价">
            {{ detail.avg_price ? `¥${detail.avg_price.toFixed(2)}` : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="手续费">
            {{ detail.commission ? `¥${detail.commission.toFixed(2)}` : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="拒绝原因" :span="2" v-if="detail.reject_reason">
            <span style="color: #F56C6C">{{ detail.reject_reason }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useOrders, statusConfig } from '@/composables/useOrders'

const {
  allOrders,
  paginatedOrders,
  total,
  loading,
  filterStatus,
  filterDirection,
  currentPage,
  pageSize,
  fetchOrders,
  cancelOrder,
  getOrderDetail,
  resetFilter,
  onFilterChange,
} = useOrders()

const detailVisible = ref(false)
const detail = ref<any>(null)

async function showDetail(row: any) {
  try {
    detail.value = await getOrderDetail(row.order_id)
    detailVisible.value = true
  } catch {
    ElMessage.error('获取订单详情失败')
  }
}

async function handleCancel(orderId: string) {
  try {
    await ElMessageBox.confirm('确认撤销该订单？', '撤单确认', {
      type: 'warning',
      confirmButtonText: '确认撤单',
      cancelButtonText: '取消',
    })
    await cancelOrder(orderId)
    ElMessage.success('撤单成功')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '撤单失败')
    }
  }
}

function formatDateTime(d: string): string {
  if (!d) return '-'
  return d.replace('T', ' ').substring(0, 19)
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.orders-table {
  cursor: pointer;
}

.order-id {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  color: #606266;
}

.stock-code {
  font-weight: 600;
  color: #303133;
}

.stock-name {
  margin-left: 6px;
  color: #909399;
  font-size: 13px;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
