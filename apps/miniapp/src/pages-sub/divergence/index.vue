<script lang="ts" setup>
import type { AssetDetailResponse, ChartType } from '@/api/types/trend'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { getAssetCatalog, getAssetDetail } from '@/api/trend'
import AnnualTrendView from '@/components/trend/AnnualTrendView.vue'
import ComponentTrendView from '@/components/trend/ComponentTrendView.vue'
import RightDistributionView from '@/components/trend/RightDistributionView.vue'
import SolarTermsView from '@/components/trend/SolarTermsView.vue'
import TrendCurveView from '@/components/trend/TrendCurveView.vue'
import TrendDashboardView from '@/components/trend/TrendDashboardView.vue'
import TrendMenuDropdown from '@/components/trend/TrendMenuDropdown.vue'
import TrendTierView from '@/components/trend/TrendTierView.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, isExpiredSnapshotError, isNotFoundError, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceSharePath, buildFinanceUrl, decodeFinanceQuery, toApiCurrency, toDisplayCurrency } from '@/utils/financeRoute'

definePage({
  style: {
    navigationBarTitleText: '七图研究工作台',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    enablePullDownRefresh: false,
    navigationStyle: 'custom'
  }
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

interface ChartTabMeta {
  key: ChartType
  label: string
  shortLabel: string
  row: 1 | 2
}

const CHART_TABS: ChartTabMeta[] = [
  {
    key: 'curve',
    label: '趋势曲线',
    shortLabel: '曲线',
    row: 1
  },
  {
    key: 'annual',
    label: '年度同期走势',
    shortLabel: '年度',
    row: 1
  },
  {
    key: 'components',
    label: '品种成分趋势',
    shortLabel: '成分',
    row: 1
  },
  {
    key: 'dashboard',
    label: '趋势仪表盘',
    shortLabel: '仪表盘',
    row: 1
  },
  {
    key: 'season',
    label: '趋势节气图',
    shortLabel: '节气图',
    row: 2
  },
  {
    key: 'distribution',
    label: '右侧分布图',
    shortLabel: '右侧分布',
    row: 2
  },
  {
    key: 'tier',
    label: '趋势梯队榜',
    shortLabel: '梯队榜',
    row: 2
  }
]

const quickAssets = ref<Array<{ id: string, label: string }>>([])

const currentAssetId = ref<string>('universe_macro')
const currentChart = ref<ChartType>('curve')
const currentRange = ref<string>(financeStore.defaultRange || '1Y')
const currentCurrency = ref<string>(financeStore.defaultCurrency || 'CNY')
const currentScopeId = ref<string>('')
const currentBenchmarkId = ref<string>('')
const snapshotMode = ref<'latest' | 'fixed'>('latest')
const requestedSnapshotId = ref<string>('')
const resolvedSnapshotId = ref<string>('')
const customStartDate = ref<string>('')
const customEndDate = ref<string>('')

const assetDetail = ref<AssetDetailResponse | null>(null)
const detailLoading = ref(false)
const detailError = ref('')
const detailExpired = ref(false)
const showDropdown = ref(false)
const showConfigModal = ref(false)
const showFavoriteSheet = ref(false)
const favoriteNote = ref('')
const detailRace = createAsyncRace()
const catalogRace = createAsyncRace()

// 安全区与胶囊适配
const statusBarHeight = ref(24)
const navContentHeight = ref(44)
const capsuleRightPadding = ref(12)
const safeBottomPx = ref(0)

function measureSafeArea() {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 24
    safeBottomPx.value = sys.safeAreaInsets?.bottom ?? 0

    // #ifdef MP-WEIXIN
    if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
      const rect = uni.getMenuButtonBoundingClientRect()
      if (rect && rect.top && rect.height) {
        const gap = Math.max(4, rect.top - statusBarHeight.value)
        navContentHeight.value = rect.height + gap * 2
        const winW = sys.windowWidth || 375
        capsuleRightPadding.value = Math.max(12, winW - rect.left + 8)
      }
    }
    // #endif
  }
  catch {
    statusBarHeight.value = 24
    navContentHeight.value = 44
  }
}

onMounted(() => {
  measureSafeArea()
  loadCatalogQuickAssets()
})

onUnmounted(() => {
  detailRace.invalidate()
  catalogRace.invalidate()
})

