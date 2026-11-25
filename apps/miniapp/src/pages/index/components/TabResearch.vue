<script lang="ts" setup>
import type { BacktestOutput, CombinedSignal } from '@/api/cycle'
import type { FavoriteChartItem } from '@/stores/finance'
import { computed, ref } from 'vue'
import BacktestSummary from '@/components/finance/BacktestSummary.vue'
import RotationIndicator from '@/components/finance/RotationIndicator.vue'
import StrategyHandbook from '@/components/finance/StrategyHandbook.vue'
import StrategyScoreGauge from '@/components/finance/StrategyScoreGauge.vue'
import CButton from '@/components/ui/CButton.vue'
import CCard from '@/components/ui/CCard.vue'
import { useFinanceStore } from '@/stores/finance'
import { buildFinanceUrl } from '@/utils/financeRoute'
import { SHARED_ROTATION_RULES } from '@/utils/strategyRules'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps<{
  strategyData: CombinedSignal
  isBacktesting: boolean
  backtestResult: BacktestOutput | null
  backtestError: string
  isSampleResult: boolean
}>()

const emit = defineEmits<{
  (e: 'runBacktest', params: { low_line: number, high_line: number, weekly_amount: number, days: number }): void
  (e: 'showSampleBacktest'): void
}>()

const financeStore = useFinanceStore()

// 可配置回测表单参数
const lowLineInput = ref<string>(String(SHARED_ROTATION_RULES.lowThreshold))
const highLineInput = ref<string>(String(SHARED_ROTATION_RULES.highThreshold))
const weeklyAmountInput = ref<string>('1000')
const selectedDays = ref<number>(800)
const formValidationError = ref<string>('')

const dayPresets = [250, 500, 800, 1200] as const

const recentCharts = computed(() => financeStore.favoriteCharts.slice(0, 4))

function isValidPositiveNumber(val: unknown): val is number {
  return typeof val === 'number' && Number.isFinite(val) && val > 0
}

const dividendPrice = computed(() => {
  const close = props.strategyData?.scoring?.close
  return isValidPositiveNumber(close)
? close
: 11878.27
})

const safeRatio = computed(() => {
  const ratio = props.strategyData?.rotation?.ratio
  return isValidPositiveNumber(ratio)
? ratio
: 0.1846
})

const growthPrice = computed(() => Number((dividendPrice.value * safeRatio.value).toFixed(2)))

function submitCustomBacktest() {
  formValidationError.value = ''
  const lowLine = Number(lowLineInput.value)
  const highLine = Number(highLineInput.value)
  const weeklyAmount = Number(weeklyAmountInput.value)

  if (!Number.isFinite(lowLine) || !Number.isFinite(highLine) || lowLine <= 0 || lowLine >= highLine) {
    formValidationError.value = '阈值设置无效：请确保 0 < 低区线 < 高区线'
    return
  }
  if (!Number.isFinite(weeklyAmount) || weeklyAmount < 100 || weeklyAmount > 100000) {
    formValidationError.value = '每周定投金额建议在 100 ~ 100,000 元之间'
    return
  }
  emit('runBacktest', {
    low_line: lowLine,
    high_line: highLine,
    weekly_amount: weeklyAmount,
    days: selectedDays.value
  })
}

function openWorkbench(chart = 'curve') {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: financeStore.defaultHomePool || 'universe_macro',
      scope_id: financeStore.defaultHomePool || 'universe_macro',
      chart: chart as any,
      currency: financeStore.defaultCurrency,
      range: financeStore.defaultRange
    })
  })
}

function openMatrixView() {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/index', {
      scope_id: financeStore.defaultHomePool || 'universe_macro',
      view_mode: 'matrix',
      currency: financeStore.defaultCurrency
    })
  })
}

function openFavorites() {
  uni.navigateTo({ url: '/pages-sub/finance/favorites' })
}

