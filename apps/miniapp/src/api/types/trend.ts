import type {
  TrendConfirmation,
  TrendDirection,
  TrendSide,
  TrendTemperature,
  TrendTier
} from '@/utils/trendDict'

export type ChartType =
  | 'curve'
  | 'annual'
  | 'components'
  | 'dashboard'
  | 'season'
  | 'distribution'
  | 'tier'

export type SnapshotStatus =
  | 'ok'
  | 'degraded'
  | 'stale'
  | 'offline_seed'
  | 'error'
  | 'insufficient'
  | 'not_covered'

export interface SourceDateMeta {
  market: string
  as_of: string
  source: string
  status?: SnapshotStatus
}

export interface SnapshotMeta {
  snapshot_id: string
  as_of: string
  computed_at: string
  algorithm_version: string
  currency: string
  scope_id: string
  scope_label: string
  candidate_count: number
  valid_count: number
  coverage: number | null
  mode: 'latest_available' | 'aligned'
  status: SnapshotStatus
  source_dates: SourceDateMeta[]
  warnings: string[]
}

export interface StrengthInfo {
  scope_id: string
  scope_label: string
  score: number | null
  rank: number | null
  raw_factor: number | null
  valid_count: number
  candidate_count: number
  delta: number | null
  reason?: string
}

export interface TrendEventItem {
  event_id: string
  asset_id: string
  event_type: string
  old_value: string
  new_value: string
  as_of: string
  detected_at: string
  snapshot_id: string
  algorithm_version: string
  scope_id: string
}

export interface TrendMetrics {
  temperature: TrendTemperature
  temperature_score: number | null
  direction: TrendDirection
  confirmation: TrendConfirmation
  side: TrendSide
  started_at: string | null
  confirmed_at: string | null
  duration_trading_days: number | null
  duration_calendar_days: number | null
  phase: string | null
  strength: StrengthInfo | null
  tier: TrendTier | null
  events: TrendEventItem[]
  reason?: string
}

export interface AssetSummaryItem {
  asset_id: string
  market: string
  symbol: string
  name: string
  short_name: string
  asset_type: 'index' | 'etf' | 'stock' | 'future' | 'bond' | 'fx' | 'category' | 'universe'
  measure_type: 'price' | 'nav' | 'yield_pct' | 'fx_rate' | 'index'
  unit: string
  currency: string
  as_of: string | null
  latest_value: number | null
  day_change_pct: number | null
  turnover: number | null
  scale_value: number | null
  scale_label: string
  group_name?: string
  has_members: boolean
  member_count: number
  right_ratio_pct: number | null
  right_count: number
  valid_count: number
  license_status: string
  quality_status: SnapshotStatus
  metrics: TrendMetrics
  feature_tags: string[]
}

export interface CatalogResponse {
  snapshot: SnapshotMeta
  universes: Array<{
    universe_id: string
    name: string
    description: string
    member_count: number
  }>
  categories: Array<{
    category_id: string
    parent_id: string | null
    name: string
    level: number
    has_index: boolean
    index_asset_id: string | null
  }>
  overview: {
    right_count: number
    left_count: number
    other_count: number
    valid_count: number
    total_count: number
    right_ratio_pct: number | null
    left_ratio_pct: number | null
    other_ratio_pct: number | null
  }
}

export interface MembersResponse {
  snapshot: SnapshotMeta
  parent_id: string
  parent_name: string
  items: AssetSummaryItem[]
  next_cursor: string | null
  total: number
}

export interface AssetDetailResponse {
  snapshot: SnapshotMeta
  asset: AssetSummaryItem
  breadcrumbs: Array<{ id: string, name: string }>
  available_charts: ChartType[]
  available_ranges: string[]
  available_currencies: string[]
  benchmark_id: string | null
  benchmark_name: string | null
  period_changes: {
    month_1_pct: number | null
    year_1_pct: number | null
  }
  recent_events: TrendEventItem[]
}

export interface SeriesPoint {
  ts: string
  trading_date: string
  value: number | null
  return_pct: number | null
  turnover: number | null
  status: string
}

export interface ExtremumInfo {
  max_drawdown_pct: number | null
  drawdown_peak_date: string | null
  drawdown_trough_date: string | null
  max_gain_pct: number | null
  gain_trough_date: string | null
  gain_peak_date: string | null
  high_value: number | null
  high_date: string | null
  low_value: number | null
  low_date: string | null
}

export interface StateInterval {
  start_date: string
  end_date: string
  side: TrendSide
}