async function loadCatalogQuickAssets() {
  const token = catalogRace.next()
  try {
    const res = await getAssetCatalog({
currency: toApiCurrency(currentCurrency.value)
})
    if (!catalogRace.isLatest(token))
      return
    const cat = unwrapApiResponse(res, '目录加载失败')
    quickAssets.value = (cat.universes || [])
      .filter(u => u.member_count > 0)
      .map(u => ({ id: u.universe_id, label: u.name }))
  }
  catch {
    if (!catalogRace.isLatest(token))
      return
    quickAssets.value = []
  }
}

const chartState = computed<'loading' | 'error' | 'empty' | 'ready'>(() => {
  if (detailLoading.value && !assetDetail.value)
    return 'loading'
  if (detailError.value)
    return 'error'
  if (!assetDetail.value)
    return 'loading'
  if (!assetDetail.value.available_charts || assetDetail.value.available_charts.length === 0)
    return 'empty'
  return 'ready'
})

const availableCharts = computed<ChartType[]>(() => {
  if (chartState.value !== 'ready')
    return []
  return assetDetail.value?.available_charts ?? []
})

const visibleTabs = computed(() => CHART_TABS.filter(t => availableCharts.value.includes(t.key)))
const row1Tabs = computed(() => visibleTabs.value.filter(t => t.row === 1))
const row2Tabs = computed(() => visibleTabs.value.filter(t => t.row === 2))

const currentChartTitle = computed(() => {
  return CHART_TABS.find(t => t.key === currentChart.value)?.label || '趋势研究'
})

const displayAssetName = computed(() => {
  return assetDetail.value?.asset?.name
    || quickAssets.value.find(a => a.id === currentAssetId.value)?.label
    || currentAssetId.value
})

const displayAsOfDate = computed(() => {
  return assetDetail.value?.snapshot?.as_of || assetDetail.value?.asset?.as_of || '—'
})

const activeSnapshotIdForQuery = computed(() => {
  if (snapshotMode.value === 'fixed')
    return requestedSnapshotId.value || resolvedSnapshotId.value
  return resolvedSnapshotId.value
})

function isChartAvailable(chart: ChartType): boolean {
  return availableCharts.value.includes(chart)
}

function isRangeDisabled(r: string): boolean {
  const list = assetDetail.value?.range_availability
  if (!list?.length)
    return false
  const item = list.find(it => it.range === r)
  return item
? !item.available
: false
}

async function fetchAssetDetail() {
  if (!currentAssetId.value)
    return
  const token = detailRace.next()
  detailLoading.value = true
  detailError.value = ''
  detailExpired.value = false
  try {
    const querySnapshotId = snapshotMode.value === 'fixed'
? requestedSnapshotId.value
: ''
    const res = await getAssetDetail(currentAssetId.value, {
      scope_id: currentScopeId.value || undefined,
      currency: toApiCurrency(currentCurrency.value),
      snapshot_id: querySnapshotId || undefined
    })
    if (!detailRace.isLatest(token))
      return
    const data = unwrapApiResponse(res, '资产详情加载失败')
    assetDetail.value = data

    if (data.snapshot?.snapshot_id)
      resolvedSnapshotId.value = data.snapshot.snapshot_id

    if (data.benchmark_id && !currentBenchmarkId.value)
      currentBenchmarkId.value = data.benchmark_id

    if (data.available_charts?.length && !data.available_charts.includes(currentChart.value)) {
      const firstChart = data.available_charts[0]
      if (firstChart)
        currentChart.value = firstChart
    }
    if (data.available_ranges?.length && !data.available_ranges.includes(currentRange.value)) {
      const fallbackRange = data.available_ranges.includes('ALL')
        ? 'ALL'
        : data.available_ranges[data.available_ranges.length - 1]
      if (fallbackRange)
        currentRange.value = fallbackRange
    }
  }
  catch (err: any) {
    if (!detailRace.isLatest(token))
      return
    assetDetail.value = null
    detailExpired.value = isExpiredSnapshotError(err)
    if (isNotFoundError(err))
      detailError.value = `未找到标的「${currentAssetId.value}」，可返回重选其他资产`
    else
      detailError.value = err?.message || '资产详情加载失败，请重试'
  }
  finally {
    if (detailRace.isLatest(token))
      detailLoading.value = false
  }
}

