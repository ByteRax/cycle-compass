import type { ChartType } from '@/api/types/trend'

export type FinanceSnapshotMode = 'latest' | 'fixed'
export type FinanceCurrency = 'CNY' | 'USD' | 'HKD' | 'RAW' | 'ORIGINAL'
export type FinanceRange = '1M' | '3M' | '6M' | '1Y' | '3Y' | '5Y' | 'ALL'
export type FinanceReportTheme = 'light' | 'dark'
export type FinanceOverviewMode = 'quotes' | 'matrix'

export interface FinanceRouteState {
  asset_id: string
  scope_id: string
  currency: FinanceCurrency
  range: FinanceRange
  chart: ChartType
  benchmark_id: string
  snapshot_mode: FinanceSnapshotMode
  snapshot_id: string
  start_date?: string
  end_date?: string
  view_mode?: FinanceOverviewMode
  sort?: string
  order?: 'asc' | 'desc'
  market?: string
  asset_type?: string
  status?: string
  theme?: FinanceReportTheme
  title?: string
}

export interface FinanceSourceContext {
  asset_id?: string
  scope_id?: string
  currency?: string
  range?: string
  chart?: string
  benchmark_id?: string
  snapshot_mode?: string
  snapshot_id?: string
  start_date?: string
  end_date?: string
  view_mode?: string
  sort?: string
  order?: string
  market?: string
  asset_type?: string
  status?: string
  theme?: string
}

export interface FinanceUserPrefs {
  defaultHomePool?: string
  defaultCurrency?: string
  defaultRange?: string
  reportThemePref?: string
}

export const VALID_FINANCE_CHARTS: readonly ChartType[] = [
  'matrix',
  'curve',
  'annual',
  'components',
  'dashboard',
  'season',
  'distribution',
  'tier'
]

export const VALID_FINANCE_RANGES: readonly FinanceRange[] = [
'1M',
'3M',
'6M',
'1Y',
'3Y',
'5Y',
'ALL'
]
export const VALID_FINANCE_CURRENCIES: readonly FinanceCurrency[] = ['CNY', 'USD', 'HKD', 'RAW', 'ORIGINAL']

export const LEGACY_VIEW_TO_CHART: Record<number, ChartType> = {
  0: 'curve',
  1: 'curve',
  2: 'annual',
  3: 'components',
  4: 'dashboard',
  5: 'season',
  6: 'distribution',
  7: 'tier'
}

export const LEGACY_ID_TO_TARGET: Record<string, { asset_id: string, scope_id: string }> = {
  a_shares: { asset_id: 'CN.SSE.000300', scope_id: 'universe_a_sector' },
  etf_macro: { asset_id: 'CN.SSE.510300', scope_id: 'universe_macro' },
  dividend_etf: { asset_id: 'CN.CSI.930955', scope_id: 'universe_dividend_etf' },
  a_shares_long: { asset_id: 'CN.SSE.000300', scope_id: 'universe_a_sector' },
  universe_a_shares: { asset_id: 'universe_a_sector', scope_id: 'universe_a_sector' }
}

const PRIVATE_SHARE_KEYS = new Set([
  'note',
  'private_note',
  'remark',
  'favorite_id',
  'memo',
  'custom_note'
])

export function safeDecodeURIComponent(raw: unknown): string {
  if (raw === null || raw === undefined)
    return ''
  const str = String(raw).trim()
  if (!str)
    return ''
  try {
    return decodeURIComponent(str)
  }
  catch {
    return str
  }
}

export function normalizeCurrency(raw: unknown, fallback: FinanceCurrency = 'CNY'): FinanceCurrency {
  const upper = safeDecodeURIComponent(raw).toUpperCase()
  if (upper === 'RAW' || upper === 'ORIGINAL')
    return 'ORIGINAL'
  if (upper === 'CNY' || upper === 'USD' || upper === 'HKD')
    return upper
  return fallback
}

export function toApiCurrency(currency: string): string {
  const norm = normalizeCurrency(currency)
  return norm === 'RAW'
? 'ORIGINAL'
: norm
}

export function toDisplayCurrency(currency: string): 'CNY' | 'USD' | 'HKD' | 'RAW' {
  const norm = normalizeCurrency(currency)
  return norm === 'ORIGINAL'
? 'RAW'
: norm
}

export function normalizeRange(raw: unknown, fallback: FinanceRange = '1Y'): FinanceRange {
  const upper = safeDecodeURIComponent(raw).toUpperCase()
  if ((VALID_FINANCE_RANGES as readonly string[]).includes(upper))
    return upper as FinanceRange
  return fallback
}

