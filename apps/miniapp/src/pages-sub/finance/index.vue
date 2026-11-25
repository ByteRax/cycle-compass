<script lang="ts" setup>
import type {
  AssetSummaryItem,
  CatalogOverview,
  CatalogResponse,
  PeriodMatrixResponse,
  SnapshotMeta
} from '@/api/types/trend'
import { onLoad, onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { computed, onUnmounted, ref } from 'vue'
import { getAssetCatalog, getAssetMembers, getPeriodMatrix } from '@/api/trend'
import AssetRow from '@/components/finance/AssetRow.vue'
import FinanceShell from '@/components/finance/FinanceShell.vue'
import PeriodMatrixTable from '@/components/finance/PeriodMatrixTable.vue'
import { useFinanceStore } from '@/stores/finance'
import { createAsyncRace, isExpiredSnapshotError, isStaleCursorError, unwrapApiResponse } from '@/utils/asyncRace'
import { buildFinanceUrl, decodeFinanceQuery, toApiCurrency, toDisplayCurrency } from '@/utils/financeRoute'
import { getOverviewSession, saveOverviewSession } from '@/utils/financeSession'

definePage({
  style: {
    navigationBarTitleText: '金融资产',
    navigationStyle: 'custom',
    enablePullDownRefresh: true,
    backgroundColor: '#090C0F'
  }
})

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

// 路由参数与当前观察池
const currentPoolId = ref<string>(financeStore.defaultHomePool || 'universe_macro')
const currentPoolTitle = ref<string>('大类资产')
const breadcrumbText = ref<string>('金融资产 / 精选观察池')
const snapshotMode = ref<'latest' | 'fixed'>('latest')
const requestedSnapshotId = ref<string>('')
const boundSnapshotId = ref<string>('')

// 视图切换：行情表 vs 周期矩阵
type OverviewMode = 'quotes' | 'matrix'
const viewMode = ref<OverviewMode>('quotes')

// 状态机
const loading = ref<boolean>(true)
const loadingMore = ref<boolean>(false)
const errorMsg = ref<string>('')
const staleCursorBanner = ref<boolean>(false)
const overviewUnavailable = ref<boolean>(false)
const overviewCollapsed = ref<boolean>(false)
const showFilterSheet = ref<boolean>(false)
const showCurrencySheet = ref<boolean>(false)
const showFavoriteSheet = ref<boolean>(false)
const favoriteMode = ref<'latest' | 'fixed'>('latest')
const favoriteNote = ref<string>('')
const anchorAssetId = ref<string>('')

// 数据
const snapshot = ref<SnapshotMeta | null>(null)
const catalog = ref<CatalogResponse | null>(null)
const poolOverview = ref<CatalogOverview | null>(null)
const serverMarkets = ref<string[]>([])
const items = ref<AssetSummaryItem[]>([])
const matrixData = ref<PeriodMatrixResponse | null>(null)
const nextCursor = ref<string | null>(null)
const totalCount = ref<number>(0)
const matchedTotal = ref<number>(0)
const poolTotal = ref<number>(0)
const restoreLimit = ref<number>(40)

// 筛选与排序
type QuoteSortField = 'strength' | 'day_change_pct' | 'turnover' | 'duration_trading_days'
type StatusFilter = 'all' | 'right' | 'left' | 'forming'

const sortField = ref<string>('strength')
const sortOrder = ref<'asc' | 'desc'>('desc')
const currency = ref<'CNY' | 'USD' | 'HKD' | 'RAW'>(financeStore.defaultCurrency || 'CNY')
const assetTypeFilter = ref<string>('all')
const statusFilter = ref<StatusFilter>('all')
const marketFilter = ref<string>('all')

const poolRace = createAsyncRace()

onUnmounted(() => {
  poolRace.invalidate()
})

// 分组折叠状态
const collapsedGroups = ref<Record<string, boolean>>({})

const assetTypeChips = [
  { key: 'all', label: '全部' },
  { key: 'stock', label: '股票' },
  { key: 'etf', label: 'ETF' },
  { key: 'bond', label: '债券' },
  { key: 'future', label: '期货' },
  { key: 'fx', label: '外汇' }
] as const

const quoteSortOptions: Array<{ key: QuoteSortField, label: string }> = [
  { key: 'strength', label: '强度' },
  { key: 'day_change_pct', label: '日涨幅' },
  { key: 'turnover', label: '成交额' },
  { key: 'duration_trading_days', label: '持续天数' }
]

const matrixSortOptions: Array<{ key: string, label: string }> = [
  { key: 'ytd', label: '今年来' },
  { key: 'rolling_1m', label: '近1月' },
  { key: 'wtd', label: '本周' },
  { key: 'mtd', label: '本月' },
  { key: 'qtd', label: '本季' },
  { key: 'prev_year', label: '去年全年' }
]

const activeSortOptions = computed(() =>
  viewMode.value === 'matrix'
? matrixSortOptions
: quoteSortOptions
)

const statusOptions: Array<{ key: StatusFilter, label: string }> = [
  { key: 'all', label: '全部状态' },
  { key: 'right', label: '右侧确认' },
  { key: 'left', label: '下行确认' },
  { key: 'forming', label: '形成中' }
]

const currencyChoices: Array<{ key: 'CNY' | 'USD' | 'HKD' | 'RAW', title: string, desc: string }> = [
  { key: 'CNY', title: 'CNY · 人民币', desc: '外币资产按起止交易日汇率折算为人民币口径' },
  { key: 'USD', title: 'USD · 美元', desc: '非美元资产按汇率折算为美元口径（缺汇率时标明原因）' },
  { key: 'HKD', title: 'HKD · 港元', desc: '折算为港元口径，无完整汇率链时不模拟估算' },
  { key: 'RAW', title: 'RAW · 原币种', desc: '保持各资产挂牌原币种收益率，不叠加汇率波动' }
]

const isPoolFavorite = computed(() => financeStore.isAssetFavorite(currentPoolId.value))
const isStale = computed(() => snapshot.value?.status === 'stale' || snapshot.value?.status === 'degraded')

const statusBadgeText = computed(() => {
  const st = snapshot.value?.status
  if (st === 'stale')
    return '数据已过期'
  if (st === 'degraded')
    return '部分降级'
  if (st === 'offline_seed')
    return '离线快照'
  return '非实时数据'
})

const currentSortLabel = computed(() => {
  const allOpts = [
...quoteSortOptions,
...matrixSortOptions,
{ key: 'name', label: '名称' }
]
  const found = allOpts.find(o => o.key === sortField.value)
  return `${found?.label || '排序'}${sortOrder.value === 'desc'
    ? '降序 ↓'
    : '升序 ↑'}`
})

const hasActiveFilters = computed(() =>
  assetTypeFilter.value !== 'all' || statusFilter.value !== 'all' || marketFilter.value !== 'all'
)

const filteredItems = computed(() => items.value)
const matrixLoadedCount = computed(() =>
  (matrixData.value?.groups ?? []).reduce((sum, g) => sum + g.rows.length, 0)
)
const activeLoadedCount = computed(() =>
  viewMode.value === 'matrix'
? matrixLoadedCount.value
: filteredItems.value.length
)

const availableMarkets = computed(() => {
  return ['all', ...serverMarkets.value]
})

interface GroupedSection {
  groupName: string
  rows: AssetSummaryItem[]
}

const hasQuoteGroups = computed(() => filteredItems.value.some(i => Boolean(i.group_name)))
const hasMatrixGroups = computed(() => (matrixData.value?.groups ?? []).some(g => Boolean(g.group_name)))
const hasGroups = computed(() => viewMode.value === 'matrix'
? hasMatrixGroups.value
: hasQuoteGroups.value)

const groupedSections = computed<GroupedSection[]>(() => {
  if (!hasQuoteGroups.value) {
    return [
{ groupName: '', rows: filteredItems.value }
]
}

  const map = new Map<string, AssetSummaryItem[]>()
  for (const row of filteredItems.value) {
    const key = row.group_name || '其他'
    const list = map.get(key)
    if (list)
      list.push(row)
    else
      map.set(key, [row])
  }
  return Array.from(map.entries()).map(([groupName, rows]) => ({ groupName, rows }))
})

const allGroupsCollapsed = computed(() => {
  if (viewMode.value === 'matrix') {
    const grps = matrixData.value?.groups ?? []
    if (!grps.length)
      return false
    return grps.every(g => collapsedGroups.value[g.group_name])
  }
  if (!hasQuoteGroups.value || !groupedSections.value.length)
    return false
  return groupedSections.value.every(g => collapsedGroups.value[g.groupName])
})

function toggleAllGroups() {
  const next = !allGroupsCollapsed.value
  const map: Record<string, boolean> = {}
  if (viewMode.value === 'matrix') {
    for (const g of matrixData.value?.groups ?? [])
      map[g.group_name] = next
  }
  else {
    for (const g of groupedSections.value)
      map[g.groupName] = next
  }
  collapsedGroups.value = map
  persistSessionState()
}

function toggleGroup(name: string) {
  collapsedGroups.value = {
    ...collapsedGroups.value,
    [name]: !collapsedGroups.value[name]
  }
  persistSessionState()
}

const overviewStats = computed(() => {
  const ov = poolOverview.value ?? catalog.value?.overview
  if (!ov) {
    return {
      unavailable: overviewUnavailable.value,
      validCount: 0,
      totalCount: 0,
      rightCount: 0,
      leftCount: 0,
      formingCount: 0,
      otherCount: 0,
      uncoveredCount: 0,
      rightPctText: '—',
      leftPctText: '—',
      otherPctText: '—',
      rightPctNum: 0,
      leftPctNum: 0,
      otherPctNum: 0
    }
  }
  const hasValid = ov.valid_count > 0
  const fmtRatio = (val: number | null | undefined, count: number) => {
    if (!hasValid)
      return '—'
    if (val !== null && val !== undefined)
      return `${Math.round(val)}%`
    return `${Math.round((count / ov.valid_count) * 100)}%`
  }
  const rPct = hasValid
    ? (ov.right_ratio_pct ?? (ov.right_count / ov.valid_count) * 100)
    : 0
  const lPct = hasValid
    ? (ov.left_ratio_pct ?? (ov.left_count / ov.valid_count) * 100)
    : 0
  const fPct = hasValid
    ? ((ov.forming_count ?? 0) / ov.valid_count) * 100
    : 0
  return {
    unavailable: false,
    validCount: ov.valid_count,
    totalCount: ov.total_count,
    rightCount: ov.right_count,
    leftCount: ov.left_count,
    formingCount: ov.forming_count ?? 0,
    otherCount: ov.other_count,
    uncoveredCount: ov.uncovered_count ?? Math.max(0, ov.total_count - ov.valid_count),
    rightPctText: fmtRatio(ov.right_ratio_pct, ov.right_count),
    leftPctText: fmtRatio(ov.left_ratio_pct, ov.left_count),
    otherPctText: hasValid
? `${Math.round(fPct)}%`
: '—',
    rightPctNum: Math.max(0, Math.min(100, rPct)),
    leftPctNum: Math.max(0, Math.min(100, lPct)),
    otherPctNum: Math.max(0, Math.min(100, fPct))
  }
})

function persistSessionState(overrideAnchorId?: string) {
  saveOverviewSession({
    poolId: currentPoolId.value,
    viewMode: viewMode.value,
    sortField: sortField.value,
    sortOrder: sortOrder.value,
    currency: currency.value,
    marketFilter: marketFilter.value,
    assetTypeFilter: assetTypeFilter.value,
    statusFilter: statusFilter.value,
    collapsedGroups: collapsedGroups.value,
    loadedCount: Math.max(40, activeLoadedCount.value),
    anchorAssetId: overrideAnchorId ?? anchorAssetId.value
  })
}

function mergeMatrixPages(prev: PeriodMatrixResponse | null, next: PeriodMatrixResponse): PeriodMatrixResponse {
  if (!prev)
    return next
  const groupMap = new Map<string, PeriodMatrixResponse['groups'][number]['rows']>()
  const orderKeys: string[] = []
  const pushRows = (grps: PeriodMatrixResponse['groups']) => {
    for (const grp of grps) {
      let arr = groupMap.get(grp.group_name)
      if (!arr) {
        arr = []
        groupMap.set(grp.group_name, arr)
        orderKeys.push(grp.group_name)
      }
      arr.push(...grp.rows)
    }
  }
  pushRows(prev.groups || [])
  pushRows(next.groups || [])
  return {
    ...next,
    groups: orderKeys.map(k => ({
      group_name: k,
      rows: groupMap.get(k) || []
    }))
  }
}

async function fetchPoolData(append = false, forceRefreshSnapshot = false) {
  const token = poolRace.next()
  staleCursorBanner.value = false
  if (!append) {
    loading.value = true
    errorMsg.value = ''
    if (forceRefreshSnapshot && snapshotMode.value === 'latest')
      boundSnapshotId.value = ''
  }
  else {
    loadingMore.value = true
  }

  const reqSnapId = append
    ? (boundSnapshotId.value || undefined)
    : (snapshotMode.value === 'fixed'
        ? (requestedSnapshotId.value || boundSnapshotId.value || undefined)
        : undefined)

  const pageLimit = append
    ? 40
    : Math.min(200, Math.max(40, restoreLimit.value))

  const commonParams = {
    scope_id: currentPoolId.value,
    sort: sortField.value,
    order: sortOrder.value,
    cursor: append
? (nextCursor.value || undefined)
: undefined,
    limit: pageLimit,
    currency: toApiCurrency(currency.value),
    snapshot_id: reqSnapId,
    market: marketFilter.value !== 'all'
? marketFilter.value
: undefined,
    asset_type: assetTypeFilter.value !== 'all'
? assetTypeFilter.value
: undefined,
    status: statusFilter.value !== 'all'
? statusFilter.value
: undefined
  }

  try {
    const catalogPromise = append
      ? Promise.resolve(null)
      : getAssetCatalog({
          scope_id: currentPoolId.value,
          currency: toApiCurrency(currency.value)
        }).catch(() => null)

    if (viewMode.value === 'matrix') {
      const [catalogRes, matrixRes] = await Promise.all([
        catalogPromise,
        getPeriodMatrix(currentPoolId.value, commonParams)
      ])
      if (!poolRace.isLatest(token))
        return

      if (catalogRes && catalogRes.code === 0 && catalogRes.data)
        catalog.value = catalogRes.data

      const mData = unwrapApiResponse(matrixRes, '周期矩阵加载失败')
      snapshot.value = mData.snapshot
      if (mData.snapshot?.snapshot_id)
        boundSnapshotId.value = mData.snapshot.snapshot_id
      if (mData.pool_name)
        currentPoolTitle.value = mData.pool_name
      if (mData.overview) {
        poolOverview.value = mData.overview
        overviewUnavailable.value = false
      }
      else if (!catalogRes?.data?.overview) {
        overviewUnavailable.value = true
      }
      if (mData.available_markets)
        serverMarkets.value = mData.available_markets

      matrixData.value = append
? mergeMatrixPages(matrixData.value, mData)
: mData
      nextCursor.value = mData.next_cursor
      matchedTotal.value = mData.matched_total ?? mData.total
      poolTotal.value = mData.pool_total ?? mData.total
      totalCount.value = matchedTotal.value
    }
    else {
      const [catalogRes, membersRes] = await Promise.all([
        catalogPromise,
        getAssetMembers(currentPoolId.value, commonParams)
      ])
      if (!poolRace.isLatest(token))
        return

      if (catalogRes && catalogRes.code === 0 && catalogRes.data)
        catalog.value = catalogRes.data

      const membersData = unwrapApiResponse(membersRes, '资产数据响应为空')
      snapshot.value = membersData.snapshot
      if (membersData.snapshot?.snapshot_id)
        boundSnapshotId.value = membersData.snapshot.snapshot_id
      if (membersData.parent_name)
        currentPoolTitle.value = membersData.parent_name

      if (membersData.overview) {
        poolOverview.value = membersData.overview
        overviewUnavailable.value = false
      }
      else if (!catalogRes?.data?.overview) {
        overviewUnavailable.value = true
      }
      if (membersData.available_markets)
        serverMarkets.value = membersData.available_markets

      items.value = append
        ? [
...items.value,
...membersData.items
]
        : membersData.items
      nextCursor.value = membersData.next_cursor
      matchedTotal.value = membersData.matched_total ?? membersData.total
      poolTotal.value = membersData.pool_total ?? membersData.total
      totalCount.value = matchedTotal.value
    }

    if (!append)
      restoreLimit.value = 40
    persistSessionState()
  }
  catch (err: any) {
    if (!poolRace.isLatest(token))
      return
    if (isStaleCursorError(err)) {
      staleCursorBanner.value = true
      nextCursor.value = null
      uni.showToast({ title: '数据快照已更新，请刷新列表', icon: 'none' })
      return
    }
    if (!append) {
      overviewUnavailable.value = true
      errorMsg.value = isExpiredSnapshotError(err)
        ? '请求的快照已过期，请刷新获取最新快照'
        : (err?.message || '资产数据加载失败，请稍后重试')
    }
    else {
      uni.showToast({ title: err?.message || '加载更多失败，请重试', icon: 'none' })
    }
  }
  finally {
    if (poolRace.isLatest(token)) {
      loading.value = false
      loadingMore.value = false
      uni.stopPullDownRefresh()
    }
  }
}

onLoad((query: any) => {
  const parsed = decodeFinanceQuery(query, undefined, {
    defaultHomePool: financeStore.defaultHomePool,
    defaultCurrency: financeStore.defaultCurrency,
    defaultRange: financeStore.defaultRange
  }, {
    scope_id: financeStore.defaultHomePool || 'universe_macro'
  })
  currentPoolId.value = parsed.scope_id || parsed.asset_id || 'universe_macro'
  if (parsed.title)
    currentPoolTitle.value = parsed.title

  const saved = getOverviewSession(currentPoolId.value)
  if (!query?.view_mode && saved) {
    viewMode.value = saved.viewMode
    collapsedGroups.value = { ...saved.collapsedGroups }
    restoreLimit.value = saved.loadedCount || 40
    anchorAssetId.value = saved.anchorAssetId || ''
  }
  if (parsed.view_mode)
    viewMode.value = parsed.view_mode

  currency.value = toDisplayCurrency(query?.currency
? parsed.currency
: (saved?.currency || parsed.currency))
  snapshotMode.value = parsed.snapshot_mode
  requestedSnapshotId.value = parsed.snapshot_id
  boundSnapshotId.value = parsed.snapshot_id

  if (parsed.sort) {
    sortField.value = parsed.sort === 'duration'
? 'duration_trading_days'
: parsed.sort
  }
  else if (saved?.sortField && !query?.view_mode) {
    sortField.value = saved.sortField === 'duration'
? 'duration_trading_days'
: saved.sortField
  }
  else {
    sortField.value = viewMode.value === 'matrix'
? 'ytd'
: 'strength'
  }

  sortOrder.value = parsed.order || saved?.sortOrder || 'desc'
  marketFilter.value = parsed.market || saved?.marketFilter || 'all'
  assetTypeFilter.value = parsed.asset_type || saved?.assetTypeFilter || 'all'
  if (parsed.status === 'right' || parsed.status === 'left' || parsed.status === 'forming')
    statusFilter.value = parsed.status
  else if (saved?.statusFilter === 'right' || saved?.statusFilter === 'left' || saved?.statusFilter === 'forming')
    statusFilter.value = saved.statusFilter

  if (query?.category_id)
    breadcrumbText.value = `金融资产 / ${currentPoolTitle.value}`
  fetchPoolData(false, false)
})

onShow(() => {
  const saved = getOverviewSession(currentPoolId.value)
  if (saved?.anchorAssetId)
    anchorAssetId.value = saved.anchorAssetId
})

onPullDownRefresh(() => {
  if (snapshotMode.value === 'fixed')
    snapshotMode.value = 'latest'
  requestedSnapshotId.value = ''
  nextCursor.value = null
  anchorAssetId.value = ''
  fetchPoolData(false, true)
})

function switchViewMode(mode: OverviewMode) {
  if (viewMode.value === mode)
    return
  viewMode.value = mode
  nextCursor.value = null
  anchorAssetId.value = ''
  if (mode === 'matrix') {
    const validMatrixSorts = new Set([
'wtd',
'mtd',
'qtd',
'ytd',
'prev_year',
'rolling_1m',
'name'
])
    if (!validMatrixSorts.has(sortField.value)) {
      sortField.value = 'ytd'
      sortOrder.value = 'desc'
    }
  }
  else {
    const validQuoteSorts = new Set([
      'strength',
      'day_change_pct',
      'turnover',
      'duration_trading_days'
    ])
    if (sortField.value === 'duration')
      sortField.value = 'duration_trading_days'
    if (!validQuoteSorts.has(sortField.value)) {
      sortField.value = 'strength'
      sortOrder.value = 'desc'
    }
  }
  fetchPoolData(false, false)
}

function selectPool(poolId: string, name: string) {
  if (poolId === currentPoolId.value)
    return
  currentPoolId.value = poolId
  currentPoolTitle.value = name
  breadcrumbText.value = poolId.startsWith('universe_')
    ? '金融资产 / 精选观察池'
    : `金融资产 / ${name}`
  nextCursor.value = null
  marketFilter.value = 'all'
  anchorAssetId.value = ''
  collapsedGroups.value = {}
  fetchPoolData(false, true)
}

function toggleSort(field: string) {
  if (sortField.value === field) {
    sortOrder.value = sortOrder.value === 'desc'
? 'asc'
: 'desc'
  }
  else {
    sortField.value = field
    sortOrder.value = field === 'name'
? 'asc'
: 'desc'
  }
  nextCursor.value = null
  fetchPoolData(false, false)
}

function selectAssetType(key: string) {
  if (assetTypeFilter.value === key)
    return
  assetTypeFilter.value = key
  nextCursor.value = null
  fetchPoolData(false, false)
}

function selectStatusFilter(key: StatusFilter) {
  statusFilter.value = statusFilter.value === key && key !== 'all'
? 'all'
: key
  nextCursor.value = null
  fetchPoolData(false, false)
}

function selectMarketFilter(mkt: string) {
  if (marketFilter.value === mkt)
    return
  marketFilter.value = mkt
  nextCursor.value = null
  fetchPoolData(false, false)
}

function chooseCurrency(nextCcy: 'CNY' | 'USD' | 'HKD' | 'RAW') {
  showCurrencySheet.value = false
  if (currency.value === nextCcy)
    return
  currency.value = nextCcy
  nextCursor.value = null
  fetchPoolData(false, true)
}

function resetFilters() {
  assetTypeFilter.value = 'all'
  statusFilter.value = 'all'
  marketFilter.value = 'all'
  nextCursor.value = null
  fetchPoolData(false, false)
}

function handleToggleFavorite(assetId: string) {
  const added = financeStore.toggleFavoriteAsset(assetId)
  uni.showToast({
    title: added ? '已加入收藏' : '已取消收藏',
    icon: 'none',
    duration: 1200
  })
}

function openFavoriteViewSheet() {
  favoriteMode.value = snapshotMode.value === 'fixed'
? 'fixed'
: 'latest'
  favoriteNote.value = ''
  showFavoriteSheet.value = true
}

function confirmSaveFavoriteView() {
  const targetChart = viewMode.value === 'matrix'
? 'matrix'
: 'dashboard'
  const snapId = favoriteMode.value === 'fixed'
? boundSnapshotId.value
: ''
  financeStore.saveFavoriteChart({
    title: `${currentPoolTitle.value} · ${viewMode.value === 'matrix'
? '周期矩阵'
: '行情全景'}`,
    chart: targetChart,
    asset_id: currentPoolId.value,
    asset_name: currentPoolTitle.value,
    snapshot_mode: favoriteMode.value,
    snapshot_id: snapId || undefined,
    note: favoriteNote.value.trim() || undefined,
    context: {
      asset_id: currentPoolId.value,
      scope_id: currentPoolId.value,
      pool_id: currentPoolId.value,
      view_mode: viewMode.value,
      chart: targetChart,
      currency: toApiCurrency(currency.value),
      range: financeStore.defaultRange,
      snapshot_mode: favoriteMode.value,
      snapshot_id: snapId,
      sort: sortField.value,
      order: sortOrder.value,
      market: marketFilter.value !== 'all'
? marketFilter.value
: undefined,
      asset_type: assetTypeFilter.value !== 'all'
? assetTypeFilter.value
: undefined,
      status: statusFilter.value !== 'all'
? statusFilter.value
: undefined
    }
  })
  showFavoriteSheet.value = false
  uni.showToast({
    title: favoriteMode.value === 'fixed'
? '已保存本期快照视图'
: '已加入每日跟踪视图',
    icon: 'success'
  })
}

function openSearch() {
  uni.navigateTo({ url: '/pages-sub/finance/search' })
}

function openAssetDetail(item: AssetSummaryItem) {
  anchorAssetId.value = item.asset_id
  persistSessionState(item.asset_id)
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/asset', {
      asset_id: item.asset_id,
      scope_id: currentPoolId.value,
      currency: currency.value,
      range: financeStore.defaultRange,
      view_mode: viewMode.value,
      sort: sortField.value,
      order: sortOrder.value,
      snapshot_mode: boundSnapshotId.value ? snapshotMode.value : 'latest',
      snapshot_id: snapshotMode.value === 'fixed'
? boundSnapshotId.value
: ''
    })
  })
}

