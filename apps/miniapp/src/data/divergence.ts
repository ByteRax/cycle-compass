// -*- coding: utf-8 -*-
/**
 * 资产与行业同期涨跌分化数据集
 * 支持两类核心场景：
 * 1. 结构性同期涨跌分化（如 红利ETF内部、ETF大类、A股一级行业 1M/3M/6M/1Y）
 * 2. 长期超级周期“同期均上涨，谁在领先？”（如 A股27年超长牛熊领跑 5Y/10Y/20Y/27Y）
 */

// 7 级趋势温度：趋势的绝对体感体系（买卖方与策略核心信号）
export type TrendTemperature = 'boiling' | 'hot' | 'warm' | 'flat' | 'cool' | 'cold' | 'freezing'

export interface TrendTempConfig {
  key: TrendTemperature
  label: string
  nature: string // 预警 | 确认 | 泡沫区 | 超跌区 | 横盘
  actionBuy: string // 买方操作
  actionSell: string // 卖方操作
  color: string
  bg: string
  border: string
  desc: string
}

export const TREND_TEMP_DICT: Record<TrendTemperature, TrendTempConfig> = {
  boiling: {
    key: 'boiling',
    label: '沸',
    nature: '泡沫区',
    actionBuy: '警惕见顶反转，坚决不追买',
    actionSell: '分批止盈离场，或反转做空',
    color: '#FF2E51',
    bg: 'rgba(255, 46, 81, 0.20)',
    border: 'rgba(255, 46, 81, 0.45)',
    desc: '快速拉升、波动放大，极度拥挤'
  },
  hot: {
    key: 'hot',
    label: '热',
    nature: '上涨确认',
    actionBuy: '趋势启动，右侧果断建仓',
    actionSell: '顺势持有，享受主升浪',
    color: '#FF736A',
    bg: 'rgba(255, 115, 106, 0.18)',
    border: 'rgba(255, 115, 106, 0.38)',
    desc: '处于上涨行情，右侧动能稳健'
  },
  warm: {
    key: 'warm',
    label: '温',
    nature: '买点预警',
    actionBuy: '密切监测买点，准备建仓',
    actionSell: '空头减仓防守',
    color: '#E5B352',
    bg: 'rgba(229, 179, 82, 0.18)',
    border: 'rgba(229, 179, 82, 0.38)',
    desc: '盘整蓄势，即将启动上涨'
  },
  flat: {
    key: 'flat',
    label: '平',
    nature: '横盘观望',
    actionBuy: '耐心观望，等待方向选择',
    actionSell: '耐心观望，等待方向选择',
    color: '#94A3B8',
    bg: 'rgba(148, 163, 184, 0.14)',
    border: 'rgba(148, 163, 184, 0.30)',
    desc: '趋势横盘，多空胶着无明显倾向'
  },
  cool: {
    key: 'cool',
    label: '凉',
    nature: '卖点预警',
    actionBuy: '多头提高警惕，收紧止损',
    actionSell: '密切监测卖出机会，准备斩仓',
    color: '#2EC4B6',
    bg: 'rgba(46, 196, 182, 0.18)',
    border: 'rgba(46, 196, 182, 0.38)',
    desc: '转弱盘整，即将启动下跌'
  },
  cold: {
    key: 'cold',
    label: '寒',
    nature: '下跌确认',
    actionBuy: '离场观望，绝不盲目抄底',
    actionSell: '下跌开启，尽快坚决斩仓',
    color: '#38BDF8',
    bg: 'rgba(56, 189, 248, 0.18)',
    border: 'rgba(56, 189, 248, 0.38)',
    desc: '处于下跌行情，空头主导'
  },
  freezing: {
    key: 'freezing',
    label: '冻',
    nature: '超跌区',
    actionBuy: '反转策略左侧分批埋伏做多',
    actionSell: '警惕下跌尾声，避免恐慌割肉',
    color: '#818CF8',
    bg: 'rgba(129, 140, 248, 0.20)',
    border: 'rgba(129, 140, 248, 0.45)',
    desc: '快速下跌、波动放大，极度悲观'
  }
}

export const getTempConfig = (temp?: string): TrendTempConfig => {
  if (!temp) return TREND_TEMP_DICT.flat
  if (temp === 'fair') return TREND_TEMP_DICT.flat
  return TREND_TEMP_DICT[temp as TrendTemperature] || TREND_TEMP_DICT.flat
}

export interface DivergenceItem {
  id: string
  name: string
  short_name?: string
  temp: TrendTemperature | 'warm' | 'fair' | 'cool' // 7级趋势温度 (沸/热/温/平/凉/寒/冻)
  price: number
  price_display?: string
  day_change_pct: number // 短周期为日涨幅，月频下为月涨幅
  turnover_desc: string // e.g. "46亿", "8561万", "36.0万亿"
  total_change_pct: number // 累计涨跌幅，如 +9.9, +1775
  color: string // 走势图线条代表色
  is_benchmark?: boolean // 是否为主要参照物（如 A股基准，红利ETF基准）
  series: number[] // 归一化时序百分比，长度一致，用于画线
}

export interface DivergencePeriodData {
  period_label: string
  timeline: string[]
  benchmark_val: number
  items: DivergenceItem[]
}

export interface DivergenceDataset {
  id: string
  type: 'sub_variety' | 'macro_sector' // 细分品种 | 大类行业
  freq_type?: 'day' | 'month' // 截面频度：日频(默认) | 月频
  default_period?: string
  period_tabs: string[]
  title: string
  subtitle: string
  category_path: string
  as_of_date: string
  period_label: string
  benchmark_name?: string
  benchmark_val?: number
  timeline: string[]
  items: DivergenceItem[]
  periods: Record<string, DivergencePeriodData>
}

