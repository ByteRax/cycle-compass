<script lang="ts" setup>
import type { AssetDetailResponse, ChartType } from '@/api/types/trend'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { computed, onMounted, ref, watch } from 'vue'
import { getAssetDetail } from '@/api/trend'
import AnnualTrendView from '@/components/trend/AnnualTrendView.vue'
import ComponentTrendView from '@/components/trend/ComponentTrendView.vue'
import RightDistributionView from '@/components/trend/RightDistributionView.vue'
import SolarTermsView from '@/components/trend/SolarTermsView.vue'
import TrendCurveView from '@/components/trend/TrendCurveView.vue'
import TrendDashboardView from '@/components/trend/TrendDashboardView.vue'
import TrendMenuDropdown from '@/components/trend/TrendMenuDropdown.vue'
import TrendTierView from '@/components/trend/TrendTierView.vue'
import { useFinanceStore } from '@/stores/finance'

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

const LEGACY_VIEW_MAP: Record<number, ChartType> = {
  0: 'curve',
  1: 'curve',
  2: 'annual',
  3: 'components',
  4: 'dashboard',
  5: 'season',
  6: 'distribution',
  7: 'tier'
}

const LEGACY_ID_MAP: Record<string, string> = {
  a_shares: 'universe_a_shares',
  etf_macro: 'universe_macro',
  dividend_etf: 'dividend_etf',
  a_shares_long: 'a_shares_long'
}

const quickAssets = [
  { id: 'universe_a_shares', label: 'A股行业' },
  { id: 'universe_macro', label: '宏观大类' },
  { id: 'dividend_etf', label: '红利ETF' },
  { id: 'a_shares_long', label: 'A股27年' }
]

const currentAssetId = ref<string>('universe_a_shares')
const currentChart = ref<ChartType>('curve')
const currentRange = ref<string>(financeStore.defaultRange || '1Y')
const currentCurrency = ref<string>(financeStore.defaultCurrency || 'CNY')
const currentScopeId = ref<string>('')
const currentBenchmarkId = ref<string>('')
const currentSnapshotId = ref<string>('')

const assetDetail = ref<AssetDetailResponse | null>(null)
const detailLoading = ref(false)
const showDropdown = ref(false)
const showConfigModal = ref(false)

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
})

const availableCharts = computed<ChartType[]>(() => {
  const list = assetDetail.value?.available_charts
  if (list && list.length > 0)
    return list
  return CHART_TABS.map(t => t.key)
})

const currentChartTitle = computed(() => {
  return CHART_TABS.find(t => t.key === currentChart.value)?.label || '趋势曲线'
})

const displayAssetName = computed(() => {
  return assetDetail.value?.asset?.name
    || quickAssets.find(a => a.id === currentAssetId.value)?.label
    || currentAssetId.value
})

const displayAsOfDate = computed(() => {
  return assetDetail.value?.snapshot?.as_of || assetDetail.value?.asset?.as_of || '—'
})

const row1Tabs = computed(() => CHART_TABS.filter(t => t.row === 1))
const row2Tabs = computed(() => CHART_TABS.filter(t => t.row === 2))

function isChartAvailable(chart: ChartType): boolean {
  return availableCharts.value.includes(chart)
}

async function fetchAssetDetail() {
  if (!currentAssetId.value)
    return
  detailLoading.value = true
  try {
    const res = await getAssetDetail(currentAssetId.value, {
      scope_id: currentScopeId.value || undefined,
      currency: currentCurrency.value || undefined,
      snapshot_id: currentSnapshotId.value || undefined
    })
    const data = res.data ?? null
    assetDetail.value = data
    if (data) {
      if (data.benchmark_id && !currentBenchmarkId.value)
        currentBenchmarkId.value = data.benchmark_id
      if (data.snapshot?.snapshot_id && !currentSnapshotId.value)
        currentSnapshotId.value = data.snapshot.snapshot_id
      if (data.available_charts?.length && !data.available_charts.includes(currentChart.value)) {
        const firstChart = data.available_charts[0]
        if (firstChart)
          currentChart.value = firstChart
      }
    }
  }
  catch {
    assetDetail.value = null
  }
  finally {
    detailLoading.value = false
  }
}

watch(
  () => [currentAssetId.value, currentCurrency.value, currentScopeId.value],
  () => {
    fetchAssetDetail()
  }
)

