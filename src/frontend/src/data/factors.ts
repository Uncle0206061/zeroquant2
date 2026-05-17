/**
 * 策略因子数据定义
 * 包含财务指标、技术指标、板块概念三类因子
 */

export type FactorParamType = 'number' | 'select' | 'tags'

export interface FactorOption {
  label: string
  value: string
}

export interface FactorParam {
  key: string
  label: string
  type: FactorParamType
  default?: any
  options?: FactorOption[]
  min?: number
  max?: number
  step?: number
  precision?: number
  unit?: string
}

export interface Factor {
  key: string
  name: string
  category: 'financial' | 'technical' | 'sector'
  params: FactorParam[]
}

// ========== 财务指标 ==========
export const financialFactors: Factor[] = [
  {
    key: 'pe',
    name: '市盈率 PE',
    category: 'financial',
    params: [
      { key: 'min', label: '最小值', type: 'number', default: 0, min: -1000, max: 10000, step: 1 },
      { key: 'max', label: '最大值', type: 'number', default: 50, min: -1000, max: 10000, step: 1 },
    ],
  },
  {
    key: 'pb',
    name: '市净率 PB',
    category: 'financial',
    params: [
      { key: 'min', label: '最小值', type: 'number', default: 0, min: -10, max: 100, step: 0.1, precision: 2 },
      { key: 'max', label: '最大值', type: 'number', default: 5, min: -10, max: 100, step: 0.1, precision: 2 },
    ],
  },
  {
    key: 'market_cap',
    name: '市值（亿元）',
    category: 'financial',
    params: [
      { key: 'min', label: '最小值', type: 'number', default: 0, min: 0, max: 100000, step: 10, unit: '亿' },
      { key: 'max', label: '最大值', type: 'number', default: 10000, min: 0, max: 100000, step: 10, unit: '亿' },
    ],
  },
  {
    key: 'roe',
    name: '净资产收益率 ROE',
    category: 'financial',
    params: [
      { key: 'min', label: '最小值', type: 'number', default: 0, min: -100, max: 100, step: 1, unit: '%' },
    ],
  },
  {
    key: 'revenue_growth',
    name: '营收增长率',
    category: 'financial',
    params: [
      { key: 'min', label: '最小值', type: 'number', default: 0, min: -100, max: 1000, step: 1, unit: '%' },
    ],
  },
]

// ========== 技术指标 ==========
export const technicalFactors: Factor[] = [
  {
    key: 'ma',
    name: 'MA 均线',
    category: 'technical',
    params: [
      { key: 'period', label: '周期', type: 'number', default: 20, min: 5, max: 250, step: 1, unit: '日' },
      {
        key: 'cross_type',
        label: '突破类型',
        type: 'select',
        default: 'golden_cross',
        options: [
          { label: '金叉买入', value: 'golden_cross' },
          { label: '死叉卖出', value: 'dead_cross' },
          { label: '价格突破上方', value: 'break_above' },
          { label: '价格跌破下方', value: 'break_below' },
        ],
      },
    ],
  },
  {
    key: 'macd',
    name: 'MACD',
    category: 'technical',
    params: [
      { key: 'fast', label: '快线周期', type: 'number', default: 12, min: 2, max: 60, step: 1 },
      { key: 'slow', label: '慢线周期', type: 'number', default: 26, min: 5, max: 200, step: 1 },
      { key: 'signal', label: '信号线周期', type: 'number', default: 9, min: 2, max: 60, step: 1 },
    ],
  },
  {
    key: 'rsi',
    name: 'RSI',
    category: 'technical',
    params: [
      { key: 'period', label: '周期', type: 'number', default: 14, min: 2, max: 100, step: 1, unit: '日' },
      { key: 'oversold', label: '超卖阈值', type: 'number', default: 30, min: 0, max: 100, step: 1 },
      { key: 'overbought', label: '超买阈值', type: 'number', default: 70, min: 0, max: 100, step: 1 },
    ],
  },
  {
    key: 'boll',
    name: '布林带',
    category: 'technical',
    params: [
      { key: 'period', label: '周期', type: 'number', default: 20, min: 5, max: 250, step: 1, unit: '日' },
      { key: 'std_dev', label: '标准差倍数', type: 'number', default: 2, min: 0.5, max: 5, step: 0.1, precision: 1 },
    ],
  },
  {
    key: 'volume',
    name: '成交量',
    category: 'technical',
    params: [
      { key: 'ratio', label: '量比阈值', type: 'number', default: 1.5, min: 0.1, max: 20, step: 0.1, precision: 1 },
    ],
  },
]

// ========== 板块概念 ==========
export const sectorFactors: Factor[] = [
  {
    key: 'industry',
    name: '行业板块',
    category: 'sector',
    params: [
      {
        key: 'sectors',
        label: '板块列表',
        type: 'tags',
        default: [],
        options: [
          { label: '银行', value: '银行' },
          { label: '医药', value: '医药' },
          { label: '科技', value: '科技' },
          { label: '消费', value: '消费' },
          { label: '新能源', value: '新能源' },
          { label: '地产', value: '地产' },
          { label: '军工', value: '军工' },
          { label: '半导体', value: '半导体' },
        ],
      },
    ],
  },
  {
    key: 'concept',
    name: '概念题材',
    category: 'sector',
    params: [
      {
        key: 'concepts',
        label: '题材列表',
        type: 'tags',
        default: [],
        options: [
          { label: 'AI', value: 'AI' },
          { label: '芯片', value: '芯片' },
          { label: '机器人', value: '机器人' },
          { label: '新能源汽车', value: '新能源汽车' },
          { label: '光伏', value: '光伏' },
          { label: '锂电', value: '锂电' },
          { label: '数字经济', value: '数字经济' },
          { label: '元宇宙', value: '元宇宙' },
        ],
      },
    ],
  },
]

// ========== 汇总 ==========
export const allFactors: Factor[] = [
  ...financialFactors,
  ...technicalFactors,
  ...sectorFactors,
]

/** 择时条件选项 */
export const timingOptions = [
  { label: 'MA 金叉', value: 'ma_golden_cross' },
  { label: 'MA 死叉', value: 'ma_dead_cross' },
  { label: 'MACD 底背离', value: 'macd_divergence' },
  { label: 'MACD 顶背离', value: 'macd_top_divergence' },
  { label: 'RSI 超卖反弹', value: 'rsi_oversold' },
  { label: '布林带下轨支撑', value: 'boll_support' },
  { label: '放量突破', value: 'volume_breakout' },
]

/** 需要周期参数的择时类型 */
export const timingNeedPeriod = ['ma_golden_cross', 'ma_dead_cross', 'boll_support', 'volume_breakout']
