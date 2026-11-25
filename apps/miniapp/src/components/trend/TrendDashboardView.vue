<script lang="ts" setup>
import type { DashboardRow, TrendDashboardResponse } from '@/api/types/trend'
import { computed, onMounted, ref, watch } from 'vue'
import { getTrendDashboard } from '@/api/trend'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import { useFinanceStore } from '@/stores/finance'
import { formatDurationBadge } from '@/utils/trendDict'

interface Props {
  assetId: string
  snapshotId?: string
  range?: string
  currency?: string
  scopeId?: string
}

const props = withDefaults(defineProps<Props>(), {
  snapshotId: undefined,
  range: undefined,
  currency: undefined,
  scopeId: undefined
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

type PresetKey = 'daily' | 'weekly' | 'monthly' | 'custom'
type SortKey = 'strength' | 'change' | 'turnover' | 'duration'

const REPORT_PRESETS: Array<{ key: Exclude<PresetKey, 'custom'>, label: string, changePeriod: string, sparkRange: string }> = [
  {
key: 'daily',
label: '每日报告',
changePeriod: '1D',
sparkRange: '1M'
},
  {
key: 'weekly',
label: '每周报告',
changePeriod: '1W',
sparkRange: '3M'
},
  {
key: 'monthly',
label: '每月报告',
changePeriod: '1M',
sparkRange: '6M'
}
]

const CHANGE_PERIOD_OPTIONS = [
  { key: '1D', label: '1日' },
  { key: '1W', label: '1周' },
  { key: '1M', label: '1月' },
  { key: '3M', label: '3月' },
  { key: '1Y', label: '1年' }
]

const SPARK_RANGE_OPTIONS = [
  { key: '1M', label: '1M走势' },
  { key: '3M', label: '3M走势' },
  { key: '6M', label: '6M走势' },
  { key: '1Y', label: '1Y走势' }
]

const SORT_OPTIONS: Array<{ key: SortKey, label: string }> = [
  { key: 'strength', label: '相对强度' },
  { key: 'change', label: '涨跌幅' },
  { key: 'turnover', label: '成交额' },
  { key: 'duration', label: '持续天数' }
]

const activePreset = ref<PresetKey>('daily')
const changePeriod = ref<string>('1D')
const sparkRange = ref<string>('1M')
const sortBy = ref<SortKey>('strength')
const sortOrder = ref<'desc' | 'asc'>('desc')

const loading = ref(false)
const errorMsg = ref('')
const data = ref<TrendDashboardResponse | null>(null)
const collapsedGroups = ref<Set<string>>(new Set())

async function fetchData () {
  if (!props.assetId)
    return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendDashboard(props.assetId, {
      change_period: changePeriod.value,
      spark_range: sparkRange.value,
      sort: sortBy.value,
      order: sortOrder.value,
      scope_id: props.scopeId,
      currency: props.currency,
      snapshot_id: props.snapshotId
    })
    data.value = res.data ?? null
  }
  catch (err: any) {
    errorMsg.value = err?.message || '趋势仪表盘加载失败'
    data.value = null
  }
  finally {
    loading.value = false
  }
}

function applyPreset (presetKey: Exclude<PresetKey, 'custom'>) {
  const found = REPORT_PRESETS.find(p => p.key === presetKey)
  if (!found)
    return
  activePreset.value = presetKey
  changePeriod.value = found.changePeriod
  sparkRange.value = found.sparkRange
}

function selectChangePeriod (period: string) {
  activePreset.value = 'custom'
  changePeriod.value = period
}

function selectSparkRange (sr: string) {
  activePreset.value = 'custom'
  sparkRange.value = sr
}

function toggleSort (key: SortKey) {
  if (sortBy.value === key) {
    sortOrder.value = sortOrder.value === 'desc'
? 'asc'
: 'desc'
  }
  else {
    sortBy.value = key
    sortOrder.value = 'desc'
  }
}

function toggleGroup (groupName: string) {
  const next = new Set(collapsedGroups.value)
  if (next.has(groupName))
    next.delete(groupName)
  else
    next.add(groupName)
  collapsedGroups.value = next
}

function isGroupExpanded (groupName: string): boolean {
  return !collapsedGroups.value.has(groupName)
}

function getRowSortValue (row: DashboardRow, key: SortKey): number | null {
  if (key === 'strength')
    return row.asset.metrics.strength?.score ?? null
  if (key === 'change')
    return row.change_pct
  if (key === 'turnover')
    return row.asset.turnover
  return row.asset.metrics.duration_trading_days
}

const sortedGroups = computed(() => {
  if (!data.value?.groups)
    return []
  return data.value.groups.map((g) => {
    const rows = [...g.rows].sort((a, b) => {
      const va = getRowSortValue(a, sortBy.value)
      const vb = getRowSortValue(b, sortBy.value)
      if (va === null && vb === null)
        return 0
      if (va === null)
        return 1
      if (vb === null)
        return -1
      return sortOrder.value === 'desc'
? vb - va
: va - vb
    })
    return {
      group_name: g.group_name,
      rows
    }
  })
})

function formatChangePct (val: number | null): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(2)}%`
}

function getChangeColorClass (val: number | null): string {
  if (val === null || val === undefined || !Number.isFinite(val) || val === 0)
    return isDark.value ? 'text-[#8D99A4]' : 'text-[#64748B]'
  const isUp = val > 0
  const cnRedUp = financeStore.colorScheme === 'cn_red_up'
  if (isUp)
    return cnRedUp ? 'text-[#FF5A5F]' : 'text-[#2EC4B6]'
  return cnRedUp ? 'text-[#2EC4B6]' : 'text-[#FF5A5F]'
}

interface SparkBar {
  heightPct: number
  isNull: boolean
  isPositive: boolean
}

function buildSparkBars (row: DashboardRow): SparkBar[] {
  const pts = row.sparkline || []
  if (pts.length === 0)
    return []
  const valid = pts.filter((v): v is number => v !== null && Number.isFinite(v))
  if (valid.length === 0)
    return pts.map(() => ({ heightPct: 0, isNull: true, isPositive: false }))

  const min = row.spark_min !== null && Number.isFinite(row.spark_min)
? row.spark_min
: Math.min(...valid)
  const max = row.spark_max !== null && Number.isFinite(row.spark_max)
? row.spark_max
: Math.max(...valid)
  const span = max - min
  const firstValid = valid[0] ?? 0
  const lastValid = valid[valid.length - 1] ?? 0
  const isPositive = lastValid >= firstValid

  // 采样至最多 16 根对齐柱，保持宽度一致
  const maxBars = 16
  const step = pts.length > maxBars
? pts.length / maxBars
: 1
  const bars: SparkBar[] = []
  for (let i = 0; i < (pts.length > maxBars
? maxBars
: pts.length); i++) {
    const rawIdx = Math.min(pts.length - 1, Math.floor(i * step))
    const val = pts[rawIdx]
    if (val === null || val === undefined || !Number.isFinite(val)) {
      bars.push({ heightPct: 0, isNull: true, isPositive })
    }
    else {
      const ratio = span > 0
? (val - min) / span
: 0.5
      bars.push({
        heightPct: Math.max(12, Math.round(ratio * 100)),
        isNull: false,
        isPositive
      })
    }
  }
  return bars
}

function getSparkBarColor (isPositive: boolean): string {
  const cnRedUp = financeStore.colorScheme === 'cn_red_up'
  if (isPositive)
    return cnRedUp ? '#FF736A' : '#2EC4B6'
  return cnRedUp ? '#2EC4B6' : '#FF736A'
}

function getStrengthWidth (score: number | null | undefined): string {
  if (score === null || score === undefined || !Number.isFinite(score))
    return '0%'
  return `${Math.max(0, Math.min(100, score))}%`
}

function navigateToAsset (assetId: string) {
  uni.navigateTo({
    url: `/pages-sub/finance/asset?id=${encodeURIComponent(assetId)}`
  })
}

watch(
  () => [
props.assetId,
props.snapshotId,
props.currency,
props.scopeId
],
  () => {
    fetchData()
  }
)

watch([changePeriod, sparkRange], () => {
  fetchData()
})

onMounted(() => {
  fetchData()
})
</script>

<template>
  <view class="w-full">
    <!-- 头部标题与快照元信息 -->
    <view class="mb-3">
      <view class="flex items-center justify-between gap-2">
        <text
          class="text-lg font-black tracking-tight truncate"
          :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'"
        >
          {{ data?.asset_name ? `${data.asset_name} · 趋势仪表盘` : '趋势仪表盘' }}
        </text>
        <text
          v-if="data?.snapshot"
          class="text-[11px] font-mono shrink-0"
          :class="isDark ? 'text-[#E5B352]' : 'text-[#9A6B16]'"
        >
          {{ data.snapshot.as_of }}
        </text>
      </view>
      <text
        class="text-[11px] font-mono mt-0.5 block"
        :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'"
      >
        多周期变动 · 对齐迷你走势 · 相对强度排名
      </text>
    </view>

    <!-- 报告预设条 + 双周期独立切换 -->
    <view
      class="rounded-2xl border p-3 mb-3 space-y-2.5"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <!-- 报告预设 -->
      <view class="flex items-center justify-between gap-2">
        <text class="text-[11px] font-medium shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          报告预设
        </text>
        <view class="flex items-center gap-1.5 flex-wrap">
          <view
            v-for="preset in REPORT_PRESETS"
            :key="preset.key"
            class="px-2.5 py-1 rounded-lg text-[11px] font-medium border cursor-pointer transition-all"
            :class="activePreset === preset.key
              ? (isDark ? 'bg-[#E5B352]/20 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#FEF3C7] border-[#D97706] text-[#9A6B16] font-bold')
              : (isDark ? 'bg-white/[0.03] border-white/[0.08] text-[#8D99A4]' : 'bg-[#F8FAFC] border-[#E2E8F0] text-[#475569]')"
            @tap="applyPreset(preset.key)"
          >
            {{ preset.label }}
          </view>
        </view>
      </view>

      <!-- 统计周期 change_period 与迷你图区间 spark_range 独立切换 -->
      <view class="flex items-center justify-between gap-2 pt-2 border-t" :class="isDark ? 'border-white/[0.06]' : 'border-[#F1F5F9]'">
        <text class="text-[11px] shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          涨跌周期
        </text>
        <view class="flex items-center gap-1 flex-wrap justify-end">
          <text
            v-for="opt in CHANGE_PERIOD_OPTIONS"
            :key="opt.key"
            class="px-2 py-0.5 rounded text-[11px] font-mono cursor-pointer"
            :class="changePeriod === opt.key
              ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#0F172A] text-white font-bold')
              : (isDark ? 'bg-white/[0.04] text-[#8D99A4]' : 'bg-[#F1F5F9] text-[#475569]')"
            @tap="selectChangePeriod(opt.key)"
          >
            {{ opt.label }}
          </text>
        </view>
      </view>

      <view class="flex items-center justify-between gap-2">
        <text class="text-[11px] shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          迷你走势
        </text>
        <view class="flex items-center gap-1 flex-wrap justify-end">
          <text
            v-for="opt in SPARK_RANGE_OPTIONS"
            :key="opt.key"
            class="px-2 py-0.5 rounded text-[11px] font-mono cursor-pointer"
            :class="sparkRange === opt.key
              ? (isDark ? 'bg-[#38BDF8]/25 text-[#38BDF8] border border-[#38BDF8]/50 font-bold' : 'bg-[#E0F2FE] text-[#0369A1] border border-[#7DD3FC] font-bold')
              : (isDark ? 'bg-white/[0.04] text-[#8D99A4]' : 'bg-[#F1F5F9] text-[#475569]')"
            @tap="selectSparkRange(opt.key)"
          >
            {{ opt.label }}
          </text>
        </view>
      </view>

      <!-- 排序工具条 -->
      <view class="flex items-center justify-between gap-2 pt-2 border-t" :class="isDark ? 'border-white/[0.06]' : 'border-[#F1F5F9]'">
        <text class="text-[11px] shrink-0" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
          组内排序
        </text>
        <view class="flex items-center gap-1 flex-wrap justify-end">
          <view
            v-for="opt in SORT_OPTIONS"
            :key="opt.key"
            class="px-2 py-0.5 rounded text-[11px] cursor-pointer flex items-center gap-0.5"
            :class="sortBy === opt.key
              ? (isDark ? 'bg-white/[0.12] text-[#F1EFEA] font-bold' : 'bg-[#E2E8F0] text-[#0F172A] font-bold')
              : (isDark ? 'text-[#8D99A4]' : 'text-[#64748B]')"
            @tap="toggleSort(opt.key)"
          >
            <text>{{ opt.label }}</text>
            <text v-if="sortBy === opt.key" class="font-mono text-[10px]">
              {{ sortOrder === 'desc' ? '↓' : '↑' }}
            </text>
          </view>
        </view>
      </view>
    </view>

    <!-- 加载态 -->
    <view
      v-if="loading"
      class="rounded-2xl border p-8 flex flex-col items-center justify-center"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        正在加载趋势仪表盘...
      </text>
    </view>

    <!-- 失败态 -->
    <view
      v-else-if="errorMsg"
      class="rounded-2xl border p-6 flex flex-col items-center justify-center gap-3"
      :class="isDark ? 'bg-[#12181F] border-[#FF5A5F]/30' : 'bg-white border-[#FECACA]'"
    >
      <text class="text-xs text-[#FF5A5F] text-center">{{ errorMsg }}</text>
      <view
        class="px-3 py-1.5 rounded-lg text-xs font-bold cursor-pointer"
        :class="isDark ? 'bg-white/[0.08] text-[#F1EFEA]' : 'bg-[#F1F5F9] text-[#0F172A]'"
        @tap="fetchData"
      >
        重新加载
      </view>
    </view>

    <!-- 空态 -->
    <view
      v-else-if="!sortedGroups.length"
      class="rounded-2xl border p-8 flex flex-col items-center justify-center"
      :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
    >
      <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
        暂无可展示的仪表盘分组数据
      </text>
    </view>

    <!-- 分组列表 -->
    <view v-else class="space-y-2.5">
      <view
        v-for="group in sortedGroups"
        :key="group.group_name"
        class="rounded-2xl border overflow-hidden"
        :class="isDark ? 'bg-[#12181F] border-white/[0.08]' : 'bg-white border-[#E2E8F0]'"
      >
        <!-- 分组头 -->
        <view
          class="px-3 py-2.5 flex items-center justify-between cursor-pointer"
          :class="isDark ? 'bg-white/[0.03]' : 'bg-[#F8FAFC]'"
          @tap="toggleGroup(group.group_name)"
        >
          <view class="flex items-center gap-2">
            <text class="text-xs font-bold" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'">
              {{ group.group_name }}
            </text>
            <text
              class="text-[11px] font-mono px-1.5 py-0.5 rounded"
              :class="isDark ? 'bg-white/[0.06] text-[#8D99A4]' : 'bg-[#E2E8F0] text-[#475569]'"
            >
              {{ group.rows.length }}
            </text>
          </view>
          <text class="text-xs font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#64748B]'">
            {{ isGroupExpanded(group.group_name) ? '收起 ▲' : '展开 ▼' }}
          </text>
        </view>

        <!-- 列头 -->
        <view
          v-if="isGroupExpanded(group.group_name)"
          class="px-3 py-1.5 grid grid-cols-12 gap-1.5 text-[10px] font-mono border-t border-b"
          :class="isDark ? 'border-white/[0.06] text-[#64707C]' : 'border-[#F1F5F9] text-[#94A3B8]'"
        >
          <text class="col-span-4">品种 / 状态</text>
          <text class="col-span-2 text-right">{{ changePeriod }}涨跌</text>
          <text class="col-span-3 text-center">{{ sparkRange }}走势</text>
          <text class="col-span-3 text-right">相对强度</text>
        </view>

        <!-- 数据行 -->
        <view
          v-if="isGroupExpanded(group.group_name)"
          class="divide-y"
          :class="isDark ? 'divide-white/[0.05]' : 'divide-[#F1F5F9]'"
        >
          <view
            v-for="row in group.rows"
            :key="row.asset.asset_id"
            class="px-3 py-2.5 grid grid-cols-12 gap-1.5 items-center cursor-pointer active:opacity-80"
            @tap="navigateToAsset(row.asset.asset_id)"
          >
            <!-- 品种名称与温度/右侧标记 -->
            <view class="col-span-4 min-w-0">
              <view class="flex items-center gap-1">
                <TrendTempBadge :temperature="row.asset.metrics.temperature" size="sm" />
                <text
                  class="text-xs font-bold truncate"
                  :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'"
                >
                  {{ row.asset.short_name || row.asset.name }}
                </text>
              </view>
              <view class="flex items-center gap-1 mt-1">
                <text
                  class="text-[10px] font-mono px-1 rounded"
                  :class="formatDurationBadge(row.asset.metrics.side, row.asset.metrics.confirmation, row.asset.metrics.duration_trading_days).isRight
                    ? (isDark ? 'bg-[#E5B352]/20 text-[#E5B352]' : 'bg-[#FEF3C7] text-[#9A6B16]')
                    : (isDark ? 'bg-white/[0.05] text-[#8D99A4]' : 'bg-[#F1F5F9] text-[#64748B]')"
                >
                  {{ formatDurationBadge(row.asset.metrics.side, row.asset.metrics.confirmation, row.asset.metrics.duration_trading_days).text }}
                </text>
              </view>
            </view>

            <!-- 周期涨跌 -->
            <view class="col-span-2 text-right">
              <text
                class="text-xs font-mono font-bold tabular-nums"
                :class="getChangeColorClass(row.change_pct)"
              >
                {{ formatChangePct(row.change_pct) }}
              </text>
            </view>

            <!-- 对齐迷你走势微图 (纯 CSS 柱状微图，无 SVG) -->
            <view class="col-span-3 flex items-end justify-center gap-[2px] h-7 px-1">
              <template v-if="buildSparkBars(row).length > 0">
                <view
                  v-for="(bar, bIdx) in buildSparkBars(row)"
                  :key="bIdx"
                  class="flex-1 rounded-t-[1px]"
                  :style="{
                    height: bar.isNull ? '0%' : `${bar.heightPct}%`,
                    backgroundColor: bar.isNull ? 'transparent' : getSparkBarColor(bar.isPositive),
                    opacity: bar.isNull ? 0 : 0.85
                  }"
                />
              </template>
              <text v-else class="text-[10px] font-mono" :class="isDark ? 'text-[#64707C]' : 'text-[#94A3B8]'">
                —
              </text>
            </view>

            <!-- 相对强度进度条 -->
            <view class="col-span-3 flex flex-col items-end gap-1">
              <text
                class="text-xs font-mono font-bold tabular-nums"
                :class="isDark ? 'text-[#F1EFEA]' : 'text-[#0F172A]'"
              >
                {{ row.asset.metrics.strength?.score !== null && row.asset.metrics.strength?.score !== undefined
                  ? row.asset.metrics.strength.score.toFixed(0)
                  : '—' }}
              </text>
              <view
                class="w-full h-1.5 rounded-full overflow-hidden"
                :class="isDark ? 'bg-white/[0.08]' : 'bg-[#E2E8F0]'"
              >
                <view
                  class="h-full rounded-full transition-all"
                  :style="{
                    width: getStrengthWidth(row.asset.metrics.strength?.score),
                    backgroundColor: (row.asset.metrics.strength?.score ?? 0) >= 70
                      ? '#E5B352'
                      : (row.asset.metrics.strength?.score ?? 0) >= 30
                        ? '#38BDF8'
                        : '#818CF8'
                  }"
                />
              </view>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
