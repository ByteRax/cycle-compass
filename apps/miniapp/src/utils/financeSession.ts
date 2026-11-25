import type { FinanceCurrency, FinanceOverviewMode } from '@/utils/financeRoute'

export interface OverviewSessionSnapshot {
  poolId: string
  viewMode: FinanceOverviewMode
  sortField: string
  sortOrder: 'asc' | 'desc'
  currency: FinanceCurrency | 'RAW'
  marketFilter: string
  assetTypeFilter: string
  statusFilter: string
  collapsedGroups: Record<string, boolean>
  loadedCount: number
  anchorAssetId: string
  updatedAt: number
}

const sessionByPool = new Map<string, OverviewSessionSnapshot>()

export function saveOverviewSession(
  snapshot: Omit<OverviewSessionSnapshot, 'updatedAt'> & { updatedAt?: number }
): void {
  sessionByPool.set(snapshot.poolId, {
    ...snapshot,
    collapsedGroups: { ...snapshot.collapsedGroups },
    updatedAt: snapshot.updatedAt ?? Date.now()
  })
}

export function getOverviewSession(poolId: string): OverviewSessionSnapshot | null {
  return sessionByPool.get(poolId) ?? null
}

export function clearOverviewSession(poolId?: string): void {
  if (poolId)
    sessionByPool.delete(poolId)
  else
    sessionByPool.clear()
}
