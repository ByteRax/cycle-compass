<script lang="ts" setup>
import type {
  AssetDetailResponse,
  AssetSummaryItem,
  TrendCurveResponse,
  TrendEventItem
} from '@/api/types/trend'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { computed, onUnmounted, ref } from 'vue'
import { getAssetDetail, getAssetMembers, getTrendCurve } from '@/api/trend'
import AssetRow from '@/components/finance/AssetRow.vue'
import FinanceShell from '@/components/finance/FinanceShell.vue'
import TrendTempBadge from '@/components/finance/TrendTempBadge.vue'
import TrendEventsTimeline from '@/components/trend/TrendEventsTimeline.vue'
import TrendLineCanvas from '@/components/trend/TrendLineCanvas.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, isExpiredSnapshotError, isNotFoundError, isStaleCursorError, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceUrl, decodeFinanceQuery, toApiCurrency, toDisplayCurrency } from '@/utils/financeRoute'
import { formatDurationBadge, getTrendTempMeta } from '@/utils/trendDict'

definePage({
  style: {
    navigationBarTitleText: '资产详情',
    navigationStyle: 'custom',
    enablePullDownRefresh: true,
    backgroundColor: '#090C0F'
  }
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

const assetId = ref<string>('')
const scopeId = ref<string>('')
const currency = ref<string>(financeStore.defaultCurrency || 'CNY')
const benchmarkId = ref<string>('')
const snapshotMode = ref<'latest' | 'fixed'>('latest')
const requestedSnapshotId = ref<string>('')
const resolvedSnapshotId = ref<string>('')
const selectedRange = ref<string>(financeStore.defaultRange || '1Y')
const customStartDate = ref<string>('')
const customEndDate = ref<string>('')

const loading = ref<boolean>(true)
const chartLoading = ref<boolean>(false)
const errorMsg = ref<string>('')
const isExpiredSnapshot = ref<boolean>(false)
const isUnknownAsset = ref<boolean>(false)

const detail = ref<AssetDetailResponse | null>(null)
const curve = ref<TrendCurveResponse | null>(null)
const members = ref<AssetSummaryItem[]>([])
const memberNextCursor = ref<string | null>(null)
const memberTotal = ref<number>(0)
const loadingMoreMembers = ref<boolean>(false)
const membersLoadError = ref<string>('')
const memberSortOrder = ref<'desc' | 'asc'>('desc')

const detailRace = createAsyncRace()
const curveRace = createAsyncRace()
const membersRace = createAsyncRace()

onUnmounted(() => {
  detailRace.invalidate()
  curveRace.invalidate()
  membersRace.invalidate()
})

const asset = computed(() => detail.value?.asset || null)
const isFav = computed(() => financeStore.isAssetFavorite(assetId.value))
const isHomePool = computed(() => financeStore.defaultHomePool === assetId.value)
const isStale = computed(() => detail.value?.snapshot?.status === 'stale' || detail.value?.snapshot?.status === 'degraded')

const activeSnapshotId = computed(() => {
  if (snapshotMode.value === 'fixed')
    return requestedSnapshotId.value || resolvedSnapshotId.value
  return resolvedSnapshotId.value
})

const sourcePoolName = computed(() => {
  if (asset.value?.metrics?.strength?.scope_label)
    return asset.value.metrics.strength.scope_label
  if (detail.value?.breadcrumbs?.length)
    return detail.value.breadcrumbs[0]?.name || scopeId.value
  return scopeId.value || '全市场观察池'
})

const strengthRankText = computed(() => {
  const st = asset.value?.metrics?.strength
  if (!st || st.rank === null || st.rank === undefined || !st.valid_count)
    return '—'
  return `第 ${st.rank} / ${st.valid_count} 名`
})

const confirmedDateText = computed(() => {
  const m = asset.value?.metrics
  if (!m)
    return ''
  if (m.confirmed_at && m.duration_trading_days)
    return `确认于 ${m.confirmed_at} · 持续 ${m.duration_trading_days} 交易日`
  if (m.confirmed_at)
    return `确认于 ${m.confirmed_at}`
  if (m.started_at && m.confirmation === 'forming')
    return `形成于 ${m.started_at}`
  return ''
})

const rangeOptions = [
  { key: '1M', label: '近1月' },
  { key: '3M', label: '近3月' },
  { key: '6M', label: '近6月' },
  { key: '1Y', label: '近1年' },
  { key: '3Y', label: '近3年' },
  { key: 'ALL', label: '全部' }
]

function isRangeDisabled (r: string): boolean {
  const list = detail.value?.range_availability
  if (!list?.length)
    return false
  const found = list.find(it => it.range === r)
  return found
? !found.available
: false
}

function getRangeReason (r: string): string {
  const found = detail.value?.range_availability?.find(it => it.range === r)
  return found?.reason || ''
}

const tempMeta = computed(() => getTrendTempMeta(asset.value?.metrics.temperature))

const statusTitleText = computed(() => {
  const m = asset.value?.metrics
  if (!m)
    return '—'
  if (m.side === 'right' && m.confirmation === 'confirmed')
    return '右侧'
  if (m.side === 'left' && m.confirmation === 'confirmed')
    return '下行'
  if (m.confirmation === 'forming')
    return '形成中'
  if (m.side === 'neutral')
    return '震荡'
  return '待确认'
})

const statusSubText = computed(() => {
  const m = asset.value?.metrics
  if (!m)
    return '—'
  if (m.confirmation === 'confirmed' && m.duration_trading_days)
    return `持续${m.duration_trading_days}交易日`
  if (m.confirmation === 'forming')
    return '趋势形成中'
  return '未确认'
})

const durationBadge = computed(() => {
  const m = asset.value?.metrics
  if (!m)
    return { text: '—', isRight: false, isLeft: false }
  return formatDurationBadge(m.side, m.confirmation, m.duration_trading_days)
})

const formattedMainValue = computed(() => {
  const a = asset.value
  if (!a || a.latest_value === null || a.latest_value === undefined)
    return '—'
  if (a.measure_type === 'yield_pct')
    return `${a.latest_value.toFixed(2)}%`
  if (a.measure_type === 'fx_rate')
    return a.latest_value.toFixed(4)
  return a.latest_value.toFixed(2)
})

function formatSignedPct (val: number | null | undefined, digits = 1): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(digits)}%`
}

function pctColorClass (val: number | null | undefined) {
  if (val === null || val === undefined || !Number.isFinite(val) || val === 0)
    return isDark.value ? 'text-[#9AA5B1]' : 'text-[#64707C]'
  const isUp = val > 0
  const cnRedUp = financeStore.colorScheme === 'cn_red_up'
  const useRise = cnRedUp
? isUp
: !isUp
  if (useRise)
    return isDark.value ? 'text-[#F05A50]' : 'text-[#B93842]'
  return isDark.value ? 'text-[#2AD1A3]' : 'text-[#14745E]'
}

const latestAssetReturn = computed(() => {
  const pts = curve.value?.points || []
  for (let i = pts.length - 1; i >= 0; i--) {
    const pt = pts[i]
    if (pt && pt.return_pct !== null)
      return pt.return_pct
  }
  return null
})

const latestBenchReturn = computed(() => {
  const pts = curve.value?.benchmark_points || []
  for (let i = pts.length - 1; i >= 0; i--) {
    const pt = pts[i]
    if (pt && pt.return_pct !== null)
      return pt.return_pct
  }
  return null
})

const previewDates = computed<string[]>(() => {
  return (curve.value?.points || []).map(p => p.trading_date)
})

const previewSeries = computed(() => {
  if (!curve.value?.points?.length)
    return []
  const list = [
    {
      id: assetId.value,
      name: detail.value?.asset?.short_name || detail.value?.asset?.name || assetId.value,
      color: isDark.value ? '#E5B352' : '#8A641F',
      values: curve.value.points.map(p => p.return_pct),
      lineWidth: 2.2
    }
  ]
  if (curve.value.benchmark_points?.length) {
    const benchMap = new Map(curve.value.benchmark_points.map(p => [p.trading_date, p.return_pct]))
    list.push({
      id: curve.value.benchmark_id || 'benchmark',
      name: detail.value?.benchmark_name || curve.value.benchmark_name || '基准',
      color: isDark.value ? '#72A7D9' : '#326C9D',
      values: previewDates.value.map(d => benchMap.get(d) ?? null),
      lineWidth: 1.5
    })
  }
  return list
})

async function fetchCurve () {
  if (!assetId.value)
    return
  const token = curveRace.next()
  chartLoading.value = true
  try {
    const res = await getTrendCurve(assetId.value, {
      range: selectedRange.value,
      currency: toApiCurrency(currency.value),
      benchmark_id: benchmarkId.value || detail.value?.benchmark_id || undefined,
      snapshot_id: activeSnapshotId.value || undefined,
      start_date: customStartDate.value || undefined,
      end_date: customEndDate.value || undefined
    })
    if (!curveRace.isLatest(token))
      return
    curve.value = unwrapApiResponse(res, '趋势曲线加载失败')
  }
  catch {
    if (!curveRace.isLatest(token))
      return
    curve.value = null
  }
  finally {
    if (curveRace.isLatest(token))
      chartLoading.value = false
  }
}

async function fetchMembers (append = false) {
  if (!detail.value?.asset.has_members)
    return
  const token = membersRace.next()
  if (append)
    loadingMoreMembers.value = true
  membersLoadError.value = ''
  try {
    const res = await getAssetMembers(assetId.value, {
      scope_id: scopeId.value || assetId.value,
      sort: 'strength',
      order: memberSortOrder.value,
      cursor: append
? (memberNextCursor.value || undefined)
: undefined,
      limit: 30,
      currency: toApiCurrency(currency.value),
      snapshot_id: activeSnapshotId.value || undefined
    })
    if (!membersRace.isLatest(token))
      return
    const memData = unwrapApiResponse(res, '成员列表加载失败')
    members.value = append
      ? [
...members.value,
...memData.items
]
      : memData.items
    memberNextCursor.value = memData.next_cursor
    memberTotal.value = memData.matched_total ?? memData.total
  }
  catch (err: any) {
    if (!membersRace.isLatest(token))
      return
    if (append) {
      membersLoadError.value = isStaleCursorError(err)
        ? '快照已更新，请下拉刷新列表'
        : (err?.message || '继续加载失败，可重试')
    }
    else {
      members.value = []
      membersLoadError.value = err?.message || '成员加载失败'
    }
  }
  finally {
    if (membersRace.isLatest(token))
      loadingMoreMembers.value = false
  }
}

async function fetchAll (forceLatest = false) {
  if (!assetId.value)
    return
  const token = detailRace.next()
  if (forceLatest && snapshotMode.value === 'latest')
    resolvedSnapshotId.value = ''
  loading.value = true
  errorMsg.value = ''
  isExpiredSnapshot.value = false
  isUnknownAsset.value = false
  try {
    const reqSnap = snapshotMode.value === 'fixed'
? requestedSnapshotId.value
: ''
    const res = await getAssetDetail(assetId.value, {
      scope_id: scopeId.value || undefined,
      currency: toApiCurrency(currency.value),
      snapshot_id: reqSnap || undefined
    })
    if (!detailRace.isLatest(token))
      return
    const detailData = unwrapApiResponse(res, '资产详情响应为空')
    detail.value = detailData
    if (detailData.snapshot?.snapshot_id)
      resolvedSnapshotId.value = detailData.snapshot.snapshot_id
    if (detailData.benchmark_id && !benchmarkId.value)
      benchmarkId.value = detailData.benchmark_id
    if (detailData.available_ranges?.length && !detailData.available_ranges.includes(selectedRange.value)) {
      const fallback = detailData.available_ranges.includes('ALL')
        ? 'ALL'
        : detailData.available_ranges[detailData.available_ranges.length - 1]
      if (fallback)
        selectedRange.value = fallback
    }

    if (detailData.asset.has_members) {
      await fetchMembers(false)
    }
    else {
      members.value = []
      memberNextCursor.value = null
      memberTotal.value = 0
    }
    await fetchCurve()
  }
  catch (err: any) {
    if (!detailRace.isLatest(token))
      return
    if (isExpiredSnapshotError(err) && snapshotMode.value === 'fixed' && requestedSnapshotId.value) {
      snapshotMode.value = 'latest'
      requestedSnapshotId.value = ''
      resolvedSnapshotId.value = ''
      uni.showToast({ title: '原锁定快照已过期，已自动切换为最新快照', icon: 'none' })
      return fetchAll(true)
    }
    isExpiredSnapshot.value = isExpiredSnapshotError(err)
    isUnknownAsset.value = isNotFoundError(err)
    errorMsg.value = isUnknownAsset.value
      ? `未知标的「${assetId.value}」，请返回重选`
      : (err?.message || '资产详情加载失败')
  }
  finally {
    if (detailRace.isLatest(token)) {
      loading.value = false
      uni.stopPullDownRefresh()
    }
  }
}

onLoad((query: any) => {
  const parsed = decodeFinanceQuery(query, undefined, {
    defaultHomePool: financeStore.defaultHomePool,
    defaultCurrency: financeStore.defaultCurrency,
    defaultRange: financeStore.defaultRange
  })
  assetId.value = parsed.asset_id
  scopeId.value = parsed.scope_id
  currency.value = toDisplayCurrency(parsed.currency)
  selectedRange.value = parsed.range
  benchmarkId.value = parsed.benchmark_id
  snapshotMode.value = parsed.snapshot_mode
  requestedSnapshotId.value = parsed.snapshot_id
  resolvedSnapshotId.value = parsed.snapshot_id
  customStartDate.value = parsed.start_date || ''
  customEndDate.value = parsed.end_date || ''
  fetchAll(false)
})

onPullDownRefresh(() => {
  fetchAll(true)
})

function switchToLatestSnapshot () {
  snapshotMode.value = 'latest'
  requestedSnapshotId.value = ''
  resolvedSnapshotId.value = ''
  fetchAll(true)
}

function goBackOrReselect () {
  uni.navigateBack({
    fail: () => {
      uni.redirectTo({ url: '/pages-sub/finance/index' })
    }
  })
}

function switchRange (r: string) {
  if (isRangeDisabled(r)) {
    uni.showToast({ title: getRangeReason(r) || `历史不足 ${r}`, icon: 'none' })
    return
  }
  if (selectedRange.value === r && !customStartDate.value)
    return
  customStartDate.value = ''
  customEndDate.value = ''
  selectedRange.value = r
  fetchCurve()
}

async function toggleMemberSort () {
  memberSortOrder.value = memberSortOrder.value === 'desc'
    ? 'asc'
    : 'desc'
  memberNextCursor.value = null
  await fetchMembers(false)
}

function handleToggleFavorite (id: string) {
  const added = financeStore.toggleFavoriteAsset(id)
  uni.showToast({
    title: added ? '已加入收藏' : '已取消收藏',
    icon: 'none',
    duration: 1200
  })
}

function setAsHomePool () {
  financeStore.defaultHomePool = assetId.value
  uni.showToast({
    title: '已设为金融全景默认主页',
    icon: 'none',
    duration: 1500
  })
}

function openWorkbench (chart = 'curve') {
  const supported = detail.value?.available_charts || []
  const targetChart = supported.includes(chart as any)
? chart
: (supported[0] || 'curve')
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: assetId.value,
      scope_id: scopeId.value,
      currency: currency.value as any,
      range: selectedRange.value as any,
      chart: targetChart as any,
      benchmark_id: benchmarkId.value || detail.value?.benchmark_id || '',
      snapshot_mode: activeSnapshotId.value ? 'fixed' : snapshotMode.value,
      snapshot_id: activeSnapshotId.value,
      start_date: customStartDate.value,
      end_date: customEndDate.value
    })
  })
}

function openReport () {
  if (!detail.value || errorMsg.value || isExpiredSnapshot.value) {
    uni.showToast({ title: '数据未就绪，暂不可导出报告', icon: 'none' })
    return
  }
  const supported = detail.value.available_charts || []
  const targetChart = supported.includes('curve')
? 'curve'
: (supported[0] || 'curve')
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/report', {
      asset_id: assetId.value,
      scope_id: scopeId.value,
      currency: currency.value as any,
      range: selectedRange.value as any,
      chart: targetChart as any,
      benchmark_id: benchmarkId.value || detail.value?.benchmark_id || '',
      snapshot_mode: activeSnapshotId.value ? 'fixed' : snapshotMode.value,
      snapshot_id: activeSnapshotId.value
    })
  })
}

function openMemberDetail (item: AssetSummaryItem) {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: item.asset_id,
      scope_id: scopeId.value || assetId.value,
      currency: currency.value as any,
      range: selectedRange.value as any,
      snapshot_mode: activeSnapshotId.value ? 'fixed' : snapshotMode.value,
      snapshot_id: activeSnapshotId.value
    })
  })
}

function openBreadcrumb (id: string) {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/index', {
      scope_id: id,
      currency: currency.value as any
    })
  })
}

function handleSelectTrendEvent (ev: TrendEventItem) {
  const endDate = ev.as_of || ''
  let startDate = ''
  if (endDate) {
    const dt = new Date(`${endDate}T00:00:00Z`)
    if (!Number.isNaN(dt.getTime())) {
      dt.setUTCMonth(dt.getUTCMonth() - 3)
      startDate = dt.toISOString().slice(0, 10)
    }
  }
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: assetId.value,
      scope_id: scopeId.value,
      currency: currency.value as any,
      range: selectedRange.value as any,
      chart: 'curve',
      benchmark_id: benchmarkId.value || detail.value?.benchmark_id || '',
      snapshot_mode: activeSnapshotId.value ? 'fixed' : snapshotMode.value,
      snapshot_id: activeSnapshotId.value,
      start_date: startDate,
      end_date: endDate
    })
  })
}
</script>

<template>
  <FinanceShell
    :title="asset?.name || '资产详情'"
    active-tab="overview"
    :asset-id="assetId"
    :scope-id="scopeId"
    :currency="currency"
    :range="selectedRange"
    :snapshot-mode="snapshotMode"
    :snapshot-id="activeSnapshotId"
    :show-bottom-nav="false"
  >
    <view v-if="loading" class="px-3 py-4 space-y-3">
      <view
        v-for="n in 4"
        :key="n"
        class="h-[110px] rounded-xl animate-pulse"
        :class="isDark ? 'bg-[#12161B]' : 'bg-[#EDECE6]'"
      />
    </view>

    <view v-else-if="errorMsg" class="px-4 py-12 text-center space-y-3">
      <text class="text-[15px] font-bold block text-[#F05A50]">
        {{ isExpiredSnapshot ? '快照已过期' : (isUnknownAsset ? '未找到对应标的' : '无法获取资产详情') }}
      </text>
      <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
        {{ errorMsg }}
      </text>
      <view class="flex items-center justify-center gap-2 pt-1">
        <view
          v-if="isExpiredSnapshot"
          class="inline-flex items-center justify-center px-5 min-h-[44px] rounded-xl text-[13px] font-bold"
          :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-white'"
          @tap="switchToLatestSnapshot"
        >
          切回最新快照
        </view>
        <view
          v-if="!isUnknownAsset"
          class="inline-flex items-center justify-center px-5 min-h-[44px] rounded-xl border text-[13px] font-bold"
          :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
          @tap="fetchAll(false)"
        >
          重试
        </view>
        <view
          class="inline-flex items-center justify-center px-5 min-h-[44px] rounded-xl border text-[13px]"
          :class="isDark ? 'border-[#29313A] text-[#A9B3BD]' : 'border-[#DCDDD8] text-[#515F6A]'"
          @tap="goBackOrReselect"
        >
          返回重选
        </view>
      </view>
    </view>

    <view v-else-if="asset" class="px-3 py-3 space-y-3">
      <!-- 面包屑 -->
      <view v-if="detail?.breadcrumbs?.length" class="flex flex-wrap items-center gap-1">
        <text class="text-[11px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">金融资产</text>
        <template v-for="bc in detail.breadcrumbs" :key="bc.id">
          <text class="text-[11px]" :class="isDark ? 'text-[#64707C]' : 'text-[#8D99A4]'">/</text>
          <text
            class="text-[11px] active:opacity-70"
            :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'"
            @tap="openBreadcrumb(bc.id)"
          >
            {{ bc.name }}
          </text>
        </template>
      </view>

      <!-- 标题 + 收藏 + 分享 -->
      <view class="flex items-center justify-between gap-2">
        <view class="min-w-0 flex-1">
          <text class="text-[24px] font-bold tracking-tight block truncate">{{ asset.name }}</text>
          <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ asset.symbol }} · {{ asset.market }}
          </text>
        </view>
        <view class="flex items-center shrink-0">
          <view
            class="min-w-[44px] min-h-[44px] flex items-center justify-center"
            @tap="handleToggleFavorite(asset.asset_id)"
          >
            <text
              class="text-[20px]"
              :class="isFav ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]') : (isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]')"
            >
              {{ isFav ? '★' : '☆' }}
            </text>
          </view>
          <view
            class="min-w-[44px] min-h-[44px] flex items-center justify-center"
            @tap="openReport"
          >
            <text class="text-[18px]" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">↗</text>
          </view>
        </view>
      </view>

      <!-- 快照日期与状态标签 + 来自观察池 + 设主页 -->
      <view class="flex items-center justify-between gap-2">
        <view class="flex items-center gap-1.5 min-w-0">
          <view
            class="w-2 h-2 rounded-full shrink-0"
            :class="isStale ? 'bg-[#F05A50]' : (isDark ? 'bg-[#E5B352]' : 'bg-[#8A641F]')"
          />
          <text class="text-[11px] font-mono cc-tabular truncate" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">
            快照 · {{ detail?.snapshot.as_of || asset.as_of || '—' }} · 来自 {{ sourcePoolName }}
          </text>
        </view>
        <text
          class="px-2 py-0.5 rounded border text-[11px] shrink-0"
          :class="isDark ? 'border-[#E5B352]/35 text-[#E5B352] bg-[#E5B352]/10' : 'border-[#8A641F]/35 text-[#8A641F] bg-[#8A641F]/10'"
        >
          {{ detail?.snapshot?.is_current === false ? '历史快照' : '非实时数据' }}
        </text>
      </view>
      <text
        v-if="confirmedDateText"
        class="text-[11px] font-mono cc-tabular block"
        :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'"
      >
        {{ confirmedDateText }}
      </text>

      <view class="flex items-center justify-between gap-2">
        <view class="flex flex-wrap items-center gap-1.5">
          <text
            class="px-2 py-0.5 rounded text-[11px] font-medium border"
            :class="durationBadge.isRight
              ? (isDark ? 'bg-[#E5B352]/14 border-[#E5B352]/40 text-[#E5B352]' : 'bg-[#8A641F]/12 border-[#8A641F]/40 text-[#8A641F]')
              : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]')"
          >
            {{ durationBadge.text }}
          </text>
          <text
            v-if="asset.has_members && asset.right_ratio_pct !== null"
            class="px-2 py-0.5 rounded text-[11px] font-medium border"
            :class="isDark ? 'bg-[#E5B352]/14 border-[#E5B352]/40 text-[#E5B352]' : 'bg-[#8A641F]/12 border-[#8A641F]/40 text-[#8A641F]'"
          >
            右侧占比{{ Math.round(asset.right_ratio_pct) }}%
          </text>
          <text
            v-for="tag in asset.feature_tags.slice(0, 3)"
            :key="tag"
            class="px-2 py-0.5 rounded text-[11px] border"
            :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#72A7D9]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#326C9D]'"
          >
            {{ tag }}
          </text>
        </view>
        <view
          class="px-2.5 py-1 rounded-lg border text-[11px] shrink-0 min-h-[30px] flex items-center"
          :class="isHomePool
            ? (isDark ? 'border-[#E5B352] text-[#E5B352] bg-[#E5B352]/12' : 'border-[#8A641F] text-[#8A641F] bg-[#8A641F]/10')
            : (isDark ? 'border-[#29313A] text-[#A9B3BD] bg-[#12161B]' : 'border-[#DCDDD8] text-[#515F6A] bg-[#FFFFFF]')"
          @tap="setAsHomePool"
        >
          {{ isHomePool ? '已设主页' : '设为主页' }}
        </view>
      </view>

      <!-- 核心指标 3x2 网格 -->
      <view
        class="rounded-xl border p-3.5 grid grid-cols-3 gap-x-2 gap-y-4"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <view>
          <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ asset.has_members ? '综合指数' : '最新价格' }}
          </text>
          <text
            class="text-[20px] font-bold font-mono cc-tabular block mt-1 leading-tight"
            :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'"
          >
            {{ formattedMainValue }}
          </text>
          <text class="text-[11px] block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ asset.unit || asset.currency || '—' }}
          </text>
        </view>

        <view>
          <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">趋势状态</text>
          <text class="text-[20px] font-bold block mt-1 leading-tight">{{ statusTitleText }}</text>
          <text class="text-[11px] font-mono cc-tabular block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ statusSubText }}
          </text>
        </view>

        <view>
          <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">趋势温度</text>
          <view class="mt-1.5">
            <TrendTempBadge :temperature="asset.metrics.temperature" size="md" />
          </view>
          <text class="text-[11px] block mt-1" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ tempMeta.shortDesc }}
          </text>
        </view>

        <view>
          <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">相对强度</text>
          <text class="text-[20px] font-bold font-mono cc-tabular block mt-1 leading-tight">
            {{ asset.metrics.strength?.score !== null && asset.metrics.strength?.score !== undefined ? Math.round(asset.metrics.strength.score) : '—' }}
          </text>
          <text class="text-[11px] block mt-0.5 truncate font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ strengthRankText }}
          </text>
        </view>

        <view v-if="asset.has_members">
          <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">右侧占比</text>
          <text
            class="text-[20px] font-bold font-mono cc-tabular block mt-1 leading-tight"
            :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'"
          >
            {{ asset.right_ratio_pct !== null ? `${Math.round(asset.right_ratio_pct)}%` : '—' }}
          </text>
          <text class="text-[11px] font-mono cc-tabular block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ asset.right_count }}/{{ asset.valid_count }}有效
          </text>
        </view>
        <view v-else>
          <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">近1月变动</text>
          <text
            class="text-[20px] font-bold font-mono cc-tabular block mt-1 leading-tight"
            :class="pctColorClass(detail?.period_changes.month_1_pct)"
          >
            {{ formatSignedPct(detail?.period_changes.month_1_pct) }}
          </text>
          <text class="text-[11px] block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">滚动月度</text>
        </view>

        <view>
          <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">近1年变动</text>
          <text
            class="text-[20px] font-bold font-mono cc-tabular block mt-1 leading-tight"
            :class="pctColorClass(detail?.period_changes.year_1_pct)"
          >
            {{ formatSignedPct(detail?.period_changes.year_1_pct, 0) }}
          </text>
          <text class="text-[11px] block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">滚动年度</text>
        </view>
      </view>

      <!-- 趋势曲线预览卡 -->
      <view class="flex items-center justify-between pt-1">
        <text class="text-[16px] font-bold">
          {{ rangeOptions.find(r => r.key === selectedRange)?.label || '近1年' }}累计收益
        </text>
        <view
          v-if="detail?.available_charts?.length"
          class="min-h-[36px] flex items-center gap-0.5 active:opacity-70"
          @tap="openWorkbench('curve')"
        >
          <text class="text-[13px] font-medium" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">更多趋势图</text>
          <text class="text-[13px]" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">›</text>
        </view>
      </view>

      <view
        class="rounded-xl border p-3 space-y-3"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <!-- 图例 -->
        <view class="flex items-center gap-4 text-[11px]">
          <view class="flex items-center gap-1.5">
            <view class="w-3 h-[3px] rounded-full" :class="isDark ? 'bg-[#E5B352]' : 'bg-[#8A641F]'" />
            <text>{{ asset.short_name || asset.name }}</text>
          </view>
          <view v-if="detail?.benchmark_name" class="flex items-center gap-1.5">
            <view class="w-3 h-[3px] rounded-full" :class="isDark ? 'bg-[#72A7D9]' : 'bg-[#326C9D]'" />
            <text :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">{{ detail.benchmark_name }}</text>
          </view>
        </view>

        <!-- Canvas 预览图 -->
        <view v-if="chartLoading" class="h-[170px] rounded-lg animate-pulse" :class="isDark ? 'bg-[#181E25]' : 'bg-[#EDECE6]'" />
        <view v-else-if="!curve?.points?.length" class="h-[140px] flex items-center justify-center">
          <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">暂无区间曲线数据</text>
        </view>
        <TrendLineCanvas
          v-else
          canvas-id="assetPreviewCanvas"
          :dates="previewDates"
          :series="previewSeries"
          :height="180"
          value-format="pct"
          :show-right-labels="false"
        />

        <!-- 区间极值摘要 -->
        <view class="grid grid-cols-4 gap-2 pt-2 border-t" :class="isDark ? 'border-[#29313A]' : 'border-[#DCDDD8]'">
          <view>
            <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">当前累计</text>
            <text class="text-[13px] font-bold font-mono cc-tabular block mt-0.5" :class="pctColorClass(latestAssetReturn)">
              {{ formatSignedPct(latestAssetReturn, 0) }}
            </text>
          </view>
          <view>
            <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">比较基准</text>
            <text class="text-[13px] font-bold font-mono cc-tabular block mt-0.5" :class="pctColorClass(latestBenchReturn)">
              {{ formatSignedPct(latestBenchReturn, 0) }}
            </text>
          </view>
          <view>
            <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">最大回撤</text>
            <text class="text-[13px] font-bold font-mono cc-tabular block mt-0.5" :class="pctColorClass(curve?.extremum.max_drawdown_pct)">
              {{ formatSignedPct(curve?.extremum.max_drawdown_pct, 1) }}
            </text>
          </view>
          <view>
            <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">最大上涨</text>
            <text class="text-[13px] font-bold font-mono cc-tabular block mt-0.5" :class="pctColorClass(curve?.extremum.max_gain_pct)">
              {{ formatSignedPct(curve?.extremum.max_gain_pct, 0) }}
            </text>
          </view>
        </view>
      </view>

      <!-- 区间选择器（不足区间禁用并标注起始日） -->
      <view class="space-y-1.5">
        <view class="grid grid-cols-6 gap-1.5">
          <view
            v-for="r in rangeOptions"
            :key="r.key"
            class="min-h-[36px] rounded-lg border flex items-center justify-center text-[11px]"
            :class="isRangeDisabled(r.key)
              ? (isDark ? 'bg-[#12161B]/40 border-[#29313A]/40 text-[#64707C] opacity-45' : 'bg-[#EDECE6]/50 border-[#DCDDD8]/60 text-[#8D99A4] opacity-50')
              : selectedRange === r.key
                ? (isDark ? 'bg-[#E5B352]/16 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
            @tap="switchRange(r.key)"
          >
            {{ r.label }}
          </view>
        </view>
        <text
          v-if="detail?.earliest_date"
          class="text-[11px] font-mono cc-tabular block"
          :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'"
        >
          可用历史起始日：{{ detail.earliest_date }}（不足区间的选项已禁用）
        </text>
      </view>

      <!-- 近期趋势事件（支持分页、筛选与点击回溯区间曲线） -->
      <TrendEventsTimeline
        :asset-id="assetId"
        :snapshot-id="activeSnapshotId"
        :initial-events="detail?.recent_events || []"
        @select-event="handleSelectTrendEvent"
      />

      <!-- 成员表（仅父级/分类节点展示，叶节点严格隐藏，首次30条支持继续加载与失败重试） -->
      <view v-if="asset.has_members" class="space-y-2 pt-1">
        <view class="flex items-center justify-between">
          <text class="text-[16px] font-bold">
            成员品种 ({{ members.length }}{{ memberTotal > members.length ? `/${memberTotal}` : '' }})
          </text>
          <view
            class="px-2.5 py-1 rounded-lg border text-[12px]"
            :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#E5B352]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#8A641F]'"
            @tap="toggleMemberSort"
          >
            强度{{ memberSortOrder === 'desc' ? '降序 ↓' : '升序 ↑' }}
          </view>
        </view>

        <view
          class="rounded-xl border overflow-hidden"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
        >
          <view v-if="!members.length && !membersLoadError" class="p-6 text-center">
            <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">暂无下属成员数据</text>
          </view>
          <AssetRow
            v-for="mem in members"
            :key="mem.asset_id"
            :item="mem"
            @select="openMemberDetail"
            @toggle-favorite="handleToggleFavorite"
          />
        </view>

        <view v-if="membersLoadError" class="rounded-xl border border-[#F05A50]/40 bg-[#F05A50]/10 px-3 py-2.5 flex items-center justify-between gap-2">
          <text class="text-[12px] text-[#F05A50] flex-1">{{ membersLoadError }}</text>
          <text
            class="px-3 py-1 rounded-lg border border-[#F05A50] text-[12px] font-bold text-[#F05A50] shrink-0"
            @tap="fetchMembers(members.length > 0)"
          >
            重试
          </text>
        </view>

        <view v-if="memberNextCursor" class="pt-1 flex justify-center">
          <view
            class="px-5 min-h-[44px] rounded-xl border flex items-center justify-center text-[13px]"
            :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]'"
            @tap="!loadingMoreMembers && fetchMembers(true)"
          >
            {{ loadingMoreMembers ? '加载中...' : `继续加载 (${members.length}/${memberTotal})` }}
          </view>
        </view>
      </view>
    </view>
  </FinanceShell>
</template>
