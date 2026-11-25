import { http } from '@/http'

export interface DataSourceMeta {
  source_name: string
  symbol: string
  as_of: string
  status: 'ok' | 'degraded' | 'offline_seed' | 'error'
  latency_ms?: number
  warning?: string
  updated_at: string
}

export interface MarketQuote {
  symbol: string
  name: string
  category: string
  price: number
  change: number
  change_pct: number
  currency: string
  as_of: string
  meta: DataSourceMeta
}

export interface IndustryMetric {
  sector_id: string
  name: string
  icon?: string
  stage: string
  prosperity_score: number
  price_index: number
  price_change_pct: number
  unit: string
  catalyst: string
  as_of: string
  data_date?: string
  is_reference?: boolean
  meta: DataSourceMeta
}

export interface CycleOverview {
  current_phase: string
  phase_code: number
  dominant_asset: string
  equity_bond_spread: number
  style_preference: string
  summary: string
  data_date?: string
  source_name?: string
  is_static_reference?: boolean
  evidence_notes?: string[]
}

export interface RotationSignal {
  date: string
  ratio: number
  zone: string
  quantile_pct: number
  action: string
  to_high_pct: number
  to_low_pct: number
}

export interface ScoreFactor {
  key: string
  name: string
  value: number
  unit: string
  weight_pct: number
  score: number
  as_of: string
  source: string
  status: 'computed' | 'live' | 'static_reference' | 'stale' | 'unavailable' | string
  note?: string
}

export interface ScoreResult {
  tech: number
  spread: number
  value: number
  style: number
  spread_pp: number
  total: number
  action: string
  multiplier: number
  amount: number
  reduce_ratio: number
  close?: number
  as_of?: string
  has_stale_input?: boolean
  pe?: number
  pe_percentile?: number
  dividend_yield?: number
  bond10y?: number
  rsi70?: number
  b_pct100?: number
  style_percentile?: number
  data_source_note?: string
  rule_summary?: string
  factors?: ScoreFactor[]
}

export interface CombinedSignal {
  date: string
  rotation: RotationSignal
  scoring: ScoreResult
  combined_action: string
  summary_tips: string
}

export interface BacktestSummary {
  start_date: string
  end_date: string
  trading_days: number
  total_invest: number
  final_asset: number
  net_profit: number
  total_return_pct: number
  annualized_return: number
  xirr: number
  max_drawdown: number
  sharpe_ratio: number
  calmar_ratio: number
  zone_distribution: Record<string, number>
  total_rebuy_count: number
  ledger_entries_count: number
  rule_version?: string
  data_source?: string
  execution_timing?: string
  clip_reason?: string
}

export interface BacktestOutput {
  summary: BacktestSummary
  daily_records: any[]
  ledger_entries: any[]
}

export interface PortfolioPosition {
  symbol: string
  signal_symbol?: string
  name: string
  shares: number
  cost_price: number
  current_price: number
  unit_price_1_share?: number
  price_status?: 'live' | 'last_known' | 'cost_fallback' | 'unavailable' | string
  valuation_date?: string
  valuation_source?: string
  market_value: number
  total_cost: number
  unrealized_profit: number
  unrealized_profit_pct: number
}

export interface PortfolioTransaction {
  id?: string
  date: string
  week_key?: string
  event_type:
    | 'account_init'
    | 'cash_deposit'
    | 'cash_pool_deposit'
    | 'weekly_invest'
    | 'cash_pool_rebuy'
    | 'manual_adjustment'
    | 'manual_rebuy_override'
    | string
  target_name: string
  symbol?: string | null
  amount: number
  description: string
  shares?: number
  price?: number
  cash_delta?: number
  cash_pool_delta?: number
  cash_balance_after?: number
  cash_pool_after?: number
  source?: 'user_action' | 'strategy_auto' | 'legacy_v1_migrated' | string
  unverified?: boolean
  [key: string]: any
}

export interface PortfolioOverview {
  account_id: string
  is_demo?: boolean
  total_asset: number
  available_cash: number
  cash_pool: number
  positions_value: number
  total_cost: number
  total_profit: number
  total_profit_pct: number
  valuation_date?: string
  valuation_note?: string
  warnings?: string[]
  positions: PortfolioPosition[]
  recent_transactions: PortfolioTransaction[]
}

// 内存 TTL 缓存与并发请求合并机制（杜绝短时间内的重复网络风暴）
interface CacheEntry<T> {
  data: T
  timestamp: number
}

const memoryCache = new Map<string, CacheEntry<any>>()
const pendingPromises = new Map<string, Promise<any>>()
const bgRefreshingKeys = new Set<string>()
const STORAGE_PREFIX = 'CC_SWR_'

function getLocalCache<T>(key: string): CacheEntry<T> | null {
  try {
    const raw = uni.getStorageSync(STORAGE_PREFIX + key)
    if (raw && typeof raw === 'object' && 'data' in raw && 'timestamp' in raw) {
      return raw as CacheEntry<T>
    }
  } catch {
    // ignore
  }
  return null
}

