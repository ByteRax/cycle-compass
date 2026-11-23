<script lang="ts" setup>
import { computed } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

interface Props {
  ratio?: number
  growthPrice?: number
  dividendPrice?: number
  zone?: 'ZONE_LOW' | 'ZONE_MID' | 'ZONE_HIGH' | string
  zoneName?: string
  percentile1500?: number
  toHighPct?: number
  toLowPct?: number
  actionText?: string
  asOf?: string
}

const props = withDefaults(defineProps<Props>(), {
  ratio: 0.1846,
  growthPrice: 2192.8,
  dividendPrice: 11878.27,
  zone: 'ZONE_MID',
  zoneName: '中间区',
  percentile1500: 0.2433,
  toHighPct: 51.67,
  toLowPct: 13.33,
  actionText: '空仓 · 进现金池',
  asOf: '2026-11-20'
})

const needlePositionPct = computed(() => {
  const minVal = 0.1
  const maxVal = 0.5
  const clamped = Math.max(minVal, Math.min(maxVal, props.ratio))
  const rawPct = ((clamped - minVal) / (maxVal - minVal)) * 100
  return Math.max(10, Math.min(90, rawPct))
})

const zonePillClass = computed(() => {
  if (props.zone === 'ZONE_LOW' || props.zoneName === '低区') return 'bg-[#2ad1a3]/16 text-[#2ad1a3]'
  if (props.zone === 'ZONE_HIGH' || props.zoneName === '高区') return 'bg-[#f05a50]/18 text-[#ff736a]'
  return 'bg-[#e5b352]/18 text-[#e5b352]'
})
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-4">
    <!-- 头部标题与状态胶囊 -->
    <view class="flex items-start justify-between gap-2">
      <view class="min-w-0">
        <view class="text-base font-bold text-[#f1efea] tracking-tight">
          创业板50 ÷ 红利低波100 · 轮动标尺
        </view>
        <view class="text-[11px] text-[#8d99a4] mt-0.5 font-mono">
          {{ asOf }} 收盘 · 横轴 0.10 ~ 0.50 等比例
        </view>
      </view>
      <view :class="['shrink-0 rounded-full px-2.5 py-1 text-[11px] font-bold whitespace-nowrap', zonePillClass]">
        {{ zoneName }}
      </view>
    </view>

    <!-- 等比例三区间游标尺（彻底避免标注重叠） -->
    <view class="pt-1">
      <view class="relative h-7 text-[10px] font-mono">
        <view
          class="absolute bottom-0 left-0 flex flex-col items-center z-20 whitespace-nowrap will-change-transform"
          :style="{ transform: `translate3d(${needlePositionPct}%, 0, 0) translate3d(-50%, 0, 0)` }"
        >
          <view class="inline-flex items-center px-2 py-0.5 rounded-full bg-[#e5b352] text-[#090c0f] text-[11px] font-bold cc-tabular">
            现在 {{ ratio.toFixed(4) }}
          </view>
          <text class="text-[#e5b352] text-[9px] block leading-none -mt-0.5">▼</text>
        </view>
      </view>

      <!-- 三区间语义色带（低区 15% / 中间区 30% / 高区 55%） -->
      <view class="relative h-14 rounded-xl overflow-hidden flex border border-white/[0.12]">
        <view class="w-[15%] bg-[#2ad1a3]/16 border-r border-white/[0.12] flex flex-col justify-center items-center p-1">
          <text class="text-[11px] font-bold text-[#2ad1a3] leading-none">低区</text>
          <text class="text-[9px] text-[#2ad1a3]/80 font-mono mt-1">{{ '<0.16' }}</text>
        </view>

        <view class="w-[30%] bg-[#e5b352]/14 border-r border-white/[0.12] flex flex-col justify-center items-center p-1">
          <text class="text-[11px] font-bold text-[#e5b352] leading-none">中间区</text>
          <text class="text-[9px] text-[#e5b352]/80 font-mono mt-1">0.16~0.28</text>
          <text class="text-[9px] text-[#8d99a4] mt-0.5">空仓 268 周</text>
        </view>

        <view class="w-[55%] bg-[#f05a50]/14 flex flex-col justify-center items-center p-1">
          <text class="text-[11px] font-bold text-[#ff736a] leading-none">高区 · 买红利低波100</text>
          <text class="text-[9px] text-[#ff736a]/80 font-mono mt-1">≥ 0.28 · 10年 106 周</text>
        </view>
      </view>

      <!-- 刻度数值行 -->
      <view class="flex justify-between text-[10px] text-[#64707c] font-mono mt-1.5 px-0.5">
        <text>0.10</text>
        <text class="text-[#2ad1a3] font-semibold">0.16</text>
        <text class="text-[#e5b352] font-semibold">0.28</text>
        <text>0.40</text>
        <text>0.50</text>
      </view>

      <!-- 历史极值参考行（独立分层零碰撞） -->
      <view class="flex items-center justify-between text-[10px] font-mono mt-2 pt-2 border-t border-white/[0.07] text-[#8d99a4]">
        <text>历史低谷 <text class="text-[#2ad1a3] font-semibold">0.1204</text> (2019-01)</text>
        <text>历史峰值 <text class="text-[#ff736a] font-semibold">0.4399</text> (2026-06)</text>
      </view>
    </view>

    <!-- 3 列发丝线核心读数 -->
    <view class="grid grid-cols-3 border-t border-b border-white/[0.08] py-2.5">
      <view>
        <text class="text-[10px] text-[#8d99a4] block">距低区线</text>
        <text class="text-sm font-bold font-mono cc-tabular text-[#2ad1a3] mt-0.5 block">跌 {{ toLowPct.toFixed(1) }}%</text>
      </view>
      <view class="text-center border-l border-r border-white/[0.08]">
        <text class="text-[10px] text-[#8d99a4] block">距高区线</text>
        <text class="text-sm font-bold font-mono cc-tabular text-[#ff736a] mt-0.5 block">涨 {{ toHighPct.toFixed(1) }}%</text>
      </view>
      <view class="text-right">
        <text class="text-[10px] text-[#8d99a4] block">1500日分位</text>
        <text class="text-sm font-bold font-mono cc-tabular text-[#e5b352] mt-0.5 block">{{ (percentile1500 * 100).toFixed(1) }}%</text>
      </view>
    </view>

    <!-- 距离边界进度条 -->
    <view class="space-y-2">
      <view class="flex items-center justify-between text-[11px]">
        <text class="text-[#9aa5b1] w-28 shrink-0">距低区（买创50）</text>
        <view class="flex-1 mx-2 h-1.5 bg-white/[0.08] rounded-full overflow-hidden">
          <view class="h-full bg-[#2ad1a3] rounded-full" :style="{ width: `${Math.min(100, (toLowPct / 50) * 100)}%` }" />
        </view>
        <text class="font-mono cc-tabular text-[#2ad1a3] font-semibold shrink-0">-{{ toLowPct.toFixed(1) }}%</text>
      </view>

      <view class="flex items-center justify-between text-[11px]">
        <text class="text-[#9aa5b1] w-28 shrink-0">距高区（买红利）</text>
        <view class="flex-1 mx-2 h-1.5 bg-white/[0.08] rounded-full overflow-hidden">
          <view class="h-full bg-[#e5b352] rounded-full" :style="{ width: `${Math.min(100, (toHighPct / 50) * 100)}%` }" />
        </view>
        <text class="font-mono cc-tabular text-[#e5b352] font-semibold shrink-0">+{{ toHighPct.toFixed(1) }}%</text>
      </view>
    </view>

    <!-- 规则速查 -->
    <view class="rounded-xl bg-[#181e25] p-3 border border-white/[0.07] space-y-1.5 text-[11px]">
      <view class="flex items-center justify-between">
        <text class="font-mono text-[#2ad1a3] font-semibold">{{ '< 0.16 低区' }}</text>
        <text class="text-[#9aa5b1]">定投创业板50 (159949)</text>
      </view>
      <view class="flex items-center justify-between">
        <text class="font-mono text-[#e5b352] font-semibold">0.16~0.28 中间</text>
        <text class="text-[#9aa5b1]">暂停买入，定投额进现金池</text>
      </view>
      <view class="flex items-center justify-between">
        <text class="font-mono text-[#ff736a] font-semibold">≥ 0.28 高区</text>
        <text class="text-[#9aa5b1]">定投红利低波100 (515100)</text>
      </view>
    </view>
  </view>
</template>
