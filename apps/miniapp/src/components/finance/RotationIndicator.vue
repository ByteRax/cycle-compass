<script lang="ts" setup>
import { computed, ref } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

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

const showEli5 = ref(false)

const needlePositionPct = computed(() => {
  const minVal = 0.1
  const maxVal = 0.5
  const clamped = Math.max(minVal, Math.min(maxVal, props.ratio))
  const rawPct = ((clamped - minVal) / (maxVal - minVal)) * 100
  return Math.max(10, Math.min(90, rawPct))
})

const zonePillClass = computed(() => {
  if (props.zone === 'ZONE_LOW' || props.zoneName === '低区') {
    return 'bg-[#2ad1a3]/16 text-[#2ad1a3]'
  }
  if (props.zone === 'ZONE_HIGH' || props.zoneName === '高区') {
    return 'bg-[#f05a50]/18 text-[#ff736a]'
  }
  return 'bg-[#e5b352]/18 text-[#e5b352]'
})
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-4">
    <!-- 头部标题与状态胶囊 -->
    <view class="flex items-start justify-between gap-2 pb-3 border-b border-white/[0.08]">
      <view class="min-w-0 flex-1">
        <view class="flex items-center cursor-pointer active:opacity-75" @tap="showEli5 = true">
          <view class="text-base font-bold text-[#f1efea] tracking-tight">
            创业板50 ÷ 红利低波100 · 轮动标尺
          </view>
        </view>
        <view class="flex items-center space-x-1.5 mt-1 text-[10px] text-[#8d99a4] font-mono">
          <text class="text-[#e5b352]">📅</text>
          <text>{{ asOf.length === 10 ? `${asOf} 15:30` : asOf }}</text>
          <text class="text-white/[0.15]">|</text>
          <text>横轴 0.10 ~ 0.50 等比例</text>
        </view>
      </view>
      <view
        class="shrink-0 rounded-full px-2.5 py-1 text-[11px] font-bold whitespace-nowrap cursor-pointer active:opacity-75"
        :class="[zonePillClass]"
        @tap="showEli5 = true"
      >
        {{ zoneName }} ⓘ
      </view>
    </view>

    <!-- 等比例三区间游标尺（彻底避免标注重叠） -->
    <view class="pt-1">
      <view class="relative h-7 text-[10px] font-mono">
        <view
          class="absolute bottom-0 left-0 flex flex-col items-center z-20 whitespace-nowrap will-change-transform cursor-pointer"
          :style="{ transform: `translate3d(${needlePositionPct}%, 0, 0) translate3d(-50%, 0, 0)` }"
          @tap="showEli5 = true"
        >
          <view class="inline-flex items-center px-2 py-0.5 rounded-full bg-[#e5b352] text-[#090c0f] text-[11px] font-bold cc-tabular">
            现在 {{ ratio.toFixed(4) }} ⓘ
          </view>
          <text class="text-[#e5b352] text-[9px] block leading-none -mt-0.5">▼</text>
        </view>
      </view>

      <!-- 连续比例色带轨条：0.10 - 0.16 (低区绿) / 0.16 - 0.28 (中间区金) / 0.28 - 0.50 (高区红) -->
      <view class="h-3 w-full rounded-full overflow-hidden flex border border-white/[0.12] bg-[#090c0f]">
        <!-- 低区：占宽 15% (对应 0.10 ~ 0.16) -->
        <view class="h-full bg-gradient-to-r from-[#2ad1a3]/40 to-[#2ad1a3] w-[15%]" />
        <!-- 中间区：占宽 30% (对应 0.16 ~ 0.28) -->
        <view class="h-full bg-gradient-to-r from-[#e5b352]/70 to-[#e5b352] w-[30%]" />
        <!-- 高区：占宽 55% (对应 0.28 ~ 0.50) -->
        <view class="h-full bg-gradient-to-r from-[#f05a50]/70 to-[#f05a50] w-[55%]" />
      </view>

      <!-- 刻度值对齐注脚 -->
      <view class="mt-1.5 flex justify-between text-[10px] font-mono cc-tabular text-[#64707c]">
        <text>0.10 极底</text>
        <text class="text-[#2ad1a3]">0.16 低区线</text>
        <text class="text-[#e5b352]">0.28 高区线</text>
        <text>0.50 极高</text>
      </view>
    </view>

    <!-- 关键数据三列指标网格 -->
    <view class="grid grid-cols-3 border-y border-white/[0.08] py-2.5">
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

    <!-- 标尺小白大白话弹窗 (ELI5 Modal) -->
    <view
      v-if="showEli5"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4 backdrop-blur-[2px]"
      @tap="showEli5 = false"
    >
      <view
        class="w-full max-w-sm bg-[#151a21] border border-white/[0.14] rounded-2xl p-4 space-y-3 shadow-2xl text-[#f1efea] max-h-[85vh] flex flex-col"
        @tap.stop
      >
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2 shrink-0">
          <view class="flex items-center space-x-1.5 min-w-0">
            <text class="text-base shrink-0">💡</text>
            <text class="text-sm font-bold text-[#e5b352] truncate">轮动标尺：大白话拆解</text>
          </view>
          <text
            class="text-[#8d99a4] text-base px-2 py-0.5 cursor-pointer active:opacity-70 shrink-0"
            @tap="showEli5 = false"
          >
            ✕
          </text>
        </view>

        <!-- 数据时间：统一格式 2026-11-20 15:30 -->
        <view class="bg-[#0e1216] px-3 py-1.5 rounded-xl flex items-center justify-between border border-white/[0.06] text-xs shrink-0">
          <text class="text-[#8d99a4]">数据时间</text>
          <text class="font-bold font-mono text-[#e5b352]">{{ asOf.length === 10 ? `${asOf} 15:30` : asOf }}</text>
        </view>

        <!-- 弹窗可滚动主体区 -->
        <view class="overflow-y-auto space-y-2.5 flex-1 pr-0.5 text-xs">
          <view class="bg-[#0e1216] px-3 py-2 rounded-xl flex items-center justify-between border border-white/[0.06]">
            <text class="text-[11px] text-[#8d99a4]">当前比值与区间</text>
            <text class="text-xs font-bold font-mono text-[#e5b352]">{{ ratio.toFixed(4) }} ({{ zoneName }})</text>
          </view>

          <!-- 📐 数值怎么算的 -->
          <view class="space-y-1 bg-[#2ad1a3]/10 p-2.5 rounded-xl border border-[#2ad1a3]/20">
            <text class="text-[11px] font-bold text-[#2ad1a3] block">📐 数值怎么算的：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block font-mono cc-tabular">
              创业板50指数 (代码 399673) 收盘价 ({{ growthPrice }}) ÷ 红利低波100指数 (代码 930955) 收盘价 ({{ dividendPrice }}) = {{ ratio.toFixed(4) }}。
              落在 0.16 ~ 0.28 区间即为中间区。（对应定投ETF：创50 159949 / 红利低波 515100）
            </text>
          </view>

          <view class="space-y-1.5">
            <text class="text-[11px] font-bold text-[#e5b352] block">🗣️ 大白话说了啥：</text>
            <text class="text-xs text-[#d1d5db] leading-relaxed block">
              这是“冲劲小伙（创业板50指数 399673，涨得猛跌得狠）”和“收租大爷（红利低波100指数 930955，抗跌拿分红）”的拉扯拔河标尺：
            </text>
            <view class="text-[11px] space-y-1 text-[#9aa5b1] pl-1">
              <view><text class="text-[#2ad1a3] font-bold">🟢 低区(&lt;0.16)</text>：小伙摔惨了大减价，定投创业板50 (399673 / ETF 159949) 抄底。</view>
              <view><text class="text-[#e5b352] font-bold">🟡 中间区(0.16~0.28)</text>：势均力敌谁也没便宜占，暂停买入，钱存现金池攒子弹。</view>
              <view><text class="text-[#ff736a] font-bold">🔴 高区(&ge;0.28)</text>：小伙冲太疯泡沫大，转攻为守，定投红利低波100 (930955 / ETF 515100) 避险。</view>
            </view>
          </view>

          <view class="space-y-1 bg-[#e5b352]/10 p-2.5 rounded-xl border border-[#e5b352]/20">
            <text class="text-[11px] font-bold text-[#e5b352] block">🎯 我现在该干嘛：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block">
              当前比值 {{ ratio.toFixed(4) }} 刚好卡在中间区，本周什么都不要买，把定投资金存进零钱罐！
            </text>
          </view>
        </view>

        <view class="pt-1 shrink-0">
          <view
            class="py-2.5 text-center rounded-xl bg-[#e5b352] text-[#090c0f] font-bold text-xs active:opacity-85 cursor-pointer"
            @tap="showEli5 = false"
          >
            我知道了
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