function setLocalCache<T>(key: string, entry: CacheEntry<T>) {
  try {
    uni.setStorage?.({
      key: STORAGE_PREFIX + key,
      data: entry
    })
  } catch {
    // ignore
  }
}

/**
 * 带有时间戳单调比对的缓存写入器（防慢后台请求倒灌覆盖较新的前台数据）
 */
function updateCacheIfFresher<T>(cacheKey: string, data: T, requestTime: number) {
  const current = memoryCache.get(cacheKey)
  if (!current || requestTime >= current.timestamp) {
    const entry = { data, timestamp: requestTime }
    memoryCache.set(cacheKey, entry)
    setLocalCache(cacheKey, entry)
  }
}

/**
 * 带 TTL（毫秒）缓存、本地 SWR 离线秒开与防重复并发的 fetcher 包装器
 */
async function fetchWithCache<T>(
  cacheKey: string,
  ttlMs: number,
  fetcher: () => Promise<{ data?: T }>,
  forceRefresh = false
): Promise<{ data: T }> {
  const now = Date.now()
  if (!forceRefresh) {
    // 1. 优先从内存缓存读取（0ms）
    const cached = memoryCache.get(cacheKey)
    if (cached && now - cached.timestamp < ttlMs) {
      return { data: cached.data }
    }
    // 2. 检查是否有在途请求（防并发重复请求）
    const pending = pendingPromises.get(cacheKey)
    if (pending) {
      return pending
    }
    // 3. 检查本地二级持久化缓存（二级冷启动 SWR）
    const localCached = getLocalCache<T>(cacheKey)
    if (localCached && now - localCached.timestamp < ttlMs * 10) {
      memoryCache.set(cacheKey, localCached)
      // 仅在无在途后台刷新时触发静默更新，不污染 pendingPromises 返回契约
      if (!bgRefreshingKeys.has(cacheKey)) {
        bgRefreshingKeys.add(cacheKey)
        const reqTime = Date.now()
        ;(async () => {
          try {
            const res = await fetcher()
            if (res && res.data !== undefined) {
              updateCacheIfFresher(cacheKey, res.data, reqTime)
            }
          } catch {
            // ignore background fetch error
          } finally {
            bgRefreshingKeys.delete(cacheKey)
          }
        })()
      }
      return { data: localCached.data }
    }
  }

  const reqTime = Date.now()
  const promise = (async () => {
    try {
      const res = await fetcher()
      if (res && res.data !== undefined) {
        updateCacheIfFresher(cacheKey, res.data, reqTime)
      }
      return res as { data: T }
    } finally {
      pendingPromises.delete(cacheKey)
    }
  })()

  pendingPromises.set(cacheKey, promise)
  return promise
}

export function clearApiCache () {
  memoryCache.clear()
  pendingPromises.clear()
  bgRefreshingKeys.clear()
}

export function getMarketQuotes (category?: string, forceRefresh = false) {
  const key = `market_quotes_${category || 'all'}`
  return fetchWithCache<MarketQuote[]>(
    key,
    30_000, // 30秒缓存
    () => http.get<MarketQuote[]>('/api/v1/markets/quotes', category
? { category }
: undefined, {
meta: { loading: false, toast: false }
}),
    forceRefresh
  )
}

export function getIndustryMetrics (forceRefresh = false) {
  return fetchWithCache<IndustryMetric[]>(
    'industry_metrics',
    60_000, // 1分钟缓存
    () => http.get<IndustryMetric[]>('/api/v1/sectors/metrics', undefined, {
meta: { loading: false, toast: false }
}),
    forceRefresh
  )
}

export function getCycleOverview (forceRefresh = false) {
  return fetchWithCache<CycleOverview>(
    'cycle_overview',
    60_000, // 1分钟缓存
    () => http.get<CycleOverview>('/api/v1/cycles/overview', undefined, {
meta: { loading: false, toast: false }
}),
    forceRefresh
  )
}

export function getCombinedStrategySignal (forceRefresh = false) {
  return fetchWithCache<CombinedSignal>(
    'combined_strategy_signal',
    60_000, // 1分钟缓存
    () => http.get<CombinedSignal>('/api/v1/strategies/combined', undefined, {
meta: { loading: false, toast: false }
}),
    forceRefresh
  )
}

export function runBacktest (params: {
  low_line: number
  high_line: number
  weekly_amount: number
  days: number
}) {
  return http.post<BacktestOutput>('/api/v1/backtests/run', params, {
timeout: 25000,
meta: { loading: false, toast: false }
})
}

export function getPortfolioOverview (forceRefresh = false) {
  return fetchWithCache<PortfolioOverview>(
    'portfolio_overview',
    30_000, // 30秒缓存
    () => http.get<PortfolioOverview>('/api/v1/portfolios/overview', undefined, {
meta: { loading: false, toast: false }
}),
    forceRefresh
  )
}
