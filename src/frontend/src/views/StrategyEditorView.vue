<template>
  <div class="strategy-editor">
    <!-- 顶部操作栏 -->
    <div class="editor-header">
      <div class="header-left">
        <el-button text @click="$router.back()">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h2>策略编辑器</h2>
      </div>
      <div class="header-right">
        <el-button @click="resetForm">重置</el-button>
        <el-button @click="saveDraft" :disabled="submitting">保存草稿</el-button>
        <el-button type="primary" @click="submitStrategy" :loading="submitting">
          提交执行
        </el-button>
      </div>
    </div>

    <!-- 策略名称 -->
    <div class="strategy-name-row">
      <el-input
        v-model="formData.name"
        placeholder="输入策略名称"
        size="large"
        clearable
        maxlength="50"
        show-word-limit
      >
        <template #prefix>
          <el-icon><EditPen /></el-icon>
        </template>
      </el-input>
    </div>

    <!-- 三栏主体 -->
    <div class="editor-body">
      <!-- 左侧：因子选择树 -->
      <div class="factor-panel">
        <div class="panel-title">选股因子</div>
        <el-tabs v-model="activeTab">
          <el-tab-pane label="财务指标" name="financial">
            <el-tree
              :data="financialTreeData"
              show-checkbox
              node-key="key"
              :default-checked-keys="selectedKeys"
              :props="{ label: 'name', children: 'children' }"
              @check="onFactorCheck"
              class="factor-tree"
            />
          </el-tab-pane>
          <el-tab-pane label="技术指标" name="technical">
            <el-tree
              :data="technicalTreeData"
              show-checkbox
              node-key="key"
              :default-checked-keys="selectedKeys"
              :props="{ label: 'name', children: 'children' }"
              @check="onFactorCheck"
              class="factor-tree"
            />
          </el-tab-pane>
          <el-tab-pane label="板块概念" name="sector">
            <el-tree
              :data="sectorTreeData"
              show-checkbox
              node-key="key"
              :default-checked-keys="selectedKeys"
              :props="{ label: 'name', children: 'children' }"
              @check="onFactorCheck"
              class="factor-tree"
            />
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 中间：参数表单 -->
      <div class="form-panel">
        <div class="panel-title">参数配置</div>

        <!-- 选股条件 -->
        <template v-if="selectedFactors.length > 0">
          <el-divider content-position="left">选股条件</el-divider>
          <el-form label-width="120px" size="default">
            <template v-for="factor in selectedFactors" :key="factor.key">
              <div class="factor-section">
                <div class="factor-label">{{ factor.name }}</div>
                <template v-for="param in factor.params" :key="param.key">
                  <el-form-item :label="param.label">
                    <!-- 数值型 -->
                    <el-input-number
                      v-if="param.type === 'number'"
                      v-model="formData.stock_filter[factor.key][param.key]"
                      :min="param.min"
                      :max="param.max"
                      :step="param.step || 1"
                      :precision="param.precision"
                      controls-position="right"
                    />
                    <!-- 下拉选择 -->
                    <el-select
                      v-else-if="param.type === 'select'"
                      v-model="formData.stock_filter[factor.key][param.key]"
                    >
                      <el-option
                        v-for="opt in param.options"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </el-select>
                    <!-- 标签选择 -->
                    <div v-else-if="param.type === 'tags'" class="tags-wrapper">
                      <el-check-tag
                        v-for="opt in param.options"
                        :key="opt.value"
                        :checked="(formData.stock_filter[factor.key][param.key] || []).includes(opt.value)"
                        @change="toggleTag(factor.key, param.key, opt.value)"
                      >
                        {{ opt.label }}
                      </el-check-tag>
                    </div>
                  </el-form-item>
                </template>
              </div>
            </template>
          </el-form>

          <!-- 择时条件 -->
          <el-divider content-position="left">择时条件</el-divider>
          <el-form label-width="120px" size="default">
            <el-form-item label="买入时机">
              <el-select v-model="formData.timing.type" placeholder="请选择" clearable>
                <el-option
                  v-for="opt in timingOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              v-if="formData.timing.type && timingNeedPeriod.includes(formData.timing.type)"
              label="周期"
            >
              <el-input-number
                v-model="formData.timing.period"
                :min="5"
                :max="250"
                :step="1"
                controls-position="right"
              />
              <span class="input-unit">日</span>
            </el-form-item>
          </el-form>

          <!-- 风控参数 -->
          <el-divider content-position="left">风控参数</el-divider>
          <el-form label-width="120px" size="default">
            <el-form-item label="最大仓位">
              <el-slider
                v-model="formData.risk.max_position"
                :min="0.05"
                :max="0.8"
                :step="0.05"
                :format-tooltip="(v: number) => `${Math.round(v * 100)}%`"
                show-input
              />
            </el-form-item>
            <el-form-item label="止损">
              <el-input-number
                v-model="formData.risk.stop_loss"
                :min="0.01"
                :max="0.5"
                :step="0.01"
                :precision="2"
                controls-position="right"
              />
              <span class="input-unit">%</span>
            </el-form-item>
            <el-form-item label="止盈">
              <el-input-number
                v-model="formData.risk.stop_profit"
                :min="0.01"
                :max="1.0"
                :step="0.01"
                :precision="2"
                controls-position="right"
              />
              <span class="input-unit">%</span>
            </el-form-item>
            <el-form-item label="最大持仓天数">
              <el-input-number
                v-model="formData.risk.max_hold_days"
                :min="1"
                :max="365"
                :step="1"
                controls-position="right"
              />
              <span class="input-unit">天</span>
            </el-form-item>
          </el-form>
        </template>

        <!-- 空状态 -->
        <div v-else class="empty-tip">
          <el-empty description="请在左侧选择因子开始配置策略" />
        </div>
      </div>

      <!-- 右侧：实时 JSON 预览 -->
      <div class="preview-panel">
        <div class="panel-title">策略 JSON 预览</div>
        <pre class="json-preview">{{ jsonPreview }}</pre>
        <el-button class="copy-btn" size="small" @click="copyJson">
          <el-icon><CopyDocument /></el-icon>
          复制 JSON
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ArrowLeft, EditPen, CopyDocument } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  financialFactors,
  technicalFactors,
  sectorFactors,
  timingOptions,
  timingNeedPeriod,
} from '@/data/factors'
import { useStrategyEditor } from '@/composables/useStrategyEditor'

