export const BOTTOM_TAB_KEYS = [
'overview',
'markets',
'research',
'portfolio'
] as const

export type BottomTabKey = (typeof BOTTOM_TAB_KEYS)[number]

export type LegacyTabKey = 'cycle' | 'markets' | 'sectors' | 'portfolio' | BottomTabKey

export interface ResolvedTabTarget {
  tab: BottomTabKey
  marketSegment?: 'quotes' | 'industries'
}

/**
 * 将历史 tab 查询参数映射到新版四栏导航（总览 / 市场 / 研究 / 账本）
 */
export function resolveBottomTabKey(raw?: string | null): ResolvedTabTarget {
  const clean = (raw || '').trim().toLowerCase()
  switch (clean) {
    case 'markets':
      return { tab: 'markets', marketSegment: 'quotes' }
    case 'sectors':
      return { tab: 'markets', marketSegment: 'industries' }
    case 'research':
      return { tab: 'research' }
    case 'portfolio':
      return { tab: 'portfolio' }
    case 'overview':
    case 'cycle':
    default:
      return { tab: 'overview' }
  }
}

/**
 * 评估横向滑动手势是否应切换一级底部 Tab
 */
export function evaluateTabSwipe<T extends string>(
  currentTab: T,
  tabKeys: readonly T[],
  deltaX: number,
  deltaY: number,
  deltaTime: number
): T {
  // 1. 横向滑动位移 >= 30px
  // 2. 横向位移明确主导纵向分量（Math.abs(deltaX) > Math.abs(deltaY) * 1.25，严格防纵向滚动误触）
  // 3. 有效滑动手势时长（50ms <= deltaTime <= 1000ms）
  if (
    Math.abs(deltaX) >= 30
    && Math.abs(deltaX) > Math.abs(deltaY) * 1.25
    && deltaTime >= 50
    && deltaTime <= 1000
  ) {
    const curIdx = tabKeys.indexOf(currentTab)
    if (curIdx === -1)
      return currentTab

    if (deltaX < 0 && curIdx < tabKeys.length - 1)
      return tabKeys[curIdx + 1] ?? currentTab

    if (deltaX > 0 && curIdx > 0)
      return tabKeys[curIdx - 1] ?? currentTab
  }
  return currentTab
}

/**
 * 计算横向滚动容器中选中项居中时的目标 scrollLeft
 */
export function computeTabCenterScrollLeft(
  tabOffsetLeft: number,
  tabWidth: number,
  scrollViewWidth: number
): number {
  const tabCenter = tabOffsetLeft + tabWidth / 2
  const target = tabCenter - scrollViewWidth / 2
  return Math.max(0, target)
}
