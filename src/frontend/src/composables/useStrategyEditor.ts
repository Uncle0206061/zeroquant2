import { reactive, computed, ref, watch } from 'vue'
import type { Factor } from '@/data/factors'
import { allFactors } from '@/data/factors'
import backend from '@/api/backend'
import { safeGetWebSocket } from '@/utils/websocket'
import { ElMessage } from 'element-plus'

/** 策略配置 JSON 结构（与后端 biz_strategy 一致） */
export interface StrategyConfig {
  name: string
  stock_filter: Record<string, any>
  timing: { type: string; period?: number }
  risk: {
    max_position: number
    stop_loss: number
    stop_profit: number
    max_hold_days?: number
  }
}

export function useStrategyEditor() {
  // 选中的因子 key 列表
  const selectedKeys = ref<string[]>([])

  // 表单数据
  const formData = reactive({
    name: '',
    stock_filter: {} as Record<string, any>,
    timing: { type: '', period: 20 },
    risk: {
      max_position: 0.3,
      stop_loss: 0.05,
      stop_profit: 0.1,
      max_hold_days: 30,
    },
  })

  // 当前任务 ID（提交执行后获得）
  const taskId = ref('')
  const submitting = ref(false)

  // 选中的因子对象列表
  const selectedFactors = computed<Factor[]>(() => {
    return allFactors.filter((f) => selectedKeys.value.includes(f.key))
  })

  // 实时 JSON 预览
  const jsonPreview = computed<string>(() => {
    const config = buildConfig()
    return JSON.stringify(config, null, 2)
  })

  // 构建最终配置
  function buildConfig(): StrategyConfig {
    const config: StrategyConfig = {
      name: formData.name || '未命名策略',
      stock_filter: {},
      timing: { ...formData.timing },
      risk: { ...formData.risk },
    }
    // 清理空的 stock_filter
    for (const factor of selectedFactors.value) {
      const val = formData.stock_filter[factor.key]
      if (val && Object.keys(val).length > 0) {
        config.stock_filter[factor.key] = val
      }
    }
    // 清理不需要的 timing 字段
    if (!config.timing.period) {
      delete config.timing.period
    }
    if (!config.timing.type) {
      delete config.timing.type
    }
    return config
  }

  // 因子选中/取消回调
  function onFactorCheck(_data: any, checkStatus: { checkedKeys: string[] }) {
    const newKeys = checkStatus.checkedKeys.filter((k) => typeof k === 'string')
    const removedKeys = selectedKeys.value.filter((k) => !newKeys.includes(k))
    const addedKeys = newKeys.filter((k) => !selectedKeys.value.includes(k))

    // 移除取消选中的因子参数
    for (const key of removedKeys) {
      delete formData.stock_filter[key]
    }

    // 初始化新增因子的默认参数
    for (const key of addedKeys) {
      const factor = allFactors.find((f) => f.key === key)
      if (factor && !formData.stock_filter[key]) {
        const defaults: Record<string, any> = {}
        for (const p of factor.params) {
          defaults[p.key] = p.default ?? (p.type === 'tags' ? [] : '')
        }
        formData.stock_filter[key] = defaults
      }
    }

    selectedKeys.value = newKeys
  }

  // 保存草稿
  async function saveDraft() {
    const config = buildConfig()
    try {
      await backend.post('/strategy/save', {
        name: formData.name || '未命名策略',
        config,
      })
      ElMessage.success('草稿保存成功')
    } catch (e: any) {
      ElMessage.error(e.message || '保存失败')
    }
  }

  // 提交执行
  async function submitStrategy() {
    if (!formData.name?.trim()) {
      ElMessage.warning('请输入策略名称')
      return
    }
    if (selectedKeys.value.length === 0) {
      ElMessage.warning('请至少选择一个因子')
      return
    }
    if (!formData.timing.type) {
      ElMessage.warning('请选择择时条件')
      return
    }

    submitting.value = true
    try {
      const config = buildConfig()
      const res: any = await backend.post('/strategy/submit', config)
      taskId.value = res?.task_id || res?.data?.task_id || ''
      ElMessage.success(`提交成功${taskId.value ? `，task_id: ${taskId.value}` : ''}`)

      // 监听回测进度
      safeGetWebSocket().on('backtest_progress', (data: any) => {
        if (!taskId.value || data?.task_id !== taskId.value) return
        if (data.progress !== undefined) {
          ElMessage.info(`回测进度: ${data.progress}%`)
        }
      })
    } catch (e: any) {
      ElMessage.error(e.message || '提交失败')
    } finally {
      submitting.value = false
    }
  }

  // 重置表单
  function resetForm() {
    formData.name = ''
    formData.stock_filter = {}
    formData.timing = { type: '', period: 20 }
    formData.risk = { max_position: 0.3, stop_loss: 0.05, stop_profit: 0.1, max_hold_days: 30 }
    selectedKeys.value = []
    taskId.value = ''
  }

  return {
    formData,
    selectedKeys,
    selectedFactors,
    jsonPreview,
    taskId,
    submitting,
    onFactorCheck,
    saveDraft,
    submitStrategy,
    resetForm,
    buildConfig,
  }
}