function restoreChart(item: FavoriteChartItem) {
  const isFixed = item.snapshot_mode === 'fixed' && Boolean(item.snapshot_id)
  const target = (item.chart === 'matrix' || item.context?.view_mode === 'quotes')
    ? '/pages-sub/finance/index'
    : '/pages-sub/divergence/index'
  uni.navigateTo({
    url: buildFinanceUrl(target, {
      asset_id: item.asset_id,
      scope_id: item.context?.scope_id || item.context?.pool_id || financeStore.defaultHomePool || 'universe_macro',
      chart: item.chart,
      range: (item.context?.range as any) || financeStore.defaultRange,
      currency: (item.context?.currency as any) || financeStore.defaultCurrency,
      benchmark_id: item.context?.benchmark_id || '',
      snapshot_mode: isFixed ? 'fixed' : 'latest',
      snapshot_id: isFixed
? (item.snapshot_id || '')
: ''
    })
  })
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 研究工作台与矩阵入口 -->
    <view class="grid grid-cols-2 gap-2.5">
      <view
        class="rounded-2xl bg-[#12161b] border border-[#e5b352]/30 p-3 active:opacity-80 cursor-pointer"
        @tap="openWorkbench('curve')"
      >
        <text class="text-sm font-bold text-[#e5b352] block">趋势分析工作台</text>
        <text class="text-[11px] text-[#8d99a4] mt-1 block">七类深度研究图表 · 节气与分布</text>
      </view>
      <view
        class="rounded-2xl bg-[#12161b] border border-white/[0.1] p-3 active:opacity-80 cursor-pointer"
        @tap="openMatrixView"
      >
        <text class="text-sm font-bold text-[#f1efea] block">周期收益矩阵</text>
        <text class="text-[11px] text-[#8d99a4] mt-1 block">WTD/MTD/QTD/YTD 全池对比</text>
      </view>
    </view>

    <!-- 最近研究 / 收藏图表 -->
    <view class="rounded-2xl bg-[#12161b] border border-white/[0.09] p-3.5 space-y-2">
      <view class="flex items-center justify-between">
        <text class="text-sm font-bold text-[#f1efea]">最近收藏的研究视角</text>
        <text class="text-[11px] text-[#e5b352] cursor-pointer" @tap="openFavorites">
          全部收藏 ({{ financeStore.favoriteCharts.length }}) ›
        </text>
      </view>
      <view v-if="!recentCharts.length" class="rounded-xl bg-[#181e25] p-3 text-[11px] text-[#8d99a4]">
        暂无已保存图表视角；可在工作台或周期矩阵中点击“收藏视图”锁定当前参数或历史快照。
      </view>
      <view v-else class="space-y-1.5">
        <view
          v-for="item in recentCharts"
          :key="item.id"
          class="rounded-xl bg-[#181e25] border border-white/[0.06] px-3 py-2 flex items-center justify-between gap-2 active:opacity-80"
          @tap="restoreChart(item)"
        >
          <view class="min-w-0 flex-1">
            <text class="text-xs font-bold text-[#f1efea] truncate block">
              {{ item.title || `${item.asset_name} · ${item.chart}` }}
            </text>
            <text class="text-[10px] text-[#8d99a4] block mt-0.5">
              {{ item.snapshot_mode === 'fixed' ? '锁定快照' : '每日跟踪' }} · {{ item.context?.currency || 'CNY' }}
            </text>
          </view>
          <text class="text-xs text-[#e5b352] shrink-0">还原 ›</text>
        </view>
      </view>
    </view>

    <!-- 三区间轮动标尺 -->
    <RotationIndicator
      :ratio="safeRatio"
      :growth-price="growthPrice"
      :dividend-price="dividendPrice"
      :zone="strategyData.rotation.zone === '低区' ? 'ZONE_LOW' : strategyData.rotation.zone === '高区' ? 'ZONE_HIGH' : 'ZONE_MID'"
      :zone-name="strategyData.rotation.zone"
      :percentile1500="strategyData.rotation.quantile_pct / 100"
      :to-high-pct="strategyData.rotation.to_high_pct"
      :to-low-pct="strategyData.rotation.to_low_pct"
      :action-text="strategyData.rotation.action"
      :as-of="strategyData.date"
    />

    <!-- 红利低波100 真实因子评分证据 -->
    <StrategyScoreGauge
      :total-score="strategyData.scoring.total"
      :tech-score="strategyData.scoring.tech"
      :spread-score="strategyData.scoring.spread"
      :valuation-score="strategyData.scoring.value"
      :style-score="strategyData.scoring.style"
      :multiplier-text="strategyData.scoring.action"
      :gated="strategyData.scoring.total <= 6"
      :effective-amount="strategyData.scoring.amount"
      :close-price="strategyData.scoring.close"
      :pe="strategyData.scoring.pe"
      :pe-percentile="strategyData.scoring.pe_percentile"
      :dividend-yield="strategyData.scoring.dividend_yield"
      :bond10y="strategyData.scoring.bond10y"
      :spread-p-p="strategyData.scoring.spread_pp"
      :rsi70="strategyData.scoring.rsi70"
      :b-pct100="strategyData.scoring.b_pct100"
      :style-percentile="strategyData.scoring.style_percentile"
      :data-source-note="strategyData.scoring.data_source_note"
      :rule-summary="strategyData.scoring.rule_summary"
      :has-stale-input="strategyData.scoring.has_stale_input"
      :factors="strategyData.scoring.factors"
      :as-of="strategyData.scoring.as_of || strategyData.date"
    />

    <!-- 规则说明书 -->
    <StrategyHandbook />

    <!-- 共用规则的可配置历史回测模块 -->
    <CCard
      title="双策略历史回测实验室"
      :subtitle="`规则 ${SHARED_ROTATION_RULES.version} · 与账本严格共口径（离开中间区于次周首周触发）`"
    >
      <view class="text-xs space-y-3">
        <view class="grid grid-cols-3 gap-2">
          <view class="rounded-xl bg-[#181e25] border border-white/[0.08] p-2">
            <text class="text-[10px] text-[#8d99a4] block">低区线</text>
            <input
              v-model="lowLineInput"
              type="digit"
              class="text-sm font-mono font-bold text-[#e5b352] mt-0.5"
            >
          </view>
          <view class="rounded-xl bg-[#181e25] border border-white/[0.08] p-2">
            <text class="text-[10px] text-[#8d99a4] block">高区线</text>
            <input
              v-model="highLineInput"
              type="digit"
              class="text-sm font-mono font-bold text-[#e5b352] mt-0.5"
            >
          </view>
          <view class="rounded-xl bg-[#181e25] border border-white/[0.08] p-2">
            <text class="text-[10px] text-[#8d99a4] block">周定投(元)</text>
            <input
              v-model="weeklyAmountInput"
              type="number"
              class="text-sm font-mono font-bold text-[#f1efea] mt-0.5"
            >
          </view>
        </view>

        <!-- 回测交易日数 -->
        <view class="flex items-center gap-2 flex-wrap">
          <text class="text-[11px] text-[#8d99a4]">测试窗口：</text>
          <view
            v-for="d in dayPresets"
            :key="d"
            class="px-2.5 py-1 rounded-lg border text-[11px] font-mono cursor-pointer"
            :class="selectedDays === d
              ? 'bg-[#e5b352]/18 border-[#e5b352] text-[#e5b352] font-bold'
              : 'bg-[#181e25] border-white/[0.08] text-[#9aa5b1]'"
            @tap="selectedDays = d"
          >
            {{ d }} 日
          </view>
        </view>

        <text v-if="formValidationError" class="text-[11px] text-[#ff736a] block">
          {{ formValidationError }}
        </text>

        <CButton block variant="primary" :loading="isBacktesting" @click="submitCustomBacktest">
          运行实时策略回测（{{ selectedDays }} 交易日）
        </CButton>

        <!-- 回测报错绝不静默伪造：提供重试与单列的静态历史样例入口 -->
        <view
          v-if="backtestError"
          class="rounded-xl border border-[#f05a50]/40 bg-[#f05a50]/12 p-3 space-y-2"
        >
          <text class="text-xs font-bold text-[#ff736a] block">实时回测请求失败</text>
          <text class="text-[11px] text-[#d1d5db] block leading-relaxed">{{ backtestError }}</text>
          <view class="flex items-center gap-2 pt-0.5">
            <view
              class="px-3 py-1.5 rounded-lg border border-[#f05a50] text-[11px] font-bold text-[#ff736a]"
              @tap="submitCustomBacktest"
            >
              重新请求
            </view>
            <view
              class="px-3 py-1.5 rounded-lg border border-[#e5b352]/50 text-[11px] text-[#e5b352]"
              @tap="emit('showSampleBacktest')"
            >
              查看带日期历史样例 (2023-08~2026-11)
            </view>
          </view>
        </view>
      </view>
    </CCard>

    <view v-if="backtestResult" class="space-y-2">
      <BacktestSummary :summary="backtestResult.summary" :is-sample="isSampleResult" />
    </view>

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">本程序展示所有公开查询数据，仅供参考，不构成任何投资建议</text>
    </view>
  </view>
</template>
