<script lang="ts" setup>
import type { TrendCurveResponse } from '@/api/types/trend'
import type { TrendLineSeries, TrendStateInterval } from '@/components/trend/TrendLineCanvas.vue'
import { computed, ref, watch } from 'vue'
import { getTrendCurve } from '@/api/trend'
import TrendLineCanvas from '@/components/trend/TrendLineCanvas.vue'
import { useFinanceStore } from '@/stores/finance'

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

const rangeOptions = ['3M', '6M', '1Y', '3Y', 'ALL'] as const
const freqOptions = [
  { key: 'day', label: '日线' },
  { key: 'week', label: '周线' },
  { key: 'month', label: '月线' }
] as const

const activeRange = ref<string>(props.range || '1Y')
const activeFreq = ref<'day' | 'week' | 'month'>('day')
const scaleMode = ref<'linear' | 'log'>('linear')

const loading = ref(false)
const errorMsg = ref('')
const curveData = ref<TrendCurveResponse | null>(null)

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
  const abs = Math.abs(val)
  if (abs >= 1000)
    return val.toFixed(1)
  if (abs >= 1)
    return val.toFixed(2)
  return val.toFixed(4)
}

function formatTurnover(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  if (val >= 1e8)
    return `${(val / 1e8).toFixed(2)}亿`
  if (val >= 1e4)
    return `${(val / 1e4).toFixed(1)}万`
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

async function fetchCurveData() {
  if (!props.assetId)
    return
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await getTrendCurve(props.assetId, {
      range: activeRange.value,
      freq: activeFreq.value,
      currency: props.currency || undefined,
      benchmark_id: props.benchmarkId || undefined,
      snapshot_id: props.snapshotId || undefined
    })
    curveData.value = res.data ?? null
  }
  catch (err: any) {
    errorMsg.value = err?.message || '趋势曲线加载失败，请稍后重试'
  }
  finally {
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
props.benchmarkId,
activeRange.value,
activeFreq.value
],
  () => {
    fetchCurveData()
  },
  { immediate: true }
)

const dates = computed(() => (curveData.value?.points || []).map(p => p.trading_date))

const chartSeries = computed<TrendLineSeries[]>(() => {
  const data = curveData.value
  if (!data || !data.points.length)
    return []

  const isLog = scaleMode.value === 'log'
  const assetColor = isDark.value ? '#E5B352' : '#8A641F'
  const benchColor = isDark.value ? '#72A7D9' : '#326C9D'

  const mainValues = data.points.map(p => (isLog ? p.value : p.return_pct))
  const list: TrendLineSeries[] = [
    {
      id: data.asset_id,
      name: data.asset_name || '标的',
      color: assetColor,
      values: mainValues,
      highlighted: true,
      lineWidth: 2.4
    }
  ]

  if (!isLog && data.benchmark_points && data.benchmark_points.length > 0) {
    const benchByDate = new Map<string, number | null>()
    for (const bp of data.benchmark_points)
      benchByDate.set(bp.trading_date, bp.return_pct)
    list.push({
      id: data.benchmark_id || 'benchmark',
      name: data.benchmark_name || '基准',
      color: benchColor,
      values: dates.value.map(d => benchByDate.get(d) ?? null),
      dashed: true,
      lineWidth: 1.6
    })
  }

  return list
})

const stateIntervals = computed<TrendStateInterval[]>(() => {
  return (curveData.value?.state_intervals || []).map(s => ({
    startDate: s.start_date,
    endDate: s.end_date,
    side: s.side
  }))
})

const latestPoint = computed(() => {
  const pts = curveData.value?.points || []
  for (let i = pts.length - 1; i >= 0; i--) {
    if (pts[i]?.value !== null && pts[i]?.value !== undefined)
      return pts[i]
  }
  return null
})

const avgTurnover = computed(() => {
  const pts = (curveData.value?.points || []).filter(p => p.turnover !== null && Number.isFinite(p.turnover))
  if (!pts.length)
    return null
  const sum = pts.reduce((acc, p) => acc + (p.turnover || 0), 0)
  return sum / pts.length
})