export function normalizeChart(raw: unknown, legacyView: unknown, fallback: ChartType = 'curve'): ChartType {
  const lower = safeDecodeURIComponent(raw).toLowerCase()
  if ((VALID_FINANCE_CHARTS as readonly string[]).includes(lower))
    return lower as ChartType

  if (legacyView !== undefined && legacyView !== null && String(legacyView).trim() !== '') {
    const viewNum = Number.parseInt(safeDecodeURIComponent(legacyView), 10)
    if (!Number.isNaN(viewNum) && LEGACY_VIEW_TO_CHART[viewNum])
      return LEGACY_VIEW_TO_CHART[viewNum]
  }
  return fallback
}

export function normalizeSnapshotMode(
  rawMode: unknown,
  snapshotId: string,
  fallback: FinanceSnapshotMode = 'latest'
): FinanceSnapshotMode {
  const lower = safeDecodeURIComponent(rawMode).toLowerCase()
  if (lower === 'fixed' || lower === 'latest')
    return lower
  if (snapshotId)
    return 'fixed'
  return fallback
}

export function resolveLegacyAssetAndScope(
  rawAssetId: string,
  rawScopeId: string
): { asset_id: string, scope_id: string } {
  const cleanAsset = rawAssetId.trim()
  const cleanScope = rawScopeId.trim()
  const legacyAsset = LEGACY_ID_TO_TARGET[cleanAsset]
  const legacyScope = LEGACY_ID_TO_TARGET[cleanScope]

  const asset_id = legacyAsset ? legacyAsset.asset_id : cleanAsset
  const scope_id = cleanScope
    ? (legacyScope ? legacyScope.scope_id : cleanScope)
    : (legacyAsset ? legacyAsset.scope_id : '')

  return { asset_id, scope_id }
}

/**
 * 统一路由参数解码：优先级 显式 query > 来源页面 context > 用户偏好 > 安全默认值。
 * 对所有枚举和百分号编码做安全校验，并兼容旧 id / pool_id / category_id / view 参数。
 */
export function decodeFinanceQuery(
  query: Record<string, any> | undefined | null,
  sourceContext?: FinanceSourceContext,
  userPrefs?: FinanceUserPrefs,
  defaults?: Partial<FinanceRouteState>
): FinanceRouteState {
  const q = query || {}
  const src = sourceContext || {}

  const pickRaw = (...keys: string[]) => {
    for (const k of keys) {
      const val = q[k]
      if (val !== undefined && val !== null && String(val).trim() !== '')
        return safeDecodeURIComponent(val)
    }
    return ''
  }

  const defaultPool = defaults?.scope_id || userPrefs?.defaultHomePool || 'universe_macro'
  const rawAsset = pickRaw('asset_id', 'id') || src.asset_id || defaults?.asset_id || defaultPool
  const rawScope = pickRaw('scope_id', 'pool_id', 'category_id') || src.scope_id || defaults?.scope_id || ''
  const resolvedIds = resolveLegacyAssetAndScope(rawAsset, rawScope)

  const prefCurrency = normalizeCurrency(userPrefs?.defaultCurrency, 'CNY')
  const defaultCurrency = normalizeCurrency(defaults?.currency, prefCurrency)
  const currency = normalizeCurrency(
    pickRaw('currency') || src.currency,
    defaultCurrency
  )

  const prefRange = normalizeRange(userPrefs?.defaultRange, '1Y')
  const defaultRange = normalizeRange(defaults?.range, prefRange)
  const range = normalizeRange(
    pickRaw('range') || src.range,
    defaultRange
  )

  const chart = normalizeChart(
    pickRaw('chart') || src.chart,
    q['view'],
    defaults?.chart || 'curve'
  )

  const benchmark_id = pickRaw('benchmark_id') || src.benchmark_id || defaults?.benchmark_id || ''

  const rawMode = pickRaw('snapshot_mode') || src.snapshot_mode
  const rawSnapshotId = pickRaw('snapshot_id') || src.snapshot_id || defaults?.snapshot_id || ''
  const snapshot_mode = normalizeSnapshotMode(rawMode, rawSnapshotId, defaults?.snapshot_mode || 'latest')
  const snapshot_id = snapshot_mode === 'fixed'
? rawSnapshotId
: ''

  const start_date = pickRaw('start_date') || src.start_date || defaults?.start_date || ''
  const end_date = pickRaw('end_date') || src.end_date || defaults?.end_date || ''

  const rawViewMode = pickRaw('view_mode') || src.view_mode || defaults?.view_mode
  const view_mode: FinanceOverviewMode | undefined = rawViewMode === 'matrix' || rawViewMode === 'quotes'
    ? rawViewMode
    : undefined

  const sort = pickRaw('sort') || src.sort || defaults?.sort || ''
  const rawOrder = pickRaw('order') || src.order || defaults?.order
  const order: 'asc' | 'desc' | undefined = rawOrder === 'asc' || rawOrder === 'desc'
? rawOrder
: undefined

  const market = pickRaw('market') || src.market || defaults?.market || ''
  const asset_type = pickRaw('asset_type') || src.asset_type || defaults?.asset_type || ''
  const status = pickRaw('status') || src.status || defaults?.status || ''

  const rawTheme = pickRaw('theme') || src.theme || defaults?.theme || userPrefs?.reportThemePref
  const theme: FinanceReportTheme | undefined = rawTheme === 'light' || rawTheme === 'dark'
? rawTheme
: undefined
  const title = pickRaw('title') || defaults?.title || ''

  return {
    asset_id: resolvedIds.asset_id,
    scope_id: resolvedIds.scope_id || defaultPool,
    currency,
    range,
    chart,
    benchmark_id,
    snapshot_mode,
    snapshot_id,
    ...(start_date
? { start_date }
: {}),
    ...(end_date
? { end_date }
: {}),
    ...(view_mode
? { view_mode }
: {}),
    ...(sort
? { sort }
: {}),
    ...(order
? { order }
: {}),
    ...(market
? { market }
: {}),
    ...(asset_type
? { asset_type }
: {}),
    ...(status
? { status }
: {}),
    ...(theme
? { theme }
: {}),
    ...(title
? { title }
: {})
  }
}

