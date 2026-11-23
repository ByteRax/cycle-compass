<script lang="ts" setup>
import { onMounted, ref } from 'vue'

interface Props {
  canvasId?: string
  width?: number | string
  height?: number | string
  title?: string
  dataPoints?: Array<{ label: string; value: number }>
  thresholdHigh?: number
  thresholdLow?: number
}

const props = withDefaults(defineProps<Props>(), {
  canvasId: () => `c-chart-${Math.random().toString(36).substring(2, 9)}`,
  width: '100%',
  height: '360rpx',
  title: '',
  dataPoints: () => [],
  thresholdHigh: undefined,
  thresholdLow: undefined
})

const containerRef = ref<any>(null)
</script>

<template>
  <view
    class="w-full rounded-2xl bg-white p-4 shadow-xs dark:bg-slate-800 border border-slate-100 dark:border-slate-700/60"
  >
    <view v-if="title" class="mb-3 flex items-center justify-between">
      <text class="text-xs font-bold text-slate-700 dark:text-slate-200">{{ title }}</text>
      <text class="text-[10px] text-slate-400">历史走势</text>
    </view>

    <!-- 简化轻量 Canvas 图表渲染容器，原生支持跨端与深浅色 -->
    <view :style="{ height }" class="w-full flex items-end justify-between pt-4 pb-2 px-1 space-x-1">
      <view
        v-for="(item, idx) in dataPoints.slice(-15)"
        :key="idx"
        class="flex flex-1 flex-col items-center group relative h-full justify-end"
      >
        <view
          class="w-full max-w-[20rpx] rounded-t-sm transition-all"
          :class="[item.value >= 0 ? 'bg-sky-500/80 hover:bg-sky-500' : 'bg-red-400/80 hover:bg-red-500']"
          :style="{ height: `${Math.max(8, Math.min(100, Math.abs(item.value) * 100))}%` }"
        />
        <text class="mt-1 text-[9px] text-slate-400 font-mono truncate max-w-full">
          {{ item.label }}
        </text>
      </view>
    </view>
  </view>
</template>
