<template>
  <div class="strategy">
    <div class="page-header">
      <h2>策略管理</h2>
      <el-button type="primary" @click="$router.push('/strategy/editor')">
        新建策略
      </el-button>
    </div>

    <!-- 策略表格 -->
    <el-table :data="strategies" v-loading="loading" stripe>
      <el-table-column prop="name" label="策略名称" min-width="150" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">
            {{ statusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="returnRate" label="收益率" width="120">
        <template #default="{ row }">
          <span :class="row.return_rate >= 0 ? 'rate-up' : 'rate-down'">
            {{ row.return_rate ? `${(row.return_rate * 100).toFixed(2)}%` : '--' }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="180">
        <template #default="{ row }">
          {{ formatDateTime(row.updated_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="editStrategy(row)">
            编辑
          </el-button>
          <el-button size="small" type="primary" link @click="runBacktest(row)">
            回测
          </el-button>
          <el-button
            size="small"
            type="danger"
            link
            @click="deleteStrategy(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 空状态 -->
    <el-empty v-if="!loading && strategies.length === 0" description="暂无策略，点击新建创建" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import backend from '@/api/backend'

const router = useRouter()
const strategies = ref<any[]>([])
const loading = ref(false)

// 状态映射
const statusMap: Record<string, { label: string; type: string }> = {
  draft: { label: '草稿', type: 'info' },
  running: { label: '运行中', type: 'success' },
  stopped: { label: '已停止', type: '' },
}

function statusLabel(status: string) {
  return statusMap[status]?.label || status
}

function statusType(status: string) {
  return statusMap[status]?.type || 'info'
}

function formatDateTime(d: string) {
  if (!d) return '-'
  return d.replace('T', ' ').substring(0, 19)
}

// 加载策略列表
async function fetchStrategies() {
  loading.value = true
  try {
    const res: any = await backend.get('/strategy/list')
    strategies.value = res?.data?.list ?? res?.list ?? []
  } catch (e) {
    console.error('[Strategy] fetch failed', e)
    strategies.value = []
  } finally {
    loading.value = false
  }
}

// 编辑策略
function editStrategy(row: any) {
  router.push(`/strategy/editor?id=${row.id}`)
}

// 回测
function runBacktest(row: any) {
  router.push(`/backtest?strategy=${row.id}`)
}

// 删除策略
async function deleteStrategy(row: any) {
  try {
    await ElMessageBox.confirm(`确定删除策略 "${row.name}"？`, '删除确认', {
      type: 'warning',
    })
    await backend.delete(`/strategy/${row.id}`)
    ElMessage.success('删除成功')
    fetchStrategies()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '删除失败')
    }
  }
}

onMounted(() => {
  fetchStrategies()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.rate-up {
  color: #F56C6C;
  font-weight: 600;
}
.rate-down {
  color: #67C23A;
  font-weight: 600;
}
</style>