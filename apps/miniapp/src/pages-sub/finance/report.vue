<script lang="ts" setup>
import type {
  AnnualTrendResponse,
  AssetDetailResponse,
  ChartType,
  ComponentTrendResponse,
  PeriodMatrixResponse,
  TrendCurveResponse,
  TrendDashboardResponse,
  TrendDistributionResponse,
  TrendSeasonResponse,
  TrendTierResponse
} from '@/api/types/trend'
import type {ReportDataBundle} from '@/utils/reportRegistry';
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { computed, getCurrentInstance, onUnmounted, ref } from 'vue'
import {
  getAssetDetail,
  getPeriodMatrix,
  getTrendAnnual,
  getTrendComponents,
  getTrendCurve,
  getTrendDashboard,
  getTrendDistribution,
  getTrendSeason,
  getTrendTier
} from '@/api/trend'
import CSafeArea from '@/components/ui/CSafeArea.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, isExpiredSnapshotError, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceSharePath, decodeFinanceQuery, toApiCurrency, toDisplayCurrency } from '@/utils/financeRoute'
import {
  buildReportPages,
  buildReportTitle,
  canExportReportBundle,
  drawReportPageToCanvas,
  getReportExporterMeta,
  
  resolveReportAsOf,
  resolveReportAssetName,
  resolveReportSnapshotId
} from '@/utils/reportRegistry'
import { getTrendTempMeta } from '@/utils/trendDict'

definePage({
  style: {
    navigationBarTitleText: '研究报告',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    navigationStyle: 'custom'
  }
})

const financeStore = useFinanceStore()
const instance = getCurrentInstance()

const assetId = ref('universe_macro')
const scopeId = ref('universe_macro')
const chartType = ref<ChartType>('curve')
const rangeParam = ref('1Y')
const currencyParam = ref('CNY')
const benchmarkIdParam = ref('')
const snapshotMode = ref<'latest' | 'fixed'>('latest')
const snapshotIdParam = ref('')
const startDateParam = ref('')
const endDateParam = ref('')
const sortParam = ref('')
const orderParam = ref<'asc' | 'desc' | ''>('')
const marketParam = ref('')
const assetTypeParam = ref('')
const statusParam = ref('')
const reportTheme = ref<'light' | 'dark'>(financeStore.reportThemePref)

const loading = ref(false)
const exporting = ref(false)
const errorMsg = ref('')
const isExpiredSnapshot = ref(false)
const assetDetail = ref<AssetDetailResponse | null>(null)
const curveData = ref<TrendCurveResponse | null>(null)
const matrixData = ref<PeriodMatrixResponse | null>(null)
const tierData = ref<TrendTierResponse | null>(null)
const seasonData = ref<TrendSeasonResponse | null>(null)
const annualData = ref<AnnualTrendResponse | null>(null)
const componentsData = ref<ComponentTrendResponse | null>(null)
const dashboardData = ref<TrendDashboardResponse | null>(null)
const distributionData = ref<TrendDistributionResponse | null>(null)
const exportedPages = ref<string[]>([])
const activePreviewPage = ref(1)
const reportRace = createAsyncRace()

onUnmounted(() => {
  reportRace.invalidate()
})

const exporterMeta = computed(() => getReportExporterMeta(chartType.value))

const reportBundle = computed<ReportDataBundle>(() => ({
  chart: chartType.value,
  assetId: assetId.value,
  scopeId: scopeId.value,
  currency: toApiCurrency(currencyParam.value),
  range: rangeParam.value,
  detail: assetDetail.value,
  curve: curveData.value,
  matrix: matrixData.value,
  tier: tierData.value,
  season: seasonData.value,
  annual: annualData.value,
  components: componentsData.value,
  dashboard: dashboardData.value,
  distribution: distributionData.value
}))

const reportPages = computed(() => buildReportPages(reportBundle.value))
const exportCheck = computed(() => {
  if (loading.value)
    return { ok: false, reason: '数据加载中，请稍候' }
  if (isExpiredSnapshot.value)
    return { ok: false, reason: '快照已过期，请切回最新快照后重试' }
  if (errorMsg.value)
    return { ok: false, reason: errorMsg.value }
  return canExportReportBundle(reportBundle.value)
})