const {
  formData,
  selectedKeys,
  selectedFactors,
  jsonPreview,
  submitting,
  onFactorCheck,
  saveDraft,
  submitStrategy,
  resetForm,
} = useStrategyEditor()

const activeTab = ref('financial')

// 将因子数组转为 el-tree 格式
function factorsToTree(factors: typeof financialFactors) {
  return factors.map((f) => ({
    key: f.key,
    name: f.name,
    // 叶子节点不需要 children，但 el-tree 需要 isLeaf 或空 children
    children: f.params.map((p) => ({ key: `${f.key}__${p.key}`, name: p.label })),
  }))
}

const financialTreeData = factorsToTree(financialFactors)
const technicalTreeData = factorsToTree(technicalFactors)
const sectorTreeData = factorsToTree(sectorFactors)

// 标签切换
function toggleTag(factorKey: string, paramKey: string, value: string) {
  const arr = formData.stock_filter[factorKey]?.[paramKey] || []
  const idx = arr.indexOf(value)
  if (idx >= 0) {
    arr.splice(idx, 1)
  } else {
    arr.push(value)
  }
  if (!formData.stock_filter[factorKey]) {
    formData.stock_filter[factorKey] = {}
  }
  formData.stock_filter[factorKey][paramKey] = [...arr]
}

// 复制 JSON
function copyJson() {
  navigator.clipboard.writeText(jsonPreview.value).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}
</script>

<style scoped>
.strategy-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.header-left h2 {
  margin: 0;
  font-size: 18px;
}
.header-right {
  display: flex;
  gap: 8px;
}

.strategy-name-row {
  max-width: 400px;
}

.editor-body {
  flex: 1;
  display: grid;
  grid-template-columns: 260px 1fr 360px;
  gap: 12px;
  overflow: hidden;
}

.factor-panel,
.form-panel,
.preview-panel {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  overflow-y: auto;
  border: 1px solid #e8e8e8;
}

.panel-title {
  font-weight: 600;
  font-size: 15px;
  margin-bottom: 12px;
  color: #333;
}

.factor-tree {
  font-size: 13px;
}

.factor-section {
  margin-bottom: 8px;
}
.factor-label {
  font-size: 13px;
  color: #666;
  margin-bottom: 4px;
  padding-left: 120px;
}

.tags-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.input-unit {
  margin-left: 8px;
  color: #999;
  font-size: 13px;
}

.empty-tip {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 300px;
}

.json-preview {
  background: #f5f7fa;
  border-radius: 6px;
  padding: 16px;
  font-size: 12px;
  line-height: 1.6;
  overflow: auto;
  max-height: calc(100% - 100px);
  margin: 0;
  color: #333;
}

.copy-btn {
  margin-top: 8px;
}

/* 响应式：窄屏下切换为竖向布局 */
@media (max-width: 1200px) {
  .editor-body {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr auto;
  }
}
</style>
