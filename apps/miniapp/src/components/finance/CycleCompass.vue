<script lang="ts" setup>
import { computed } from 'vue'

interface Props {
  currentStage?: 'recovery' | 'expansion' | 'slowdown' | 'contraction'
  stageName?: string
  confidence?: number
  equityBondRatio?: number
  inflationState?: string
  interestState?: string
  ratio?: number
  percentile1500?: number
  totalScore?: number
  zoneName?: string
  actionSummary?: string
}

const props = withDefaults(defineProps<Props>(), {
  currentStage: 'recovery',
  stageName: '弱复苏与海外结构性繁荣分化',
  confidence: 88,
  equityBondRatio: 2.8097,
  inflationState: '温和回升',
  interestState: '流动性充裕',
  ratio: 0.1846,
  percentile1500: 24.33,
  totalScore: 6.57,
  zoneName: '中间区',
  actionSummary: '暂存现金池'
})

const confPct = computed(() => Math.min(100, Math.max(0, props.confidence)))
const quantilePct = computed(() => {
  const val = props.percentile1500 <= 1 ? props.percentile1500 * 100 : props.percentile1500
  return Math.min(100, Math.max(0, val))
})
const scorePct = computed(() => Math.min(100, Math.max(0, (props.totalScore / 20) * 100)))

const outerRingStyle = computed(() => ({
  background: `conic-gradient(#e5b352 0% ${confPct.value}%, rgba(229,179,82,0.16) ${confPct.value}% 100%)`
}))
const midRingStyle = computed(() => ({
  background: `conic-gradient(#2ad1a3 0% ${quantilePct.value}%, rgba(42,209,163,0.16) ${quantilePct.value}% 100%)`
}))
const innerRingStyle = computed(() => ({
  background: `conic-gradient(#f05a50 0% ${scorePct.value}%, rgba(240,90,80,0.16) ${scorePct.value}% 100%)`
}))
</script>

<template>
  <view class="rounded-2xl bg-[#12161b] border border-white/[0.09] p-4 text-[#f1efea]">
    <!-- 顶部状态条：方向四语义胶囊 -->
    <view class="flex items-start justify-between gap-2 pb-3 border-b border-white/[0.08]">
      <view class="min-w-0 flex-1">
        <text class="text-[11px] font-medium text-[#8d99a4] block">宏观周期与策略三环体征</text>
        <text class="text-[15px] font-bold text-[#f1efea] tracking-tight leading-snug mt-0.5 block">
          {{ stageName }}
        </text>
      </view>
      <view class="shrink-0 mt-0.5 rounded-full bg-[#e5b352]/18 px-2.5 py-1 text-[11px] font-bold text-[#e5b352] whitespace-nowrap">
        {{ zoneName }} · {{ actionSummary }}
      </view>
    </view>

    <!-- 签名核心：左三同心闭环（方向四） + 右精密数字刻度（方向一） -->
    <view class="mt-4 flex items-center justify-between gap-4">
      <!-- 纯 CSS 三同心环（100% 兼容微信小程序 WXML/WXSS） -->
      <view class="relative w-36 h-36 shrink-0 flex items-center justify-center">
        <!-- 外环：宏观置信度 -->
        <view class="w-36 h-36 rounded-full flex items-center justify-center" :style="outerRingStyle">
          <view class="w-[120px] h-[120px] rounded-full bg-[#12161b] flex items-center justify-center">
            <!-- 中环：1500 日轮动分位 -->
            <view class="w-[110px] h-[110px] rounded-full flex items-center justify-center" :style="midRingStyle">
              <view class="w-[88px] h-[88px] rounded-full bg-[#12161b] flex items-center justify-center">
                <!-- 内环：红利四维评分 -->
                <view class="w-[78px] h-[78px] rounded-full flex items-center justify-center" :style="innerRingStyle">
                  <view class="w-[58px] h-[58px] rounded-full bg-[#12161b] flex flex-col items-center justify-center">
                    <text class="text-[13px] font-bold font-mono cc-tabular text-[#f1efea] leading-none">
                      {{ ratio.toFixed(4) }}
                    </text>
                    <text class="text-[9px] font-semibold text-[#e5b352] mt-1 leading-none">
                      {{ zoneName }}
                    </text>
                  </view>
                </view>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 右侧三环图例与精密等宽读数 -->
      <view class="flex-1 space-y-2.5 min-w-0">
        <view class="flex items-center justify-between pb-2 border-b border-white/[0.07]">
          <view class="flex items-center space-x-1.5">
            <view class="w-2 h-2 rounded-full bg-[#e5b352] shrink-0" />
            <text class="text-xs text-[#9aa5b1]">宏观置信度</text>
          </view>
          <text class="text-base font-bold font-mono cc-tabular text-[#e5b352]">{{ confPct }}%</text>
        </view>

        <view class="flex items-center justify-between pb-2 border-b border-white/[0.07]">
          <view class="flex items-center space-x-1.5">
            <view class="w-2 h-2 rounded-full bg-[#2ad1a3] shrink-0" />
            <text class="text-xs text-[#9aa5b1]">1500日分位</text>
          </view>
          <text class="text-base font-bold font-mono cc-tabular text-[#2ad1a3]">{{ quantilePct.toFixed(1) }}%</text>
        </view>

        <view class="flex items-center justify-between">
          <view class="flex items-center space-x-1.5">
            <view class="w-2 h-2 rounded-full bg-[#f05a50] shrink-0" />
            <text class="text-xs text-[#9aa5b1]">红利四维分</text>
          </view>
          <view class="flex items-baseline">
            <text class="text-base font-bold font-mono cc-tabular text-[#ff736a]">{{ totalScore.toFixed(2) }}</text>
            <text class="text-[10px] font-mono text-[#64707c] ml-0.5">/20</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部发丝线宏观指标带 -->
    <view class="mt-4 pt-3 border-t border-white/[0.08] grid grid-cols-2 gap-3">
      <view class="bg-[#181e25] rounded-xl px-3 py-2.5 border border-white/[0.06]">
        <text class="text-[10px] text-[#8d99a4] block">股债相对强弱 (ERP)</text>
        <view class="mt-1 flex items-center justify-between">
          <text class="text-base font-bold font-mono cc-tabular text-[#f1efea]">{{ equityBondRatio }}%</text>
          <text class="rounded-full bg-[#2ad1a3]/16 px-2 py-0.5 text-[10px] font-semibold text-[#2ad1a3]">偏向权益</text>
        </view>
      </view>

      <view class="bg-[#181e25] rounded-xl px-3 py-2.5 border border-white/[0.06]">
        <text class="text-[10px] text-[#8d99a4] block">通胀与流动性</text>
        <view class="mt-1 flex items-center justify-between gap-1">
          <text class="text-sm font-bold text-[#e5b352] truncate">{{ inflationState }}</text>
          <text class="text-[10px] text-[#9aa5b1] shrink-0">{{ interestState }}</text>
        </view>
      </view>
    </view>
  </view>
</template>