// 动态获取今日/最新日期 YYYY/MM/DD
export const getTodayDateStr = (d = new Date()): string => {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}/${m}/${day}`
}

const TODAY = getTodayDateStr()

// 统一的色彩面板（贴合当前项目暗夜风格，兼具高辨识度）
export const DIVERGENCE_COLORS = [
  '#2EC4B6', // 蒂芙尼青绿
  '#E5B352', // 周期金
  '#FF736A', // 珊瑚亮红
  '#A78BFA', // 熏衣紫
  '#38BDF8', // 天空蓝
  '#F472B6', // 樱花粉
  '#FB923C', // 暖橙
  '#4ADE80', // 荧光绿
  '#94A3B8', // 中性灰蓝
  '#CBD5E1'  // 浅白银
]

// 模拟时序节点生成曲线
const makeSeries = (start: number, end: number, midJitter: number[]): number[] => {
  return [start, ...midJitter, end]
}

// ==========================================
// 1. 红利 ETF 细分品种多周期数据
// ==========================================
const DIVIDEND_ITEMS_BASE = [
  { id: '512890', name: '红利低波50ETF南方', short_name: '红利低波50ETF...', price: 1.505, day_change_pct: 0.2, turnover_desc: '4.0亿', color: '#2EC4B6' },
  { id: '512530', name: '300红利低波ETF嘉实', short_name: '300红利低波ET...', price: 1.420, day_change_pct: 0.3, turnover_desc: '1.7亿', color: '#38BDF8' },
  { id: '512890_hb', name: '红利低波ETF华泰柏瑞', short_name: '红利低波ETF华...', price: 1.255, day_change_pct: -0.1, turnover_desc: '8.9亿', color: '#FF736A' },
  { id: '515450', name: '红利低波ETF易方达', short_name: '红利低波ETF易...', price: 1.231, day_change_pct: -0.1, turnover_desc: '3.1亿', color: '#A78BFA' },
  { id: '515100', name: '红利低波100ETF博时', short_name: '红利低波100ET...', price: 1.055, day_change_pct: -0.6, turnover_desc: '8561万', color: '#F472B6' },
  { id: '510880', name: '红利ETF', short_name: '红利ETF', price: 1.333, day_change_pct: -0.2, turnover_desc: '46亿', color: '#E5B352', is_benchmark: true },
  { id: '501029', name: '标普A股红利ETF华宝', short_name: '标普A股红利ET...', price: 0.606, day_change_pct: -0.5, turnover_desc: '7553万', color: '#4ADE80' },
  { id: '510880_ht', name: '红利ETF华泰柏瑞', short_name: '红利ETF华泰柏瑞', price: 3.380, day_change_pct: -0.4, turnover_desc: '5.1亿', color: '#FB923C' },
  { id: '515890', name: '红利ETF易方达', short_name: '红利ETF易方达', price: 1.350, day_change_pct: -0.5, turnover_desc: '5.2亿', color: '#94A3B8' },
  { id: '515080', name: '中证红利ETF招商', short_name: '中证红利ETF招商', price: 1.541, day_change_pct: -0.6, turnover_desc: '4.6亿', color: '#CBD5E1' }
]

const DIVIDEND_PERIODS: Record<string, DivergencePeriodData> = {
  '近1个月': {
    period_label: '近1个月趋势 2026/10-2026/11',
    timeline: ['2026/10/24', '2026/11/04', '2026/11/14', TODAY],
    benchmark_val: 0.9,
    items: [
      { ...DIVIDEND_ITEMS_BASE[0], temp: 'warm', total_change_pct: 2.6, series: makeSeries(0, 2.6, [0.8, 0.4, 1.2, 1.8, 1.5, 2.2]) },
      { ...DIVIDEND_ITEMS_BASE[1], temp: 'warm', total_change_pct: 2.1, series: makeSeries(0, 2.1, [0.6, 0.2, 1.0, 1.5, 1.3, 1.8]) },
      { ...DIVIDEND_ITEMS_BASE[2], temp: 'warm', total_change_pct: 1.5, series: makeSeries(0, 1.5, [0.4, -0.1, 0.7, 1.1, 0.9, 1.3]) },
      { ...DIVIDEND_ITEMS_BASE[3], temp: 'warm', total_change_pct: 1.4, series: makeSeries(0, 1.4, [0.5, 0.0, 0.8, 1.0, 0.8, 1.2]) },
      { ...DIVIDEND_ITEMS_BASE[5], temp: 'fair', total_change_pct: 0.9, series: makeSeries(0, 0.9, [0.2, -0.3, 0.4, 0.7, 0.5, 0.8]) },
      { ...DIVIDEND_ITEMS_BASE[4], temp: 'fair', total_change_pct: 0.8, series: makeSeries(0, 0.8, [0.1, -0.4, 0.3, 0.6, 0.4, 0.7]) },
      { ...DIVIDEND_ITEMS_BASE[6], temp: 'fair', total_change_pct: 0.6, series: makeSeries(0, 0.6, [0.0, -0.5, 0.2, 0.5, 0.3, 0.5]) },
      { ...DIVIDEND_ITEMS_BASE[7], temp: 'cool', total_change_pct: -0.5, series: makeSeries(0, -0.5, [-0.2, -0.8, -0.3, 0.0, -0.2, -0.4]) },
      { ...DIVIDEND_ITEMS_BASE[8], temp: 'cool', total_change_pct: -0.7, series: makeSeries(0, -0.7, [-0.3, -1.0, -0.5, -0.2, -0.4, -0.6]) },
      { ...DIVIDEND_ITEMS_BASE[9], temp: 'cool', total_change_pct: -0.8, series: makeSeries(0, -0.8, [-0.4, -1.2, -0.6, -0.3, -0.5, -0.7]) }
    ]
  },
  '近3个月': {
    period_label: '近3个月趋势 2026/08-2026/11',
    timeline: ['2026/09', '2026/10', '2026/11', TODAY],
    benchmark_val: 2.8,
    items: [
      { ...DIVIDEND_ITEMS_BASE[0], temp: 'warm', total_change_pct: 9.9, series: makeSeries(0, 9.9, [2.1, -1.2, 4.3, 3.5, 6.2, 7.8, 6.9, 8.4, 9.9]) },
      { ...DIVIDEND_ITEMS_BASE[1], temp: 'warm', total_change_pct: 8.7, series: makeSeries(0, 8.7, [1.8, -0.8, 3.5, 2.9, 5.4, 6.8, 6.2, 7.5, 8.7]) },
      { ...DIVIDEND_ITEMS_BASE[2], temp: 'warm', total_change_pct: 5.6, series: makeSeries(0, 5.6, [0.9, -1.5, 2.1, 1.8, 3.6, 4.5, 4.1, 5.0, 5.6]) },
      { ...DIVIDEND_ITEMS_BASE[3], temp: 'warm', total_change_pct: 5.5, series: makeSeries(0, 5.5, [1.1, -1.3, 2.3, 1.6, 3.4, 4.2, 3.9, 4.8, 5.5]) },
      { ...DIVIDEND_ITEMS_BASE[4], temp: 'warm', total_change_pct: 4.1, series: makeSeries(0, 4.1, [0.5, -2.0, 1.5, 1.0, 2.8, 3.4, 3.0, 3.8, 4.1]) },
      { ...DIVIDEND_ITEMS_BASE[5], temp: 'warm', total_change_pct: 2.8, series: makeSeries(0, 2.8, [0.3, -2.5, 0.8, 0.4, 1.9, 2.2, 1.8, 2.4, 2.8]) },
      { ...DIVIDEND_ITEMS_BASE[6], temp: 'fair', total_change_pct: 2.5, series: makeSeries(0, 2.5, [0.2, -2.8, 0.6, 0.2, 1.6, 2.0, 1.6, 2.1, 2.5]) },
      { ...DIVIDEND_ITEMS_BASE[7], temp: 'fair', total_change_pct: -1.7, series: makeSeries(0, -1.7, [-0.5, -4.2, -1.2, -1.8, -0.4, -0.8, -1.2, -1.5, -1.7]) },
      { ...DIVIDEND_ITEMS_BASE[9], temp: 'cool', total_change_pct: -2.9, series: makeSeries(0, -2.9, [-0.8, -5.0, -1.8, -2.5, -1.2, -1.8, -2.2, -2.6, -2.9]) },
      { ...DIVIDEND_ITEMS_BASE[8], temp: 'cool', total_change_pct: -2.9, series: makeSeries(0, -2.9, [-0.9, -5.2, -2.0, -2.8, -1.4, -2.0, -2.4, -2.8, -2.9]) }
    ]
  },
  '近6个月': {
    period_label: '近6个月趋势 2026/05-2026/11',
    timeline: ['2026/05', '2026/07', '2026/09', TODAY],
    benchmark_val: 6.5,
    items: [
      { ...DIVIDEND_ITEMS_BASE[0], temp: 'warm', total_change_pct: 18.2, series: makeSeries(0, 18.2, [4.5, 3.2, 7.8, 9.5, 12.4, 15.2, 16.8]) },
      { ...DIVIDEND_ITEMS_BASE[1], temp: 'warm', total_change_pct: 15.6, series: makeSeries(0, 15.6, [3.8, 2.5, 6.5, 8.2, 10.8, 13.5, 14.8]) },
      { ...DIVIDEND_ITEMS_BASE[2], temp: 'warm', total_change_pct: 11.2, series: makeSeries(0, 11.2, [2.5, 1.2, 4.8, 6.5, 8.2, 9.8, 10.5]) },
      { ...DIVIDEND_ITEMS_BASE[3], temp: 'warm', total_change_pct: 10.8, series: makeSeries(0, 10.8, [2.2, 1.0, 4.5, 6.2, 7.8, 9.5, 10.2]) },
      { ...DIVIDEND_ITEMS_BASE[4], temp: 'warm', total_change_pct: 7.5, series: makeSeries(0, 7.5, [1.5, 0.2, 3.2, 4.8, 5.8, 6.8, 7.2]) },
      { ...DIVIDEND_ITEMS_BASE[5], temp: 'warm', total_change_pct: 6.5, series: makeSeries(0, 6.5, [1.2, -0.2, 2.5, 4.0, 5.0, 5.8, 6.2]) },
      { ...DIVIDEND_ITEMS_BASE[6], temp: 'fair', total_change_pct: 5.2, series: makeSeries(0, 5.2, [0.8, -0.8, 1.8, 3.2, 4.2, 4.8, 5.0]) },
      { ...DIVIDEND_ITEMS_BASE[7], temp: 'fair', total_change_pct: -1.2, series: makeSeries(0, -1.2, [-0.5, -2.5, -0.8, 0.2, -0.5, -1.0, -1.1]) },
      { ...DIVIDEND_ITEMS_BASE[9], temp: 'cool', total_change_pct: -3.5, series: makeSeries(0, -3.5, [-1.0, -3.8, -1.8, -1.0, -2.2, -3.0, -3.2]) },
      { ...DIVIDEND_ITEMS_BASE[8], temp: 'cool', total_change_pct: -3.8, series: makeSeries(0, -3.8, [-1.2, -4.2, -2.0, -1.2, -2.5, -3.2, -3.5]) }
    ]
  },
  '近1年': {
    period_label: '近1年趋势 2025/11-2026/11',
    timeline: ['2025/11', '2026/03', '2026/07', TODAY],
    benchmark_val: 12.4,
    items: [
      { ...DIVIDEND_ITEMS_BASE[0], temp: 'warm', total_change_pct: 26.8, series: makeSeries(0, 26.8, [6.5, 9.8, 14.2, 18.5, 22.0, 25.2]) },
      { ...DIVIDEND_ITEMS_BASE[1], temp: 'warm', total_change_pct: 22.4, series: makeSeries(0, 22.4, [5.2, 8.4, 12.0, 15.6, 18.8, 21.2]) },
      { ...DIVIDEND_ITEMS_BASE[2], temp: 'warm', total_change_pct: 16.5, series: makeSeries(0, 16.5, [3.8, 6.2, 9.5, 12.2, 14.5, 15.8]) },
      { ...DIVIDEND_ITEMS_BASE[3], temp: 'warm', total_change_pct: 15.8, series: makeSeries(0, 15.8, [3.5, 5.8, 9.0, 11.8, 13.8, 15.2]) },
      { ...DIVIDEND_ITEMS_BASE[4], temp: 'warm', total_change_pct: 13.2, series: makeSeries(0, 13.2, [2.8, 4.8, 7.5, 10.0, 11.8, 12.6]) },
      { ...DIVIDEND_ITEMS_BASE[5], temp: 'warm', total_change_pct: 12.4, series: makeSeries(0, 12.4, [2.5, 4.2, 7.0, 9.2, 10.8, 11.8]) },
      { ...DIVIDEND_ITEMS_BASE[6], temp: 'fair', total_change_pct: 9.6, series: makeSeries(0, 9.6, [1.8, 3.2, 5.5, 7.2, 8.5, 9.2]) },
      { ...DIVIDEND_ITEMS_BASE[7], temp: 'fair', total_change_pct: 1.8, series: makeSeries(0, 1.8, [0.2, 1.0, 0.5, 1.2, 1.5, 1.6]) },
      { ...DIVIDEND_ITEMS_BASE[9], temp: 'cool', total_change_pct: -1.5, series: makeSeries(0, -1.5, [-0.5, -1.8, -1.0, -0.8, -1.2, -1.4]) },
      { ...DIVIDEND_ITEMS_BASE[8], temp: 'cool', total_change_pct: -2.0, series: makeSeries(0, -2.0, [-0.8, -2.2, -1.5, -1.2, -1.6, -1.8]) }
    ]
  }
}

// ==========================================
// 2. ETF 基金大类资产分化多周期数据
// ==========================================
const ETF_MACRO_BASE = [
  { id: 'etf_dividend', name: '红利ETF', short_name: '红利ETF', price: 1.333, day_change_pct: -0.2, turnover_desc: '46亿', color: '#E5B352' },
  { id: 'etf_money', name: '货币ETF', short_name: '货币ETF', price: 100.22, day_change_pct: 0.0, turnover_desc: '428亿', color: '#2EC4B6' },
  { id: 'etf_tech_bond', name: '科创债ETF', short_name: '科创债ETF', price: 102.08, day_change_pct: 0.0, turnover_desc: '1319亿', color: '#38BDF8' },
  { id: 'etf_us_stock', name: '美股ETF', short_name: '美股ETF', price: 2.231, day_change_pct: 3.3, turnover_desc: '100亿', color: '#FB923C' },
  { id: 'etf_hs300', name: '沪深300ETF', short_name: '沪深300ETF', price: 2.595, day_change_pct: 0.2, turnover_desc: '152亿', color: '#FF736A', is_benchmark: true },
  { id: 'etf_a500', name: 'A500ETF', short_name: 'A500ETF', price: 1.037, day_change_pct: 0.0, turnover_desc: '241亿', color: '#F472B6' },
  { id: 'etf_gold', name: '黄金商品ETF', short_name: '黄金商品ETF', price: 7.544, day_change_pct: -0.1, turnover_desc: '96亿', color: '#A78BFA' },
  { id: 'etf_star50', name: '科创板ETF', short_name: '科创板ETF', price: 1.275, day_change_pct: 1.5, turnover_desc: '223亿', color: '#4ADE80' },
  { id: 'etf_semi', name: '半导体芯片ETF', short_name: '半导体芯片E...', price: 0.959, day_change_pct: 1.3, turnover_desc: '531亿', color: '#CBD5E1' },
  { id: 'etf_chinext', name: '创业板ETF', short_name: '创业板ETF', price: 1.145, day_change_pct: -0.1, turnover_desc: '266亿', color: '#94A3B8' }
]

const ETF_MACRO_PERIODS: Record<string, DivergencePeriodData> = {
  '近1个月': {
    period_label: '近1个月趋势 2026/10-2026/11',
    timeline: ['2026/10/24', '2026/11/04', '2026/11/14', TODAY],
    benchmark_val: -1.5,
    items: [
      { ...ETF_MACRO_BASE[6], temp: 'warm', total_change_pct: 2.4, series: makeSeries(0, 2.4, [0.8, 0.4, 1.2, 1.8, 1.5, 2.1]) },
      { ...ETF_MACRO_BASE[3], temp: 'warm', total_change_pct: 1.2, series: makeSeries(0, 1.2, [0.5, 0.2, 0.8, 1.0, 0.7, 1.1]) },
      { ...ETF_MACRO_BASE[0], temp: 'fair', total_change_pct: 0.9, series: makeSeries(0, 0.9, [0.3, 0.1, 0.5, 0.7, 0.6, 0.8]) },
      { ...ETF_MACRO_BASE[2], temp: 'fair', total_change_pct: 0.25, series: makeSeries(0, 0.25, [0.05, 0.08, 0.12, 0.18, 0.20, 0.22]) },
      { ...ETF_MACRO_BASE[1], temp: 'fair', total_change_pct: 0.15, series: makeSeries(0, 0.15, [0.03, 0.05, 0.08, 0.10, 0.12, 0.14]) },
      { ...ETF_MACRO_BASE[4], temp: 'cool', total_change_pct: -1.5, series: makeSeries(0, -1.5, [-0.4, -0.8, -1.2, -1.6, -1.3, -1.4]) },
      { ...ETF_MACRO_BASE[5], temp: 'cool', total_change_pct: -1.8, series: makeSeries(0, -1.8, [-0.5, -0.9, -1.4, -1.9, -1.6, -1.7]) },
      { ...ETF_MACRO_BASE[8], temp: 'cool', total_change_pct: -2.8, series: makeSeries(0, -2.8, [-0.8, -1.5, -2.2, -3.1, -2.6, -2.7]) },
      { ...ETF_MACRO_BASE[7], temp: 'cool', total_change_pct: -3.2, series: makeSeries(0, -3.2, [-1.0, -1.8, -2.5, -3.5, -3.0, -3.1]) },
      { ...ETF_MACRO_BASE[9], temp: 'cool', total_change_pct: -4.5, series: makeSeries(0, -4.5, [-1.2, -2.2, -3.4, -4.8, -4.2, -4.4]) }
    ]
  },
  '近3个月': {
    period_label: '近3个月趋势 2026/08-2026/11',
    timeline: ['2026/09', '2026/10', '2026/11', TODAY],
    benchmark_val: -13.0,
    items: [
      { ...ETF_MACRO_BASE[0], temp: 'warm', total_change_pct: 2.8, series: makeSeries(0, 2.8, [1.5, 3.2, 2.4, 3.8, 2.9, 3.5, 2.8]) },
      { ...ETF_MACRO_BASE[1], temp: 'warm', total_change_pct: 0.1, series: makeSeries(0, 0.1, [0.03, 0.05, 0.07, 0.08, 0.09, 0.1, 0.1]) },
      { ...ETF_MACRO_BASE[2], temp: 'fair', total_change_pct: 0.1, series: makeSeries(0, 0.1, [0.02, 0.04, 0.06, 0.07, 0.08, 0.09, 0.1]) },
      { ...ETF_MACRO_BASE[3], temp: 'cool', total_change_pct: -3.0, series: makeSeries(0, -3.0, [0.5, 1.2, -0.8, -2.5, -1.8, -3.5, -3.0]) },
      { ...ETF_MACRO_BASE[4], temp: 'cool', total_change_pct: -13.0, series: makeSeries(0, -13.0, [-2.5, -5.8, -8.2, -14.5, -11.8, -15.2, -13.0]) },
      { ...ETF_MACRO_BASE[5], temp: 'cool', total_change_pct: -16.0, series: makeSeries(0, -16.0, [-3.2, -7.0, -10.5, -17.8, -14.2, -18.5, -16.0]) },
      { ...ETF_MACRO_BASE[6], temp: 'cool', total_change_pct: -23.0, series: makeSeries(0, -23.0, [-4.5, -10.2, -15.8, -25.5, -20.8, -26.2, -23.0]) },
      { ...ETF_MACRO_BASE[7], temp: 'cool', total_change_pct: -23.0, series: makeSeries(0, -23.0, [-4.8, -10.8, -16.2, -26.0, -21.2, -26.8, -23.0]) },
      { ...ETF_MACRO_BASE[8], temp: 'cool', total_change_pct: -28.0, series: makeSeries(0, -28.0, [-5.5, -12.5, -19.5, -31.0, -25.8, -32.5, -28.0]) },
      { ...ETF_MACRO_BASE[9], temp: 'cool', total_change_pct: -34.0, series: makeSeries(0, -34.0, [-6.8, -15.2, -24.0, -38.5, -31.5, -39.0, -34.0]) }
    ]
  },
  '近6个月': {
    period_label: '近6个月趋势 2026/05-2026/11',
    timeline: ['2026/05', '2026/07', '2026/09', TODAY],
    benchmark_val: -6.8,
    items: [
      { ...ETF_MACRO_BASE[6], temp: 'warm', total_change_pct: 14.2, series: makeSeries(0, 14.2, [3.2, 6.5, 9.8, 12.0, 13.5, 14.0]) },
      { ...ETF_MACRO_BASE[0], temp: 'warm', total_change_pct: 6.5, series: makeSeries(0, 6.5, [1.5, 3.0, 4.5, 5.8, 6.2, 6.4]) },
      { ...ETF_MACRO_BASE[3], temp: 'warm', total_change_pct: 3.5, series: makeSeries(0, 3.5, [1.0, 2.2, 1.8, 3.0, 3.2, 3.4]) },
      { ...ETF_MACRO_BASE[2], temp: 'fair', total_change_pct: 1.8, series: makeSeries(0, 1.8, [0.4, 0.8, 1.2, 1.5, 1.6, 1.7]) },
      { ...ETF_MACRO_BASE[1], temp: 'fair', total_change_pct: 0.8, series: makeSeries(0, 0.8, [0.2, 0.4, 0.5, 0.6, 0.7, 0.8]) },
      { ...ETF_MACRO_BASE[4], temp: 'cool', total_change_pct: -6.8, series: makeSeries(0, -6.8, [-1.5, -3.2, -5.0, -8.2, -7.0, -6.9]) },
      { ...ETF_MACRO_BASE[5], temp: 'cool', total_change_pct: -8.5, series: makeSeries(0, -8.5, [-1.8, -4.0, -6.2, -10.0, -8.8, -8.6]) },
      { ...ETF_MACRO_BASE[7], temp: 'cool', total_change_pct: -14.2, series: makeSeries(0, -14.2, [-3.0, -6.8, -10.5, -16.8, -14.8, -14.4]) },
      { ...ETF_MACRO_BASE[8], temp: 'cool', total_change_pct: -18.5, series: makeSeries(0, -18.5, [-4.0, -8.5, -13.2, -21.0, -19.0, -18.6]) },
      { ...ETF_MACRO_BASE[9], temp: 'cool', total_change_pct: -22.0, series: makeSeries(0, -22.0, [-5.0, -10.5, -16.0, -25.2, -22.8, -22.2]) }
    ]
  },
  '近1年': {
    period_label: '近1年趋势 2025/11-2026/11',
    timeline: ['2025/11', '2026/03', '2026/07', TODAY],
    benchmark_val: -8.2,
    items: [
      { ...ETF_MACRO_BASE[6], temp: 'warm', total_change_pct: 28.5, series: makeSeries(0, 28.5, [6.5, 12.0, 18.5, 24.0, 27.2, 28.0]) },
      { ...ETF_MACRO_BASE[3], temp: 'warm', total_change_pct: 18.2, series: makeSeries(0, 18.2, [4.2, 8.5, 12.8, 16.0, 17.5, 18.0]) },
      { ...ETF_MACRO_BASE[0], temp: 'warm', total_change_pct: 12.4, series: makeSeries(0, 12.4, [2.8, 5.5, 8.2, 10.8, 11.9, 12.2]) },
      { ...ETF_MACRO_BASE[2], temp: 'fair', total_change_pct: 4.2, series: makeSeries(0, 4.2, [0.9, 1.8, 2.6, 3.5, 3.9, 4.1]) },
      { ...ETF_MACRO_BASE[1], temp: 'fair', total_change_pct: 2.1, series: makeSeries(0, 2.1, [0.5, 0.9, 1.4, 1.8, 1.9, 2.0]) },
      { ...ETF_MACRO_BASE[4], temp: 'cool', total_change_pct: -8.2, series: makeSeries(0, -8.2, [-1.8, -4.0, -6.5, -9.8, -8.5, -8.3]) },
      { ...ETF_MACRO_BASE[5], temp: 'cool', total_change_pct: -9.8, series: makeSeries(0, -9.8, [-2.2, -4.8, -7.8, -11.5, -10.2, -9.9]) },
      { ...ETF_MACRO_BASE[7], temp: 'cool', total_change_pct: -16.5, series: makeSeries(0, -16.5, [-3.5, -7.8, -12.5, -19.2, -17.0, -16.6]) },
      { ...ETF_MACRO_BASE[8], temp: 'cool', total_change_pct: -21.0, series: makeSeries(0, -21.0, [-4.5, -9.8, -15.8, -24.5, -21.8, -21.2]) },
      { ...ETF_MACRO_BASE[9], temp: 'cool', total_change_pct: -27.8, series: makeSeries(0, -27.8, [-6.0, -13.0, -20.5, -31.8, -28.5, -28.0]) }
    ]
  }
}

// ==========================================
// 3. A 股一级产业分化多周期数据（短周期 1M/3M/6M/1Y）
// ==========================================
const A_SHARES_BASE = [
  { id: 'sec_bank', name: '银行', short_name: '银行', price: 11.52, day_change_pct: 0.7, turnover_desc: '414亿', color: '#FF736A' },
  { id: 'sec_oil_gas', name: '油气能源', short_name: '油气能源', price: 12.05, day_change_pct: -1.8, turnover_desc: '133亿', color: '#E5B352' },
  { id: 'sec_transport', name: '交通运输', short_name: '交通运输...', price: 9.99, day_change_pct: -0.1, turnover_desc: '200亿', color: '#2EC4B6' },
  { id: 'sec_securities', name: '证券保险', short_name: '证券保险', price: 12.88, day_change_pct: 0.1, turnover_desc: '319亿', color: '#38BDF8' },
  { id: 'sec_healthcare', name: '医疗服务', short_name: '医疗服务', price: 41.66, day_change_pct: -1.2, turnover_desc: '183亿', color: '#A78BFA' },
  { id: 'sec_pharma', name: '化学制药', short_name: '化学制药', price: 24.54, day_change_pct: -0.8, turnover_desc: '224亿', color: '#F472B6' },
  { id: 'sec_a_shares', name: 'A股', short_name: 'A股', price: 30.58, day_change_pct: 0.1, turnover_desc: '1.7万亿', color: '#94A3B8', is_benchmark: true },
  { id: 'sec_consumer_elec', name: '消费电子', short_name: '消费电子', price: 36.19, day_change_pct: 3.8, turnover_desc: '571亿', color: '#FB923C' },
  { id: 'sec_semi', name: '半导体', short_name: '半导体', price: 129.74, day_change_pct: 1.8, turnover_desc: '2258亿', color: '#4ADE80' },
  { id: 'sec_comm', name: '通信设备', short_name: '通信设备', price: 67.81, day_change_pct: 0.3, turnover_desc: '1233亿', color: '#CBD5E1' }
]

const A_SHARES_PERIODS: Record<string, DivergencePeriodData> = {
  '近1个月': {
    period_label: '近1个月趋势 2026/10-2026/11',
    timeline: ['2026/10/24', '2026/11/04', '2026/11/14', TODAY],
    benchmark_val: -1.8,
    items: [
      { ...A_SHARES_BASE[0], temp: 'warm', total_change_pct: 3.2, series: makeSeries(0, 3.2, [1.0, 0.5, 1.6, 2.4, 2.0, 2.8]) },
      { ...A_SHARES_BASE[7], temp: 'warm', total_change_pct: 1.5, series: makeSeries(0, 1.5, [0.5, 0.2, 0.9, 1.2, 1.0, 1.4]) },
      { ...A_SHARES_BASE[1], temp: 'warm', total_change_pct: 1.2, series: makeSeries(0, 1.2, [0.4, 0.1, 0.7, 1.0, 0.8, 1.1]) },
      { ...A_SHARES_BASE[2], temp: 'fair', total_change_pct: 0.6, series: makeSeries(0, 0.6, [0.2, -0.1, 0.4, 0.5, 0.4, 0.5]) },
      { ...A_SHARES_BASE[3], temp: 'fair', total_change_pct: -1.0, series: makeSeries(0, -1.0, [-0.3, -0.6, -0.8, -1.2, -0.9, -1.0]) },
      { ...A_SHARES_BASE[6], temp: 'cool', total_change_pct: -1.8, series: makeSeries(0, -1.8, [-0.5, -0.9, -1.4, -2.1, -1.7, -1.8]) },
      { ...A_SHARES_BASE[4], temp: 'cool', total_change_pct: -2.4, series: makeSeries(0, -2.4, [-0.7, -1.3, -1.8, -2.8, -2.3, -2.4]) },
      { ...A_SHARES_BASE[5], temp: 'cool', total_change_pct: -2.6, series: makeSeries(0, -2.6, [-0.8, -1.4, -2.0, -3.0, -2.5, -2.6]) },
      { ...A_SHARES_BASE[8], temp: 'cool', total_change_pct: -3.1, series: makeSeries(0, -3.1, [-0.9, -1.6, -2.3, -3.5, -2.9, -3.0]) },
      { ...A_SHARES_BASE[9], temp: 'cool', total_change_pct: -4.8, series: makeSeries(0, -4.8, [-1.3, -2.4, -3.6, -5.2, -4.5, -4.7]) }
    ]
  },
  '近3个月': {
    period_label: '近3个月趋势 2026/08-2026/11',
    timeline: ['2026/09', '2026/10', '2026/11', TODAY],
    benchmark_val: -12.0,
    items: [
      { ...A_SHARES_BASE[0], temp: 'warm', total_change_pct: 16.0, series: makeSeries(0, 16.0, [3.2, 1.8, 6.8, 5.5, 10.2, 12.8, 14.5, 16.0]) },
      { ...A_SHARES_BASE[1], temp: 'fair', total_change_pct: 4.3, series: makeSeries(0, 4.3, [0.8, -0.5, 1.8, 1.2, 2.8, 3.6, 3.9, 4.3]) },
      { ...A_SHARES_BASE[2], temp: 'warm', total_change_pct: 1.4, series: makeSeries(0, 1.4, [0.3, -0.8, 0.9, 0.5, 1.2, 1.1, 1.3, 1.4]) },
      { ...A_SHARES_BASE[3], temp: 'cool', total_change_pct: -6.2, series: makeSeries(0, -6.2, [-1.2, -3.0, -4.2, -7.5, -5.8, -7.8, -6.2]) },
      { ...A_SHARES_BASE[4], temp: 'cool', total_change_pct: -9.3, series: makeSeries(0, -9.3, [-1.8, -4.2, -6.5, -10.8, -8.5, -11.2, -9.3]) },
      { ...A_SHARES_BASE[5], temp: 'cool', total_change_pct: -12.0, series: makeSeries(0, -12.0, [-2.5, -5.5, -8.1, -13.2, -10.5, -14.2, -12.0]) },
      { ...A_SHARES_BASE[6], temp: 'cool', total_change_pct: -12.0, series: makeSeries(0, -12.0, [-2.8, -6.0, -8.8, -13.8, -11.0, -14.5, -12.0]) },
      { ...A_SHARES_BASE[7], temp: 'cool', total_change_pct: -25.0, series: makeSeries(0, -25.0, [-4.2, -9.5, -16.8, -26.5, -22.1, -28.4, -25.0]) },
      { ...A_SHARES_BASE[8], temp: 'cool', total_change_pct: -26.0, series: makeSeries(0, -26.0, [-4.8, -10.8, -18.2, -28.2, -23.5, -30.1, -26.0]) },
      { ...A_SHARES_BASE[9], temp: 'cool', total_change_pct: -44.0, series: makeSeries(0, -44.0, [-8.5, -18.2, -29.5, -42.1, -38.5, -48.2, -44.0]) }
    ]
  },
  '近6个月': {
    period_label: '近6个月趋势 2026/05-2026/11',
    timeline: ['2026/05', '2026/07', '2026/09', TODAY],
    benchmark_val: -8.5,
    items: [
      { ...A_SHARES_BASE[0], temp: 'warm', total_change_pct: 28.5, series: makeSeries(0, 28.5, [6.5, 12.0, 18.5, 23.2, 26.5, 28.2]) },
      { ...A_SHARES_BASE[1], temp: 'warm', total_change_pct: 11.2, series: makeSeries(0, 11.2, [2.5, 5.0, 7.8, 9.5, 10.8, 11.0]) },
      { ...A_SHARES_BASE[2], temp: 'warm', total_change_pct: 5.8, series: makeSeries(0, 5.8, [1.2, 2.5, 3.8, 4.8, 5.5, 5.7]) },
      { ...A_SHARES_BASE[3], temp: 'fair', total_change_pct: -3.5, series: makeSeries(0, -3.5, [-0.8, -1.8, -2.5, -4.2, -3.8, -3.6]) },
      { ...A_SHARES_BASE[5], temp: 'cool', total_change_pct: -7.2, series: makeSeries(0, -7.2, [-1.5, -3.5, -5.2, -8.5, -7.5, -7.3]) },
      { ...A_SHARES_BASE[6], temp: 'cool', total_change_pct: -8.5, series: makeSeries(0, -8.5, [-1.8, -4.2, -6.5, -10.0, -8.8, -8.6]) },
      { ...A_SHARES_BASE[4], temp: 'cool', total_change_pct: -11.0, series: makeSeries(0, -11.0, [-2.2, -5.2, -8.0, -12.5, -11.5, -11.2]) },
      { ...A_SHARES_BASE[7], temp: 'cool', total_change_pct: -18.2, series: makeSeries(0, -18.2, [-3.8, -8.5, -13.2, -21.0, -19.0, -18.5]) },
      { ...A_SHARES_BASE[8], temp: 'cool', total_change_pct: -19.5, series: makeSeries(0, -19.5, [-4.0, -9.0, -14.0, -22.5, -20.2, -19.8]) },
      { ...A_SHARES_BASE[9], temp: 'cool', total_change_pct: -32.6, series: makeSeries(0, -32.6, [-7.0, -15.5, -23.5, -36.5, -33.8, -33.0]) }
    ]
  },
  '近1年': {
    period_label: '近1年趋势 2025/11-2026/11',
    timeline: ['2025/11', '2026/03', '2026/07', TODAY],
    benchmark_val: -10.4,
    items: [
      { ...A_SHARES_BASE[0], temp: 'warm', total_change_pct: 38.6, series: makeSeries(0, 38.6, [8.5, 16.2, 25.0, 31.8, 36.2, 38.0]) },
      { ...A_SHARES_BASE[1], temp: 'warm', total_change_pct: 18.4, series: makeSeries(0, 18.4, [4.2, 8.5, 12.8, 15.6, 17.5, 18.0]) },
      { ...A_SHARES_BASE[2], temp: 'warm', total_change_pct: 8.2, series: makeSeries(0, 8.2, [1.8, 3.8, 5.8, 7.2, 7.9, 8.1]) },
      { ...A_SHARES_BASE[3], temp: 'fair', total_change_pct: -4.2, series: makeSeries(0, -4.2, [-1.0, -2.2, -3.2, -5.0, -4.5, -4.3]) },
      { ...A_SHARES_BASE[5], temp: 'cool', total_change_pct: -8.8, series: makeSeries(0, -8.8, [-2.0, -4.5, -6.8, -10.2, -9.2, -8.9]) },
      { ...A_SHARES_BASE[6], temp: 'cool', total_change_pct: -10.4, series: makeSeries(0, -10.4, [-2.4, -5.2, -8.0, -12.0, -10.8, -10.5]) },
      { ...A_SHARES_BASE[4], temp: 'cool', total_change_pct: -14.5, series: makeSeries(0, -14.5, [-3.2, -7.0, -10.5, -16.5, -15.0, -14.7]) },
      { ...A_SHARES_BASE[7], temp: 'cool', total_change_pct: -16.8, series: makeSeries(0, -16.8, [-3.5, -7.8, -12.0, -19.2, -17.5, -17.0]) },
      { ...A_SHARES_BASE[8], temp: 'cool', total_change_pct: -22.4, series: makeSeries(0, -22.4, [-4.8, -10.5, -16.2, -25.5, -23.2, -22.8]) },
      { ...A_SHARES_BASE[9], temp: 'cool', total_change_pct: -38.5, series: makeSeries(0, -38.5, [-8.5, -18.2, -27.8, -42.8, -39.5, -38.8]) }
    ]
  }
}

// ==========================================
// 4. A 股长周期“同期均上涨，谁在领先？”（27年超长牛熊领跑）
// 100% 像素级复刻 media_1795534769701_ffaaae5c.jpg
// ==========================================
const A_SHARES_LONG_BASE = [
  { id: 'sec_semi', name: '半导体', short_name: '半导体', price: 129.74, day_change_pct: -1.5, turnover_desc: '4.7万亿', color: '#4ADE80' },
  { id: 'sec_healthcare', name: '医疗服务', short_name: '医疗服务', price: 41.66, day_change_pct: -0.1, turnover_desc: '5441亿', color: '#A78BFA' },
  { id: 'sec_consumer_elec', name: '消费电子', short_name: '消费电子', price: 36.19, day_change_pct: -8.5, turnover_desc: '1.0万亿', color: '#FB923C' },
  { id: 'sec_pharma', name: '化学制药', short_name: '化学制药', price: 24.54, day_change_pct: -4.5, turnover_desc: '6250亿', color: '#F472B6' },
  { id: 'sec_comm', name: '通信设备', short_name: '通信设备', price: 67.81, day_change_pct: -22.0, turnover_desc: '2.3万亿', color: '#CBD5E1' },
  { id: 'sec_bank', name: '银行', short_name: '银行', price: 11.52, day_change_pct: 6.1, turnover_desc: '8364亿', color: '#FF736A' },
  { id: 'sec_oil_gas', name: '油气能源', short_name: '油气能源', price: 12.05, day_change_pct: -5.2, turnover_desc: '3146亿', color: '#E5B352' },
  { id: 'sec_a_shares', name: 'A股', short_name: 'A股', price: 30.58, day_change_pct: -3.7, turnover_desc: '36.0万亿', color: '#94A3B8', is_benchmark: true },
  { id: 'sec_securities', name: '证券保险', short_name: '证券保险', price: 12.88, day_change_pct: -0.9, turnover_desc: '6646亿', color: '#38BDF8' },
  { id: 'sec_transport', name: '交通运输与...', short_name: '交通运输...', price: 9.99, day_change_pct: 1.5, turnover_desc: '4853亿', color: '#2EC4B6' }
]

const A_SHARES_LONG_PERIODS: Record<string, DivergencePeriodData> = {
  '近5年': {
    period_label: '近5年趋势 2021/11-2026/11',
    timeline: ['2021', '2023', '2024', '2025', '2026/11/01'],
    benchmark_val: 18,
    items: [
      { ...A_SHARES_LONG_BASE[5], temp: 'warm', total_change_pct: 68, series: makeSeries(0, 68, [12, 28, 45, 58, 62]) },
      { ...A_SHARES_LONG_BASE[6], temp: 'warm', total_change_pct: 54, series: makeSeries(0, 54, [10, 24, 38, 46, 50]) },
      { ...A_SHARES_LONG_BASE[0], temp: 'fair', total_change_pct: 32, series: makeSeries(0, 32, [45, 20, 15, 28, 30]) },
      { ...A_SHARES_LONG_BASE[7], temp: 'cool', total_change_pct: 18, series: makeSeries(0, 18, [5, -4, 2, 12, 16]) },
      { ...A_SHARES_LONG_BASE[9], temp: 'fair', total_change_pct: 15, series: makeSeries(0, 15, [4, 6, 8, 12, 14]) },
      { ...A_SHARES_LONG_BASE[4], temp: 'fair', total_change_pct: 12, series: makeSeries(0, 12, [-5, 8, 14, 18, 10]) },
      { ...A_SHARES_LONG_BASE[8], temp: 'cool', total_change_pct: 8, series: makeSeries(0, 8, [-2, -8, 0, 6, 7]) },
      { ...A_SHARES_LONG_BASE[2], temp: 'cool', total_change_pct: 5, series: makeSeries(0, 5, [25, 5, -8, 2, 4]) },
      { ...A_SHARES_LONG_BASE[3], temp: 'cool', total_change_pct: -6, series: makeSeries(0, -6, [10, -5, -12, -8, -5]) },
      { ...A_SHARES_LONG_BASE[1], temp: 'cool', total_change_pct: -15, series: makeSeries(0, -15, [20, -10, -22, -18, -16]) }
    ]
  },
  '近10年': {
    period_label: '近10年趋势 2016/11-2026/11',
    timeline: ['2016', '2019', '2021', '2024', '2026/11/01'],
    benchmark_val: 72,
    items: [
      { ...A_SHARES_LONG_BASE[0], temp: 'warm', total_change_pct: 385, series: makeSeries(0, 385, [45, 120, 480, 320, 360]) },
      { ...A_SHARES_LONG_BASE[1], temp: 'warm', total_change_pct: 210, series: makeSeries(0, 210, [35, 95, 310, 180, 200]) },
      { ...A_SHARES_LONG_BASE[2], temp: 'warm', total_change_pct: 185, series: makeSeries(0, 185, [30, 85, 260, 150, 175]) },
      { ...A_SHARES_LONG_BASE[5], temp: 'warm', total_change_pct: 145, series: makeSeries(0, 145, [20, 45, 60, 110, 135]) },
      { ...A_SHARES_LONG_BASE[3], temp: 'fair', total_change_pct: 130, series: makeSeries(0, 130, [25, 60, 180, 115, 125]) },
      { ...A_SHARES_LONG_BASE[4], temp: 'fair', total_change_pct: 115, series: makeSeries(0, 115, [15, 40, 95, 100, 110]) },
      { ...A_SHARES_LONG_BASE[6], temp: 'fair', total_change_pct: 98, series: makeSeries(0, 98, [12, 25, 40, 75, 90]) },
      { ...A_SHARES_LONG_BASE[7], temp: 'cool', total_change_pct: 72, series: makeSeries(0, 72, [10, 30, 85, 55, 68]) },
      { ...A_SHARES_LONG_BASE[8], temp: 'cool', total_change_pct: 58, series: makeSeries(0, 58, [8, 25, 70, 45, 52]) },
      { ...A_SHARES_LONG_BASE[9], temp: 'cool', total_change_pct: 42, series: makeSeries(0, 42, [5, 15, 30, 35, 38]) }
    ]
  },
  '近20年': {
    period_label: '近20年趋势 2006/11-2026/11',
    timeline: ['2006', '2011', '2016', '2021', '2026/11/01'],
    benchmark_val: 186,
    items: [
      { ...A_SHARES_LONG_BASE[0], temp: 'warm', total_change_pct: 950, series: makeSeries(0, 950, [110, 80, 240, 1200, 880]) },
      { ...A_SHARES_LONG_BASE[1], temp: 'warm', total_change_pct: 560, series: makeSeries(0, 560, [95, 120, 280, 750, 520]) },
      { ...A_SHARES_LONG_BASE[2], temp: 'warm', total_change_pct: 420, series: makeSeries(0, 420, [80, 65, 210, 580, 390]) },
      { ...A_SHARES_LONG_BASE[3], temp: 'warm', total_change_pct: 380, series: makeSeries(0, 380, [70, 90, 220, 490, 350]) },
      { ...A_SHARES_LONG_BASE[4], temp: 'warm', total_change_pct: 340, series: makeSeries(0, 340, [60, 55, 180, 390, 320]) },
      { ...A_SHARES_LONG_BASE[5], temp: 'warm', total_change_pct: 260, series: makeSeries(0, 260, [85, 95, 130, 180, 240]) },
      { ...A_SHARES_LONG_BASE[6], temp: 'fair', total_change_pct: 220, series: makeSeries(0, 220, [120, 70, 95, 140, 200]) },
      { ...A_SHARES_LONG_BASE[7], temp: 'cool', total_change_pct: 186, series: makeSeries(0, 186, [140, 60, 110, 210, 175]) },
      { ...A_SHARES_LONG_BASE[8], temp: 'cool', total_change_pct: 165, series: makeSeries(0, 165, [150, 45, 105, 190, 155]) },
      { ...A_SHARES_LONG_BASE[9], temp: 'cool', total_change_pct: 98, series: makeSeries(0, 98, [75, 40, 65, 85, 92]) }
    ]
  },
  '近27年': {
    period_label: '近27年趋势 2000/01-2026/11',
    timeline: ['2001', '2007', '2013', '2019', '2026/11/01'],
    benchmark_val: 303,
    items: [
      {
        ...A_SHARES_LONG_BASE[0],
        temp: 'cool',
        total_change_pct: 1775,
        series: makeSeries(0, 1775, [28, -15, 180, 420, 280, 380, 850, 2150, 1620, 1775])
      },
      {
        ...A_SHARES_LONG_BASE[1],
        temp: 'cool',
        total_change_pct: 965,
        series: makeSeries(0, 965, [15, -8, 120, 280, 240, 320, 580, 1350, 920, 965])
      },
      {
        ...A_SHARES_LONG_BASE[2],
        temp: 'cool',
        total_change_pct: 683,
        series: makeSeries(0, 683, [20, -18, 95, 210, 160, 260, 480, 920, 640, 683])
      },
      {
        ...A_SHARES_LONG_BASE[3],
        temp: 'cool',
        total_change_pct: 606,
        series: makeSeries(0, 606, [12, -10, 110, 240, 210, 270, 420, 810, 580, 606])
      },
      {
        ...A_SHARES_LONG_BASE[4],
        temp: 'cool',
        total_change_pct: 569,
        series: makeSeries(0, 569, [40, -25, 80, 170, 130, 240, 360, 720, 540, 569])
      },
      {
        ...A_SHARES_LONG_BASE[5],
        temp: 'warm',
        total_change_pct: 328,
        series: makeSeries(0, 328, [18, -5, 140, 90, 85, 140, 190, 250, 305, 328])
      },
      {
        ...A_SHARES_LONG_BASE[6],
        temp: 'fair',
        total_change_pct: 312,
        series: makeSeries(0, 312, [25, 5, 220, 95, 75, 110, 150, 220, 290, 312])
      },
      {
        ...A_SHARES_LONG_BASE[7],
        temp: 'cool',
        total_change_pct: 303,
        series: makeSeries(0, 303, [15, -20, 185, 65, 55, 120, 160, 270, 285, 303])
      },
      {
        ...A_SHARES_LONG_BASE[8],
        temp: 'cool',
        total_change_pct: 255,
        series: makeSeries(0, 255, [30, -12, 240, 50, 45, 130, 145, 240, 235, 255])
      },
      {
        ...A_SHARES_LONG_BASE[9],
        temp: 'warm',
        total_change_pct: 144,
        series: makeSeries(0, 144, [10, -15, 125, 40, 35, 75, 95, 120, 135, 144])
      }
    ]
  }
}

// 汇总导出数据集列表
export const DIVERGENCE_DATASETS: Record<string, DivergenceDataset> = {
  dividend_etf: {
    id: 'dividend_etf',
    type: 'sub_variety',
    freq_type: 'day',
    default_period: '近3个月',
    period_tabs: ['近1个月', '近3个月', '近6个月', '近1年'],
    title: '红利ETF, 同期涨跌分化',
    subtitle: '红利主题各ETF产品跟踪分化监测',
    category_path: '金融资产 > ETF基金 > 红利ETF',
    as_of_date: TODAY,
    benchmark_name: '红利ETF',
    period_label: DIVIDEND_PERIODS['近3个月'].period_label,
    benchmark_val: DIVIDEND_PERIODS['近3个月'].benchmark_val,
    timeline: DIVIDEND_PERIODS['近3个月'].timeline,
    items: DIVIDEND_PERIODS['近3个月'].items,
    periods: DIVIDEND_PERIODS
  },
  etf_macro: {
    id: 'etf_macro',
    type: 'macro_sector',
    freq_type: 'day',
    default_period: '近3个月',
    period_tabs: ['近1个月', '近3个月', '近6个月', '近1年'],
    title: 'ETF基金, 同期涨跌分化',
    subtitle: '多资产ETF工具大类分化与资金流向',
    category_path: '金融资产 > ETF基金',
    as_of_date: TODAY,
    benchmark_name: '沪深300ETF',
    period_label: ETF_MACRO_PERIODS['近3个月'].period_label,
    benchmark_val: ETF_MACRO_PERIODS['近3个月'].benchmark_val,
    timeline: ETF_MACRO_PERIODS['近3个月'].timeline,
    items: ETF_MACRO_PERIODS['近3个月'].items,
    periods: ETF_MACRO_PERIODS
  },
  a_shares: {
    id: 'a_shares',
    type: 'macro_sector',
    freq_type: 'day',
    default_period: '近3个月',
    period_tabs: ['近1个月', '近3个月', '近6个月', '近1年'],
    title: 'A股, 同期涨跌分化',
    subtitle: '核心一级产业周期分化与景气截面',
    category_path: '金融资产 > A股',
    as_of_date: TODAY,
    benchmark_name: 'A股',
    period_label: A_SHARES_PERIODS['近3个月'].period_label,
    benchmark_val: A_SHARES_PERIODS['近3个月'].benchmark_val,
    timeline: A_SHARES_PERIODS['近3个月'].timeline,
    items: A_SHARES_PERIODS['近3个月'].items,
    periods: A_SHARES_PERIODS
  },
  a_shares_long: {
    id: 'a_shares_long',
    type: 'macro_sector',
    freq_type: 'month',
    default_period: '近27年',
    period_tabs: ['近5年', '近10年', '近20年', '近27年'],
    title: 'A股，同期均上涨，谁在领先？',
    subtitle: '核心一级产业超长牛熊周期领跑与复合回报',
    category_path: '金融资产 > A股',
    as_of_date: '2026/11/01',
    benchmark_name: 'A股',
    period_label: A_SHARES_LONG_PERIODS['近27年'].period_label,
    benchmark_val: A_SHARES_LONG_PERIODS['近27年'].benchmark_val,
    timeline: A_SHARES_LONG_PERIODS['近27年'].timeline,
    items: A_SHARES_LONG_PERIODS['近27年'].items,
    periods: A_SHARES_LONG_PERIODS
  }
}
