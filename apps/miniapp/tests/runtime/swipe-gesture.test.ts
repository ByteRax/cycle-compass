import { describe, expect, it } from 'vitest'
import {
  BOTTOM_TAB_KEYS,
  computeTabCenterScrollLeft,
  evaluateTabSwipe,
  resolveBottomTabKey
} from '@/utils/swipe'

describe('底部 4 Tab（总览 / 市场 / 研究 / 账本）导航与手势算法', () => {
  it('向左滑动手势：从总览切换到市场', () => {
    const next = evaluateTabSwipe('overview', BOTTOM_TAB_KEYS, -70, 10, 250)
    expect(next).toBe('markets')
  })

  it('向左滑动手势：连续滑动到研究和账本', () => {
    let tab = evaluateTabSwipe('markets', BOTTOM_TAB_KEYS, -80, 5, 200)
    expect(tab).toBe('research')
    tab = evaluateTabSwipe(tab, BOTTOM_TAB_KEYS, -60, 8, 220)
    expect(tab).toBe('portfolio')
  })

  it('向右滑动手势：从账本返回到研究', () => {
    const prev = evaluateTabSwipe('portfolio', BOTTOM_TAB_KEYS, 90, 12, 300)
    expect(prev).toBe('research')
  })

  it('兼容旧版一级 Tab 查询参数：cycle/markets/sectors/portfolio 正确映射（AUDIT 09）', () => {
    expect(resolveBottomTabKey('cycle')).toEqual({ tab: 'overview' })
    expect(resolveBottomTabKey('overview')).toEqual({ tab: 'overview' })
    expect(resolveBottomTabKey('markets')).toEqual({ tab: 'markets', marketSegment: 'quotes' })
    expect(resolveBottomTabKey('sectors')).toEqual({ tab: 'markets', marketSegment: 'industries' })
    expect(resolveBottomTabKey('research')).toEqual({ tab: 'research' })
    expect(resolveBottomTabKey('portfolio')).toEqual({ tab: 'portfolio' })
    expect(resolveBottomTabKey('')).toEqual({ tab: 'overview' })
  })

  it('防误触：纵向滚动页面时（deltaY 远大于 deltaX）不应误切 Tab', () => {
    const result = evaluateTabSwipe('overview', BOTTOM_TAB_KEYS, -30, 150, 400)
    expect(result).toBe('overview')
  })

  it('防误触：临界纵向位移（纵向 40px，横向 35px，纵向大于横向）不应误切 Tab', () => {
    const result = evaluateTabSwipe('overview', BOTTOM_TAB_KEYS, -35, 40, 200)
    expect(result).toBe('overview')
  })

  it('防误触：滑动距离过小（低于 30px）不应触发切换', () => {
    const result = evaluateTabSwipe('overview', BOTTOM_TAB_KEYS, -20, 5, 200)
    expect(result).toBe('overview')
  })

  it('防误触：极短触碰（低于 50ms）或慢速拖动（超过 1000ms）不触发', () => {
    expect(evaluateTabSwipe('overview', BOTTOM_TAB_KEYS, -100, 10, 30)).toBe('overview')
    expect(evaluateTabSwipe('overview', BOTTOM_TAB_KEYS, -100, 10, 1200)).toBe('overview')
  })

  it('边界保护：已经在第一个 Tab（总览）向右滑不会越界', () => {
    const result = evaluateTabSwipe('overview', BOTTOM_TAB_KEYS, 120, 5, 200)
    expect(result).toBe('overview')
  })

  it('边界保护：已经在最后一个 Tab（账本）向左滑不会越界', () => {
    const result = evaluateTabSwipe('portfolio', BOTTOM_TAB_KEYS, -120, 5, 200)
    expect(result).toBe('portfolio')
  })

  it('顶部滚动居中算法：居中偏移计算符合视口边界', () => {
    const offset = computeTabCenterScrollLeft(150, 80, 300)
    expect(offset).toBe(40)
    expect(computeTabCenterScrollLeft(20, 60, 300)).toBe(0)
  })
})
