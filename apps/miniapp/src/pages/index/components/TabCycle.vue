<script lang="ts" setup>
import { computed } from 'vue'
import CCard from '@/components/ui/CCard.vue'
import CButton from '@/components/ui/CButton.vue'
import CTag from '@/components/ui/CTag.vue'
import CycleCompass from '@/components/finance/CycleCompass.vue'
import RotationIndicator from '@/components/finance/RotationIndicator.vue'
import StrategyScoreGauge from '@/components/finance/StrategyScoreGauge.vue'
import StrategyHandbook from '@/components/finance/StrategyHandbook.vue'
import BacktestSummary from '@/components/finance/BacktestSummary.vue'
import type { CycleOverview, CombinedSignal, BacktestOutput } from '@/api/cycle'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps<{
  cycleData: CycleOverview
  strategyData: CombinedSignal
  lastUpdatedTime: string
  isBacktesting: boolean
  backtestResult: BacktestOutput | null
}>()

const emit = defineEmits<{
  (e: 'switchTab', tabKey: string): void
  (e: 'runBacktest'): void
}>()

const isValidPositiveNumber = (val: unknown): val is number => {
  return typeof val === 'number' && Number.isFinite(val) && val > 0
}

const dividendPrice = computed(() => {
  const close = props.strategyData?.scoring?.close
  return isValidPositiveNumber(close) ? close : 11878.27
})

const safeRatio = computed(() => {
  const ratio = props.strategyData?.rotation?.ratio
  return isValidPositiveNumber(ratio) ? ratio : 0.1846
})

const growthPrice = computed(() => {
  return Number((dividendPrice.value * safeRatio.value).toFixed(2))
})
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 本周指令行动横幅 -->
    <view class="rounded-xl bg-[#e5b352]/12 border border-[#e5b352]/35 px-3.5 py-2.5 flex items-center justify-between gap-2.5">
      <view class="min-w-0 flex-1">
        <text class="text-[11px] font-bold text-[#e5b352] block">本周指令 · {{ strategyData.rotation.zone }}</text>
        <text class="text-xs text-[#f1efea] leading-snug mt-0.5 block">{{ strategyData.rotation.action }}</text>
      </view>
      <view
        class="shrink-0 rounded-lg bg-[#e5b352] px-3 py-1.5 text-xs font-bold text-[#090c0f] active:opacity-85 cursor-pointer"
        @tap="emit('switchTab', 'portfolio')"
      >
        去打卡
      </view>
    </view>

    <!-- 签名核心：三同心策略体征环 + 精密读数仪表 -->
    <CycleCompass
      :stage-name="cycleData.current_phase"
      :confidence="88"
      :equity-bond-ratio="cycleData.equity_bond_spread"
      inflation-state="温和回升"
      interest-state="流动性充裕"
      :ratio="strategyData.rotation.ratio"
      :percentile1500="strategyData.rotation.quantile_pct"
      :total-score="strategyData.scoring.total"
      :zone-name="strategyData.rotation.zone"
      :action-summary="strategyData.rotation.zone === '中间区' ? '暂存现金池' : strategyData.rotation.zone === '高区' ? '定投红利' : '定投成长'"
    />

    <!-- 资产配置与周期结论 -->
    <CCard title="资产配置与周期结论" :subtitle="`已更新 ${lastUpdatedTime}`">
      <view class="space-y-2.5 text-xs">
        <view class="flex items-center justify-between pb-2 border-b border-white/[0.07]">
          <text class="text-[#8d99a4]">占优核心资产</text>
          <CTag variant="up">{{ cycleData.dominant_asset }}</CTag>
        </view>
        <view class="flex items-center justify-between pb-2 border-b border-white/[0.07]">
          <text class="text-[#8d99a4]">风格主线倾向</text>
          <text class="font-semibold text-[#f1efea]">{{ cycleData.style_preference }}</text>
        </view>
        <view class="p-3 bg-[#181e25] rounded-xl text-[#9aa5b1] leading-relaxed text-[11px] border border-white/[0.06]">
          {{ cycleData.summary }}
        </view>
      </view>
    </CCard>

    <!-- 等比例三区间轮动标尺 -->
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

    <!-- 红利低波100 四维评分 -->
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
      :as-of="strategyData.date"
    />

    <!-- 三区间执行手册与 10 年实测 -->
    <StrategyHandbook />

    <!-- 双策略联合决策与 800 日回测入口 -->
    <CCard title="双策略联合决策" subtitle="轮动区间过滤打分信号，保护资金安全">
      <view class="text-xs space-y-2.5">
        <view class="p-3 bg-[#e5b352]/12 border border-[#e5b352]/30 rounded-xl text-[#f1efea] font-medium leading-relaxed">
          {{ strategyData.combined_action }}
        </view>
        <view class="text-[11px] text-[#8d99a4]">
          {{ strategyData.summary_tips }}
        </view>
        <view class="pt-1">
          <CButton block variant="primary" :loading="isBacktesting" @click="emit('runBacktest')">
            启动全历史回测计算（800 交易日）
          </CButton>
        </view>
      </view>
    </CCard>

    <view v-if="backtestResult" class="space-y-2">
      <BacktestSummary :summary="backtestResult.summary" />
    </view>

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">本程序展示所有公开查询数据，仅供参考，不构成任何投资建议</text>
    </view>
  </view>
</template>