const displayAssetName = computed(() => resolveReportAssetName(reportBundle.value))
const displayReportTitle = computed(() => buildReportTitle(reportBundle.value))
const displayAsOf = computed(() => resolveReportAsOf(reportBundle.value))
const displayAlgo = computed(() => assetDetail.value?.snapshot?.algorithm_version || matrixData.value?.snapshot?.algorithm_version || 'trend-v1.0.0')
const resolvedSnapId = computed(() => resolveReportSnapshotId(reportBundle.value) || snapshotIdParam.value)
const displayLatestValue = computed(() => {
  const v = assetDetail.value?.asset?.latest_value
  return v !== null && v !== undefined
? v.toFixed(2)
: '—'
})
const displayTempMeta = computed(() => getTrendTempMeta(assetDetail.value?.asset?.metrics?.temperature))
const displayStrength = computed(() => assetDetail.value?.asset?.metrics?.strength?.score ?? '—')
const displayReturnPct = computed(() => {
  const pts = curveData.value?.points || []
  for (let i = pts.length - 1; i >= 0; i--) {
    const pt = pts[i]
    if (pt && pt.return_pct !== null && pt.return_pct !== undefined)
      return pt.return_pct
  }
  return null
})
const displayBenchmarkReturnPct = computed(() => {
  const pts = curveData.value?.benchmark_points || []
  for (let i = pts.length - 1; i >= 0; i--) {
    const pt = pts[i]
    if (pt && pt.return_pct !== null && pt.return_pct !== undefined)
      return pt.return_pct
  }
  return null
})

async function loadReportData() {
  if (!assetId.value)
    return
  const token = reportRace.next()
  loading.value = true
  errorMsg.value = ''
  isExpiredSnapshot.value = false
  exportedPages.value = []
  assetDetail.value = null
  curveData.value = null
  matrixData.value = null
  tierData.value = null
  seasonData.value = null
  annualData.value = null
  componentsData.value = null
  dashboardData.value = null
  distributionData.value = null

  const apiCurrency = toApiCurrency(currencyParam.value)
  const reqSnap = snapshotMode.value === 'fixed'
? snapshotIdParam.value
: ''

  try {
    if (chartType.value === 'matrix') {
      const mRes = await getPeriodMatrix(scopeId.value || assetId.value, {
        currency: apiCurrency,
        snapshot_id: reqSnap || undefined,
        sort: sortParam.value || undefined,
        order: orderParam.value || undefined,
        market: marketParam.value || undefined,
        asset_type: assetTypeParam.value || undefined,
        status: statusParam.value || undefined,
        limit: 100
      })
      if (!reportRace.isLatest(token))
        return
      matrixData.value = unwrapApiResponse(mRes, '周期矩阵数据加载失败')
      activePreviewPage.value = 1
      return
    }

    const detailRes = await getAssetDetail(assetId.value, {
      scope_id: scopeId.value || undefined,
      currency: apiCurrency,
      snapshot_id: reqSnap || undefined
    })
    if (!reportRace.isLatest(token))
      return
    const detail = unwrapApiResponse(detailRes, '资产详情加载失败')
    assetDetail.value = detail
    const boundSnap = reqSnap || detail.snapshot?.snapshot_id || ''

    if (chartType.value === 'curve') {
      const cRes = await getTrendCurve(assetId.value, {
        range: rangeParam.value,
        currency: apiCurrency,
        benchmark_id: benchmarkIdParam.value || detail.benchmark_id || undefined,
        snapshot_id: boundSnap || undefined,
        start_date: startDateParam.value || undefined,
        end_date: endDateParam.value || undefined
      })
      if (!reportRace.isLatest(token))
        return
      curveData.value = unwrapApiResponse(cRes, '趋势曲线加载失败')
    }
    else if (chartType.value === 'tier') {
      const tRes = await getTrendTier(scopeId.value || assetId.value, {
        scope_id: scopeId.value || undefined,
        snapshot_id: boundSnap || undefined
      })
      if (!reportRace.isLatest(token))
        return
      tierData.value = unwrapApiResponse(tRes, '梯队榜加载失败')
    }
    else if (chartType.value === 'season') {
      const sRes = await getTrendSeason(scopeId.value || assetId.value, {
        snapshot_id: boundSnap || undefined
      })
      if (!reportRace.isLatest(token))
        return
      seasonData.value = unwrapApiResponse(sRes, '趋势节气图加载失败')
    }
    else if (chartType.value === 'annual') {
      const aRes = await getTrendAnnual(assetId.value, {
        currency: apiCurrency,
        snapshot_id: boundSnap || undefined
      })
      if (reportRace.isLatest(token) && aRes?.code === 0)
        annualData.value = aRes.data ?? null
    }
    else if (chartType.value === 'components') {
      const cpRes = await getTrendComponents(scopeId.value || assetId.value, {
        scope_id: scopeId.value || undefined,
        range: rangeParam.value,
        currency: apiCurrency,
        snapshot_id: boundSnap || undefined
      })
      if (reportRace.isLatest(token) && cpRes?.code === 0)
        componentsData.value = cpRes.data ?? null
    }
    else if (chartType.value === 'dashboard') {
      const dRes = await getTrendDashboard(scopeId.value || assetId.value, {
        scope_id: scopeId.value || undefined,
        currency: apiCurrency,
        snapshot_id: boundSnap || undefined
      })
      if (reportRace.isLatest(token) && dRes?.code === 0)
        dashboardData.value = dRes.data ?? null
    }
    else if (chartType.value === 'distribution') {
      const distRes = await getTrendDistribution(scopeId.value || assetId.value, {
        snapshot_id: boundSnap || undefined
      })
      if (reportRace.isLatest(token) && distRes?.code === 0)
        distributionData.value = distRes.data ?? null
    }
  }
  catch (err: any) {
    if (!reportRace.isLatest(token))
      return
    isExpiredSnapshot.value = isExpiredSnapshotError(err)
    errorMsg.value = err?.message || '报告快照加载失败'
  }
  finally {
    if (reportRace.isLatest(token))
      loading.value = false
  }
}

