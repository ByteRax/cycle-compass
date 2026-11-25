<script lang="ts" setup>
import { computed } from 'vue'

export interface BacktestSummaryData {
  start_date?: string
  end_date?: string
  trading_days?: number
  total_invest?: number
  final_asset?: number
  net_profit?: number
  total_return_pct?: number
  annualized_return?: number
  xirr?: number
  max_drawdown?: number
  sharpe_ratio?: number
  calmar_ratio?: number
  zone_distribution?: Record<string, number>
  total_rebuy_count?: number
  ledger_entries_count?: number
  rule_version?: string
  data_source?: string
  execution_timing?: string
  clip_reason?: string
}

interface Props {
  summary?: BacktestSummaryData
  isSample?: boolean
  totalInvest?: number
  finalAsset?: number
  netProfit?: number
  totalReturn?: number
  annualizedReturn?: number
  xirr?: number
  maxDrawdown?: number
  sharpeRatio?: number
  calmarRatio?: number
  rebuyCount?: number
  tradingDays?: number
  lowZoneDays?: number
  midZoneDays?: number
  highZoneDays?: number
}

const props = withDefaults(defineProps<Props>(), {
  summary: undefined,
  isSample: false,
  totalInvest: undefined,
  finalAsset: undefined,
  netProfit: undefined,
  totalReturn: undefined,
  annualizedReturn: undefined,
  xirr: undefined,
  maxDrawdown: undefined,
  sharpeRatio: undefined,
  calmarRatio: undefined,
  rebuyCount: undefined,
  tradingDays: undefined,
  lowZoneDays: undefined,
  midZoneDays: undefined,
  highZoneDays: undefined
})

const displayTotalInvest = computed(() =>
  Number(props.summary?.total_invest ?? props.totalInvest ?? 165000).toLocaleString('zh-CN', { maximumFractionDigits: 2 })
)
const displayFinalAsset = computed(() =>
  Number(props.summary?.final_asset ?? props.finalAsset ?? 186016.47).toLocaleString('zh-CN', { maximumFractionDigits: 2 })
)
const displayNetProfit = computed(() => {
  const num = Number(props.summary?.net_profit ?? props.netProfit ?? 21016.47)
  return `${num >= 0
? '+'
: ''}¥${num.toLocaleString('zh-CN', { maximumFractionDigits: 2 })}`
})
const displayReturnPct = computed(() => {
  const num = Number(props.summary?.total_return_pct ?? props.totalReturn ?? 12.74)
  return `${num >= 0
? '+'
: ''}${num.toFixed(2)}%`
})
const displayXirr = computed(() => `${Number(props.summary?.xirr ?? props.xirr ?? 7.39).toFixed(2)}%`)
const displayMaxDrawdown = computed(() => `-${Math.abs(Number(props.summary?.max_drawdown ?? props.maxDrawdown ?? 24.94)).toFixed(2)}%`)
const displaySharpe = computed(() => Number(props.summary?.sharpe_ratio ?? props.sharpeRatio ?? 2.42).toFixed(2))
const displayTradingDays = computed(() => props.summary?.trading_days ?? props.tradingDays ?? 800)
const displayDateRange = computed(() => `${props.summary?.start_date ?? '2023-08-02'} ~ ${props.summary?.end_date ?? '2026-11-20'}`)
const displayRebuyCount = computed(() => props.summary?.total_rebuy_count ?? props.rebuyCount ?? 6)

const lowDays = computed(() => props.summary?.zone_distribution?.['低区'] ?? props.lowZoneDays ?? 163)
const midDays = computed(() => props.summary?.zone_distribution?.['中间区'] ?? props.midZoneDays ?? 435)
const highDays = computed(() => props.summary?.zone_distribution?.['高区'] ?? props.highZoneDays ?? 202)
const totalDays = computed(() => Math.max(1, lowDays.value + midDays.value + highDays.value))

