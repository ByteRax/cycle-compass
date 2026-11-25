<script lang="ts">
import type { ChartType } from '@/api/types/trend'

export interface ChartRegistryItem {
  type: ChartType
  index: number
  label: string
  desc: string
}

export const CHART_REGISTRY: ChartRegistryItem[] = [
  {
    type: 'curve',
    index: 1,
    label: '趋势曲线',
    desc: '价格走势 · 基准 · 极值区间'
  },
  {
    type: 'annual',
    index: 2,
    label: '年度价格走势',
    desc: '历年同周季节性对比'
  },
  {
    type: 'components',
    index: 3,
    label: '品种成分趋势',
    desc: '成分归一化与截面表'
  },
  {
    type: 'dashboard',
    index: 4,
    label: '趋势仪表盘',
    desc: '报告预设与双周期强度'
  },
  {
    type: 'season',
    index: 5,
    label: '趋势节气图',
    desc: '趋势阶段分布与成员'
  },
  {
    type: 'distribution',
    index: 6,
    label: '右侧分布图',
    desc: '市值/成交额树图分布'
  },
  {
    type: 'tier',
    index: 7,
    label: '趋势梯队榜',
    desc: 'S~D 五级强弱梯队'
  }
]
</script>

<script lang="ts" setup>
/* eslint-disable import/first */
import { computed } from 'vue'
import { useFinanceStore } from '@/stores/finance'

interface Props {
  visible?: boolean
  currentChart?: ChartType
  availableCharts?: ChartType[]
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  currentChart: 'curve'
})

const emit = defineEmits<{
  (e: 'select', chart: ChartType): void
  (e: 'close'): void
}>()

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

function isAvailable (chart: ChartType): boolean {
  if (!props.availableCharts || props.availableCharts.length === 0)
    return true
  return props.availableCharts.includes(chart)
}

function handleSelect (item: ChartRegistryItem) {
  if (!isAvailable(item.type))
    return
  emit('select', item.type)
  emit('close')
}
</script>

<template>
  <view v-if="visible" class="relative z-50">
    <!-- 遮罩背景 -->
    <view
      class="fixed inset-0 transition-opacity"
      :class="isDark ? 'bg-black/60' : 'bg-black/35'"
      @tap="emit('close')"
    />

    <!-- 下拉悬浮菜单 -->
    <view
      class="fixed top-12 left-4 w-68 rounded-2xl border p-2 shadow-2xl z-50"
      :class="isDark ? 'bg-[#141A22] border-white/[0.12]' : 'bg-white border-[#E2E8F0]'"
    >
      <view
        class="divide-y"
        :class="isDark ? 'divide-white/[0.04]' : 'divide-[#F1F5F9]'"
      >
        <view
          v-for="item in CHART_REGISTRY"
          :key="item.type"
          class="flex items-center justify-between px-3 py-2.5 rounded-xl transition-all"
          :class="[
            !isAvailable(item.type)
              ? 'opacity-40 cursor-not-allowed'
              : currentChart === item.type
                ? (isDark ? 'bg-white/[0.08] text-white font-bold' : 'bg-[#F8FAFC] text-[#0F172A] font-bold')
                : (isDark ? 'text-[#8D99A4] active:bg-white/[0.04]' : 'text-[#475569] active:bg-[#F8FAFC]')
          ]"
          @tap="handleSelect(item)"
        >
          <view class="flex items-center gap-2.5 min-w-0">
            <text
              class="text-xs font-mono w-4 text-center shrink-0"
              :class="currentChart === item.type ? 'text-[#E5B352] font-bold' : (isDark ? 'text-[#64707C]' : 'text-[#94A3B8]')"
            >
              {{ item.index }}
            </text>
            <view class="min-w-0">
              <text class="text-[13px] tracking-wide block truncate">{{ item.label }}</text>
              <text
                class="text-[10px] block truncate mt-0.5"
                :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'"
              >
                {{ isAvailable(item.type) ? item.desc : '当前资产无成员能力' }}
              </text>
            </view>
          </view>
          <text
            v-if="currentChart === item.type"
            class="text-sm font-black shrink-0 ml-2"
            :class="isDark ? 'text-[#E5B352]' : 'text-[#9A6B16]'"
          >
            ✓
          </text>
          <text
            v-else-if="!isAvailable(item.type)"
            class="text-[10px] px-1.5 py-0.5 rounded shrink-0 ml-2"
            :class="isDark ? 'bg-white/[0.06] text-[#64707C]' : 'bg-[#F1F5F9] text-[#94A3B8]'"
          >
            不可用
          </text>
        </view>
      </view>
    </view>
  </view>
</template>