onLoad((query: any) => {
  const parsed = decodeFinanceQuery(query, undefined, {
    defaultHomePool: financeStore.defaultHomePool,
    defaultCurrency: financeStore.defaultCurrency,
    defaultRange: financeStore.defaultRange,
    reportThemePref: financeStore.reportThemePref
  })
  assetId.value = parsed.asset_id
  scopeId.value = parsed.scope_id
  chartType.value = parsed.chart
  rangeParam.value = parsed.range
  currencyParam.value = toDisplayCurrency(parsed.currency)
  benchmarkIdParam.value = parsed.benchmark_id
  snapshotMode.value = parsed.snapshot_mode
  snapshotIdParam.value = parsed.snapshot_id
  startDateParam.value = parsed.start_date || ''
  endDateParam.value = parsed.end_date || ''
  sortParam.value = parsed.sort || ''
  orderParam.value = parsed.order || ''
  marketParam.value = parsed.market || ''
  assetTypeParam.value = parsed.asset_type || ''
  statusParam.value = parsed.status || ''
  if (parsed.theme)
    reportTheme.value = parsed.theme
  loadReportData()
})

onShareAppMessage(() => {
  const fixedId = resolvedSnapId.value
  const path = buildFinanceSharePath('/pages-sub/finance/report', {
    asset_id: assetId.value,
    scope_id: scopeId.value,
    chart: chartType.value,
    range: rangeParam.value as any,
    currency: currencyParam.value as any,
    benchmark_id: benchmarkIdParam.value,
    snapshot_mode: fixedId ? 'fixed' : snapshotMode.value,
    snapshot_id: fixedId,
    start_date: startDateParam.value,
    end_date: endDateParam.value,
    sort: sortParam.value,
    order: orderParam.value || undefined,
    market: marketParam.value,
    asset_type: assetTypeParam.value,
    status: statusParam.value,
    theme: reportTheme.value
  }, fixedId || undefined)

  return {
    title: displayReportTitle.value,
    path
  }
})

function switchToLatestAndReload() {
  snapshotMode.value = 'latest'
  snapshotIdParam.value = ''
  loadReportData()
}

function handleBack() {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function setReportTheme(theme: 'light' | 'dark') {
  reportTheme.value = theme
  financeStore.reportThemePref = theme
  exportedPages.value = []
}

function formatPct(val: number | null | undefined): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  const sign = val > 0
    ? '+'
    : ''
  return `${sign}${val.toFixed(1)}%`
}

function getPctColor(val: number | null | undefined): string {
  const isLight = reportTheme.value === 'light'
  if (val === null || val === undefined || !Number.isFinite(val) || val === 0)
    return isLight ? '#626F77' : '#8D99A4'
  const isUp = val > 0
  const cnRedUp = financeStore.colorScheme === 'cn_red_up'
  const useRise = cnRedUp ? isUp : !isUp
  if (useRise)
    return isLight ? '#B93842' : '#F05A50'
  return isLight ? '#14745E' : '#2AD1A3'
}

function getTempBadgeStyle(temp?: string | null) {
  const meta = getTrendTempMeta(temp)
  const isLight = reportTheme.value === 'light'
  return {
    color: isLight ? meta.lightColor : meta.color,
    backgroundColor: meta.bg,
    borderColor: meta.border
  }
}

