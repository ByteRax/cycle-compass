<script lang="ts" setup>
import { computed } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

interface Props {
  totalScore?: number
  techScore?: number
  spreadScore?: number
  valuationScore?: number
  styleScore?: number
  multiplierText?: string
  gated?: boolean
  effectiveAmount?: number
  asOf?: string
  dataSourceNote?: string
  closePrice?: number
  pe?: number
  pePercentile?: number
  dividendYield?: number
  bond10y?: number
  spreadPP?: number
  rsi70?: number
  bPct100?: number
  stylePercentile?: number
}

const props = withDefaults(defineProps<Props>(), {
  totalScore: 6.57,
  techScore: 22.98,
  spreadScore: 70.24,
  valuationScore: 9.23,
  styleScore: 24.27,
  multiplierText: '空仓 · 暂停定投',
  gated: false,
  effectiveAmount: 0.0,
  asOf: '2026-11-20',
  dataSourceNote: '中证指数官网官方行情',
  closePrice: 11878.27,
  pe: 9.61,
  pePercentile: 0.9077,
  dividendYield: 4.39,
  bond10y: 1.5803,
  spreadPP: 2.8097,
  rsi70: 53.57,
  bPct100: 0.962,
  stylePercentile: 0.7573
})

const scoreNeedlePct = computed(() => Math.min(94, Math.max(6, (props.totalScore / 20) * 100)))
const displayClose = computed(() => Math.round(props.closePrice).toLocaleString())
const displayPe = computed(() => props.pe.toFixed(2))
const displayPePercentile = computed(() => Math.round(props.pePercentile * 100))
const displayDividendYield = computed(() => `${props.dividendYield.toFixed(2)}%`)
const displayBond10y = computed(() => `${props.bond10y.toFixed(2)}%`)
const displaySpreadPP = computed(() => `${(props.spreadPP ?? props.dividendYield - props.bond10y).toFixed(2)}pp`)
const displayRsi = computed(() => Math.round(props.rsi70))
const displayBPct = computed(() => props.bPct100.toFixed(2))
const displayStylePercentile = computed(() => Math.round(props.stylePercentile * 100))

const getScoreTag = (score: number) => {
  if (score >= 65) return { label: '高性价比', cls: 'bg-[#e5b352]/18 text-[#e5b352]' }
  if (score >= 40) return { label: '中性', cls: 'bg-white/[0.08] text-[#9aa5b1]' }
  if (score >= 20) return { label: '偏贵', cls: 'bg-[#f05a50]/16 text-[#ff736a]' }
  return { label: '历史高位', cls: 'bg-[#f05a50]/22 text-[#ff736a]' }
}

