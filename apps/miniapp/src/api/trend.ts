import type {
  AnnualTrendResponse,
  AssetDetailResponse,
  AssetSummaryItem,
  CatalogResponse,
  ComponentTrendResponse,
  DataStatusResponse,
  MembersResponse,
  MethodologyResponse,
  SnapshotMeta,
  TrendCurveResponse,
  TrendDashboardResponse,
  TrendDistributionResponse,
  TrendEventItem,
  TrendSeasonResponse,
  TrendTierResponse
} from '@/api/types/trend'
import { http } from '@/http'

const requestOpts = {
meta: { loading: false, toast: false }
}

export function getAssetCatalog(params?: { scope_id?: string, currency?: string }) {
  return http.get<CatalogResponse>('/api/v1/assets/catalog', params, requestOpts)
}

export function getAssetUniverses() {
  return http.get<{ items: CatalogResponse['universes'], snapshot: SnapshotMeta }>('/api/v1/assets/universes', undefined, requestOpts)
}

export function searchAssets(params: { q: string, type?: string, cursor?: string, limit?: number }) {
  return http.get<{ items: AssetSummaryItem[], next_cursor: string | null, total: number }>('/api/v1/assets/search', params, requestOpts)
}

export function getAssetDetail(id: string, params?: { scope_id?: string, currency?: string, snapshot_id?: string }) {
  return http.get<AssetDetailResponse>(`/api/v1/assets/${encodeURIComponent(id)}`, params, requestOpts)
}

export function getAssetMembers(
  id: string,
  params?: {
    scope_id?: string
    sort?: string
    order?: 'asc' | 'desc'
    cursor?: string
    limit?: number
    currency?: string
  }
) {
  return http.get<MembersResponse>(`/api/v1/assets/${encodeURIComponent(id)}/members`, params, requestOpts)
}

export function getLatestTrendSnapshot(params?: { scope_id?: string, currency?: string, mode?: string }) {
  return http.get<SnapshotMeta>('/api/v1/trends/snapshots/latest', params, requestOpts)
}

export function getTrendCurve(
  id: string,
  params?: {
    range?: string
    freq?: string
    currency?: string
    benchmark_id?: string
    snapshot_id?: string
  }
) {
  return http.get<TrendCurveResponse>(`/api/v1/trends/${encodeURIComponent(id)}/curve`, params, requestOpts)
}

export function getTrendAnnual(
  id: string,
  params?: {
    years?: string
    currency?: string
    snapshot_id?: string
  }
) {
  return http.get<AnnualTrendResponse>(`/api/v1/trends/${encodeURIComponent(id)}/annual`, params, requestOpts)
}

export function getTrendComponents(
  id: string,
  params?: {
    scope_id?: string
    range?: string
    top_n?: number
    sort?: string
    currency?: string
    snapshot_id?: string
  }
) {
  return http.get<ComponentTrendResponse>(`/api/v1/trends/${encodeURIComponent(id)}/components`, params, requestOpts)
}

export function getTrendDashboard(
  id: string,
  params?: {
    change_period?: string
    spark_range?: string
    sort?: string
    order?: 'asc' | 'desc'
    scope_id?: string
    group?: string
    currency?: string
    snapshot_id?: string
  }
) {
  return http.get<TrendDashboardResponse>(`/api/v1/trends/${encodeURIComponent(id)}/dashboard`, params, requestOpts)
}

export function getTrendSeason(
  id: string,
  params?: {
    direction?: 'all' | 'up' | 'down'
    snapshot_id?: string
  }
) {
  return http.get<TrendSeasonResponse>(`/api/v1/trends/${encodeURIComponent(id)}/season`, params, requestOpts)
}

export function getTrendDistribution(
  id: string,
  params?: {
    weight_by?: string
    color_by?: string
    snapshot_id?: string
  }
) {
  return http.get<TrendDistributionResponse>(`/api/v1/trends/${encodeURIComponent(id)}/distribution`, params, requestOpts)
}

export function getTrendTier(
  id: string,
  params?: {
    scope_id?: string
    snapshot_id?: string
  }
) {
  return http.get<TrendTierResponse>(`/api/v1/trends/${encodeURIComponent(id)}/tier`, params, requestOpts)
}

export function getTrendEvents(
  id: string,
  params?: {
    types?: string
    limit?: number
    cursor?: string
    snapshot_id?: string
  }
) {
  return http.get<{ items: TrendEventItem[], next_cursor: string | null }>(`/api/v1/trends/${encodeURIComponent(id)}/events`, params, requestOpts)
}

export function getTrendMethodology() {
  return http.get<MethodologyResponse>('/api/v1/trends/methodology', undefined, requestOpts)
}

export function getTrendDataStatus() {
  return http.get<DataStatusResponse>('/api/v1/trends/data-status', undefined, requestOpts)
}