export interface TrendCurveResponse {
  snapshot: SnapshotMeta
  asset_id: string
  asset_name: string
  benchmark_id: string | null
  benchmark_name: string | null
  range: string
  freq: string
  points: SeriesPoint[]
  benchmark_points: SeriesPoint[]
  extremum: ExtremumInfo
  state_intervals: StateInterval[]
  events: TrendEventItem[]
}

export interface AnnualYearSeries {
  year: number
  is_current: boolean
  is_complete: boolean
  same_week_return_pct: number | null
  full_year_return_pct: number | null
  weekly_returns: Array<{
    week: number
    end_date: string
    return_pct: number | null
  }>
}

export interface AnnualTrendResponse {
  snapshot: SnapshotMeta
  asset_id: string
  asset_name: string
  cutoff_week: number
  years: AnnualYearSeries[]
  stats: {
    sample_count: number
    current_year_return_pct: number | null
    same_week_median_pct: number | null
    same_week_mean_pct: number | null
    current_rank: number | null
    up_ratio_pct: number | null
  }
}

export interface ComponentTrendSeries {
  asset_id: string
  name: string
  color: string
  is_benchmark: boolean
  temperature: TrendTemperature
  latest_value: number | null
  period_change_pct: number | null
  turnover: number | null
  total_return_pct: number | null
  points: Array<number | null>
  reason?: string
}

export interface ComponentTrendResponse {
  snapshot: SnapshotMeta
  asset_id: string
  asset_name: string
  range: string
  timeline: string[]
  series: ComponentTrendSeries[]
}

export interface DashboardRow {
  asset: AssetSummaryItem
  change_pct: number | null
  sparkline: Array<number | null>
  spark_min: number | null
  spark_max: number | null
}

export interface TrendDashboardResponse {
  snapshot: SnapshotMeta
  asset_id: string
  asset_name: string
  change_period: string
  spark_range: string
  groups: Array<{
    group_name: string
    rows: DashboardRow[]
  }>
}

export interface SeasonStageBucket {
  stage_key: string
  stage_label: string
  direction: 'up' | 'down' | 'neutral'
  rule_desc: string
  count: number
  members: Array<{
    asset_id: string
    name: string
    duration_trading_days: number | null
    temperature: TrendTemperature
    side: TrendSide
  }>
}

export interface TrendSeasonResponse {
  snapshot: SnapshotMeta
  asset_id: string
  asset_name: string
  up_count: number
  down_count: number
  other_count: number
  valid_count: number
  stages: SeasonStageBucket[]
}

export interface DistributionNode {
  asset_id: string
  name: string
  group_name: string
  weight_value: number
  weight_pct: number
  side: TrendSide
  temperature: TrendTemperature
  day_change_pct: number | null
  duration_trading_days: number | null
  strength_score: number | null
}

export interface TrendDistributionResponse {
  snapshot: SnapshotMeta
  asset_id: string
  asset_name: string
  weight_by: 'float_market_cap' | 'turnover' | 'equal'
  color_by: 'side' | 'temperature' | 'change'
  right_count: number
  valid_count: number
  unknown_count: number
  right_ratio_pct: number | null
  weighted_right_ratio_pct: number | null
  weight_coverage_pct: number | null
  nodes: DistributionNode[]
}

export interface TierBucket {
  tier: TrendTier
  pro_label: string
  fun_label: string
  rule_text: string
  count: number
  items: Array<{
    asset_id: string
    name: string
    score: number
    side: TrendSide
    duration_trading_days: number | null
    delta: number | null
    temperature: TrendTemperature
  }>
}

export interface TrendTierResponse {
  snapshot: SnapshotMeta
  asset_id: string
  asset_name: string
  candidate_count: number
  valid_count: number
  excluded_count: number
  tiers: TierBucket[]
  excluded_items: Array<{
    asset_id: string
    name: string
    reason: string
  }>
}

export interface ChartContext {
  asset_id: string
  pool_id: string
  snapshot_id: string
  chart: ChartType
  currency: string
  range: string
  benchmark_id: string
  change_period: string
  spark_range: string
  weight_by: 'float_market_cap' | 'turnover' | 'equal'
  color_by: 'side' | 'temperature' | 'change'
  sort: string
  order: 'asc' | 'desc'
}

export interface DataStatusResponse {
  snapshot: SnapshotMeta
  markets: Array<{
    market: string
    label: string
    as_of: string | null
    source_name: string
    status: SnapshotStatus
    covered_count: number
    total_count: number
    note: string
  }>
}

export interface MethodologyResponse {
  algorithm_version: string
  temperature_thresholds: Array<{
    key: TrendTemperature
    label: string
    min: number
    max: number
    desc: string
  }>
  tier_rules: Array<{
    tier: TrendTier
    rule: string
  }>
  disclaimers: string[]
}