function renderSinglePage(pageSpec: ReturnType<typeof buildReportPages>[number]): Promise<string> {
  return new Promise((resolve, reject) => {
    const width = 1080
    const height = 1440
    // #ifdef H5
    if (typeof document !== 'undefined') {
      const h5Canvas = document.createElement('canvas')
      h5Canvas.width = width
      h5Canvas.height = height
      const ctx = h5Canvas.getContext('2d')
      if (!ctx) {
        reject(new Error('当前环境不支持 Canvas 2D 高清导出'))
        return
      }
      drawReportPageToCanvas(
        ctx,
        width,
        height,
        reportBundle.value,
        pageSpec,
        reportTheme.value,
        financeStore.colorScheme === 'cn_red_up'
      )
      resolve(h5Canvas.toDataURL('image/png'))
      return
    }
    // #endif
    const query = uni.createSelectorQuery().in(instance?.proxy)
    query.select('#reportCanvas').fields({ node: true, size: true }, () => {}).exec((res) => {
      const node = res?.[0]?.node
      if (!node) {
        reject(new Error('当前环境不支持 Canvas 2D 高清导出'))
        return
      }
      node.width = width
      node.height = height
      const ctx = node.getContext('2d')
      drawReportPageToCanvas(
        ctx,
        width,
        height,
        reportBundle.value,
        pageSpec,
        reportTheme.value,
        financeStore.colorScheme === 'cn_red_up'
      )
      uni.canvasToTempFilePath({
        canvasId: 'reportCanvas',
        canvas: node,
        destWidth: width,
        destHeight: height,
        fileType: 'png',
        quality: 1,
        success: out => resolve(out.tempFilePath),
        fail: err => reject(err)
      } as UniApp.CanvasToTempFilePathOptions)
    })
  })
}

async function ensureRenderedPages(): Promise<string[]> {
  if (!exportCheck.value.ok)
    throw new Error(exportCheck.value.reason || '当前报告数据未就绪')
  if (exportedPages.value.length === reportPages.value.length && exportedPages.value.length > 0)
    return exportedPages.value

  const paths: string[] = []
  for (const page of reportPages.value) {
    const path = await renderSinglePage(page)
    paths.push(path)
  }
  exportedPages.value = paths
  return paths
}

async function handlePreviewImage() {
  if (exporting.value)
    return
  if (!exportCheck.value.ok) {
    uni.showToast({ title: exportCheck.value.reason, icon: 'none' })
    return
  }
  exporting.value = true
  uni.showLoading({ title: '生成高清图...' })
  try {
    const urls = await ensureRenderedPages()
    uni.hideLoading()
    uni.previewImage({
      urls,
      current: urls[Math.max(0, activePreviewPage.value - 1)] || urls[0]
    })
  }
  catch (err: any) {
    uni.hideLoading()
    uni.showToast({ title: err?.message || '生成失败', icon: 'none' })
  }
  finally {
    exporting.value = false
  }
}

function saveSingleImageToAlbum(filePath: string, index = 0): Promise<void> {
  // #ifdef H5
  return new Promise((resolve, reject) => {
    try {
      const link = document.createElement('a')
      link.href = filePath
      link.download = `cyclecompass-report-${index + 1}.png`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      resolve()
    }
    catch (err) {
      reject(err)
    }
  })
  // #endif
  // #ifndef H5
  return new Promise((resolve, reject) => {
    uni.saveImageToPhotosAlbum({
      filePath,
      success: () => resolve(),
      fail: err => reject(err)
    })
  })
  // #endif
}

async function handleSaveToAlbum() {
  if (exporting.value)
    return
  if (!exportCheck.value.ok) {
    uni.showToast({ title: exportCheck.value.reason, icon: 'none' })
    return
  }
  exporting.value = true
  uni.showLoading({ title: '正在导出 1080px 高清图...' })
  try {
    const paths = await ensureRenderedPages()
    for (let i = 0; i < paths.length; i++)
      await saveSingleImageToAlbum(paths[i]!, i)
    uni.hideLoading()
    let successTitle = paths.length > 1 ? `已保存全部 ${paths.length} 页到相册` : '已保存到相册'
    // #ifdef H5
    successTitle = paths.length > 1 ? `已导出全部 ${paths.length} 页高清报告` : '已导出高清报告图'
    // #endif
    uni.showToast({
      title: successTitle,
      icon: 'success'
    })
  }
  catch (err: any) {
    uni.hideLoading()
    const msg = String(err?.errMsg || err?.message || '')
    if (msg.includes('auth deny') || msg.includes('authorize')) {
      uni.showModal({
        title: '需要相册权限',
        content: '请在设置中允许保存图片到相册，以便导出高清研究报告。',
        confirmText: '去设置',
        success: (modalRes) => {
          if (modalRes.confirm)
            uni.openSetting({})
        }
      })
    }
    else if (msg.includes('cancel')) {
      uni.showToast({ title: '已取消保存', icon: 'none' })
    }
    else {
      uni.showToast({ title: msg || '保存失败，请重试', icon: 'none' })
    }
  }
  finally {
    exporting.value = false
  }
}
function handleShareTap() {
  // #ifdef H5
  const shareUrl = typeof window !== 'undefined' ? window.location.href : ''
  if (shareUrl) {
    uni.setClipboardData({
      data: shareUrl,
      success: () => {
        uni.showToast({ title: '已复制报告链接，可直接粘贴分享', icon: 'none' })
      }
    })
  }
  // #endif
}
</script>