export function encodeFinanceParams(
  state: Partial<FinanceRouteState>,
  options?: { forShare?: boolean, includeOptionalFilters?: boolean }
): string {
  const entries: Array<[string, string]> = []
  const mode = normalizeSnapshotMode(state.snapshot_mode, state.snapshot_id || '', 'latest')

  const pushIfPresent = (key: string, val: unknown) => {
    if (PRIVATE_SHARE_KEYS.has(key))
      return
    if (val === undefined || val === null)
      return
    const str = String(val).trim()
    if (!str)
      return
    entries.push([key, str])
  }

  pushIfPresent('asset_id', state.asset_id)
  pushIfPresent('scope_id', state.scope_id)
  if (state.currency)
    pushIfPresent('currency', toApiCurrency(state.currency))
  if (state.range)
    pushIfPresent('range', normalizeRange(state.range))
  if (state.chart)
    pushIfPresent('chart', normalizeChart(state.chart, undefined))
  pushIfPresent('benchmark_id', state.benchmark_id)
  pushIfPresent('snapshot_mode', mode)
  if (mode === 'fixed' && state.snapshot_id)
    pushIfPresent('snapshot_id', state.snapshot_id)
  pushIfPresent('start_date', state.start_date)
  pushIfPresent('end_date', state.end_date)

  if (options?.includeOptionalFilters !== false) {
    pushIfPresent('view_mode', state.view_mode)
    pushIfPresent('sort', state.sort)
    pushIfPresent('order', state.order)
    pushIfPresent('market', state.market)
    pushIfPresent('asset_type', state.asset_type)
    pushIfPresent('status', state.status)
    pushIfPresent('theme', state.theme)
    if (!options?.forShare)
      pushIfPresent('title', state.title)
  }

  return entries
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
    .join('&')
}

export function buildFinanceUrl(
  path: string,
  state: Partial<FinanceRouteState>,
  options?: { forShare?: boolean, includeOptionalFilters?: boolean }
): string {
  const qs = encodeFinanceParams(state, options)
  return qs
? `${path}?${qs}`
: path
}

export function buildFinanceSharePath(
  path: string,
  state: Partial<FinanceRouteState>,
  forceFixedSnapshotId?: string
): string {
  const shareState: Partial<FinanceRouteState> = {
    ...state
  }
  if (forceFixedSnapshotId) {
    shareState.snapshot_mode = 'fixed'
    shareState.snapshot_id = forceFixedSnapshotId
  }
  else if (shareState.snapshot_mode === 'latest') {
    shareState.snapshot_id = ''
  }
  return buildFinanceUrl(path, shareState, { forShare: true, includeOptionalFilters: true })
}
