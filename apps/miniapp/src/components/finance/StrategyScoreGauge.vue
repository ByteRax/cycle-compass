<script lang="ts" setup>
import { computed, ref } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

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

const showEli5 = ref(false)

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

function getScoreTag (score: number) {
  if (score >= 65) {
    return { label: '高性价比', cls: 'bg-[#e5b352]/18 text-[#e5b352]' }
  }
  if (score >= 40) {
    return { label: '中性', cls: 'bg-white/[0.08] text-[#9aa5b1]' }
  }
  if (score >= 20) {
    return { label: '偏贵', cls: 'bg-[#f05a50]/16 text-[#ff736a]' }
  }
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
    <!-- 顶部状态栏：分行精细布局，杜绝任何字样重叠与挤压 -->
    <view class="pb-3 border-b border-white/[0.08] space-y-2">
      <!-- 第一行：标题 + 右侧大分数 -->
      <view class="flex items-center justify-between gap-2">
        <view class="flex items-center cursor-pointer active:opacity-75 min-w-0" @tap="showEli5 = true">
          <text class="text-base font-bold text-[#f1efea] tracking-tight">红利低波100 四维评分</text>
        </view>
        <view class="flex items-baseline space-x-0.5 shrink-0 cursor-pointer active:opacity-75" @tap="showEli5 = true">
          <text class="text-2xl font-bold font-mono cc-tabular text-[#e5b352]">{{ totalScore.toFixed(2) }}</text>
          <text class="text-xs font-mono text-[#64707c]">/20 ⓘ</text>
        </view>
      </view>

      <!-- 第二行：数据时间与数据源 + 门控与定投建议状态 -->
      <view class="flex items-center justify-between gap-2 text-[10px]">
        <view class="flex items-center space-x-1.5 text-[#8d99a4] font-mono min-w-0">
          <text class="text-[#e5b352]">📅</text>
          <text>{{ asOf.length === 10 ? `${asOf} 15:30` : asOf }}</text>
          <text class="text-white/[0.15]">|</text>
          <text class="truncate">{{ dataSourceNote }}</text>
        </view>
        <view class="flex items-center space-x-1.5 shrink-0">
          <text
            class="rounded-full px-1.5 py-0.5 text-[9px] font-bold whitespace-nowrap"
            :class="gated ? 'bg-[#f05a50]/18 text-[#ff736a]' : 'bg-[#2ad1a3]/16 text-[#2ad1a3]'"
          >
            {{ gated ? '门控拦截' : '门控通行' }}
          </text>
          <text class="text-[#9aa5b1] whitespace-nowrap font-medium">{{ multiplierText }}</text>
        </view>
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
        <text class="text-[10px] text-[#64707c] block">PE分位</text>
        <text class="font-bold font-mono cc-tabular text-[#f1efea] text-[11px] mt-0.5 block">{{ displayPePercentile }}%</text>
      </view>
      <view>
        <text class="text-[10px] text-[#64707c] block">股息率</text>
        <text class="font-bold font-mono cc-tabular text-[#e5b352] text-[11px] mt-0.5 block">{{ displayDividendYield }}</text>
      </view>
      <view>
        <text class="text-[10px] text-[#64707c] block">十年国债</text>
        <text class="font-bold font-mono cc-tabular text-[#f1efea] text-[11px] mt-0.5 block">{{ displayBond10y }}</text>
      </view>
    </view>

    <!-- 连续色带评分仪表盘 -->
    <view class="pt-1">
      <view class="relative h-6 text-[10px] font-mono">
        <view
          class="absolute bottom-0 left-0 flex flex-col items-center z-10 whitespace-nowrap will-change-transform cursor-pointer"
          :style="{ transform: `translate3d(${scoreNeedlePct}%, 0, 0) translate3d(-50%, 0, 0)` }"
          @tap="showEli5 = true"
        >
          <text class="px-2 py-0.5 rounded-full bg-[#e5b352] text-[#090c0f] text-[11px] font-bold cc-tabular">
            综合 {{ totalScore.toFixed(2) }} ⓘ
          </text>
          <text class="text-[#e5b352] text-[9px] block leading-none -mt-0.5">▼</text>
        </view>
      </view>

      <!-- 渐变轨条：红(0-4) -> 灰黄(4-12) -> 金黄(12-20) -->
      <view class="h-2.5 w-full rounded-full overflow-hidden flex border border-white/[0.08] bg-[#090c0f]">
        <view class="h-full bg-gradient-to-r from-[#f05a50]/40 to-[#f05a50] w-[20%]" />
        <view class="h-full bg-gradient-to-r from-white/[0.12] to-white/[0.22] w-[40%]" />
        <view class="h-full bg-gradient-to-r from-[#e5b352]/50 to-[#e5b352] w-[40%]" />
      </view>

      <!-- 刻度分界 -->
      <view class="mt-1 flex justify-between text-[10px] font-mono cc-tabular text-[#64707c]">
        <text>0 暂停定投</text>
        <text class="text-[#ff736a]">4分 警戒线</text>
        <text class="text-[#9aa5b1]">12分 基准</text>
        <text class="text-[#e5b352]">20分 加倍定投</text>
      </view>
    </view>

    <!-- 四维卡片横排/纵排细分 -->
    <view class="space-y-2">
      <view
        v-for="item in dims"
        :key="item.name"
        class="bg-[#181e25] rounded-xl p-2.5 border border-white/[0.06]"
      >
        <view class="flex items-center justify-between">
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
            <text
              class="rounded-full px-2 py-0.5 text-[10px] font-semibold whitespace-nowrap"
              :class="item.tag.cls"
            >
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

    <!-- 四维评分小白大白话弹窗 (ELI5 Modal) -->
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
            <text class="text-sm font-bold text-[#e5b352] truncate">红利四维评分：大白话拆解</text>
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
          <!-- 当前读数 / 状态 -->
          <view class="bg-[#0e1216] px-3 py-2 rounded-xl flex items-center justify-between border border-white/[0.06]">
            <text class="text-[11px] text-[#8d99a4]">综合体检得分</text>
            <text class="text-xs font-bold font-mono text-[#ff736a]">{{ totalScore.toFixed(2) }} / 20 分 (不及格)</text>
          </view>

          <!-- 📐 数值怎么算的 -->
          <view class="space-y-1 bg-[#2ad1a3]/10 p-2.5 rounded-xl border border-[#2ad1a3]/20">
            <text class="text-[11px] font-bold text-[#2ad1a3] block">📐 数值怎么算的：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block font-mono cc-tabular">
              4门科目加权后除以5：(技术面×50% + 股息利差×25% + 估值×15% + 风格×10%) ÷ 5。
              技术面 22.98分 + 股息利差 70.24分 + 估值 9.23分 + 风格 24.27分，加权折算得出 {{ totalScore.toFixed(2) }} 分。
            </text>
          </view>

          <!-- 🗣️ 大白话说了啥 -->
          <view class="space-y-1">
            <text class="text-[11px] font-bold text-[#e5b352] block">🗣️ 大白话说了啥：</text>
            <text class="text-xs text-[#d1d5db] leading-relaxed block">
              这是“收租老大爷（红利稳健股）”的全身体检得分表。分值在 0~20 分之间，分越高代表越便宜、越值得买：
            </text>
            <view class="text-[11px] space-y-1 text-[#9aa5b1] pl-1">
              <view><text class="text-[#f1efea] font-medium">1. 技术面 (50%)</text>：近期跌得狠不狠，跌透了才有高性价比。</view>
              <view><text class="text-[#f1efea] font-medium">2. 股息利差 (25%)</text>：分红率比存银行高多少，高越多越香。</view>
              <view><text class="text-[#f1efea] font-medium">3. 估值分位 (15%)</text>：跟历史比划不划算。</view>
              <view><text class="text-[#f1efea] font-medium">4. 市场风格 (10%)</text>：资金爱不爱防守型股票。</view>
            </view>
          </view>

          <!-- 🎯 我现在该干嘛 -->
          <view class="space-y-1 bg-[#e5b352]/10 p-2.5 rounded-xl border border-[#e5b352]/20">
            <text class="text-[11px] font-bold text-[#e5b352] block">🎯 我现在该干嘛：</text>
            <text class="text-xs text-[#f1efea] leading-relaxed block">
              当前仅 {{ totalScore.toFixed(2) }} 分属于偏低区间，说明红利资产目前便宜度一般，没必要追高买入，保持观望。
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