function switchAsset(nextAssetId: string) {
  if (!nextAssetId || nextAssetId === currentAssetId.value)
    return
  // 切换标的时重验基准、自定义窗口与最新快照绑定
  currentAssetId.value = nextAssetId
  currentScopeId.value = nextAssetId.startsWith('universe_')
? nextAssetId
: currentScopeId.value
  currentBenchmarkId.value = ''
  customStartDate.value = ''
  customEndDate.value = ''
  if (snapshotMode.value === 'latest')
    resolvedSnapshotId.value = ''
  assetDetail.value = null
  fetchAssetDetail()
}

function switchToLatestSnapshot() {
  snapshotMode.value = 'latest'
  requestedSnapshotId.value = ''
  resolvedSnapshotId.value = ''
  fetchAssetDetail()
}

watch(
  () => [currentCurrency.value, currentScopeId.value],
  () => {
    if (currentAssetId.value)
      fetchAssetDetail()
  }
)

onLoad((query: any) => {
  const parsed = decodeFinanceQuery(query, undefined, {
    defaultHomePool: financeStore.defaultHomePool,
    defaultCurrency: financeStore.defaultCurrency,
    defaultRange: financeStore.defaultRange
  }, {
    asset_id: financeStore.defaultHomePool || 'universe_macro',
    chart: 'curve'
  })

  currentAssetId.value = parsed.asset_id
  currentScopeId.value = parsed.scope_id
  currentCurrency.value = toDisplayCurrency(parsed.currency)
  currentRange.value = parsed.range
  currentChart.value = parsed.chart === 'matrix'
? 'curve'
: parsed.chart
  currentBenchmarkId.value = parsed.benchmark_id
  snapshotMode.value = parsed.snapshot_mode
  requestedSnapshotId.value = parsed.snapshot_id
  resolvedSnapshotId.value = parsed.snapshot_id
  customStartDate.value = parsed.start_date || ''
  customEndDate.value = parsed.end_date || ''

  fetchAssetDetail()
})

function handleSelectChart(chart: ChartType) {
  if (!isChartAvailable(chart)) {
    uni.showToast({ title: '当前标的暂无该图表成员数据', icon: 'none' })
    return
  }
  currentChart.value = chart
  showDropdown.value = false
}

