export type TrendTemperature =
  | 'boiling'
  | 'hot'
  | 'warm'
  | 'flat'
  | 'cool'
  | 'cold'
  | 'freezing'
  | 'unavailable'

export type TrendDirection = 'up' | 'down' | 'neutral' | 'unknown'
export type TrendConfirmation = 'forming' | 'confirmed' | 'none' | 'unknown'
export type TrendSide = 'right' | 'left' | 'neutral' | 'unknown'
export type TrendTier = 'S' | 'A' | 'B' | 'C' | 'D'

export interface TrendTempMeta {
  key: TrendTemperature
  label: string
  shortDesc: string
  color: string
  lightColor: string
  bg: string
  border: string
  rangeText: string
}

export const TREND_TEMP_META: Record<TrendTemperature, TrendTempMeta> = {
  boiling: {
    key: 'boiling',
    label: '沸',
    shortDesc: '极强',
    color: '#FF2E51',
    lightColor: '#B91C1C',
    bg: 'rgba(255, 46, 81, 0.20)',
    border: 'rgba(255, 46, 81, 0.45)',
    rangeText: '86–100'
  },
  hot: {
    key: 'hot',
    label: '热',
    shortDesc: '强势',
    color: '#FF736A',
    lightColor: '#B93842',
    bg: 'rgba(255, 115, 106, 0.18)',
    border: 'rgba(255, 115, 106, 0.38)',
    rangeText: '72–86'
  },
  warm: {
    key: 'warm',
    label: '温',
    shortDesc: '偏强',
    color: '#E5B352',
    lightColor: '#8A641F',
    bg: 'rgba(229, 179, 82, 0.18)',
    border: 'rgba(229, 179, 82, 0.38)',
    rangeText: '58–72'
  },
  flat: {
    key: 'flat',
    label: '平',
    shortDesc: '中性',
    color: '#94A3B8',
    lightColor: '#64707C',
    bg: 'rgba(148, 163, 184, 0.14)',
    border: 'rgba(148, 163, 184, 0.30)',
    rangeText: '43–58'
  },
  cool: {
    key: 'cool',
    label: '凉',
    shortDesc: '偏弱',
    color: '#2EC4B6',
    lightColor: '#0F766E',
    bg: 'rgba(46, 196, 182, 0.18)',
    border: 'rgba(46, 196, 182, 0.38)',
    rangeText: '29–43'
  },
  cold: {
    key: 'cold',
    label: '寒',
    shortDesc: '下行',
    color: '#38BDF8',
    lightColor: '#0369A1',
    bg: 'rgba(56, 189, 248, 0.18)',
    border: 'rgba(56, 189, 248, 0.38)',
    rangeText: '15–29'
  },
  freezing: {
    key: 'freezing',
    label: '冻',
    shortDesc: '极弱',
    color: '#818CF8',
    lightColor: '#4338CA',
    bg: 'rgba(129, 140, 248, 0.20)',
    border: 'rgba(129, 140, 248, 0.45)',
    rangeText: '0–15'
  },
  unavailable: {
    key: 'unavailable',
    label: '—',
    shortDesc: '暂无',
    color: '#8D99A4',
    lightColor: '#626F77',
    bg: 'rgba(141, 153, 164, 0.12)',
    border: 'rgba(141, 153, 164, 0.25)',
    rangeText: '样本不足'
  }
}

export function getTrendTempMeta(temp?: string | null): TrendTempMeta {
  if (!temp)
    return TREND_TEMP_META.unavailable
  if (temp === 'fair')
    return TREND_TEMP_META.flat
  if (temp in TREND_TEMP_META)
    return TREND_TEMP_META[temp as TrendTemperature]
  return TREND_TEMP_META.unavailable
}

export interface TierMeta {
  key: TrendTier
  proLabel: string
  funLabel: string
  ruleText: string
  color: string
}

export const TIER_META_MAP: Record<TrendTier, TierMeta> = {
  S: {
    key: 'S',
    proLabel: 'S 领跑',
    funLabel: '夯',
    ruleText: '强度 ≥ 90 且右侧确认',
    color: '#E5B352'
  },
  A: {
    key: 'A',
    proLabel: 'A 强势',
    funLabel: '顶级',
    ruleText: '强度 ≥ 70（或 ≥ 90 非右侧）',
    color: '#FF736A'
  },
  B: {
    key: 'B',
    proLabel: 'B 中坚',
    funLabel: '人上人',
    ruleText: '30 ≤ 强度 < 70',
    color: '#72A7D9'
  },
  C: {
    key: 'C',
    proLabel: 'C 偏弱',
    funLabel: 'NPC',
    ruleText: '10 ≤ 强度 < 30（或 < 10 非左侧）',
    color: '#2EC4B6'
  },
  D: {
    key: 'D',
    proLabel: 'D 落后',
    funLabel: '拉完了',
    ruleText: '强度 < 10 且左侧确认',
    color: '#818CF8'
  }
}

export function formatDurationBadge(
  side: TrendSide,
  confirmation: TrendConfirmation,
  tradingDays: number | null
): { text: string, isRight: boolean, isLeft: boolean } {
  if (side === 'right' && confirmation === 'confirmed') {
    const days = tradingDays ?? 1
    return { text: `右侧+${days}`, isRight: true, isLeft: false }
  }
  if (side === 'left' && confirmation === 'confirmed') {
    const days = tradingDays ?? 1
    return { text: `下行-${days}`, isRight: false, isLeft: true }
  }
  if (confirmation === 'forming') {
    return { text: '形成中', isRight: false, isLeft: false }
  }
  if (side === 'neutral') {
    return { text: '—', isRight: false, isLeft: false }
  }
  return { text: '待确认', isRight: false, isLeft: false }
}
