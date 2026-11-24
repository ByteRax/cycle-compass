<script lang="ts" setup>
import { computed, ref } from 'vue'

interface Props {
  datasetTitle: string
  asOfDate?: string
}

const props = withDefaults(defineProps<Props>(), {
  datasetTitle: 'A股',
  asOfDate: ''
})

const displayDate = computed(() => {
  return props.asOfDate || '2026/11/25'
})

const selectedPeriod = ref<'日线' | '周线' | '月线'>('日线')

// 模拟趋势数据点 (SVG 渲染)
const svgW = 660
const svgH = 260
const padL = 40
const padR = 20
const padT = 20
const padB = 30

const pricePoints = [
  { x: 40, y: 180, ma20: 175, ma60: 165 },
  { x: 100, y: 160, ma20: 168, ma60: 163 },
  { x: 160, y: 175, ma20: 164, ma60: 162 },
  { x: 220, y: 140, ma20: 155, ma60: 160 },
  { x: 280, y: 130, ma20: 145, ma60: 158 },
  { x: 340, y: 110, ma20: 132, ma60: 152 },
  { x: 400, y: 125, ma20: 124, ma60: 146 },
  { x: 460, y: 95, ma20: 112, ma60: 138 },
  { x: 520, y: 70, ma20: 98, ma60: 128 },
  { x: 580, y: 65, ma20: 82, ma60: 118 },
  { x: 640, y: 55, ma20: 70, ma60: 108 }
]

const pricePath = computed(() => {
  return pricePoints.reduce((acc, p, i) => `${acc} ${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`, '')
})

const ma20Path = computed(() => {
  return pricePoints.reduce((acc, p, i) => `${acc} ${i === 0 ? 'M' : 'L'} ${p.x} ${p.ma20}`, '')
})

const ma60Path = computed(() => {
  return pricePoints.reduce((acc, p, i) => `${acc} ${i === 0 ? 'M' : 'L'} ${p.x} ${p.ma60}`, '')
})
</script>

<template>
  <view class="w-full">
    <!-- 头部趋势信号栏 -->
    <view class="mb-3 flex items-center justify-between">
      <view>
        <text class="text-xl font-black text-[#f1efea] tracking-tight block">{{ datasetTitle }} · 趋势曲线</text>
        <view class="flex items-center space-x-2 mt-1">
          <text class="text-[10px] font-mono text-[#e5b352]">{{ displayDate }}</text>
          <text class="text-white/20 text-[10px]">|</text>
          <text class="text-[10px] font-mono text-[#8d99a4]">最新基准价: 30.58</text>
          <text class="px-1.5 py-0.5 rounded text-[9px] font-bold bg-[#ff736a]/20 text-[#ff736a] border border-[#ff736a]/30">
            多头排列 · 趋势加速
          </text>
        </view>
      </view>
      <!-- 周期粒度切换 -->
      <view class="flex items-center space-x-1 bg-[#12181F] p-1 rounded-lg border border-white/[0.06]">
        <text
          v-for="p in ['日线', '周线', '月线']"
          :key="p"
          class="px-2 py-0.5 rounded text-[10px] font-mono cursor-pointer transition-all"
          :class="selectedPeriod === p ? 'bg-[#e5b352] text-[#090C0F] font-bold' : 'text-[#8d99a4]'"
          @tap="selectedPeriod = p as any"
        >
          {{ p }}
        </text>
      </view>
    </view>

    <!-- 趋势主图 (价格走势 + 均线系统) -->
    <view class="rounded-2xl bg-[#12181F] border border-white/[0.08] p-3 shadow-xl mb-3">
      <!-- 图例 -->
      <view class="flex items-center space-x-3 pb-2 border-b border-white/[0.05] text-[10px] font-mono">
        <view class="flex items-center space-x-1">
          <text class="w-2.5 h-0.5 bg-[#f1efea]" />
          <text class="text-[#f1efea]">收盘价</text>
        </view>
        <view class="flex items-center space-x-1">
          <text class="w-2.5 h-0.5 bg-[#e5b352]" />
          <text class="text-[#e5b352]">MA20</text>
        </view>
        <view class="flex items-center space-x-1">
          <text class="w-2.5 h-0.5 bg-[#38bdf8]" />
          <text class="text-[#38bdf8]">MA60</text>
        </view>
      </view>

      <!-- SVG 曲线走势 -->
      <view class="w-full relative mt-2">
        <svg :viewBox="`0 0 ${svgW} ${svgH}`" class="w-full h-auto block" preserveAspectRatio="xMidYMid meet">
          <!-- 背景网格虚线 -->
          <line x1="40" y1="60" x2="640" y2="60" stroke="rgba(255,255,255,0.06)" stroke-dasharray="2,2" />
          <line x1="40" y1="120" x2="640" y2="120" stroke="rgba(255,255,255,0.06)" stroke-dasharray="2,2" />
          <line x1="40" y1="180" x2="640" y2="180" stroke="rgba(255,255,255,0.06)" stroke-dasharray="2,2" />

          <!-- MA60 蓝线 -->
          <path :d="ma60Path" fill="none" stroke="#38bdf8" stroke-width="1.5" stroke-dasharray="4,2" />
          <!-- MA20 金线 -->
          <path :d="ma20Path" fill="none" stroke="#e5b352" stroke-width="2" />
          <!-- 价格白线 -->
          <path :d="pricePath" fill="none" stroke="#f1efea" stroke-width="2.5" stroke-linecap="round" />

          <!-- 当前价格圆点与水波高亮 -->
          <circle cx="640" cy="55" r="4" fill="#e5b352" />
          <circle cx="640" cy="55" r="8" fill="none" stroke="#e5b352" stroke-opacity="0.4" stroke-width="1.5" />

          <!-- X 轴刻度 -->
          <text x="40" y="240" fill="#64707c" font-size="10" font-family="monospace">08/15</text>
          <text x="240" y="240" fill="#64707c" font-size="10" font-family="monospace">09/20</text>
          <text x="440" y="240" fill="#64707c" font-size="10" font-family="monospace">10/25</text>
          <text x="640" y="240" text-anchor="end" fill="#e5b352" font-size="10" font-family="monospace" font-weight="bold">11/24 (最新)</text>
        </svg>
      </view>
    </view>

    <!-- 趋势体征诊断卡 -->
    <view class="grid grid-cols-3 gap-2">
      <view class="rounded-xl bg-[#12181F] border border-white/[0.08] p-2.5">
        <text class="text-[10px] text-[#8d99a4] block">趋势强度</text>
        <text class="text-base font-black font-mono text-[#ff736a] mt-0.5 block">82.5 <text class="text-xs">分</text></text>
        <text class="text-[9px] text-[#64707c] mt-0.5 block">主升通道中轨</text>
      </view>
      <view class="rounded-xl bg-[#12181F] border border-white/[0.08] p-2.5">
        <text class="text-[10px] text-[#8d99a4] block">均线发散率</text>
        <text class="text-base font-black font-mono text-[#e5b352] mt-0.5 block">+6.8%</text>
        <text class="text-[9px] text-[#64707c] mt-0.5 block">MA20 / MA60 扩大</text>
      </view>
      <view class="rounded-xl bg-[#12181F] border border-white/[0.08] p-2.5">
        <text class="text-[10px] text-[#8d99a4] block">支撑阻力带</text>
        <text class="text-base font-black font-mono text-[#38bdf8] mt-0.5 block">29.40</text>
        <text class="text-[9px] text-[#64707c] mt-0.5 block">强下沿防护</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
</style>
