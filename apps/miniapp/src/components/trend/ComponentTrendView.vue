<script lang="ts" setup>
import type { ComponentTrendResponse, ComponentTrendSeries } from '@/api/types/trend'
import type { TrendLineSeries } from '@/components/trend/TrendLineCanvas.vue'
import { computed, onUnmounted, ref, watch } from 'vue'
import { getTrendComponents } from '@/api/trend'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import TrendLineCanvas from '@/components/trend/TrendLineCanvas.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, unwrapApiResponse } from '@/utils/asyncRace'
import { deterministicAssetColor } from '@/utils/chartGeometry'
import { toApiCurrency } from '@/utils/financeRoute'

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

const rangeOptions = ['3M', '6M', '1Y', '3Y'] as const
const activeRange = ref<string>(props.range || '1Y')

const loading = ref(false)
const errorMsg = ref('')
const componentData = ref<ComponentTrendResponse | null>(null)

const hiddenIds = ref<string[]>([])
const focusedId = ref<string | null>(null)

type SortKey = 'period_change_pct' | 'latest_value' | 'turnover' | 'total_return_pct'
const sortKey = ref<SortKey>('period_change_pct')
const sortOrder = ref<'asc' | 'desc'>('desc')

function formatPct(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(2)}%`
}

function formatPrice(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  return Math.abs(val) >= 100
? val.toFixed(1)
: val.toFixed(2)
}

function formatTurnover(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  if (val >= 1e8)
    return `${(val / 1e8).toFixed(1)}亿`
  if (val >= 1e4)
    return `${(val / 1e4).toFixed(0)}万`
  return val.toFixed(0)
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

const race = createAsyncRace()

onUnmounted(() => {
  race.invalidate()
})

async function fetchComponentData() {
  if (!props.assetId)
    return
  const token = race.next()
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendComponents(props.assetId, {
      scope_id: props.scopeId || undefined,
      range: activeRange.value,
      currency: props.currency
? toApiCurrency(props.currency)
: undefined,
      snapshot_id: props.snapshotId || undefined
    })
    if (!race.isLatest(token))
      return
    componentData.value = unwrapApiResponse(res, '品种成分趋势加载失败')
    hiddenIds.value = []
    focusedId.value = null
  }
  catch (err: any) {
    if (!race.isLatest(token))
      return
    errorMsg.value = err?.message || '品种成分趋势加载失败'
  }
  finally {
    if (race.isLatest(token))
      loading.value = false
  }
}

watch(
  () => props.range,
  (val) => {
    if (val && val !== activeRange.value)
      activeRange.value = val
  }
)

watch(
  () => [
props.assetId,
props.snapshotId,
props.currency,
props.scopeId,
activeRange.value
],
  () => {
    fetchComponentData()
  },
  { immediate: true }
)

function resolveSeriesColor(item: ComponentTrendSeries, idx: number): string {
  if (item.is_benchmark)
    return isDark.value ? '#E5B352' : '#8A641F'
  if (item.color)
    return item.color
  return deterministicAssetColor(item.asset_id, idx)
}

const orderedRawSeries = computed<ComponentTrendSeries[]>(() => {
  const list = componentData.value?.series || []
  const benchmarks = list.filter(s => s.is_benchmark)
  const others = list.filter(s => !s.is_benchmark)
  return [
...benchmarks,
...others
]
})

const chartSeries = computed<TrendLineSeries[]>(() => {
  return orderedRawSeries.value.map((item, idx) => ({
    id: item.asset_id,
    name: item.name,
    color: resolveSeriesColor(item, idx),
    values: item.points,
    dashed: item.is_benchmark,
    highlighted: item.is_benchmark || focusedId.value === item.asset_id,
    dimmed: focusedId.value !== null && focusedId.value !== item.asset_id && !item.is_benchmark,
    hidden: hiddenIds.value.includes(item.asset_id),
    lineWidth: item.is_benchmark || focusedId.value === item.asset_id
? 2.4
: 1.6
  }))
})

function toggleHidden(assetId: string) {
  const idx = hiddenIds.value.indexOf(assetId)
  if (idx >= 0) {
    hiddenIds.value.splice(idx, 1)
  }
  else {
    hiddenIds.value.push(assetId)
    if (focusedId.value === assetId)
      focusedId.value = null
  }
}

function toggleFocus(assetId: string) {
  if (hiddenIds.value.includes(assetId))
    hiddenIds.value = hiddenIds.value.filter(id => id !== assetId)
  focusedId.value = focusedId.value === assetId
? null
: assetId
}

function handleToggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'desc'
? 'asc'
: 'desc'
  }
  else {
    sortKey.value = key
    sortOrder.value = 'desc'
  }
}

const sortedTableRows = computed<ComponentTrendSeries[]>(() => {
  const list = orderedRawSeries.value
  const benchmarks = list.filter(s => s.is_benchmark)
  const others = list.filter(s => !s.is_benchmark).sort((a, b) => {
    const va = a[sortKey.value]
    const vb = b[sortKey.value]
    const na = va ?? -Infinity
    const nb = vb ?? -Infinity
    return sortOrder.value === 'desc'
? nb - na
: na - nb
  })
  return [
...benchmarks,
...others
]
})
</script>

<template>
  <view class="w-full space-y-3">
    <!-- 标题与周期切换 -->
    <view class="flex items-start justify-between gap-2">
      <view class="min-w-0 flex-1">
        <text class="text-[18px] font-bold block truncate" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
          {{ componentData?.asset_name || assetId }} · 品种成分趋势
        </text>
        <text class="text-[11px] font-mono mt-0.5 block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          {{ componentData?.snapshot?.as_of || '—' }} · 同基期归一化走势（基准置顶）
        </text>
      </view>

      <view
        class="flex items-center p-0.5 rounded-lg border shrink-0"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-[#EDECE6] border-[#DCDDD8]'"
      >
        <text
          v-for="r in rangeOptions"
          :key="r"
          class="px-2 py-1 rounded text-[11px] font-mono transition-colors"
          :class="activeRange === r
            ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] text-white font-bold')
            : (isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]')"
          @tap="activeRange = r"
        >
          {{ r }}
        </text>
      </view>
    </view>

    <!-- 多线分化图卡片 -->
    <view
      class="rounded-2xl border p-3"
      :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
    >
      <!-- 可交互图例（点击隐藏/恢复，长按高亮） -->
      <view
        v-if="orderedRawSeries.length"
        class="flex flex-wrap items-center gap-1.5 pb-2 mb-2 border-b"
        :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'"
      >
        <view
          v-for="(item, idx) in orderedRawSeries"
          :key="item.asset_id"
          class="flex items-center gap-1 px-2 py-0.5 rounded-md border text-[11px] transition-opacity"
          :class="[
            hiddenIds.includes(item.asset_id) ? 'opacity-40' : 'opacity-100',
            focusedId === item.asset_id
              ? (isDark ? 'bg-[#E5B352]/15 border-[#E5B352] text-[#F1EFEA] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#22292E] font-bold')
              : (isDark ? 'border-[#29313A] text-[#F1EFEA]' : 'border-[#DCDDD8] text-[#22292E]')
          ]"
          @tap="toggleHidden(item.asset_id)"
          @longpress="toggleFocus(item.asset_id)"
        >
          <view class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: resolveSeriesColor(item, idx) }" />
          <text class="truncate max-w-[76px]">{{ item.name }}</text>
          <text v-if="item.is_benchmark" class="text-[9px]" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
            基准
          </text>
        </view>
      </view>

      <view v-if="loading" class="h-[230px] flex items-center justify-center">
        <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          正在加载成分趋势...
        </text>
      </view>

      <view v-else-if="errorMsg" class="h-[230px] flex flex-col items-center justify-center gap-2">
        <text class="text-[12px] text-center" :class="isDark ? 'text-[#FF736A]' : 'text-[#B93842]'">
          {{ errorMsg }}
        </text>
        <text
          class="px-3 py-1 rounded-lg text-[11px] font-bold"
          :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-white'"
          @tap="fetchComponentData"
        >
          重新加载
        </text>
      </view>

      <view v-else-if="!componentData?.timeline?.length || !orderedRawSeries.length" class="h-[230px] flex items-center justify-center">
        <text class="text-[12px] text-center px-4" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          当前标的暂无可对比的成分成员序列
        </text>
      </view>

      <TrendLineCanvas
        v-else
        canvas-id="componentTrendCanvas"
        :dates="componentData.timeline"
        :series="chartSeries"
        :height="230"
        value-format="pct"
        :show-zero-line="true"
      />

      <text class="text-[10px] mt-1.5 block" :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
        提示：点击图例隐藏/显示，长按图例高亮单条曲线，点击下方表格行也可高亮对应曲线。
      </text>
    </view>

    <!-- 截面排序表 -->
    <view
      v-if="sortedTableRows.length"
      class="rounded-2xl border p-3"
      :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
    >
      <view class="flex items-center justify-between pb-2 mb-1 border-b" :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'">
        <text class="text-[13px] font-bold" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
          成分截面排行
        </text>
        <text class="text-[11px] font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          共 {{ sortedTableRows.length }} 项
        </text>
      </view>

      <!-- 表头 -->
      <view class="flex items-center text-[11px] py-1.5 gap-1.5" :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
        <text class="w-8 text-center shrink-0">温度</text>
        <text class="flex-1 min-w-0">品种</text>
        <text class="w-14 text-right shrink-0" @tap="handleToggleSort('latest_value')">
          最新{{ sortKey === 'latest_value' ? (sortOrder === 'desc' ? '↓' : '↑') : '' }}
        </text>
        <text class="w-14 text-right shrink-0" @tap="handleToggleSort('turnover')">
          成交{{ sortKey === 'turnover' ? (sortOrder === 'desc' ? '↓' : '↑') : '' }}
        </text>
        <text class="w-16 text-right shrink-0" @tap="handleToggleSort('period_change_pct')">
          区间{{ sortKey === 'period_change_pct' ? (sortOrder === 'desc' ? '↓' : '↑') : '' }}
        </text>
      </view>

      <view class="divide-y" :class="isDark ? 'divide-[#29313A]' : 'divide-[#DCDDD8]'">
        <view
          v-for="(item, idx) in sortedTableRows"
          :key="item.asset_id"
          class="flex items-center py-2 gap-1.5 text-[12px]"
          :class="item.is_benchmark
            ? (isDark ? 'bg-[#E5B352]/[0.06]' : 'bg-[#8A641F]/[0.05]')
            : (focusedId === item.asset_id ? (isDark ? 'bg-white/[0.04]' : 'bg-[#EDECE6]') : '')"
          @tap="toggleFocus(item.asset_id)"
        >
          <view class="w-8 flex justify-center shrink-0">
            <TrendTempBadge :temperature="item.temperature" size="sm" />
          </view>

          <view class="flex-1 min-w-0 flex items-center gap-1.5">
            <view class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: resolveSeriesColor(item, idx) }" />
            <text
              class="truncate"
              :class="item.is_benchmark ? (isDark ? 'text-[#E5B352] font-bold' : 'text-[#8A641F] font-bold') : (isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]')"
            >
              {{ item.name }}
            </text>
            <text
              v-if="item.is_benchmark"
              class="text-[9px] px-1 rounded shrink-0"
              :class="isDark ? 'bg-[#E5B352]/20 text-[#E5B352]' : 'bg-[#8A641F]/15 text-[#8A641F]'"
            >
              基准
            </text>
          </view>

          <text class="w-14 text-right font-mono tabular-nums shrink-0" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
            {{ formatPrice(item.latest_value) }}
          </text>

          <text class="w-14 text-right font-mono tabular-nums text-[11px] shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
            {{ formatTurnover(item.turnover) }}
          </text>

          <text
            class="w-16 text-right font-mono tabular-nums font-bold shrink-0"
            :style="{ color: getGainColor(item.period_change_pct) }"
          >
            {{ formatPct(item.period_change_pct) }}
          </text>
        </view>
      </view>
    </view>
  </view>
</template>