function openMatrixRangeCurve(payload: { assetId: string, startDate: string, endDate: string }) {
  anchorAssetId.value = payload.assetId
  persistSessionState(payload.assetId)
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: payload.assetId,
      scope_id: currentPoolId.value,
      currency: currency.value,
      chart: 'curve',
      start_date: payload.startDate,
      end_date: payload.endDate,
      snapshot_mode: snapshotMode.value,
      snapshot_id: snapshotMode.value === 'fixed'
? boundSnapshotId.value
: ''
    })
  })
}

function openWorkbench() {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: currentPoolId.value,
      scope_id: currentPoolId.value,
      currency: currency.value,
      range: financeStore.defaultRange,
      chart: 'components',
      snapshot_mode: snapshotMode.value,
      snapshot_id: snapshotMode.value === 'fixed'
? boundSnapshotId.value
: ''
    })
  })
}

function openReportExport() {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/report', {
      asset_id: currentPoolId.value,
      scope_id: currentPoolId.value,
      chart: viewMode.value === 'matrix'
? 'matrix'
: 'dashboard',
      currency: currency.value,
      range: financeStore.defaultRange,
      sort: sortField.value,
      order: sortOrder.value,
      market: marketFilter.value !== 'all'
? marketFilter.value
: '',
      asset_type: assetTypeFilter.value !== 'all'
? assetTypeFilter.value
: '',
      status: statusFilter.value !== 'all'
? statusFilter.value
: '',
      snapshot_mode: snapshotMode.value,
      snapshot_id: snapshotMode.value === 'fixed'
? boundSnapshotId.value
: ''
    })
  })
}

