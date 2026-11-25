<script lang="ts" setup>
import type { AnnualTrendResponse } from '@/api/types/trend'
import type { TrendLineSeries } from '@/components/trend/TrendLineCanvas.vue'
import { computed, ref, watch } from 'vue'
import { getTrendAnnual } from '@/api/trend'
import TrendLineCanvas from '@/components/trend/TrendLineCanvas.vue'
import { useFinanceStore } from '@/stores/finance'
import { deterministicAssetColor } from '@/utils/chartGeometry'

const props = withDefaults(defineProps<{
  assetId: string
  snapshotId?: string
  range?: string
  currency?: string
  scopeId?: string
  benchmarkId?: string
}>(), {
  snapshotId: '',
  range: '1Y',
  currency: 'CNY',
  scopeId: '',
  benchmarkId: ''
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)
const isCnRedUp = computed(() => financeStore.colorScheme === 'cn_red_up')

const loading = ref(false)
const errorMsg = ref('')
const annualData = ref<AnnualTrendResponse | null>(null)
const focusedYear = ref<number | null>(null)
const hiddenYears = ref<number[]>([])

function formatPct(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(2)}%`
}

function getGainColor(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return isDark.value ? '#8D99A4' : '#626F77'
  if (val === 0)
    return isDark.value ? '#F1EFEA' : '#22292E'
  const isUp = val > 0
  if (isCnRedUp.value) {
    return isUp
? (isDark.value ? '#FF736A' : '#B93842')
: (isDark.value ? '#2EC4B6' : '#14745E')
}
  return isUp
? (isDark.value ? '#2EC4B6' : '#14745E')
: (isDark.value ? '#FF736A' : '#B93842')
}

async function fetchAnnualData() {
  if (!props.assetId)
    return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendAnnual(props.assetId, {
      currency: props.currency || undefined,
      snapshot_id: props.snapshotId || undefined
    })
    annualData.value = res.data ?? null
    focusedYear.value = null
    hiddenYears.value = []
  }
  catch (err: any) {
    errorMsg.value = err?.message || '年度同期走势加载失败'
  }
  finally {
    loading.value = false
  }
}

watch(
  () => [props.assetId, props.snapshotId, props.currency],
  () => {
    fetchAnnualData()
  },
  { immediate: true }
)

const weekNumbers = computed<number[]>(() => {
  const years = annualData.value?.years || []
  const set = new Set<number>()
  for (const y of years) {
    for (const w of y.weekly_returns)
      set.add(w.week)
  }
  return Array.from(set).sort((a, b) => a - b)
})

const weekLabels = computed(() => weekNumbers.value.map(w => `W${w}`))

function getYearColor(year: number, isCurrent: boolean, idx: number): string {
  if (isCurrent)
    return isDark.value ? '#E5B352' : '#8A641F'
  return deterministicAssetColor(String(year), idx + 1)
}

const chartSeries = computed<TrendLineSeries[]>(() => {
  const years = annualData.value?.years || []
  const weeks = weekNumbers.value
  if (!years.length || !weeks.length)
    return []

  return years.map((y, idx) => {
    const map = new Map<number, number | null>()
    for (const pt of y.weekly_returns)
      map.set(pt.week, pt.return_pct)

    const isFocused = focusedYear.value === y.year
    const isDimmed = focusedYear.value !== null
? !isFocused
: (!y.is_current && years.length > 4)

    return {
      id: String(y.year),
      name: y.is_current
? `${y.year}(今)`
: String(y.year),
      color: getYearColor(y.year, y.is_current, idx),
      values: weeks.map(w => map.get(w) ?? null),
      highlighted: y.is_current || isFocused,
      dimmed: isDimmed,
      hidden: hiddenYears.value.includes(y.year),
      lineWidth: y.is_current || isFocused
? 2.6
: 1.5
    }
  })
})

function toggleYearFocus(year: number) {
  if (focusedYear.value === year)
    focusedYear.value = null
  else
    focusedYear.value = year
}

const sortedYears = computed(() => {
  return [
...(annualData.value?.years || [])
].sort((a, b) => b.year - a.year)
})
</script>

<template>
  <view class="w-full space-y-3">
    <!-- 标题栏 -->
    <view class="flex items-start justify-between gap-2">
      <view>
        <text class="text-[18px] font-bold block" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
          {{ annualData?.asset_name || assetId }} · 年度同期走势
        </text>
        <text class="text-[11px] font-mono mt-0.5 block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          {{ annualData?.snapshot?.as_of || '—' }} · 自然年周次归一化对比（第 {{ annualData?.cutoff_week ?? '—' }} 周）
        </text>
      </view>
    </view>

    <!-- 核心统计指标卡 -->
    <view v-if="annualData" class="grid grid-cols-3 gap-2">
      <view
        class="rounded-xl border p-2.5"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          今年累计
        </text>
        <text
          class="text-[15px] font-bold font-mono tabular-nums mt-0.5 block"
          :style="{ color: getGainColor(annualData.stats.current_year_return_pct) }"
        >
          {{ formatPct(annualData.stats.current_year_return_pct) }}
        </text>
      </view>

      <view
        class="rounded-xl border p-2.5"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          往年同周中位
        </text>
        <text
          class="text-[15px] font-bold font-mono tabular-nums mt-0.5 block"
          :style="{ color: getGainColor(annualData.stats.same_week_median_pct) }"
        >
          {{ formatPct(annualData.stats.same_week_median_pct) }}
        </text>
      </view>

      <view
        class="rounded-xl border p-2.5"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          同周历史排名
        </text>
        <text class="text-[15px] font-bold font-mono tabular-nums mt-0.5 block" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
          {{ annualData.stats.current_rank ?? '—' }} / {{ annualData.stats.sample_count }}
        </text>
      </view>
    </view>

    <!-- 年份对比图卡片 -->
    <view
      class="rounded-2xl border p-3"
      :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
    >
      <!-- 年份高亮图例 -->
      <view v-if="sortedYears.length" class="flex flex-wrap items-center gap-1.5 pb-2 mb-2 border-b" :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'">
        <view
          v-for="(y, idx) in sortedYears"
          :key="y.year"
          class="flex items-center gap-1 px-2 py-0.5 rounded-md border text-[11px] font-mono transition-all"
          :class="focusedYear === y.year || (focusedYear === null && y.is_current)
            ? (isDark ? 'bg-[#E5B352]/15 border-[#E5B352]/50 text-[#F1EFEA] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F]/50 text-[#22292E] font-bold')
            : (isDark ? 'border-[#29313A] text-[#8D99A4]' : 'border-[#DCDDD8] text-[#515F6A]')"
          @tap="toggleYearFocus(y.year)"
        >
          <view class="w-2 h-2 rounded-full" :style="{ backgroundColor: getYearColor(y.year, y.is_current, idx) }" />
          <text>{{ y.year }}{{ y.is_current ? '(今)' : '' }}</text>
        </view>
      </view>

      <view v-if="loading" class="h-[220px] flex items-center justify-center">
        <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          正在加载年度同期走势...
        </text>
      </view>

      <view v-else-if="errorMsg" class="h-[220px] flex flex-col items-center justify-center gap-2">
        <text class="text-[12px] text-center" :class="isDark ? 'text-[#FF736A]' : 'text-[#B93842]'">
          {{ errorMsg }}
        </text>
        <text
          class="px-3 py-1 rounded-lg text-[11px] font-bold"
          :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-white'"
          @tap="fetchAnnualData"
        >
          重新加载
        </text>
      </view>

      <view v-else-if="!weekLabels.length" class="h-[220px] flex items-center justify-center">
        <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          暂无年度周频对比数据
        </text>
      </view>

      <TrendLineCanvas
        v-else
        canvas-id="annualTrendCanvas"
        :dates="weekLabels"
        :series="chartSeries"
        :height="220"
        value-format="pct"
        :show-zero-line="true"
        :max-x-labels="6"
      />
    </view>

    <!-- 同周 / 全年收益明细表 -->
    <view
      v-if="sortedYears.length"
      class="rounded-2xl border p-3"
      :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
    >
      <view class="flex items-center justify-between pb-2 mb-1 border-b" :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'">
        <text class="text-[13px] font-bold" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
          历年同周与全年收益明细
        </text>
        <text v-if="annualData?.stats.up_ratio_pct !== null && annualData?.stats.up_ratio_pct !== undefined" class="text-[11px] font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          同周正收益占比 {{ annualData.stats.up_ratio_pct.toFixed(0) }}%
        </text>
      </view>

      <view class="flex items-center text-[11px] py-1.5" :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
        <text class="w-20">年份</text>
        <text class="flex-1">状态</text>
        <text class="w-24 text-right">同周收益</text>
        <text class="w-24 text-right">全年收益</text>
      </view>

      <view class="divide-y" :class="isDark ? 'divide-[#29313A]' : 'divide-[#DCDDD8]'">
        <view
          v-for="y in sortedYears"
          :key="y.year"
          class="flex items-center py-2 text-[12px] font-mono tabular-nums"
          :class="y.is_current ? (isDark ? 'bg-[#E5B352]/[0.06]' : 'bg-[#8A641F]/[0.05]') : ''"
          @tap="toggleYearFocus(y.year)"
        >
          <text class="w-20 font-bold" :class="y.is_current ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]') : (isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]')">
            {{ y.year }}
          </text>
          <text class="flex-1 text-[11px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
            {{ y.is_current ? '进行中' : (y.is_complete ? '完整年度' : '部分年度') }}
          </text>
          <text class="w-24 text-right font-bold" :style="{ color: getGainColor(y.same_week_return_pct) }">
            {{ formatPct(y.same_week_return_pct) }}
          </text>
          <text class="w-24 text-right" :style="{ color: getGainColor(y.full_year_return_pct) }">
            {{ y.is_current ? '—' : formatPct(y.full_year_return_pct) }}
          </text>
        </view>
      </view>
    </view>
  </view>
</template>
