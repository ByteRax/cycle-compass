<script lang="ts" setup>
import { computed } from 'vue'
import { TREND_TEMP_DICT, type TrendTemperature } from '@/data/divergence'

interface Props {
  datasetTitle: string
}

const props = withDefaults(defineProps<Props>(), {
  datasetTitle: 'A股'
})

const score = 78 // 0 ~ 100 分

// 动态映射 7 级趋势温度
const currentTempKey = computed<TrendTemperature>(() => {
  if (score >= 86) return 'boiling'
  if (score >= 72) return 'hot'
  if (score >= 58) return 'warm'
  if (score >= 43) return 'flat'
  if (score >= 29) return 'cool'
  if (score >= 15) return 'cold'
  return 'freezing'
})

const currentTemp = computed(() => {
  return TREND_TEMP_DICT[currentTempKey.value]
})

// 仪表盘指针角度 (-90deg ~ +90deg)
const needleDeg = computed(() => {
  return -90 + (score / 100) * 180
})

const tempScaleList: { key: TrendTemperature; label: string; min: number; max: number }[] = [
  { key: 'freezing', label: '冻', min: 0, max: 14 },
  { key: 'cold', label: '寒', min: 15, max: 28 },
  { key: 'cool', label: '凉', min: 29, max: 42 },
  { key: 'flat', label: '平', min: 43, max: 57 },
  { key: 'warm', label: '温', min: 58, max: 71 },
  { key: 'hot', label: '热', min: 72, max: 85 },
  { key: 'boiling', label: '沸', min: 86, max: 100 }
]
</script>