function openDataStatus() {
  uni.navigateTo({ url: '/pages-sub/finance/data-status' })
}
</script>

<template>
  <FinanceShell title="金融资产" active-tab="overview" :asset-id="currentPoolId">
    <template #sub-header>
      <view class="px-3 pb-2.5 pt-1">
        <view
          class="flex items-center gap-2 rounded-xl px-3 min-h-[40px] border active:opacity-80"
          :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
          @tap="openSearch"
        >
          <text class="text-[13px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">🔍</text>
          <text class="text-[13px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            搜索市场、行业、品种或代码
          </text>
        </view>
      </view>
    </template>

    <view class="px-3 py-3 space-y-3">
      <!-- 面包屑 + 标题 + 快捷操作按钮 -->
      <view class="flex items-start justify-between gap-2">
        <view class="min-w-0 flex-1">
          <text class="text-[11px] block truncate" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ breadcrumbText }}
          </text>
          <text class="text-[22px] font-bold tracking-tight block mt-0.5 truncate">
            {{ currentPoolTitle }}
          </text>
        </view>
        <view class="flex items-center shrink-0">
          <view
            class="min-w-[40px] min-h-[40px] flex items-center justify-center active:opacity-70"
            @tap="openWorkbench"
          >
            <text class="text-[17px]" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">📈</text>
          </view>
          <view
            class="min-w-[40px] min-h-[40px] flex items-center justify-center active:opacity-70"
            @tap="openFavoriteViewSheet"
          >
            <text
              class="text-[19px]"
              :class="isPoolFavorite
                ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]')
                : (isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]')"
            >
              {{ isPoolFavorite ? '★' : '☆' }}
            </text>
          </view>
          <view
            class="min-w-[40px] min-h-[40px] flex items-center justify-center active:opacity-70"
            @tap="openReportExport"
          >
            <text class="text-[17px]" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">↗</text>
          </view>
        </view>
      </view>

      <!-- 观察池切换条 -->
      <scroll-view
        v-if="catalog?.universes?.length"
        scroll-x
        class="whitespace-nowrap"
        :show-scrollbar="false"
      >
        <view class="inline-flex items-center gap-2">
          <view
            v-for="uniItem in catalog.universes"
            :key="uniItem.universe_id"
            class="px-3 py-1.5 rounded-lg border text-[12px] font-medium shrink-0"
            :class="currentPoolId === uniItem.universe_id
              ? (isDark ? 'bg-[#E5B352]/15 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
              : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
            @tap="selectPool(uniItem.universe_id, uniItem.name)"
          >
            {{ uniItem.name }}
          </view>
        </view>
      </scroll-view>

      <!-- 行 2：截止日 + 覆盖数 + 显式币种下拉 + 状态徽标 -->
      <view class="flex items-center justify-between gap-2 flex-wrap">
        <view class="flex items-center gap-2 min-w-0">
          <view
            class="w-2 h-2 rounded-full shrink-0"
            :class="isStale ? 'bg-[#F05A50]' : (isDark ? 'bg-[#E5B352]' : 'bg-[#8A641F]')"
          />
          <text class="text-[11px] font-mono cc-tabular truncate" :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'">
            截至 {{ snapshot?.as_of || '—' }} · {{ overviewStats.unavailable ? '统计暂不可用' : `有效 ${overviewStats.validCount}/${overviewStats.totalCount}` }}
          </text>
        </view>

        <view class="flex items-center gap-1.5 shrink-0">
          <view
            class="px-2.5 py-1 rounded-lg border text-[11px] font-mono flex items-center gap-1"
            :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#F1EFEA]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#22292E]'"
            @tap="showCurrencySheet = true"
          >
            <text>{{ currency }}</text>
            <text class="text-[10px] opacity-70">▾</text>
          </view>
          <view
            class="px-2 py-1 rounded-lg border text-[11px]"
            :class="isStale
              ? 'border-[#F05A50]/40 text-[#F05A50] bg-[#F05A50]/10'
              : (isDark ? 'border-[#E5B352]/35 text-[#E5B352] bg-[#E5B352]/10' : 'border-[#8A641F]/35 text-[#8A641F] bg-[#8A641F]/10')"
            @tap="openDataStatus"
          >
            {{ statusBadgeText }}
          </view>
        </view>
      </view>

      <!-- Stale / 过期快照提醒条 -->
      <view
        v-if="isStale || staleCursorBanner"
        class="rounded-xl px-3 py-2 border border-[#F05A50]/35 bg-[#F05A50]/10 flex items-center justify-between gap-2"
      >
        <text class="text-[12px] text-[#F05A50] flex-1">
          {{ staleCursorBanner ? '观察池快照已更新，建议刷新后重新浏览。' : '当前快照存在延迟或部分市场未更新，下拉可刷新重试。' }}
        </text>
        <text class="text-[12px] font-bold text-[#F05A50]" @tap="fetchPoolData(false, true)">刷新</text>
      </view>

      <!-- 可折叠趋势概况与点击联动状态卡 -->
      <view
        class="rounded-xl border p-3"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <view
          class="flex items-center justify-between min-h-[26px]"
          @tap="overviewCollapsed = !overviewCollapsed"
        >
          <text class="text-[13px] font-bold">市场格局概况（点击可筛选）</text>
          <view class="flex items-center gap-2">
            <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              其余 {{ overviewStats.otherCount }} · 未覆盖 {{ overviewStats.uncoveredCount }}
            </text>
            <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              {{ overviewCollapsed ? '⌄' : '⌃' }}
            </text>
          </view>
        </view>

        <view v-if="!overviewCollapsed && overviewStats.unavailable" class="mt-2 py-2 text-center">
          <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            统计暂不可用，请下拉刷新重试
          </text>
        </view>

        <view v-else-if="!overviewCollapsed" class="mt-2.5 space-y-2.5">
          <view class="grid grid-cols-3 gap-2">
            <view
              class="rounded-lg border p-2 transition-colors"
              :class="statusFilter === 'right'
                ? (isDark ? 'bg-[#E5B352]/14 border-[#E5B352]' : 'bg-[#8A641F]/10 border-[#8A641F]')
                : (isDark ? 'bg-[#181E25]/60 border-[#29313A]' : 'bg-[#F8F7F4] border-[#DCDDD8]')"
              @tap="selectStatusFilter('right')"
            >
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">右侧确认</text>
              <text
                class="text-[17px] font-bold font-mono cc-tabular block mt-0.5"
                :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'"
              >
                {{ overviewStats.rightPctText }}
              </text>
              <text class="text-[10px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ overviewStats.rightCount }} / {{ overviewStats.validCount }}
              </text>
            </view>

            <view
              class="rounded-lg border p-2 transition-colors"
              :class="statusFilter === 'left'
                ? 'bg-[#2AD1A3]/14 border-[#2AD1A3]'
                : (isDark ? 'bg-[#181E25]/60 border-[#29313A]' : 'bg-[#F8F7F4] border-[#DCDDD8]')"
              @tap="selectStatusFilter('left')"
            >
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">下行确认</text>
              <text
                class="text-[17px] font-bold font-mono cc-tabular block mt-0.5"
                :class="isDark ? 'text-[#2AD1A3]' : 'text-[#14745E]'"
              >
                {{ overviewStats.leftPctText }}
              </text>
              <text class="text-[10px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ overviewStats.leftCount }} / {{ overviewStats.validCount }}
              </text>
            </view>

            <view
              class="rounded-lg border p-2 transition-colors"
              :class="statusFilter === 'forming'
                ? (isDark ? 'bg-[#72A7D9]/18 border-[#72A7D9]' : 'bg-[#3B6D9C]/12 border-[#3B6D9C]')
                : (isDark ? 'bg-[#181E25]/60 border-[#29313A]' : 'bg-[#F8F7F4] border-[#DCDDD8]')"
              @tap="selectStatusFilter('forming')"
            >
              <text class="text-[11px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">形成中</text>
              <text class="text-[17px] font-bold font-mono cc-tabular block mt-0.5">
                {{ overviewStats.otherPctText }}
              </text>
              <text class="text-[10px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ overviewStats.formingCount }} 只待确认
              </text>
            </view>
          </view>

          <!-- 三段比例条 -->
          <view
            class="h-[5px] rounded-full overflow-hidden flex gap-0.5"
            :class="isDark ? 'bg-[#181E25]' : 'bg-[#EDECE6]'"
          >
            <view
              class="h-full rounded-l-full"
              :class="isDark ? 'bg-[#E5B352]' : 'bg-[#8A641F]'"
              :style="{ width: `${overviewStats.rightPctNum}%` }"
            />
            <view
              class="h-full"
              :class="isDark ? 'bg-[#2AD1A3]' : 'bg-[#14745E]'"
              :style="{ width: `${overviewStats.leftPctNum}%` }"
            />
            <view
              class="h-full rounded-r-full"
              :class="isDark ? 'bg-[#64707C]' : 'bg-[#9AA5B1]'"
              :style="{ width: `${overviewStats.otherPctNum}%` }"
            />
          </view>
        </view>
      </view>

      <!-- 单一 segmented 控件：行情表 | 周期矩阵 -->
      <view
        class="p-1 rounded-xl border grid grid-cols-2 gap-1"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#EDECE6] border-[#DCDDD8]'"
      >
        <view
          class="min-h-[36px] rounded-lg flex items-center justify-center text-[13px] font-medium"
          :class="viewMode === 'quotes'
            ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#FFFFFF] text-[#22292E] font-bold shadow-sm')
            : (isDark ? 'text-[#A9B3BD]' : 'text-[#626F77]')"
          @tap="switchViewMode('quotes')"
        >
          行情表
        </view>
        <view
          class="min-h-[36px] rounded-lg flex items-center justify-center text-[13px] font-medium"
          :class="viewMode === 'matrix'
            ? (isDark ? 'bg-[#E5B352] text-[#090C0F] font-bold' : 'bg-[#FFFFFF] text-[#22292E] font-bold shadow-sm')
            : (isDark ? 'text-[#A9B3BD]' : 'text-[#626F77]')"
          @tap="switchViewMode('matrix')"
        >
          周期矩阵
        </view>
      </view>

      <!-- 资产类型筛选条 -->
      <scroll-view scroll-x class="whitespace-nowrap" :show-scrollbar="false">
        <view class="inline-flex items-center gap-1.5">
          <view
            v-for="chip in assetTypeChips"
            :key="chip.key"
            class="px-3 py-1.5 rounded-lg border text-[12px] shrink-0"
            :class="assetTypeFilter === chip.key
              ? (isDark ? 'bg-[#E5B352]/16 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
              : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
            @tap="selectAssetType(chip.key)"
          >
            {{ chip.label }}
          </view>
        </view>
      </scroll-view>

      <!-- 工具条：排序 + 分组折叠 + 筛选 + 收藏/导出快捷 -->
      <view class="flex items-center justify-between gap-2">
        <view class="flex items-center gap-2">
          <view
            class="px-2.5 py-1.5 rounded-lg border text-[12px] flex items-center gap-1 min-h-[34px]"
            :class="isDark ? 'bg-[#E5B352]/12 border-[#E5B352]/50 text-[#E5B352]' : 'bg-[#8A641F]/10 border-[#8A641F]/50 text-[#8A641F]'"
            @tap="toggleSort(sortField)"
          >
            <text class="font-medium">{{ currentSortLabel }}</text>
          </view>
          <view
            class="px-2.5 py-1.5 rounded-lg border text-[12px] flex items-center gap-1 min-h-[34px]"
            :class="hasActiveFilters
              ? (isDark ? 'bg-[#E5B352]/15 border-[#E5B352] text-[#E5B352]' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F]')
              : (isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]')"
            @tap="showFilterSheet = !showFilterSheet"
          >
            <text>筛选</text>
          </view>
        </view>

        <view class="flex items-center gap-2">
          <view
            v-if="hasGroups"
            class="px-2 py-1.5 text-[12px]"
            :class="isDark ? 'text-[#A9B3BD]' : 'text-[#515F6A]'"
            @tap="toggleAllGroups"
          >
            {{ allGroupsCollapsed ? '全部展开 ⌄' : '全部折叠 ⌃' }}
          </view>
          <view
            class="px-2.5 py-1.5 rounded-lg border text-[12px] min-h-[34px] flex items-center"
            :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#E5B352]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#8A641F]'"
            @tap="openReportExport"
          >
            <text>{{ viewMode === 'matrix' ? '导出矩阵' : '导出周报' }}</text>
          </view>
        </view>
      </view>

      <!-- 展开的筛选面板（排序字段/状态/市场） -->
      <view
        v-if="showFilterSheet"
        class="rounded-xl border p-3 space-y-2.5"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <view>
          <text class="text-[11px] block mb-1.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            全池排序字段（跨分页有效）
          </text>
          <view class="flex flex-wrap gap-1.5">
            <view
              v-for="opt in activeSortOptions"
              :key="opt.key"
              class="px-2.5 py-1 rounded border text-[12px]"
              :class="sortField === opt.key
                ? (isDark ? 'bg-[#E5B352]/18 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]')"
              @tap="toggleSort(opt.key)"
            >
              {{ opt.label }}
            </view>
          </view>
        </view>

        <view>
          <text class="text-[11px] block mb-1.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">趋势状态</text>
          <view class="flex flex-wrap gap-1.5">
            <view
              v-for="opt in statusOptions"
              :key="opt.key"
              class="px-2.5 py-1 rounded border text-[12px]"
              :class="statusFilter === opt.key
                ? (isDark ? 'bg-[#E5B352]/18 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]')"
              @tap="selectStatusFilter(opt.key)"
            >
              {{ opt.label }}
            </view>
          </view>
        </view>

        <view v-if="availableMarkets.length > 2">
          <text class="text-[11px] block mb-1.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">市场范围</text>
          <view class="flex flex-wrap gap-1.5">
            <view
              v-for="mkt in availableMarkets"
              :key="mkt"
              class="px-2.5 py-1 rounded border text-[12px]"
              :class="marketFilter === mkt
                ? (isDark ? 'bg-[#E5B352]/18 border-[#E5B352] text-[#E5B352] font-bold' : 'bg-[#8A641F]/12 border-[#8A641F] text-[#8A641F] font-bold')
                : (isDark ? 'bg-[#181E25] border-[#29313A] text-[#A9B3BD]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#515F6A]')"
              @tap="selectMarketFilter(mkt)"
            >
              {{ mkt === 'all' ? '全部市场' : mkt }}
            </view>
          </view>
        </view>
      </view>

      <!-- 模式 A：周期矩阵视图 -->
      <PeriodMatrixTable
        v-if="viewMode === 'matrix'"
        :matrix="matrixData"
        :loading="loading"
        :error-msg="errorMsg"
        :sort-field="sortField"
        :sort-order="sortOrder"
        :collapsed-groups="collapsedGroups"
        :anchor-asset-id="anchorAssetId"
        :has-active-filters="hasActiveFilters"
        @toggle-sort="toggleSort"
        @toggle-group="toggleGroup"
        @select-asset="openAssetDetail"
        @open-range-curve="openMatrixRangeCurve"
        @retry="fetchPoolData(false)"
        @reset-filters="resetFilters"
      />

      <!-- 模式 B：传统行情表 -->
      <view
        v-else
        class="rounded-xl border overflow-hidden"
        :class="isDark ? 'bg-[#12161B] border-[#29313A]' : 'bg-[#FFFFFF] border-[#DCDDD8]'"
      >
        <!-- 表头 -->
        <view
          class="flex items-center gap-1.5 px-3 py-2 border-b text-[11px]"
          :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#8D99A4]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#626F77]'"
        >
          <view class="w-[40px] text-center shrink-0">
            <text class="block leading-tight">温度</text>
            <text class="block leading-tight text-[10px]">天数</text>
          </view>
          <view class="flex-1 min-w-0">
            <text class="block leading-tight">品种</text>
            <text class="block leading-tight text-[10px]">名称 / 代码</text>
          </view>
          <view class="w-[62px] text-right shrink-0">
            <text class="block leading-tight">最新值</text>
            <text class="block leading-tight text-[10px]">规模</text>
          </view>
          <view class="w-[56px] text-right shrink-0" @tap="toggleSort('day_change_pct')">
            <text class="block leading-tight">日涨幅</text>
            <text class="block leading-tight text-[10px]">成交额</text>
          </view>
          <view class="w-[38px] text-right shrink-0" @tap="toggleSort('strength')">
            <text
              class="block leading-tight font-bold"
              :class="sortField === 'strength' ? (isDark ? 'text-[#E5B352]' : 'text-[#8A641F]') : ''"
            >
              强度{{ sortField === 'strength' ? (sortOrder === 'desc' ? '↓' : '↑') : '' }}
            </text>
            <text class="block leading-tight text-[10px]">排名</text>
          </view>
          <view class="w-[36px] text-center shrink-0 -mr-1">
            <text>☆</text>
          </view>
        </view>

        <!-- Loading 骨架屏 -->
        <view v-if="loading" class="p-3 space-y-3">
          <view
            v-for="n in 6"
            :key="n"
            class="h-[54px] rounded-lg animate-pulse"
            :class="isDark ? 'bg-[#181E25]' : 'bg-[#EDECE6]'"
          />
        </view>

        <!-- Error 态 -->
        <view v-else-if="errorMsg" class="px-4 py-10 text-center space-y-3">
          <text class="text-[14px] font-bold block text-[#F05A50]">加载失败</text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ errorMsg }}
          </text>
          <view
            class="inline-flex items-center justify-center px-4 min-h-[44px] rounded-xl border text-[13px] font-bold"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="fetchPoolData(false)"
          >
            重新加载
          </view>
        </view>

        <!-- Empty 态 -->
        <view v-else-if="!filteredItems.length" class="px-4 py-10 text-center space-y-3">
          <text class="text-[14px] font-bold block">
            {{ hasActiveFilters ? '当前筛选条件下无匹配资产' : '暂无资产数据' }}
          </text>
          <text class="text-[12px] block" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
            {{ hasActiveFilters ? '可尝试放宽类型、状态或市场筛选条件' : '当前观察池暂无有效快照成员' }}
          </text>
          <view
            v-if="hasActiveFilters"
            class="inline-flex items-center justify-center px-4 min-h-[44px] rounded-xl border text-[13px] font-bold"
            :class="isDark ? 'border-[#E5B352] text-[#E5B352]' : 'border-[#8A641F] text-[#8A641F]'"
            @tap="resetFilters"
          >
            重置筛选
          </view>
        </view>

        <!-- 资产行列表（支持分组折叠与锚点高亮） -->
        <view v-else>
          <template v-for="section in groupedSections" :key="section.groupName || '_flat'">
            <view
              v-if="section.groupName"
              class="flex items-center justify-between px-3 py-2 border-b min-h-[40px]"
              :class="isDark ? 'bg-[#181E25]/80 border-[#29313A]' : 'bg-[#EDECE6]/70 border-[#DCDDD8]'"
              @tap="toggleGroup(section.groupName)"
            >
              <view class="flex items-center gap-1.5">
                <text class="text-[13px] font-bold">{{ section.groupName }}</text>
                <text class="text-[11px] font-mono cc-tabular" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                  {{ section.rows.length }}
                </text>
              </view>
              <text class="text-[12px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
                {{ collapsedGroups[section.groupName] ? '›' : '⌄' }}
              </text>
            </view>

            <view v-if="!section.groupName || !collapsedGroups[section.groupName]">
              <view
                v-for="row in section.rows"
                :key="row.asset_id"
                :class="anchorAssetId === row.asset_id ? (isDark ? 'ring-1 ring-inset ring-[#E5B352]/60 bg-[#E5B352]/8' : 'ring-1 ring-inset ring-[#8A641F]/50 bg-[#8A641F]/6') : ''"
              >
                <AssetRow
                  :item="row"
                  @select="openAssetDetail"
                  @toggle-favorite="handleToggleFavorite"
                />
              </view>
            </view>
          </template>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="nextCursor && !loading" class="pt-1 pb-2 flex justify-center">
        <view
          class="px-5 min-h-[44px] rounded-xl border flex items-center justify-center text-[13px]"
          :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#A9B3BD]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#515F6A]'"
          @tap="fetchPoolData(true)"
        >
          {{ loadingMore ? '加载中...' : `加载更多 (${activeLoadedCount}/${totalCount})` }}
        </view>
      </view>
    </view>

    <!-- 显式币种切换抽屉 -->
    <view
      v-if="showCurrencySheet"
      class="fixed inset-0 z-50 flex items-end bg-black/60"
      @tap="showCurrencySheet = false"
    >
      <view
        class="w-full rounded-t-2xl border-t p-4 space-y-2.5"
        :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#F1EFEA]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#22292E]'"
        @tap.stop
      >
        <text class="text-[15px] font-bold block">选择计价币种口径</text>
        <view
          v-for="opt in currencyChoices"
          :key="opt.key"
          class="rounded-xl border p-3 flex items-center justify-between gap-2"
          :class="currency === opt.key
            ? (isDark ? 'bg-[#E5B352]/14 border-[#E5B352]' : 'bg-[#8A641F]/10 border-[#8A641F]')
            : (isDark ? 'bg-[#181E25] border-[#29313A]' : 'bg-[#F8F7F4] border-[#DCDDD8]')"
          @tap="chooseCurrency(opt.key)"
        >
          <view class="flex-1 min-w-0">
            <text class="text-[13px] font-bold block">{{ opt.title }}</text>
            <text class="text-[11px] block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              {{ opt.desc }}
            </text>
          </view>
          <text v-if="currency === opt.key" class="text-[14px] font-bold text-[#E5B352]">✓</text>
        </view>
      </view>
    </view>

    <!-- 收藏全景/矩阵视图抽屉：每日跟踪 vs 保存本期 -->
    <view
      v-if="showFavoriteSheet"
      class="fixed inset-0 z-50 flex items-end bg-black/60"
      @tap="showFavoriteSheet = false"
    >
      <view
        class="w-full rounded-t-2xl border-t p-4 space-y-3"
        :class="isDark ? 'bg-[#12161B] border-[#29313A] text-[#F1EFEA]' : 'bg-[#FFFFFF] border-[#DCDDD8] text-[#22292E]'"
        @tap.stop
      >
        <view class="flex items-center justify-between">
          <text class="text-[15px] font-bold">
            保存「{{ currentPoolTitle }} · {{ viewMode === 'matrix' ? '周期矩阵' : '行情全景' }}」
          </text>
          <text class="text-[12px] text-[#8D99A4]" @tap="showFavoriteSheet = false">关闭</text>
        </view>

        <view class="grid grid-cols-2 gap-2">
          <view
            class="rounded-xl border p-2.5"
            :class="favoriteMode === 'latest'
              ? (isDark ? 'bg-[#E5B352]/14 border-[#E5B352]' : 'bg-[#8A641F]/10 border-[#8A641F]')
              : (isDark ? 'bg-[#181E25] border-[#29313A]' : 'bg-[#F8F7F4] border-[#DCDDD8]')"
            @tap="favoriteMode = 'latest'"
          >
            <text class="text-[13px] font-bold block">每日跟踪</text>
            <text class="text-[11px] block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              保留筛选与排序，随快照日更新
            </text>
          </view>
          <view
            class="rounded-xl border p-2.5"
            :class="favoriteMode === 'fixed'
              ? (isDark ? 'bg-[#E5B352]/14 border-[#E5B352]' : 'bg-[#8A641F]/10 border-[#8A641F]')
              : (isDark ? 'bg-[#181E25] border-[#29313A]' : 'bg-[#F8F7F4] border-[#DCDDD8]')"
            @tap="favoriteMode = 'fixed'"
          >
            <text class="text-[13px] font-bold block">保存本期</text>
            <text class="text-[11px] block mt-0.5" :class="isDark ? 'text-[#8D99A4]' : 'text-[#626F77]'">
              锁定当前快照（{{ snapshot?.as_of || '—' }}）
            </text>
          </view>
        </view>

        <view>
          <input
            v-model="favoriteNote"
            class="w-full rounded-xl border px-3 py-2 text-[12px]"
            :class="isDark ? 'bg-[#181E25] border-[#29313A] text-[#F1EFEA]' : 'bg-[#F8F7F4] border-[#DCDDD8] text-[#22292E]'"
            placeholder="可选本地备注（仅本地保存，分享链接不会携带）"
            :maxlength="60"
          >
        </view>

        <view class="flex items-center gap-2 pt-1 pb-2">
          <view
            class="flex-1 min-h-[42px] rounded-xl border flex items-center justify-center text-[13px]"
            :class="isDark ? 'border-[#29313A] text-[#A9B3BD]' : 'border-[#DCDDD8] text-[#515F6A]'"
            @tap="handleToggleFavorite(currentPoolId); showFavoriteSheet = false"
          >
            {{ isPoolFavorite ? '取消星标观察池' : '仅星标观察池' }}
          </view>
          <view
            class="flex-[1.3] min-h-[42px] rounded-xl flex items-center justify-center text-[13px] font-bold"
            :class="isDark ? 'bg-[#E5B352] text-[#090C0F]' : 'bg-[#8A641F] text-[#FFFFFF]'"
            @tap="confirmSaveFavoriteView"
          >
            保存视图
          </view>
        </view>
      </view>
    </view>
  </FinanceShell>
</template>