function handleGoBack() {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function handleSaveFavoriteChart(mode: 'latest' | 'fixed') {
  const snapId = mode === 'fixed'
? (activeSnapshotIdForQuery.value || '')
: ''
  if (mode === 'fixed' && !snapId) {
    uni.showToast({ title: '暂无有效快照可固定保存', icon: 'none' })
    return
  }
  try {
    financeStore.saveFavoriteChart({
      title: `${displayAssetName.value} · ${currentChartTitle.value}`,
      chart: currentChart.value,
      asset_id: currentAssetId.value,
      asset_name: displayAssetName.value,
      snapshot_mode: mode,
      snapshot_id: snapId || undefined,
      note: favoriteNote.value.trim() || undefined,
      context: {
        asset_id: currentAssetId.value,
        scope_id: currentScopeId.value,
        chart: currentChart.value,
        range: currentRange.value,
        currency: toApiCurrency(currentCurrency.value),
        benchmark_id: currentBenchmarkId.value,
        snapshot_mode: mode,
        snapshot_id: snapId
      }
    })
    showFavoriteSheet.value = false
    favoriteNote.value = ''
    uni.showToast({
title: mode === 'latest'
? '已加入每日跟踪'
: '已保存本期快照',
icon: 'success'
})
  }
  catch {
    uni.showToast({ title: '收藏失败，请稍后重试', icon: 'none' })
  }
}

function handleExportReport() {
  if (chartState.value !== 'ready' || detailExpired.value) {
    uni.showToast({ title: '当前标的或快照不可用，无法生成报告', icon: 'none' })
    return
  }
  const snapId = activeSnapshotIdForQuery.value
  const url = buildFinanceUrl('/pages-sub/finance/report', {
    asset_id: currentAssetId.value,
    scope_id: currentScopeId.value,
    currency: currentCurrency.value as any,
    range: currentRange.value as any,
    chart: currentChart.value,
    benchmark_id: currentBenchmarkId.value,
    snapshot_mode: snapId ? 'fixed' : snapshotMode.value,
    snapshot_id: snapId,
    start_date: customStartDate.value,
    end_date: customEndDate.value
  })
  uni.navigateTo({
    url,
    fail: () => {
      uni.showToast({ title: '打开报告页失败', icon: 'none' })
    }
  })
}

function getCurrentSharePath(): string {
  const snapId = activeSnapshotIdForQuery.value
  return buildFinanceSharePath('/pages-sub/divergence/index', {
    asset_id: currentAssetId.value,
    scope_id: currentScopeId.value,
    currency: currentCurrency.value as any,
    range: currentRange.value as any,
    chart: currentChart.value,
    benchmark_id: currentBenchmarkId.value,
    snapshot_mode: snapId ? 'fixed' : snapshotMode.value,
    snapshot_id: snapId,
    start_date: customStartDate.value,
    end_date: customEndDate.value
  })
}

function handleShareTrigger() {
  const sharePath = getCurrentSharePath()
  // #ifdef H5
  if (typeof navigator !== 'undefined' && navigator.clipboard) {
    navigator.clipboard.writeText(sharePath)
      .then(() => uni.showToast({ title: '已复制分享路径', icon: 'success' }))
      .catch(() => uni.showToast({ title: '请使用右上角分享', icon: 'none' }))
    return
  }
  // #endif
  uni.showToast({ title: '请点击右上角转发分享', icon: 'none' })
}

onShareAppMessage(() => {
  return {
    title: `${displayAssetName.value} · ${currentChartTitle.value}`,
    path: getCurrentSharePath()
  }
})
</script>

<template>
  <view
    class="min-h-screen flex flex-col box-border"
    :class="isDark ? 'bg-[#090C0F] text-[#F1EFEA]' : 'cc-finance-light bg-[#F4F3EF] text-[#22292E]'"
  >
    <!-- 顶部胶囊安全区导航栏 -->
    <view
      class="sticky top-0 z-40 border-b"
      :class="isDark ? 'bg-[#090C0F]/95 border-[#29313A]' : 'bg-[#F4F3EF]/95 border-[#DCDDD8]'"
      :style="{ paddingTop: `${statusBarHeight}px` }"
    >
      <view
        class="flex items-center justify-between px-2"
        :style="{ height: `${navContentHeight}px`, paddingRight: `${capsuleRightPadding}px` }"
      >
        <view class="flex items-center min-w-0 flex-1 gap-1">
          <view
            class="min-w-[44px] min-h-[44px] flex items-center justify-center shrink-0 active:opacity-70"
            @tap="handleGoBack"
          >
            <text class="text-[20px] font-medium leading-none">‹</text>
          </view>
          <view class="min-w-0 flex-1">
            <text class="text-[15px] font-bold truncate block leading-tight">
              {{ displayAssetName }}
            </text>
            <text class="text-[10px] font-mono block leading-tight" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
              {{ displayAsOfDate }}
            </text>
          </view>
        </view>

        <!-- 中右视角下拉触发器 -->
        <view
          class="flex items-center gap-1 px-2.5 py-1 rounded-full border shrink-0 active:scale-95 transition-transform"
          :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
          @tap="showDropdown = !showDropdown"
        >
          <text class="text-[11px]" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">◎</text>
          <text class="text-[12px] font-bold">{{ currentChartTitle }}</text>
          <text class="text-[10px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">⌄</text>
        </view>
      </view>

      <!-- 标的快捷切换（仅展示目录真实存在且有成员的观察池） + 支持视图 Tab -->
      <view class="px-3 pb-2 space-y-1.5">
        <scroll-view v-if="quickAssets.length > 0" scroll-x class="whitespace-nowrap w-full">
          <view class="inline-flex items-center gap-1.5">
            <text
              v-for="asset in quickAssets"
              :key="asset.id"
              class="px-2 py-0.5 rounded-md text-[11px] font-medium transition-colors"
              :class="currentAssetId === asset.id
                ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] text-white font-bold')
                : (isDark ? 'bg-[#12181F] text-[#8D99A4] border border-[#29313A]' : 'bg-white text-[#515F6A] border border-[#DCDDD8]')"
              @tap="switchAsset(asset.id)"
            >
              {{ asset.label }}
            </text>
          </view>
        </scroll-view>

        <!-- 按 available_charts 四态渲染 -->
        <view v-if="chartState === 'loading'" class="grid grid-cols-4 gap-1">
          <view
            v-for="n in 4"
            :key="n"
            class="h-[28px] rounded-lg animate-pulse"
            :class="isDark ? 'bg-[#12181F]' : 'bg-[#EDECE6]'"
          />
        </view>
        <view v-else-if="chartState === 'ready'" class="space-y-1">
          <view v-if="row1Tabs.length" class="grid gap-1" :class="row1Tabs.length <= 2 ? 'grid-cols-2' : 'grid-cols-4'">
            <view
              v-for="tab in row1Tabs"
              :key="tab.key"
              class="py-1.5 rounded-lg text-center text-[11px] font-medium border transition-colors"
              :class="currentChart === tab.key
                ? (isDark ? 'bg-[#E5B352]/20 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/15 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#12181F] border-[#29313A] text-[#8D99A4]' : 'bg-white border-[#DCDDD8] text-[#515F6A]')"
              @tap="handleSelectChart(tab.key)"
            >
              {{ tab.shortLabel }}
            </view>
          </view>
          <view v-if="row2Tabs.length" class="grid grid-cols-3 gap-1">
            <view
              v-for="tab in row2Tabs"
              :key="tab.key"
              class="py-1.5 rounded-lg text-center text-[11px] font-medium border transition-colors"
              :class="currentChart === tab.key
                ? (isDark ? 'bg-[#E5B352]/20 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/15 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#12181F] border-[#29313A] text-[#8D99A4]' : 'bg-white border-[#DCDDD8] text-[#515F6A]')"
              @tap="handleSelectChart(tab.key)"
            >
              {{ tab.shortLabel }}
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 视角下拉菜单 -->
    <TrendMenuDropdown
      :visible="showDropdown"
      :current-chart="currentChart"
      :available-charts="availableCharts"
      @select="handleSelectChart"
      @close="showDropdown = false"
    />

    <!-- 当前图表视图（按需懒挂载，独立页面滚动，避免 swiper 与 Canvas 手势冲突） -->
    <view
      class="flex-1 px-3.5 py-3"
      :style="{ paddingBottom: `${76 + safeBottomPx}px` }"
    >
      <view v-if="chartState === 'loading'" class="py-12 text-center">
        <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          正在加载资产与支持视图...
        </text>
      </view>

      <view
        v-else-if="chartState === 'error'"
        class="rounded-2xl border p-6 text-center space-y-3"
        :class="isDark ? 'bg-[#12181F] border-[#F05A50]/40' : 'bg-white border-[#FECACA]'"
      >
        <text class="text-[15px] font-bold block text-[#F05A50]">
          {{ detailExpired ? '历史快照已过期或不存在' : '研究工作台加载失败' }}
        </text>
        <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          {{ detailError }}
        </text>
        <view class="flex items-center justify-center gap-2 pt-1">
          <view
            v-if="detailExpired"
            class="px-4 min-h-[40px] rounded-xl flex items-center justify-center text-[12px] font-bold"
            :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-white'"
            @tap="switchToLatestSnapshot"
          >
            切回最新快照
          </view>
          <view
            class="px-4 min-h-[40px] rounded-xl border flex items-center justify-center text-[12px] font-bold"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="fetchAssetDetail"
          >
            重试
          </view>
          <view
            class="px-4 min-h-[40px] rounded-xl border flex items-center justify-center text-[12px]"
            :class="isDark ? 'border-[#29313A] text-[#A9B3BD]' : 'border-[#DCDDD8] text-[#515F6A]'"
            @tap="handleGoBack"
          >
            返回重选
          </view>
        </view>
      </view>

      <view
        v-else-if="chartState === 'empty'"
        class="rounded-2xl border p-6 text-center space-y-2"
        :class="isDark ? 'bg-[#12181F] border-[#29313A]' : 'bg-white border-[#DCDDD8]'"
      >
        <text class="text-[14px] font-bold block">当前标的暂无可用研究视图</text>
        <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
          {{ assetDetail?.asset?.metrics?.reason || '暂无足够历史数据或授权覆盖' }}
        </text>
      </view>

      <template v-else>
        <TrendCurveView
          v-if="currentChart === 'curve'"
          :asset-id="currentAssetId"
          :snapshot-id="activeSnapshotIdForQuery"
          :range="currentRange"
          :currency="currentCurrency"
          :scope-id="currentScopeId"
          :benchmark-id="currentBenchmarkId"
          :start-date="customStartDate"
          :end-date="customEndDate"
          :range-availability="assetDetail?.range_availability || []"
          :earliest-date="assetDetail?.earliest_date || null"
          @range-change="r => { currentRange = r; customStartDate = ''; customEndDate = '' }"
        />
        <AnnualTrendView
          v-else-if="currentChart === 'annual'"
          :asset-id="currentAssetId"
          :snapshot-id="activeSnapshotIdForQuery"
          :range="currentRange"
          :currency="currentCurrency"
          :scope-id="currentScopeId"
          :benchmark-id="currentBenchmarkId"
        />
        <ComponentTrendView
          v-else-if="currentChart === 'components'"
          :asset-id="currentAssetId"
          :snapshot-id="activeSnapshotIdForQuery"
          :range="currentRange"
          :currency="currentCurrency"
          :scope-id="currentScopeId"
          :benchmark-id="currentBenchmarkId"
        />
        <TrendDashboardView
          v-else-if="currentChart === 'dashboard'"
          :asset-id="currentAssetId"
          :snapshot-id="activeSnapshotIdForQuery"
          :range="currentRange"
          :currency="currentCurrency"
          :scope-id="currentScopeId"
          :benchmark-id="currentBenchmarkId"
        />
        <SolarTermsView
          v-else-if="currentChart === 'season'"
          :asset-id="currentAssetId"
          :snapshot-id="activeSnapshotIdForQuery"
          :range="currentRange"
          :currency="currentCurrency"
          :scope-id="currentScopeId"
          :benchmark-id="currentBenchmarkId"
        />
        <RightDistributionView
          v-else-if="currentChart === 'distribution'"
          :asset-id="currentAssetId"
          :snapshot-id="activeSnapshotIdForQuery"
          :range="currentRange"
          :currency="currentCurrency"
          :scope-id="currentScopeId"
          :benchmark-id="currentBenchmarkId"
        />
        <TrendTierView
          v-else-if="currentChart === 'tier'"
          :asset-id="currentAssetId"
          :snapshot-id="activeSnapshotIdForQuery"
          :range="currentRange"
          :currency="currentCurrency"
          :scope-id="currentScopeId"
          :benchmark-id="currentBenchmarkId"
        />
      </template>
    </view>

    <!-- 底部工具栏 -->
    <view
      class="fixed bottom-0 left-0 right-0 z-30 border-t px-3"
      :class="isDark ? 'bg-[#090C0F]/95 border-[#29313A]' : 'bg-white/95 border-[#DCDDD8]'"
      :style="{ paddingBottom: `${safeBottomPx}px` }"
    >
      <view class="grid grid-cols-4 gap-2 h-[54px] items-center">
        <view
          class="min-h-[44px] flex flex-col items-center justify-center rounded-lg active:opacity-70"
          @tap="showFavoriteSheet = true"
        >
          <text class="text-[12px] font-medium" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
            ★ 收藏研究
          </text>
        </view>
        <view
          class="min-h-[44px] flex flex-col items-center justify-center rounded-lg active:opacity-70"
          @tap="handleShareTrigger"
        >
          <text class="text-[12px] font-medium" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
            分享
          </text>
        </view>
        <view
          class="min-h-[44px] flex flex-col items-center justify-center rounded-lg active:opacity-70"
          @tap="handleExportReport"
        >
          <text class="text-[12px] font-medium" :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'">
            导出报告
          </text>
        </view>
        <view
          class="min-h-[44px] flex flex-col items-center justify-center rounded-lg active:opacity-70"
          @tap="showConfigModal = true"
        >
          <text class="text-[12px] font-medium" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
            参数配置
          </text>
        </view>
      </view>
    </view>

    <!-- 参数配置弹层 -->
    <view v-if="showConfigModal" class="fixed inset-0 z-50 flex items-end justify-center">
      <view class="fixed inset-0 bg-black/60" @tap="showConfigModal = false" />
      <view
        class="relative w-full rounded-t-2xl border-t p-4 space-y-4 z-10"
        :class="isDark ? 'bg-[#12181F] border-[#29313A] text-[#F1EFEA]' : 'bg-white border-[#DCDDD8] text-[#22292E]'"
        :style="{ paddingBottom: `${20 + safeBottomPx}px` }"
      >
        <view class="flex items-center justify-between">
          <text class="text-[16px] font-bold">图表参数配置</text>
          <text
            class="px-3 py-1 rounded-lg text-[12px] font-bold"
            :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-white'"
            @tap="showConfigModal = false"
          >
            完成
          </text>
        </view>

        <view class="space-y-1.5">
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
            时间范围
          </text>
          <view class="flex flex-wrap gap-2">
            <text
              v-for="r in (assetDetail?.available_ranges?.length ? assetDetail.available_ranges : ['3M', '6M', '1Y', '3Y', 'ALL'])"
              :key="r"
              class="px-3 py-1.5 rounded-lg text-[12px] font-mono border"
              :class="isRangeDisabled(r)
                ? 'opacity-35 line-through'
                : currentRange === r
                  ? (isDark ? 'bg-[#E5B352] border-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] border-[#8A641F] text-white font-bold')
                  : (isDark ? 'border-[#29313A] text-[#8D99A4]' : 'border-[#DCDDD8] text-[#515F6A]')"
              @tap="!isRangeDisabled(r) && (currentRange = r)"
            >
              {{ r }}
            </text>
          </view>
        </view>

        <view class="space-y-1.5">
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
            计价币种
          </text>
          <view class="flex flex-wrap gap-2">
            <text
              v-for="c in (assetDetail?.available_currencies?.length ? assetDetail.available_currencies : ['CNY', 'USD', 'HKD', 'RAW'])"
              :key="c"
              class="px-3 py-1.5 rounded-lg text-[12px] font-mono border"
              :class="toApiCurrency(currentCurrency) === toApiCurrency(c)
                ? (isDark ? 'bg-[#E5B352] border-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] border-[#8A641F] text-white font-bold')
                : (isDark ? 'border-[#29313A] text-[#8D99A4]' : 'border-[#DCDDD8] text-[#515F6A]')"
              @tap="currentCurrency = toDisplayCurrency(c)"
            >
              {{ c === 'ORIGINAL' ? '原币' : c }}
            </text>
          </view>
        </view>
      </view>
    </view>

    <!-- 收藏方式选择弹层：每日跟踪 vs 保存本期（私密备注不入分享参数） -->
    <view v-if="showFavoriteSheet" class="fixed inset-0 z-50 flex items-end justify-center">
      <view class="fixed inset-0 bg-black/60" @tap="showFavoriteSheet = false" />
      <view
        class="relative w-full rounded-t-2xl border-t p-4 space-y-3 z-10"
        :class="isDark ? 'bg-[#12181F] border-[#29313A] text-[#F1EFEA]' : 'bg-white border-[#DCDDD8] text-[#22292E]'"
        :style="{ paddingBottom: `${20 + safeBottomPx}px` }"
      >
        <view class="flex items-center justify-between">
          <text class="text-[16px] font-bold">收藏研究配置</text>
          <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'" @tap="showFavoriteSheet = false">
            取消
          </text>
        </view>
        <view>
          <text class="text-[11px] block mb-1" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
            私密备注（仅本地保存，不会带入任何分享链接）
          </text>
          <input
            v-model="favoriteNote"
            type="text"
            placeholder="可选：记录观察理由或触发条件"
            class="w-full px-3 py-2 rounded-lg border text-[12px]"
            :class="isDark ? 'bg-[#090C0F] border-[#29313A] text-[#F1EFEA]' : 'bg-[#F4F3EF] border-[#DCDDD8] text-[#22292E]'"
          >
        </view>
        <view class="grid grid-cols-2 gap-2.5 pt-1">
          <view
            class="p-3 rounded-xl border text-center space-y-1 active:opacity-80"
            :class="isDark ? 'bg-[#090C0F] border-[#E5B352]/50' : 'bg-[#F4F3EF] border-[#8A641F]/50'"
            @tap="handleSaveFavoriteChart('latest')"
          >
            <text class="text-[13px] font-bold block" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
              每日跟踪
            </text>
            <text class="text-[10px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              下次打开解析最新快照
            </text>
          </view>
          <view
            class="p-3 rounded-xl text-center space-y-1 active:opacity-80"
            :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-white'"
            @tap="handleSaveFavoriteChart('fixed')"
          >
            <text class="text-[13px] font-bold block">保存本期</text>
            <text class="text-[10px] opacity-80 block">
              锁定 {{ displayAsOfDate }} 快照
            </text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