<template>
  <view
    class="min-h-screen box-border flex flex-col"
    :class="{ 'cc-finance-light': !financeStore.isDarkEffective }"
    style="background-color: var(--cc-bg-page); color: var(--cc-text-primary);"
  >
    <CSafeArea position="top" />

    <!-- 顶部导航 -->
    <view
      class="flex items-center justify-between px-3.5 py-2.5 border-b"
      style="border-color: var(--cc-border);"
    >
      <view
        class="min-w-[44px] min-h-[44px] flex items-center gap-2 cursor-pointer"
        @tap="handleBack"
      >
        <text class="text-lg font-bold cc-brand">‹</text>
        <text class="text-sm font-semibold cc-text-primary">研究报告</text>
      </view>
    </view>

    <view class="px-3.5 pt-3.5 pb-6 flex-1">
      <!-- 标题栏 -->
      <view class="flex items-center justify-between mb-1">
        <text class="text-[22px] font-bold tracking-tight cc-text-primary">分享预览</text>
        <view
          class="px-2.5 py-1 rounded-md border text-xs font-medium cc-brand"
          style="border-color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);"
        >
          高清报告
        </view>
      </view>
      <text class="text-xs cc-text-secondary block mb-3">
        {{ displayAssetName }} · {{ exporterMeta.label }} / {{ displayAsOf }}
      </text>

      <!-- 报告预览画布卡片（独立于页面主题，跟随 reportTheme） -->
      <view
        class="p-4 rounded-2xl border mb-4 transition-colors"
        :style="reportTheme === 'light'
          ? 'background-color: #F4F3EF; border-color: #DCDDD8; color: #22292E;'
          : 'background-color: #12161B; border-color: #29313A; color: #F1EFEA;'"
      >
        <view class="flex items-center justify-between mb-3">
          <text
            class="text-xs font-bold"
            :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }"
          >
            ◎ 周期罗盘
          </text>
          <text
            class="text-[11px]"
            :style="{ color: reportTheme === 'light' ? '#626F77' : '#8D99A4' }"
          >
            研究报告 / {{ exporterMeta.label }}
          </text>
        </view>

        <text class="text-lg font-bold block mb-1">
          {{ displayReportTitle }}
        </text>
        <text
          class="text-[11px] block mb-3"
          :style="{ color: reportTheme === 'light' ? '#515F6A' : '#A9B3BD' }"
        >
          {{ currencyParam }} · {{ displayAsOf }} · 算法 {{ displayAlgo }}
        </text>

        <view v-if="loading" class="py-8 text-center text-xs opacity-70">
          正在加载报告快照...
        </view>

        <view v-else-if="errorMsg || !exportCheck.ok" class="py-6 text-center space-y-2">
          <text class="text-xs font-bold text-[#F05A50] block">
            {{ isExpiredSnapshot ? '请求的快照已过期' : '当前快照暂不可生成报告' }}
          </text>
          <text class="text-[11px] opacity-80 block">
            {{ exportCheck.reason }}
          </text>
          <view class="flex items-center justify-center gap-2 pt-1">
            <text
              v-if="isExpiredSnapshot"
              class="px-3 py-1 rounded-lg text-xs font-bold"
              style="background-color: var(--cc-brand-primary); color: #090C0F;"
              @tap="switchToLatestAndReload"
            >
              切回最新快照
            </text>
            <text
              class="px-3 py-1 rounded-lg text-xs border"
              style="border-color: var(--cc-border);"
              @tap="loadReportData"
            >
              重试
            </text>
          </view>
        </view>

        <!-- 周期矩阵预览 -->
        <template v-else-if="chartType === 'matrix' && matrixData">
          <view class="space-y-2 mb-3">
            <view
              v-for="(slice, sIdx) in (reportPages[activePreviewPage - 1]?.matrixSlices || [])"
              :key="`${slice.groupName}-${sIdx}`"
              class="rounded-xl border overflow-hidden"
              :style="reportTheme === 'light'
                ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
                : 'background-color: #181E25; border-color: #29313A;'"
            >
              <view class="px-2.5 py-1.5 text-[11px] font-bold" :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }">
                {{ slice.groupName }}{{ slice.isContinuation ? '（续）' : '' }}
              </view>
              <view
                v-for="row in slice.rows.slice(0, 6)"
                :key="row.asset.asset_id"
                class="px-2.5 py-1.5 flex items-center justify-between text-[11px] border-t"
                :style="{ borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }"
              >
                <text class="font-medium truncate max-w-[110px]">{{ row.asset.short_name || row.asset.name }}</text>
                <view class="flex items-center gap-2 font-mono cc-tabular">
                  <view v-for="col in (matrixData.columns || []).slice(0, 3)" :key="col.key" class="flex items-center gap-0.5">
                    <text class="opacity-65">{{ col.label }}:</text>
                    <text
                      class="font-bold"
                      :style="{ color: getPctColor(row.cells[col.key]?.return_pct ?? row.cells[col.key]?.bp_change) }"
                    >
                      {{ row.cells[col.key]?.return_pct !== null && row.cells[col.key]?.return_pct !== undefined
                        ? `${row.cells[col.key]!.return_pct! > 0 ? '+' : ''}${row.cells[col.key]!.return_pct!.toFixed(1)}%`
                        : (row.cells[col.key]?.bp_change !== null && row.cells[col.key]?.bp_change !== undefined
                          ? `${row.cells[col.key]!.bp_change! > 0 ? '+' : ''}${row.cells[col.key]!.bp_change!.toFixed(1)}bp`
                          : '—') }}
                    </text>
                  </view>
                </view>
              </view>
            </view>

            <view v-if="reportPages.length > 1" class="flex items-center justify-center gap-3 pt-1">
              <text
                class="px-2.5 py-1 rounded border text-[11px]"
                :style="{ opacity: activePreviewPage > 1 ? 1 : 0.4, borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }"
                @tap="activePreviewPage > 1 && (activePreviewPage -= 1)"
              >
                ‹ 上一页
              </text>
              <text class="text-[11px] font-mono cc-tabular">{{ activePreviewPage }} / {{ reportPages.length }}</text>
              <text
                class="px-2.5 py-1 rounded border text-[11px]"
                :style="{ opacity: activePreviewPage < reportPages.length ? 1 : 0.4, borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }"
                @tap="activePreviewPage < reportPages.length && (activePreviewPage += 1)"
              >
                下一页 ›
              </text>
            </view>
          </view>
        </template>

        <!-- 梯队 / 节气预览 -->
        <template v-else-if="chartType === 'tier' && tierData">
          <view class="space-y-1.5 mb-3">
            <view
              v-for="b in tierData.tiers"
              :key="b.tier"
              class="px-2.5 py-2 rounded-lg border flex items-center justify-between text-[11px]"
              :style="reportTheme === 'light'
                ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
                : 'background-color: #181E25; border-color: #29313A;'"
            >
              <text class="font-bold">{{ b.pro_label }}</text>
              <text class="opacity-80 truncate max-w-[180px]">
                {{ b.items.map(i => i.name).slice(0, 4).join('、') || '暂无' }}
              </text>
            </view>
          </view>
        </template>

        <template v-else-if="chartType === 'season' && seasonData">
          <view class="space-y-1.5 mb-3">
            <view
              v-for="s in seasonData.stages.filter(st => st.count > 0)"
              :key="s.stage_key"
              class="px-2.5 py-2 rounded-lg border flex items-center justify-between text-[11px]"
              :style="reportTheme === 'light'
                ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
                : 'background-color: #181E25; border-color: #29313A;'"
            >
              <text class="font-bold">{{ s.stage_label }} ({{ s.count }})</text>
              <text class="opacity-80 truncate max-w-[180px]">
                {{ s.members.map(m => m.name).slice(0, 4).join('、') }}
              </text>
            </view>
          </view>
        </template>

        <!-- 品种成分趋势预览 -->
        <template v-else-if="chartType === 'components' && componentsData">
          <view class="space-y-1.5 mb-3">
            <view
              v-for="(s, idx) in componentsData.series.slice(0, 8)"
              :key="s.asset_id"
              class="px-2.5 py-1.5 rounded-lg border flex items-center justify-between text-[11px]"
              :style="reportTheme === 'light'
                ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
                : 'background-color: #181E25; border-color: #29313A;'"
            >
              <text class="font-medium truncate max-w-[140px]">{{ idx + 1 }}. {{ s.name }}</text>
              <view class="flex items-center gap-2 font-mono cc-tabular">
                <text
                  class="px-1.5 py-0.5 rounded border text-[10px] font-bold leading-none"
                  :style="getTempBadgeStyle(s.temperature)"
                >
                  {{ getTrendTempMeta(s.temperature).label }}
                </text>
                <text class="font-bold min-w-[48px] text-right" :style="{ color: getPctColor(s.total_return_pct) }">
                  {{ formatPct(s.total_return_pct) }}
                </text>
              </view>
            </view>
          </view>
        </template>

        <!-- 趋势仪表盘预览 -->
        <template v-else-if="chartType === 'dashboard' && dashboardData">
          <view class="space-y-2 mb-3">
            <view
              v-for="g in dashboardData.groups.slice(0, 3)"
              :key="g.group_name"
              class="rounded-xl border overflow-hidden"
              :style="reportTheme === 'light'
                ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
                : 'background-color: #181E25; border-color: #29313A;'"
            >
              <view class="px-2.5 py-1 text-[11px] font-bold" :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }">
                {{ g.group_name }}
              </view>
              <view
                v-for="r in g.rows.slice(0, 4)"
                :key="r.asset.asset_id"
                class="px-2.5 py-1.5 flex items-center justify-between text-[11px] border-t"
                :style="{ borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }"
              >
                <text class="font-medium truncate max-w-[140px]">{{ r.asset.short_name || r.asset.name }}</text>
                <view class="flex items-center gap-2 font-mono cc-tabular">
                  <text
                    class="px-1.5 py-0.5 rounded border text-[10px] font-bold leading-none"
                    :style="getTempBadgeStyle(r.asset.metrics?.temperature)"
                  >
                    {{ getTrendTempMeta(r.asset.metrics?.temperature).label }}
                  </text>
                  <text class="font-bold min-w-[48px] text-right" :style="{ color: getPctColor(r.change_pct) }">
                    {{ formatPct(r.change_pct) }}
                  </text>
                </view>
              </view>
            </view>
          </view>
        </template>

        <!-- 右侧分布预览 -->
        <template v-else-if="chartType === 'distribution' && distributionData">
          <view
            class="grid grid-cols-3 gap-2 p-3 rounded-xl border mb-3"
            :style="reportTheme === 'light'
              ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
              : 'background-color: #181E25; border-color: #29313A;'"
          >
            <view>
              <text class="text-[10px] opacity-70 block">右侧确认</text>
              <text class="text-base font-bold cc-tabular block">{{ distributionData.right_count }} / {{ distributionData.valid_count }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">右侧占比</text>
              <text class="text-base font-bold cc-tabular block" :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }">
                {{ formatPct(distributionData.right_ratio_pct) }}
              </text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">样本数</text>
              <text class="text-base font-bold cc-tabular block">{{ distributionData.nodes.length }}</text>
            </view>
          </view>
        </template>

        <!-- 年度走势预览 -->
        <template v-else-if="chartType === 'annual' && annualData">
          <view
            class="grid grid-cols-3 gap-2 p-3 rounded-xl border mb-3"
            :style="reportTheme === 'light'
              ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
              : 'background-color: #181E25; border-color: #29313A;'"
          >
            <view>
              <text class="text-[10px] opacity-70 block">今年以来</text>
              <text class="text-base font-bold cc-tabular block" :style="{ color: getPctColor(annualData.stats.current_year_return_pct) }">
                {{ formatPct(annualData.stats.current_year_return_pct) }}
              </text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">同期中位数</text>
              <text class="text-base font-bold cc-tabular block" :style="{ color: getPctColor(annualData.stats.same_week_median_pct) }">
                {{ formatPct(annualData.stats.same_week_median_pct) }}
              </text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">上涨年份比</text>
              <text class="text-base font-bold cc-tabular block" :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }">
                {{ formatPct(annualData.stats.up_ratio_pct) }}
              </text>
            </view>
          </view>
        </template>

        <template v-else>
          <!-- 3栏 KPI -->
          <view
            class="grid grid-cols-3 gap-2 p-3 rounded-xl border mb-3"
            :style="reportTheme === 'light'
              ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
              : 'background-color: #181E25; border-color: #29313A;'"
          >
            <view>
              <text class="text-[10px] opacity-70 block">最新值</text>
              <text
                class="text-lg font-bold cc-tabular block"
                :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }"
              >
                {{ displayLatestValue }}
              </text>
              <text class="text-[10px] opacity-60 block">{{ currencyParam }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">趋势温度</text>
              <text
                class="text-lg font-bold block"
                :style="{ color: reportTheme === 'light' ? displayTempMeta.lightColor : displayTempMeta.color }"
              >
                {{ displayTempMeta.label }}
              </text>
              <text class="text-[10px] opacity-60 block">{{ displayTempMeta.shortDesc }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">相对强度</text>
              <text
                class="text-lg font-bold cc-tabular block"
                :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }"
              >
                {{ displayStrength }}
              </text>
              <text class="text-[10px] opacity-60 block">百分位</text>
            </view>
          </view>

          <!-- 底部极值 -->
          <view class="grid grid-cols-4 gap-2 mb-3 pt-2 border-t" :style="{ borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }">
            <view>
              <text class="text-[10px] opacity-70 block">累计收益</text>
              <text class="text-xs font-bold cc-tabular" :style="{ color: getPctColor(displayReturnPct) }">{{ formatPct(displayReturnPct) }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">比较基准</text>
              <text class="text-xs font-bold cc-tabular" :style="{ color: getPctColor(displayBenchmarkReturnPct) }">{{ formatPct(displayBenchmarkReturnPct) }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">最大回撤</text>
              <text class="text-xs font-bold cc-tabular" :style="{ color: getPctColor(curveData?.extremum?.max_drawdown_pct) }">{{ formatPct(curveData?.extremum?.max_drawdown_pct) }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">最大上涨</text>
              <text class="text-xs font-bold cc-tabular" :style="{ color: getPctColor(curveData?.extremum?.max_gain_pct) }">{{ formatPct(curveData?.extremum?.max_gain_pct) }}</text>
            </view>
          </view>
        </template>

        <view
          class="pt-2.5 border-t flex items-center justify-between text-[10px] opacity-70"
          :style="{ borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }"
        >
          <text>周期罗盘 CycleCompass · 历史表现不代表未来</text>
          <text>1080px 高清</text>
        </view>
      </view>

      <!-- 报告主题与规格切换 -->
      <view class="flex items-center gap-2 mb-3 flex-wrap">
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium border cursor-pointer"
          :style="reportTheme === 'light'
            ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);'
            : 'border-color: var(--cc-border); color: var(--cc-text-secondary); background-color: var(--cc-bg-card);'"
          @tap="setReportTheme('light')"
        >
          浅色报告
        </view>
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium border cursor-pointer"
          :style="reportTheme === 'dark'
            ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);'
            : 'border-color: var(--cc-border); color: var(--cc-text-secondary); background-color: var(--cc-bg-card);'"
          @tap="setReportTheme('dark')"
        >
          深色报告
        </view>
        <view class="px-3 py-1.5 rounded-lg text-xs cc-elevated cc-text-muted">
          1080px 高清
        </view>
      </view>

      <text class="text-[11px] cc-text-muted block mb-3">
        图片由 Canvas 2D 独立渲染，非页面截图；导出后可直接保存相册。
      </text>

      <!-- 操作按钮（预览与保存共用同一渲染结果；不可导出态禁用） -->
      <view class="grid grid-cols-2 gap-2.5 mb-3">
        <view
          class="py-3 rounded-xl cc-panel text-center text-xs font-semibold cc-text-primary cursor-pointer"
          :class="!exportCheck.ok ? 'opacity-45' : ''"
          @tap="handlePreviewImage"
        >
          {{ exporterMeta.previewButtonLabel }}
        </view>
        <view
          class="py-3 rounded-xl text-center text-xs font-semibold cursor-pointer"
          :class="!exportCheck.ok ? 'opacity-45' : ''"
          style="background-color: var(--cc-brand-primary); color: #090C0F;"
          @tap="handleSaveToAlbum"
        >
          {{ exporterMeta.exportButtonLabel }}
        </view>
      </view>

      <!-- 微信转发按钮 -->
      <button
        open-type="share"
        class="w-full py-2.5 rounded-xl cc-elevated cc-text-primary text-xs font-medium border"
        style="border-color: var(--cc-border);"
        @tap="handleShareTap"
      >
        转发给微信好友（还原当前报告参数）
      </button>
    </view>

    <!-- 离屏 1080px Canvas 2D 节点 -->
    <canvas
      id="reportCanvas"
      canvas-id="reportCanvas"
      type="2d"
      style="width: 1080px; height: 1440px; position: fixed; left: -9999px; top: -9999px;"
    />

    <CSafeArea position="bottom" />
  </view>
</template>