const statusLabel = computed(() => {
  const status = curveData.value?.snapshot?.status
  if (!status || status === 'ok')
    return ''
  const map: Record<string, string> = {
    degraded: '部分降级',
    stale: '非最新快照',
    offline_seed: '离线样例',
    insufficient: '样本不足',
    not_covered: '暂无权限',
    error: '数据异常'
  }
  return map[status] || status
})
</script>

<template>
  <view class="w-full space-y-3">
    <!-- 顶部标的与日期信息 -->
    <view class="flex items-start justify-between gap-2">
      <view class="min-w-0 flex-1">
        <view class="flex items-center gap-2">
          <text class="text-[18px] font-bold truncate" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
            {{ curveData?.asset_name || assetId }} · 趋势曲线
          </text>
          <text
            v-if="statusLabel"
            class="px-1.5 py-0.5 rounded text-[10px] font-medium shrink-0"
            :class="isDark ? 'bg-[#E5B352]/15 text-[#E5B352]' : 'bg-[#8A641F]/12 text-[#8A641F]'"
          >
            {{ statusLabel }}
          </text>
        </view>
        <view class="flex items-center gap-2 mt-1 text-[11px] font-mono tabular-nums" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          <text :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
            {{ curveData?.snapshot?.as_of || '—' }}
          </text>
          <text>·</text>
          <text>最新 {{ formatPrice(latestPoint?.value) }}</text>
          <text v-if="curveData?.benchmark_name">· 基准 {{ curveData.benchmark_name }}</text>
        </view>
      </view>

      <!-- 线性/对数坐标切换 -->
      <view
        class="flex items-center p-0.5 rounded-lg border shrink-0"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-[#EDECE6] border-[#DCDDD8]'"
      >
        <text
          class="px-2 py-1 rounded text-[11px] font-mono transition-colors"
          :class="scaleMode === 'linear'
            ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] text-white font-bold')
            : (isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]')"
          @tap="scaleMode = 'linear'"
        >
          线性
        </text>
        <text
          class="px-2 py-1 rounded text-[11px] font-mono transition-colors"
          :class="scaleMode === 'log'
            ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] text-white font-bold')
            : (isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]')"
          @tap="scaleMode = 'log'"
        >
          对数
        </text>
      </view>
    </view>

    <!-- 区间与频度筛选栏 -->
    <view class="flex items-center justify-between gap-2">
      <view
        class="flex items-center p-0.5 rounded-lg border"
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

      <view
        class="flex items-center p-0.5 rounded-lg border"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-[#EDECE6] border-[#DCDDD8]'"
      >
        <text
          v-for="f in freqOptions"
          :key="f.key"
          class="px-2 py-1 rounded text-[11px] font-mono transition-colors"
          :class="activeFreq === f.key
            ? (isDark ? 'bg-white/15 text-[#F1EFEA] font-bold' : 'bg-white text-[#22292E] font-bold shadow-xs')
            : (isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]')"
          @tap="activeFreq = f.key"
        >
          {{ f.label }}
        </text>
      </view>
    </view>

    <!-- 主图卡片 -->
    <view
      class="rounded-2xl border p-3"
      :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
    >
      <!-- 图例与坐标说明 -->
      <view
        class="flex items-center justify-between pb-2 mb-2 border-b text-[11px] font-mono"
        :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'"
      >
        <view class="flex items-center gap-3">
          <view class="flex items-center gap-1.5">
            <view class="w-3 h-0.5" :style="{ backgroundColor: isDark ? '#E5B352' : '#8A641F' }" />
            <text :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
              {{ scaleMode === 'log' ? '净值/价格(对数)' : '区间累计收益' }}
            </text>
          </view>
          <view v-if="scaleMode === 'linear' && curveData?.benchmark_name" class="flex items-center gap-1.5">
            <view class="w-3 h-0.5 border-t border-dashed" :style="{ borderColor: isDark ? '#72A7D9' : '#326C9D' }" />
            <text :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
              {{ curveData.benchmark_name }}
            </text>
          </view>
        </view>
        <text :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
          长按滑动看读数
        </text>
      </view>

      <view v-if="loading" class="h-[220px] flex items-center justify-center">
        <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          正在加载趋势曲线...
        </text>
      </view>

      <view v-else-if="errorMsg" class="h-[220px] flex flex-col items-center justify-center gap-2">
        <text class="text-[12px] text-center" :class="isDark ? 'text-[#FF736A]' : 'text-[#B93842]'">
          {{ errorMsg }}
        </text>
        <text
          class="px-3 py-1 rounded-lg text-[11px] font-bold"
          :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-white'"
          @tap="fetchCurveData"
        >
          重新加载
        </text>
      </view>

      <view v-else-if="!dates.length" class="h-[220px] flex items-center justify-center">
        <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          当前区间暂无足够历史序列数据
        </text>
      </view>

      <TrendLineCanvas
        v-else
        canvas-id="trendCurveCanvas"
        :dates="dates"
        :series="chartSeries"
        :height="220"
        :value-format="scaleMode === 'log' ? 'price' : 'pct'"
        :log-scale="scaleMode === 'log'"
        :show-zero-line="scaleMode === 'linear'"
        :state-intervals="stateIntervals"
      />
    </view>

    <!-- 区间极值与成交额概况 -->
    <view v-if="curveData" class="grid grid-cols-2 gap-2">
      <view
        class="rounded-xl border p-2.5"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          区间累计收益
        </text>
        <text
          class="text-[16px] font-bold font-mono tabular-nums mt-0.5 block"
          :style="{ color: getGainColor(latestPoint?.return_pct) }"
        >
          {{ formatPct(latestPoint?.return_pct) }}
        </text>
        <text class="text-[10px] font-mono mt-0.5 block" :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
          截至 {{ latestPoint?.trading_date || curveData.snapshot.as_of }}
        </text>
      </view>

      <view
        class="rounded-xl border p-2.5"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          成交额概况
        </text>
        <text class="text-[16px] font-bold font-mono tabular-nums mt-0.5 block" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
          {{ formatTurnover(latestPoint?.turnover) }}
        </text>
        <text class="text-[10px] font-mono mt-0.5 block" :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
          区间日均 {{ formatTurnover(avgTurnover) }}
        </text>
      </view>

      <view
        class="rounded-xl border p-2.5"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          区间最高 / 最低
        </text>
        <text class="text-[13px] font-bold font-mono tabular-nums mt-0.5 block" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
          {{ formatPrice(curveData.extremum.high_value) }} / {{ formatPrice(curveData.extremum.low_value) }}
        </text>
        <text class="text-[10px] font-mono mt-0.5 block truncate" :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
          高 {{ curveData.extremum.high_date || '—' }} · 低 {{ curveData.extremum.low_date || '—' }}
        </text>
      </view>

      <view
        class="rounded-xl border p-2.5"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          最大上涨 / 最大回撤
        </text>
        <view class="flex items-center gap-1.5 mt-0.5 font-mono tabular-nums text-[13px] font-bold">
          <text :style="{ color: getGainColor(curveData.extremum.max_gain_pct) }">
            {{ formatPct(curveData.extremum.max_gain_pct) }}
          </text>
          <text :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">/</text>
          <text :style="{ color: getGainColor(curveData.extremum.max_drawdown_pct !== null ? -Math.abs(curveData.extremum.max_drawdown_pct) : null) }">
            {{ curveData.extremum.max_drawdown_pct !== null ? `-${Math.abs(curveData.extremum.max_drawdown_pct).toFixed(2)}%` : '—' }}
          </text>
        </view>
        <text class="text-[10px] font-mono mt-0.5 block truncate" :class="isDark ? 'text-[#64707C]' : 'text-[#626F77]'">
          回撤 {{ curveData.extremum.drawdown_peak_date || '—' }} → {{ curveData.extremum.drawdown_trough_date || '—' }}
        </text>
      </view>
    </view>

    <!-- 近期状态切换事件 -->
    <view
      v-if="curveData?.events?.length"
      class="rounded-2xl border p-3"
      :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
    >
      <text class="text-[13px] font-bold block mb-2" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
        区间趋势事件记录
      </text>
      <view class="space-y-1.5">
        <view
          v-for="ev in curveData.events.slice(0, 5)"
          :key="ev.event_id"
          class="flex items-center justify-between text-[11px] py-1 border-b last:border-b-0"
          :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'"
        >
          <text class="font-mono" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
            {{ ev.as_of }}
          </text>
          <text :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
            {{ ev.event_type }}：{{ ev.old_value }} → {{ ev.new_value }}
          </text>
        </view>
      </view>
    </view>
  </view>
</template>