const lowPct = computed(() => ((lowDays.value / totalDays.value) * 100).toFixed(1))
const midPct = computed(() => ((midDays.value / totalDays.value) * 100).toFixed(1))
const highPct = computed(() => ((highDays.value / totalDays.value) * 100).toFixed(1))
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-3.5">
    <view class="flex items-center justify-between gap-2">
      <view>
        <view class="flex items-center gap-1.5">
          <text class="text-sm font-bold text-[#f1efea] block">
            {{ isSample ? '历史回测参考样例' : '全历史回测实战成绩单' }}
          </text>
          <text
            v-if="isSample"
            class="text-[9px] px-1.5 py-0.5 rounded bg-[#e5b352]/20 text-[#e5b352] font-medium"
          >
            静态样例
          </text>
        </view>
        <text class="text-[11px] font-mono text-[#8d99a4] mt-0.5 block">{{ displayDateRange }}</text>
      </view>
      <view class="rounded-full bg-[#e5b352]/18 px-2.5 py-0.5 text-[11px] font-mono font-bold text-[#e5b352] shrink-0">
        {{ displayTradingDays }} 交易日
      </view>
    </view>

    <!-- 回测执行规则与截取原因说明 -->
    <view
      v-if="summary?.rule_version || summary?.execution_timing || summary?.data_source || summary?.clip_reason"
      class="rounded-xl bg-[#181e25] border border-white/[0.06] px-3 py-2 text-[10px] text-[#8d99a4] space-y-1"
    >
      <view v-if="summary?.rule_version || summary?.execution_timing" class="flex items-center justify-between gap-2 flex-wrap">
        <text v-if="summary?.rule_version" class="font-mono text-[#e5b352]">规则 {{ summary.rule_version }}</text>
        <text v-if="summary?.execution_timing">{{ summary.execution_timing }}</text>
      </view>
      <text v-if="summary?.data_source" class="block">数据来源：{{ summary.data_source }}</text>
      <text v-if="summary?.clip_reason" class="block text-[#e5b352]">区间说明：{{ summary.clip_reason }}</text>
    </view>

    <!-- 核心资产双卡 -->
    <view class="grid grid-cols-2 gap-2.5">
      <view class="p-3 bg-[#181e25] rounded-xl border border-white/[0.06]">
        <text class="text-[11px] text-[#8d99a4]">累计定投本金</text>
        <view class="mt-1 text-lg font-bold text-[#f1efea] font-mono cc-tabular">¥{{ displayTotalInvest }}</view>
      </view>
      <view class="p-3 bg-[#181e25] rounded-xl border border-[#e5b352]/25">
        <text class="text-[11px] text-[#e5b352]">期末总资产</text>
        <view class="mt-1 text-lg font-bold text-[#e5b352] font-mono cc-tabular">¥{{ displayFinalAsset }}</view>
      </view>
    </view>

    <!-- 净盈利横幅 -->
    <view class="px-3 py-2.5 bg-[#f05a50]/14 border border-[#f05a50]/30 rounded-xl flex items-center justify-between text-xs">
      <view class="flex items-center space-x-1.5">
        <text class="text-[#ff736a]">净盈利</text>
        <text class="font-bold text-[#ff736a] font-mono cc-tabular">{{ displayNetProfit }}</text>
      </view>
      <view class="flex items-center space-x-1.5">
        <text class="text-[#9aa5b1]">累计回报</text>
        <text class="font-bold text-[#ff736a] font-mono cc-tabular">{{ displayReturnPct }}</text>
      </view>
    </view>

    <!-- 3 列发丝线指标 -->
    <view class="grid grid-cols-3 py-2.5 border-t border-b border-white/[0.08] text-center">
      <view>
        <text class="text-[10px] text-[#8d99a4] block">含息年化 XIRR</text>
        <text class="text-base font-bold text-[#e5b352] font-mono cc-tabular mt-0.5 block">{{ displayXirr }}</text>
      </view>
      <view class="border-l border-r border-white/[0.08]">
        <text class="text-[10px] text-[#8d99a4] block">最大回撤</text>
        <text class="text-base font-bold text-[#2ad1a3] font-mono cc-tabular mt-0.5 block">{{ displayMaxDrawdown }}</text>
      </view>
      <view>
        <text class="text-[10px] text-[#8d99a4] block">夏普比率</text>
        <text class="text-base font-bold text-[#f1efea] font-mono cc-tabular mt-0.5 block">{{ displaySharpe }}</text>
      </view>
    </view>

    <!-- 三区间分布条 -->
    <view class="space-y-2">
      <view class="flex items-center justify-between text-xs">
        <text class="font-semibold text-[#f1efea]">三区间交易日分布</text>
        <text class="text-[11px] font-mono text-[#8d99a4]">回补触发 {{ displayRebuyCount }} 次</text>
      </view>
      <view class="h-2.5 w-full bg-white/[0.08] rounded-full overflow-hidden flex">
        <view class="bg-[#2ad1a3] h-full" :style="{ width: `${lowPct}%` }" />
        <view class="bg-[#e5b352] h-full" :style="{ width: `${midPct}%` }" />
        <view class="bg-[#f05a50] h-full" :style="{ width: `${highPct}%` }" />
      </view>
      <view class="grid grid-cols-3 gap-1.5 text-[10px] font-mono text-center">
        <view class="p-1.5 bg-[#2ad1a3]/12 rounded-lg text-[#2ad1a3]">低区 {{ lowDays }}日</view>
        <view class="p-1.5 bg-[#e5b352]/14 rounded-lg text-[#e5b352]">中间区 {{ midDays }}日</view>
        <view class="p-1.5 bg-[#f05a50]/14 rounded-lg text-[#ff736a]">高区 {{ highDays }}日</view>
      </view>
    </view>
  </view>
</template>
