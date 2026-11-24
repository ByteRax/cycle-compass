<script lang="ts" setup>
import type { BacktestOutput, CombinedSignal, CycleOverview } from '@/api/cycle'
import { computed, ref } from 'vue'
import BacktestSummary from '@/components/finance/BacktestSummary.vue'
import CycleCompass from '@/components/finance/CycleCompass.vue'
import RotationIndicator from '@/components/finance/RotationIndicator.vue'
import StrategyHandbook from '@/components/finance/StrategyHandbook.vue'
import StrategyScoreGauge from '@/components/finance/StrategyScoreGauge.vue'
import CButton from '@/components/ui/CButton.vue'
import CCard from '@/components/ui/CCard.vue'
import CTag from '@/components/ui/CTag.vue'

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

function isValidPositiveNumber (val: unknown): val is number {
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

const growthPrice = computed(() => {
  return Number((dividendPrice.value * safeRatio.value).toFixed(2))
})

// ==================== 资产配置与周期结论：小白大白话弹窗 ====================
const activeCycleTooltipKey = ref<string | null>(null)

interface CycleTooltipItem {
  title: string
  currentVal: string
  calc: string
  eli5: string
  action: string
}

const cycleTooltips = computed<Record<string, CycleTooltipItem>>(() => ({
  overview: {
    title: '资产配置与周期结论：大白话拆解',
    currentVal: `${props.cycleData.dominant_asset} · ${props.cycleData.style_preference}`,
    calc: '基于美林时钟弱复苏判定 + 股债利差 ERP 2.8097%（偏向股票） + 创50(代码 399673)/红利低波(代码 930955)轮动比值 0.1846（处在 24.3% 低分位但尚未跌破 0.16 极限低区）三维联动推导。',
    eli5: '这是你资产的“排兵布阵总指挥图”。告诉你目前在战场上应该用什么阵型：防守盾牌（红利低波100 代码 930955）举好，同时盯紧长矛（创业板50 代码 399673 低区），不盲目出击。',
    action: '坚守底仓分红防御，留好现金池弹药，静待成长股跌出大甩卖黄金坑。'
  },
  dominant_asset: {
    title: '占优核心资产：红利高股息与海外AI',
    currentVal: props.cycleData.dominant_asset,
    calc: '1. 国内十年期国债收益率低至 1.58%，红利低波100指数(代码 930955 / ETF 515100)股息率 4.39%，股债利差高达 2.81%（> 2.5%），高分红股票收益确定性远超存款与国债；2. 海外 AI 核心产业处于高景气周期。',
    eli5: '“当前战场上最能打的两个主力部队”！一个叫“抗揍稳拿现金流”（红利低波100 代码 930955 / ETF 515100），吃分红；另一个叫“海外赚钱机器”（海外AI），吃全球产业增长。',
    action: '稳健底仓坚守高分红红利资产，不碰无业绩、纯炒作的题材垃圾股。'
  },
  style_preference: {
    title: '风格主线倾向：红利防守为主，低区伺机成长',
    currentVal: props.cycleData.style_preference,
    calc: '创50(代码 399673) ÷ 红利低波(代码 930955) 轮动比值 0.1846 位于中间区（0.16 ~ 0.28），距离低区线（< 0.16）还有 13.3% 的下行缓冲空间。策略纪律严禁在未入低区时过早重仓接飞刀。',
    eli5: '“盾牌举好，长矛准备”！现在市场还没到跳楼价，红利盾牌（红利低波100 代码 930955 / ETF 515100）不能丢；等到创业板50（代码 399673 / ETF 159949）小伙跌到 0.16 以下大减价，再把长矛刺出去买成长股。',
    action: '保持定力管住手，不盲目抢跑抄底，继续把定投资金存进现金池，等低区信号触发再行动。'
  },
  summary: {
    title: '宏观周期研判依据深度拆解',
    currentVal: '十年期国债低位 · 利差极具性价比 · 处于24%低分位',
    calc: '1. 股债利差 2.8097%（红利低波100 代码 930955 股息率 4.39% - 十年期国债 1.58%）；2. 1500日创50(399673)/红利比值分位 24.33%（偏便宜区间）；3. 央行流动性充裕 + 实体通胀温和回升。',
    eli5: '大白话总结：银行存款利息太低，买高股息股票更划算；整个市场又处于过去6年偏便宜的底部，资金面水龙头开得很大，大方向不用慌，只需等买点。',
    action: '不要被短期市场震荡吓得割肉，按策略节奏存子弹，做好资金规划。'
  }
}))

const activeCycleTooltip = computed(() => {
  return activeCycleTooltipKey.value ? cycleTooltips.value[activeCycleTooltipKey.value] : null
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
      :data-time="`${strategyData.rotation.date || '2026-11-20'} 15:30`"
    />

    <!-- 资产配置与周期结论（升级为支持小白释义、推导算法与极简时间） -->
    <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-3">
      <!-- 头部标题 + 💡小白释义 + 时间 -->
      <view class="flex items-start justify-between gap-2 pb-2.5 border-b border-white/[0.08]">
        <view class="min-w-0 flex-1">
          <view class="flex items-center space-x-1.5 cursor-pointer active:opacity-75" @tap="activeCycleTooltipKey = 'overview'">
            <text class="text-sm font-bold tracking-tight text-[#f1efea]">资产配置与周期结论</text>
            <text class="text-[10px] text-[#e5b352] bg-[#e5b352]/15 px-1.5 py-0.5 rounded-full font-medium">💡小白释义</text>
          </view>
          <view class="flex items-center space-x-1.5 mt-1 text-[10px] text-[#8d99a4] font-mono">
            <text class="text-[#e5b352]">📅</text>
            <text>{{ strategyData.rotation.date || '2026-11-20' }} 15:30</text>
          </view>
        </view>
      </view>

      <!-- 核心指标行 -->
      <view class="space-y-2.5 text-xs">
        <view
          class="flex items-center justify-between pb-2 border-b border-white/[0.07] cursor-pointer active:opacity-75"
          @tap="activeCycleTooltipKey = 'dominant_asset'"
        >
          <view class="flex items-center space-x-1.5">
            <text class="text-[#8d99a4]">占优核心资产</text>
            <text class="text-[10px] text-[#8d99a4] border border-[#8d99a4]/40 rounded-full w-3.5 h-3.5 flex items-center justify-center leading-none">?</text>
          </view>
          <CTag variant="up">{{ cycleData.dominant_asset }}</CTag>
        </view>

        <view
          class="flex items-center justify-between pb-2 border-b border-white/[0.07] cursor-pointer active:opacity-75"
          @tap="activeCycleTooltipKey = 'style_preference'"
        >
          <view class="flex items-center space-x-1.5">
            <text class="text-[#8d99a4]">风格主线倾向</text>
            <text class="text-[10px] text-[#8d99a4] border border-[#8d99a4]/40 rounded-full w-3.5 h-3.5 flex items-center justify-center leading-none">?</text>
          </view>
          <text class="font-semibold text-[#f1efea]">{{ cycleData.style_preference }}</text>
        </view>

        <view
          class="p-3 bg-[#181e25] rounded-xl text-[#9aa5b1] leading-relaxed text-[11px] border border-white/[0.06] cursor-pointer active:opacity-75 space-y-1.5"
          @tap="activeCycleTooltipKey = 'summary'"
        >
          <view class="flex items-center justify-between text-[#8d99a4] text-[10px] pb-1 border-b border-white/[0.05]">
            <text class="font-medium">宏观配置依据</text>
            <text class="text-[#e5b352] text-[10px]">小白释义 ⓘ</text>
          </view>
          <text class="block text-[#d1d5db] leading-relaxed">{{ cycleData.summary }}</text>
        </view>
      </view>
    </view>

    <!-- 资产配置与周期结论：小白大白话弹窗 (ELI5 Modal) -->
    <view
      v-if="activeCycleTooltip"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-[2px]"
      @tap="activeCycleTooltipKey = null"
    >
      <view
        class="w-full max-w-sm bg-[#151a21] border border-white/[0.14] rounded-2xl p-4 space-y-3 shadow-2xl text-[#f1efea] max-h-[85vh] flex flex-col"
        @tap.stop
      >
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2 shrink-0">
          <view class="flex items-center space-x-1.5 min-w-0">
            <text class="text-base shrink-0">💡</text>
            <text class="text-sm font-bold text-[#e5b352] truncate">{{ activeCycleTooltip.title }}</text>
          </view>
          <text
            class="text-[#8d99a4] text-base px-2 py-0.5 cursor-pointer active:opacity-70 shrink-0"
            @tap="activeCycleTooltipKey = null"
          >
            ✕
          </text>
        </view>

        <!-- 数据时间：统一格式 2026-11-20 15:30 -->
        <view class="bg-[#0e1216] px-3 py-1.5 rounded-xl flex items-center justify-between border border-white/[0.06] text-xs shrink-0">
          <text class="text-[#8d99a4]">数据时间</text>
          <text class="font-bold font-mono text-[#e5b352]">{{ strategyData.rotation.date || '2026-11-20' }} 15:30</text>
        </view>

        <!-- 弹窗可滚动主体区 -->
        <view class="overflow-y-auto space-y-2.5 flex-1 pr-0.5 text-xs">
          <!-- 当前读数 / 状态：完整换行展示，不使用省略号截断 -->
          <view class="bg-[#0e1216] px-3 py-2 rounded-xl flex items-start justify-between gap-2 border border-white/[0.06]">
            <text class="text-[11px] text-[#8d99a4] shrink-0 mt-0.5">当前配置研判</text>
            <text class="text-xs font-bold font-mono text-[#f1efea] text-right flex-1 break-words leading-relaxed">
              {{ activeCycleTooltip.currentVal }}
            </text>
          </view>

          <!-- 📐 数值怎么算的 / 逻辑怎么推导的 -->
          <view class="space-y-1 bg-[#2ad1a3]/10 p-2.5 rounded-xl border border-[#2ad1a3]/20">
            <text class="text-[11px] font-bold text-[#2ad1a3] block">📐 逻辑怎么推导的：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block font-mono cc-tabular">{{ activeCycleTooltip.calc }}</text>
          </view>

          <!-- 🗣️ 大白话说了啥 -->
          <view class="space-y-1">
            <text class="text-[11px] font-bold text-[#e5b352] block">🗣️ 大白话说了啥：</text>
            <text class="text-xs text-[#d1d5db] leading-relaxed block">{{ activeCycleTooltip.eli5 }}</text>
          </view>

          <!-- 🎯 我现在该干嘛 -->
          <view class="space-y-1 bg-[#e5b352]/10 p-2.5 rounded-xl border border-[#e5b352]/20">
            <text class="text-[11px] font-bold text-[#e5b352] block">🎯 我现在该干嘛：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block">{{ activeCycleTooltip.action }}</text>
          </view>
        </view>

        <view class="pt-1 shrink-0">
          <view
            class="py-2.5 text-center rounded-xl bg-[#e5b352] text-[#090c0f] font-bold text-xs active:opacity-85 cursor-pointer"
            @tap="activeCycleTooltipKey = null"
          >
            我知道了
          </view>
        </view>
      </view>
    </view>

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