<template>
  <view class="w-full">
    <!-- 标题 -->
    <view class="mb-3">
      <text class="text-xl font-black text-[#f1efea] tracking-tight block">{{ datasetTitle }} · 趋势仪表盘</text>
      <text class="text-[10px] font-mono text-[#8d99a4] mt-0.5 block">7级趋势绝对体感量化与买卖策略全景</text>
    </view>

    <!-- 主半圆仪表盘卡片 (7 级趋势温度刻度) -->
    <view class="rounded-2xl bg-[#12181F] border border-white/[0.08] p-4 shadow-xl mb-3 flex flex-col items-center relative overflow-hidden">
      <!-- 半圆仪表盘 SVG -->
      <view class="w-64 h-34 relative flex items-end justify-center">
        <svg viewBox="0 0 220 120" class="w-full h-full block">
          <!-- 背景圆弧 -->
          <path d="M 25 110 A 85 85 0 0 1 195 110" fill="none" stroke="rgba(255,255,255,0.08)" stroke-width="14" stroke-linecap="round" />
          
          <!-- 渐变高亮进度弧 (通过 stroke-dasharray) -->
          <path
            d="M 25 110 A 85 85 0 0 1 195 110"
            fill="none"
            stroke="url(#tempGrad)"
            stroke-width="14"
            stroke-linecap="round"
            stroke-dasharray="267"
            :stroke-dashoffset="267 - (267 * score) / 100"
          />
          <defs>
            <linearGradient id="tempGrad" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stop-color="#818CF8" />
              <stop offset="25%" stop-color="#38BDF8" />
              <stop offset="50%" stop-color="#94A3B8" />
              <stop offset="75%" stop-color="#E5B352" />
              <stop offset="100%" stop-color="#FF2E51" />
            </linearGradient>
          </defs>

          <!-- 指针 -->
          <g :transform="`rotate(${needleDeg}, 110, 110)`">
            <line x1="110" y1="110" x2="110" y2="35" stroke="#f1efea" stroke-width="2.5" stroke-linecap="round" />
            <circle cx="110" cy="110" r="5.5" :fill="currentTemp.color" />
          </g>
        </svg>

        <!-- 7 级温度扇区标记 -->
        <view class="absolute inset-x-1 bottom-0 flex justify-between text-[9px] font-bold font-mono px-2 pointer-events-none">
          <text v-for="s in tempScaleList" :key="s.key" :style="{ color: s.key === currentTempKey ? currentTemp.color : '#64707c' }">
            {{ s.label }}
          </text>
        </view>
      </view>

      <!-- 核心体感读数与结论 -->
      <view class="text-center mt-2.5">
        <view class="flex items-center justify-center space-x-2">
          <text
            class="text-xs font-black px-2 py-0.5 rounded border"
            :style="{
              color: currentTemp.color,
              backgroundColor: currentTemp.bg,
              borderColor: currentTemp.border
            }"
          >
            体感：{{ currentTemp.label }} ({{ currentTemp.nature }})
          </text>
          <text class="text-2xl font-black font-mono text-[#f1efea]">{{ score }} 分</text>
        </view>
        <view class="flex items-center justify-center space-x-1.5 mt-1.5">
          <text class="w-2 h-2 rounded-full animate-pulse" :style="{ backgroundColor: currentTemp.color }" />
          <text class="text-xs font-bold" :style="{ color: currentTemp.color }">{{ currentTemp.desc }}</text>
        </view>
      </view>

      <!-- 买卖方即时行动建议条 -->
      <view class="w-full mt-3 pt-3 border-t border-white/[0.06] grid grid-cols-2 gap-2 text-left">
        <view class="bg-[#090C0F] p-2 rounded-lg border border-white/[0.05]">
          <text class="text-[9px] text-[#8d99a4] block">买方视角行动</text>
          <text class="text-[11px] font-bold text-[#ff736a] mt-0.5 block leading-tight">{{ currentTemp.actionBuy }}</text>
        </view>
        <view class="bg-[#090C0F] p-2 rounded-lg border border-white/[0.05]">
          <text class="text-[9px] text-[#8d99a4] block">卖方视角行动</text>
          <text class="text-[11px] font-bold text-[#38bdf8] mt-0.5 block leading-tight">{{ currentTemp.actionSell }}</text>
        </view>
      </view>
    </view>

    <!-- 3 项细分动能维度指标 -->
    <view class="grid grid-cols-3 gap-2 mb-3">
      <view class="rounded-xl bg-[#12181F] border border-white/[0.08] p-2.5">
        <text class="text-[10px] text-[#8d99a4] block">多头占比</text>
        <text class="text-base font-black font-mono text-[#ff736a] mt-0.5 block">68.5%</text>
        <text class="text-[9px] text-[#64707c] mt-0.5 block">站上 MA20 标的</text>
      </view>
      <view class="rounded-xl bg-[#12181F] border border-white/[0.08] p-2.5">
        <text class="text-[10px] text-[#8d99a4] block">动量加速度</text>
        <text class="text-base font-black font-mono text-[#e5b352] mt-0.5 block">+14.2%</text>
        <text class="text-[9px] text-[#64707c] mt-0.5 block">环比动能放大</text>
      </view>
      <view class="rounded-xl bg-[#12181F] border border-white/[0.08] p-2.5">
        <text class="text-[10px] text-[#8d99a4] block">拥挤度分位</text>
        <text class="text-base font-black font-mono text-[#2ec4b6] mt-0.5 block">42%</text>
        <text class="text-[9px] text-[#64707c] mt-0.5 block">未达「沸」泡沫区</text>
      </view>
    </view>

    <!-- 趋势温度核心纪律卡片 -->
    <view class="rounded-xl bg-[#12181F] border border-white/[0.08] p-3 space-y-1.5 text-[10px] leading-relaxed">
      <text class="font-bold text-[#e5b352] block text-xs mb-1">💡 趋势温度策略铁律</text>
      <text class="text-[#8d99a4] block">· <text class="text-white font-medium">双向确认链</text>：上升看「温」预警、「热」确认、「沸」泡沫；下跌看「凉」预警、「寒」确认、「冻」超跌。</text>
      <text class="text-[#8d99a4] block">· <text class="text-white font-medium">拥挤度规避</text>：坚决避免在「沸」时买入（极度拥挤追高），避免在「冻」时卖出（极致恐慌割肉）。</text>
      <text class="text-[#8d99a4] block">· <text class="text-white font-medium">级别与领头羊</text>：大类风格、板块温度的参考性远高于单一微观个体，率先转「热」的大板块即是领涨主线。</text>
    </view>
  </view>
</template>

<style scoped>
</style>