onLoad((query: any) => {
  const rawId = query?.asset_id || query?.id
  if (rawId) {
    const decoded = decodeURIComponent(String(rawId))
    currentAssetId.value = LEGACY_ID_MAP[decoded] || decoded
  }
  if (query?.chart) {
    const chartVal = String(query.chart) as ChartType
    if (CHART_TABS.some(t => t.key === chartVal))
      currentChart.value = chartVal
  }
  else if (query?.view !== undefined) {
    const v = Number.parseInt(String(query.view), 10)
    if (!Number.isNaN(v) && LEGACY_VIEW_MAP[v])
      currentChart.value = LEGACY_VIEW_MAP[v]
  }
  if (query?.range)
    currentRange.value = decodeURIComponent(String(query.range))
  if (query?.currency)
    currentCurrency.value = decodeURIComponent(String(query.currency))
  if (query?.scope_id)
    currentScopeId.value = decodeURIComponent(String(query.scope_id))
  if (query?.benchmark_id)
    currentBenchmarkId.value = decodeURIComponent(String(query.benchmark_id))
  if (query?.snapshot_id)
    currentSnapshotId.value = decodeURIComponent(String(query.snapshot_id))

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

function handleSaveFavoriteChart() {
  try {
    financeStore.saveFavoriteChart({
      title: `${displayAssetName.value} · ${currentChartTitle.value}`,
      chart: currentChart.value,
      asset_id: currentAssetId.value,
      asset_name: displayAssetName.value,
      snapshot_mode: currentSnapshotId.value ? 'fixed' : 'latest',
      snapshot_id: currentSnapshotId.value || undefined,
      context: {
        asset_id: currentAssetId.value,
        chart: currentChart.value,
        range: currentRange.value,
        currency: currentCurrency.value,
        benchmark_id: currentBenchmarkId.value,
        snapshot_id: currentSnapshotId.value
      }
    })
    uni.showToast({ title: '已收藏图表配置', icon: 'success' })
  }
  catch {
    uni.showToast({ title: '收藏失败，请稍后重试', icon: 'none' })
  }
}

function handleExportReport() {
  const params = [
    `asset_id=${encodeURIComponent(currentAssetId.value)}`,
    `chart=${encodeURIComponent(currentChart.value)}`,
    `range=${encodeURIComponent(currentRange.value)}`,
    `currency=${encodeURIComponent(currentCurrency.value)}`
  ]
  if (currentSnapshotId.value)
    params.push(`snapshot_id=${encodeURIComponent(currentSnapshotId.value)}`)
  uni.navigateTo({
    url: `/pages-sub/finance/report?${params.join('&')}`,
    fail: () => {
      uni.showToast({ title: '打开报告页失败', icon: 'none' })
    }
  })
}

function handleShareTrigger() {
  const sharePath = `/pages-sub/divergence/index?asset_id=${encodeURIComponent(currentAssetId.value)}&chart=${encodeURIComponent(currentChart.value)}&range=${encodeURIComponent(currentRange.value)}&currency=${encodeURIComponent(currentCurrency.value)}`
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
    path: `/pages-sub/divergence/index?asset_id=${encodeURIComponent(currentAssetId.value)}&chart=${encodeURIComponent(currentChart.value)}&range=${encodeURIComponent(currentRange.value)}&currency=${encodeURIComponent(currentCurrency.value)}`
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

      <!-- 标的快捷切换 + 两行七图 Tab -->
      <view class="px-3 pb-2 space-y-1.5">
        <scroll-view scroll-x class="whitespace-nowrap w-full">
          <view class="inline-flex items-center gap-1.5">
            <text
              v-for="asset in quickAssets"
              :key="asset.id"
              class="px-2 py-0.5 rounded-md text-[11px] font-medium transition-colors"
              :class="currentAssetId === asset.id
                ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] text-white font-bold')
                : (isDark ? 'bg-[#12181F] text-[#8D99A4] border border-[#29313A]' : 'bg-white text-[#515F6A] border border-[#DCDDD8]')"
              @tap="currentAssetId = asset.id"
            >
              {{ asset.label }}
            </text>
          </view>
        </scroll-view>

        <!-- 两行七图切换 -->
        <view class="space-y-1">
          <view class="grid grid-cols-4 gap-1">
            <view
              v-for="tab in row1Tabs"
              :key="tab.key"
              class="py-1.5 rounded-lg text-center text-[11px] font-medium border transition-colors"
              :class="[
                currentChart === tab.key
                  ? (isDark ? 'bg-[#E5B352]/20 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/15 border-[#8A641F] text-[#8A641F] font-bold')
                  : (isDark ? 'bg-[#12181F] border-[#29313A] text-[#8D99A4]' : 'bg-white border-[#DCDDD8] text-[#515F6A]'),
                !isChartAvailable(tab.key) ? 'opacity-40' : ''
              ]"
              @tap="handleSelectChart(tab.key)"
            >
              {{ tab.shortLabel }}
            </view>
          </view>
          <view class="grid grid-cols-3 gap-1">
            <view
              v-for="tab in row2Tabs"
              :key="tab.key"
              class="py-1.5 rounded-lg text-center text-[11px] font-medium border transition-colors"
              :class="[
                currentChart === tab.key
                  ? (isDark ? 'bg-[#E5B352]/20 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/15 border-[#8A641F] text-[#8A641F] font-bold')
                  : (isDark ? 'bg-[#12181F] border-[#29313A] text-[#8D99A4]' : 'bg-white border-[#DCDDD8] text-[#515F6A]'),
                !isChartAvailable(tab.key) ? 'opacity-40' : ''
              ]"
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
      <TrendCurveView
        v-if="currentChart === 'curve'"
        :asset-id="currentAssetId"
        :snapshot-id="currentSnapshotId"
        :range="currentRange"
        :currency="currentCurrency"
        :scope-id="currentScopeId"
        :benchmark-id="currentBenchmarkId"
      />
      <AnnualTrendView
        v-else-if="currentChart === 'annual'"
        :asset-id="currentAssetId"
        :snapshot-id="currentSnapshotId"
        :range="currentRange"
        :currency="currentCurrency"
        :scope-id="currentScopeId"
        :benchmark-id="currentBenchmarkId"
      />
      <ComponentTrendView
        v-else-if="currentChart === 'components'"
        :asset-id="currentAssetId"
        :snapshot-id="currentSnapshotId"
        :range="currentRange"
        :currency="currentCurrency"
        :scope-id="currentScopeId"
        :benchmark-id="currentBenchmarkId"
      />
      <TrendDashboardView
        v-else-if="currentChart === 'dashboard'"
        :asset-id="currentAssetId"
        :snapshot-id="currentSnapshotId"
        :range="currentRange"
        :currency="currentCurrency"
        :scope-id="currentScopeId"
        :benchmark-id="currentBenchmarkId"
      />
      <SolarTermsView
        v-else-if="currentChart === 'season'"
        :asset-id="currentAssetId"
        :snapshot-id="currentSnapshotId"
        :range="currentRange"
        :currency="currentCurrency"
        :scope-id="currentScopeId"
        :benchmark-id="currentBenchmarkId"
      />
      <RightDistributionView
        v-else-if="currentChart === 'distribution'"
        :asset-id="currentAssetId"
        :snapshot-id="currentSnapshotId"
        :range="currentRange"
        :currency="currentCurrency"
        :scope-id="currentScopeId"
        :benchmark-id="currentBenchmarkId"
      />
      <TrendTierView
        v-else-if="currentChart === 'tier'"
        :asset-id="currentAssetId"
        :snapshot-id="currentSnapshotId"
        :range="currentRange"
        :currency="currentCurrency"
        :scope-id="currentScopeId"
        :benchmark-id="currentBenchmarkId"
      />
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
          @tap="handleSaveFavoriteChart"
        >
          <text class="text-[12px] font-medium" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
            ★ 收藏图表
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
              :class="currentRange === r
                ? (isDark ? 'bg-[#E5B352] border-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] border-[#8A641F] text-white font-bold')
                : (isDark ? 'border-[#29313A] text-[#8D99A4]' : 'border-[#DCDDD8] text-[#515F6A]')"
              @tap="currentRange = r"
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
              :class="currentCurrency === c
                ? (isDark ? 'bg-[#E5B352] border-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#8A641F] border-[#8A641F] text-white font-bold')
                : (isDark ? 'border-[#29313A] text-[#8D99A4]' : 'border-[#DCDDD8] text-[#515F6A]')"
              @tap="currentCurrency = c"
            >
              {{ c }}
            </text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>
