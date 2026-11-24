<script lang="ts" setup>
import { ref } from 'vue'

interface Props {
  datasetTitle: string
}

withDefaults(defineProps<Props>(), {
  datasetTitle: 'A股'
})

const activeYear = ref<string>('2026')

const years = [
  { year: '2026', color: '#e5b352', label: '2026 (当前)' },
  { year: '2025', color: '#ff736a', label: '2025' },
  { year: '2024', color: '#38bdf8', label: '2024' },
  { year: '2023', color: '#a78bfa', label: '2023' },
  { year: '2022', color: '#64707c', label: '2022' }
]

// 12 个月份的历史上涨胜率
const monthStats = [
  { month: '1月', winRate: 48, avgReturn: 0.8 },
  { month: '2月', winRate: 75, avgReturn: 3.2, highlight: true }, // 春季躁动
  { month: '3月', winRate: 52, avgReturn: 0.5 },
  { month: '4月', winRate: 40, avgReturn: -1.2 }, // 四月决断
  { month: '5月', winRate: 58, avgReturn: 1.4 },
  { month: '6月', winRate: 45, avgReturn: -0.6 },
  { month: '7月', winRate: 62, avgReturn: 2.1 }, // 七翻身
  { month: '8月', winRate: 42, avgReturn: -1.0 },
  { month: '9月', winRate: 38, avgReturn: -1.5 },
  { month: '10月', winRate: 68, avgReturn: 1.9 }, // 金九银十
  { month: '11月', winRate: 65, avgReturn: 1.8, current: true }, // 当前月
  { month: '12月', winRate: 55, avgReturn: 0.9 }
]
</script>

<template>
  <view class="w-full">
    <!-- 标题与年度切换 -->
    <view class="mb-3 flex items-center justify-between">
      <view>
        <text class="text-xl font-black text-[#f1efea] tracking-tight block">{{ datasetTitle }} · 年度价格走势</text>
        <text class="text-[10px] font-mono text-[#8d99a4] mt-0.5 block">历年 1~12 月季节性归一化对比</text>
      </view>
      <!-- 年份高亮徽章 -->
      <view class="flex items-center space-x-1">
        <text
          v-for="y in years"
          :key="y.year"
          class="px-1.5 py-0.5 rounded text-[9px] font-mono cursor-pointer transition-all"
          :class="activeYear === y.year ? 'bg-[#e5b352] text-[#090C0F] font-bold' : 'text-[#8d99a4] bg-white/[0.04]'"
          @tap="activeYear = y.year"
        >
          {{ y.year }}
        </text>
      </view>
    </view>

    <!-- 年度走势多曲线叠加主图 (SVG) -->
    <view class="rounded-2xl bg-[#12181F] border border-white/[0.08] p-3 shadow-xl mb-3">
      <view class="flex items-center justify-between pb-2 border-b border-white/[0.05] text-[10px] font-mono">
        <text class="text-[#8d99a4]">累计收益率对比 (1月基准为0%)</text>
        <text class="text-[#e5b352] font-bold">11月当前收益: +12.4%</text>
      </view>

      <view class="w-full relative mt-2">
        <svg viewBox="0 0 660 220" class="w-full h-auto block" preserveAspectRatio="xMidYMid meet">
          <!-- 基准线 0% -->
          <line x1="40" y1="130" x2="630" y2="130" stroke="rgba(255,255,255,0.2)" stroke-dasharray="3,3" />
          <text x="34" y="133.5" fill="#8d99a4" font-size="9" text-anchor="end" font-family="monospace">0%</text>

          <!-- 2022 灰线 -->
          <path d="M 50 130 Q 150 160, 250 145 T 450 170 T 630 180" fill="none" stroke="#64707c" stroke-width="1.2" opacity="0.6" />
          <!-- 2023 紫线 -->
          <path d="M 50 130 Q 150 110, 250 120 T 450 150 T 630 155" fill="none" stroke="#a78bfa" stroke-width="1.4" opacity="0.7" />
          <!-- 2024 蓝线 -->
          <path d="M 50 130 Q 150 140, 250 135 T 450 130 T 630 110" fill="none" stroke="#38bdf8" stroke-width="1.5" opacity="0.8" />
          <!-- 2025 珊瑚红线 -->
          <path d="M 50 130 Q 150 100, 250 80 T 450 95 T 630 75" fill="none" stroke="#ff736a" stroke-width="1.8" opacity="0.85" />
          <!-- 2026 金色加粗实线 (进行中到11月) -->
          <path d="M 50 130 Q 150 90, 250 70 T 450 65 L 575 55" fill="none" stroke="#e5b352" stroke-width="3" stroke-linecap="round" />
          <circle cx="575" cy="55" r="4" fill="#e5b352" />
          <text x="582" y="58" fill="#e5b352" font-size="10" font-family="monospace" font-weight="bold">+12.4%</text>

          <!-- X 轴月份刻度 -->
          <text v-for="m in 12" :key="m" :x="50 + (m - 1) * 52" y="210" fill="#64707c" font-size="9" text-anchor="middle" font-family="monospace">
            {{ m }}月
          </text>
        </svg>
      </view>
    </view>

    <!-- 季节性胜率直方图卡片 -->
    <view class="rounded-2xl bg-[#12181F] border border-white/[0.08] p-3 shadow-xl">
      <view class="flex items-center justify-between pb-2 mb-2 border-b border-white/[0.05]">
        <text class="text-xs font-bold text-[#f1efea]">近 10 年各月份历史上涨概率</text>
        <text class="text-[10px] font-mono text-[#8d99a4]">季节性规律分析</text>
      </view>
      <view class="grid grid-cols-12 gap-1 items-end h-20 pt-2">
        <view
          v-for="st in monthStats"
          :key="st.month"
          class="flex flex-col items-center justify-end h-full"
        >
          <text class="text-[8px] font-mono mb-1" :class="st.current ? 'text-[#e5b352] font-bold' : st.winRate >= 60 ? 'text-[#ff736a]' : 'text-[#64707c]'">
            {{ st.winRate }}%
          </text>
          <view
            class="w-full rounded-t-xs transition-all"
            :style="{ height: `${st.winRate * 0.7}%` }"
            :class="st.current ? 'bg-[#e5b352]' : st.winRate >= 60 ? 'bg-[#ff736a]/80' : 'bg-white/[0.12]'"
          />
          <text class="text-[8px] font-mono text-[#8d99a4] mt-1">{{ st.month.replace('月', '') }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<style scoped>
</style>