const dims = computed(() => [
  {
    name: '技术面',
    weight: '50%',
    score: props.techScore,
    tag: getScoreTag(props.techScore),
    desc: `RSI ${displayRsi.value} · 布林%B ${displayBPct.value}`
  },
  {
    name: '股息利差',
    weight: '25%',
    score: props.spreadScore,
    tag: getScoreTag(props.spreadScore),
    desc: `股息 ${displayDividendYield.value} - 国债 ${displayBond10y.value} = ${displaySpreadPP.value}`
  },
  {
    name: '估值分位',
    weight: '15%',
    score: props.valuationScore,
    tag: getScoreTag(props.valuationScore),
    desc: `PE ${displayPe.value} · 历史分位 ${displayPePercentile.value}%`
  },
  {
    name: '市场风格',
    weight: '10%',
    score: props.styleScore,
    tag: getScoreTag(props.styleScore),
    desc: `红利/成长相对强弱 ${displayStylePercentile.value}% 分位`
  }
])
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] p-4 border border-white/[0.09] text-[#f1efea] space-y-4">
    <!-- 顶部标题 + 状态胶囊（严格单行不折行） -->
    <view class="flex items-center justify-between gap-2">
      <view class="min-w-0">
        <view class="flex items-center space-x-2">
          <text class="text-base font-bold text-[#f1efea] whitespace-nowrap">红利低波100 四维评分</text>
          <text
            class="shrink-0 rounded-full px-2 py-0.5 text-[10px] font-bold whitespace-nowrap"
            :class="gated ? 'bg-[#f05a50]/18 text-[#ff736a]' : 'bg-[#2ad1a3]/16 text-[#2ad1a3]'"
          >
            {{ gated ? '门控拦截' : '门控通行' }}
          </text>
        </view>
        <text class="text-[11px] text-[#8d99a4] font-mono mt-0.5 block">
          {{ asOf }} · {{ dataSourceNote }}
        </text>
      </view>

      <view class="text-right shrink-0">
        <view class="flex items-baseline justify-end">
          <text class="text-2xl font-bold font-mono cc-tabular text-[#e5b352]">{{ totalScore.toFixed(2) }}</text>
          <text class="text-xs font-mono text-[#64707c] ml-0.5">/20</text>
        </view>
        <text class="text-[10px] text-[#9aa5b1] block whitespace-nowrap">{{ multiplierText }}</text>
      </view>
    </view>

    <!-- 5 项基准参数发丝线横排 -->
    <view class="grid grid-cols-5 gap-1 py-2.5 px-2 rounded-xl bg-[#181e25] border border-white/[0.06] text-center">
      <view>
        <text class="text-[10px] text-[#64707c] block">收盘</text>
        <text class="font-bold font-mono cc-tabular text-[#f1efea] text-[11px] mt-0.5 block">{{ displayClose }}</text>
      </view>
      <view>
        <text class="text-[10px] text-[#64707c] block">PE</text>
        <text class="font-bold font-mono cc-tabular text-[#f1efea] text-[11px] mt-0.5 block">{{ displayPe }}</text>
      </view>
      <view>
        <text class="text-[10px] text-[#64707c] block">股息率</text>
        <text class="font-bold font-mono cc-tabular text-[#e5b352] text-[11px] mt-0.5 block">{{ displayDividendYield }}</text>
      </view>
      <view>
        <text class="text-[10px] text-[#64707c] block">10Y国债</text>
        <text class="font-bold font-mono cc-tabular text-[#f1efea] text-[11px] mt-0.5 block">{{ displayBond10y }}</text>
      </view>
      <view>
        <text class="text-[10px] text-[#64707c] block">利差</text>
        <text class="font-bold font-mono cc-tabular text-[#2ad1a3] text-[11px] mt-0.5 block">{{ displaySpreadPP }}</text>
      </view>
    </view>

    <!-- 六档评分游标尺 -->
    <view>
      <view class="relative h-5 font-mono">
        <view
          class="absolute bottom-0 left-0 flex flex-col items-center whitespace-nowrap will-change-transform"
          :style="{ transform: `translate3d(${scoreNeedlePct}%, 0, 0) translate3d(-50%, 0, 0)` }"
        >
          <text class="text-[10px] font-bold text-[#e5b352] leading-none">{{ totalScore.toFixed(2) }} ▼</text>
        </view>
      </view>
      <view class="h-6 rounded-lg overflow-hidden flex border border-white/[0.1] text-[10px] font-semibold text-center">
        <view class="w-[15%] bg-[#2ad1a3]/18 text-[#2ad1a3] flex items-center justify-center">减仓</view>
        <view class="w-[15%] bg-white/[0.08] text-[#9aa5b1] flex items-center justify-center">空仓</view>
        <view class="w-[15%] bg-[#e5b352]/14 text-[#e5b352] flex items-center justify-center">少买</view>
        <view class="w-[20%] bg-[#e5b352]/24 text-[#f5b731] flex items-center justify-center">正常</view>
        <view class="w-[15%] bg-[#f05a50]/18 text-[#ff736a] flex items-center justify-center">多买</view>
        <view class="w-[20%] bg-[#f05a50]/28 text-[#ff736a] flex items-center justify-center">加码</view>
      </view>
      <view class="flex justify-between text-[10px] text-[#64707c] font-mono mt-1">
        <text>0</text>
        <text>正常线 10 分</text>
        <text>峰值 15.5</text>
        <text>20</text>
      </view>
    </view>

    <!-- 四维拆解发丝线列表（去彩虹色，用铜金+中性灰+方向四状态胶囊） -->
    <view class="space-y-1">
      <view
        v-for="item in dims"
        :key="item.name"
        class="py-2.5 border-t border-white/[0.07]"
      >
        <view class="flex items-center justify-between gap-2">
          <view class="flex items-center space-x-1.5 min-w-0">
            <text class="text-xs font-semibold text-[#f1efea]">{{ item.name }}</text>
            <text class="text-[10px] font-mono text-[#64707c]">{{ item.weight }}</text>
          </view>
          <view class="flex items-center space-x-2 shrink-0">
            <text
              class="text-xs font-bold font-mono cc-tabular"
              :class="item.score >= 60 ? 'text-[#e5b352]' : 'text-[#f1efea]'"
            >
              {{ item.score.toFixed(1) }}
            </text>
            <text :class="['rounded-full px-2 py-0.5 text-[10px] font-semibold whitespace-nowrap', item.tag.cls]">
              {{ item.tag.label }}
            </text>
          </view>
        </view>
        <view class="h-1.5 w-full bg-white/[0.07] rounded-full overflow-hidden mt-2">
          <view
            class="h-full rounded-full"
            :class="item.score >= 60 ? 'bg-[#e5b352]' : 'bg-[#f1efea]/40'"
            :style="{ width: `${Math.min(100, Math.max(0, item.score))}%` }"
          />
        </view>
        <text class="text-[10px] text-[#8d99a4] font-mono mt-1 block">{{ item.desc }}</text>
      </view>
    </view>

    <!-- 今日建议投入 -->
    <view class="pt-2.5 border-t border-white/[0.08] flex items-center justify-between text-xs">
      <text class="text-[#9aa5b1]">今日建议新增投入（基数 200 元）</text>
      <text class="text-base font-bold font-mono cc-tabular text-[#f1efea]">¥{{ effectiveAmount.toFixed(0) }}</text>
    </view>
  </view>
</template>
