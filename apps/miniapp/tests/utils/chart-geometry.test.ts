import { describe, expect, it } from 'vitest'
import {
  buildTimeXMapper,
  computeNiceYRange,
  computeSquarifiedTreemap,
  deterministicAssetColor,
  lttbDownsample,
  resolveLabelCollisions,
  splitSegmentsByGap,
} from '@/utils/chartGeometry'

describe('chartGeometry', () => {
  describe('computeNiceYRange', () => {
    it('handles empty array and all-zero values safely', () => {
      const empty = computeNiceYRange([])
      expect(empty.yMin).toBe(0)
      expect(empty.yMax).toBeGreaterThan(0)
      expect(empty.step).toBeGreaterThan(0)
      expect(empty.ticks).toContain(0)

      const zeros = computeNiceYRange([0, 0, 0])
      expect(zeros.yMin).toBe(0)
      expect(zeros.yMax).toBeGreaterThan(0)
      expect(zeros.ticks).toContain(0)
    })

    it('covers negative, positive, and huge-span ranges cleanly', () => {
      const mixed = computeNiceYRange([-42, -10, 18])
      expect(mixed.yMin).toBeLessThanOrEqual(-42)
      expect(mixed.yMax).toBeGreaterThanOrEqual(18)
      expect(mixed.ticks).toContain(0)

      const huge = computeNiceYRange([12, 9850])
      expect(huge.yMin).toBe(0)
      expect(huge.yMax).toBeGreaterThanOrEqual(9850)
      expect(huge.ticks.length).toBeGreaterThanOrEqual(3)
      expect(huge.ticks.length).toBeLessThanOrEqual(8)
    })

    it('respects includeZero = false and ignores non-finite values', () => {
      const range = computeNiceYRange([
102,
108,
Number.NaN,
Number.POSITIVE_INFINITY
], false)
      expect(range.yMin).toBeGreaterThanOrEqual(95)
      expect(range.yMin).toBeLessThanOrEqual(102)
      expect(range.yMax).toBeGreaterThanOrEqual(108)
    })
  })

  describe('buildTimeXMapper', () => {
    it('maps dates proportionally by real timestamp even with missing dates', () => {
      const dates = ['2026-01-01', '2026-01-02', '2026-01-11']
      const mapper = buildTimeXMapper(dates, 50, 500)

      expect(mapper.mapX('2026-01-01')).toBeCloseTo(50, 6)
      expect(mapper.mapX('2026-01-11')).toBeCloseTo(550, 6)
      // 1月2日距1月1日为1天，总跨度10天 => 50 + 500 * 0.1 = 100（而非等距序号的 300）
      expect(mapper.mapX('2026-01-02')).toBeCloseTo(100, 6)
      // 未在列表中的中间日期也按真实时间戳映射
      expect(mapper.mapX('2026-01-06')).toBeCloseTo(300, 6)
    })

    it('builds sparse non-overlapping ticks including start and end dates', () => {
      const dates = Array.from({ length: 30 }, (_, i) => {
        const day = String(i + 1).padStart(2, '0')
        return `2026-03-${day}`
      })
      const mapper = buildTimeXMapper(dates, 40, 400, 5)

      expect(mapper.ticks.length).toBeGreaterThanOrEqual(2)
      expect(mapper.ticks.length).toBeLessThanOrEqual(5)
      expect(mapper.ticks[0]?.index).toBe(0)
      expect(mapper.ticks[mapper.ticks.length - 1]?.index).toBe(29)

      for (let i = 1; i < mapper.ticks.length; i++) {
        const prev = mapper.ticks[i - 1]!
        const curr = mapper.ticks[i]!
        expect(curr.x).toBeGreaterThan(prev.x)
      }
    })

    it('handles empty and single-date inputs', () => {
      const empty = buildTimeXMapper([], 20, 300)
      expect(empty.ticks).toEqual([])
      expect(empty.mapX('2026-01-01')).toBe(20)

      const single = buildTimeXMapper(['2026-01-01'], 20, 300)
      expect(single.mapX('2026-01-01')).toBe(20)
      expect(single.ticks).toHaveLength(1)
    })
  })

  describe('splitSegmentsByGap', () => {
    it('splits continuous segments around null values without connecting or zero-filling', () => {
      const points = [
        { date: 'd1', value: null },
        { date: 'd2', value: 10 },
        { date: 'd3', value: 0 },
        { date: 'd4', value: null },
        { date: 'd5', value: null },
        { date: 'd6', value: -5 },
        { date: 'd7', value: 8 },
        { date: 'd8', value: null },
      ]

      const segments = splitSegmentsByGap(points)
      expect(segments).toHaveLength(2)
      expect(segments[0]?.map((p) => p.date)).toEqual(['d2', 'd3'])
      expect(segments[1]?.map((p) => p.date)).toEqual(['d6', 'd7'])
    })

    it('returns empty array when all values are null or input is empty', () => {
      expect(splitSegmentsByGap([])).toEqual([])
      expect(splitSegmentsByGap([
{ value: null },
{ value: null }
])).toEqual([])
    })
  })

  describe('lttbDownsample', () => {
    it('returns original points when length <= threshold', () => {
      const pts = [
        { x: 0, y: 1 },
        { x: 1, y: 3 },
        { x: 2, y: 2 },
      ]
      expect(lttbDownsample(pts, 3)).toEqual(pts)
      expect(lttbDownsample(pts, 5)).toEqual(pts)
    })

    it('preserves endpoints and prominent visual extremum', () => {
      const pts = Array.from({ length: 100 }, (_, i) => ({
        x: i,
        y: i === 42
? 1000
: Math.sin(i * 0.1) * 5,
      }))

      const sampled = lttbDownsample(pts, 12)
      expect(sampled).toHaveLength(12)
      expect(sampled[0]).toEqual(pts[0])
      expect(sampled[sampled.length - 1]).toEqual(pts[pts.length - 1])
      expect(sampled.some((p) => p.x === 42 && p.y === 1000)).toBe(true)
    })
  })

  describe('resolveLabelCollisions', () => {
    it('pushes overlapping labels apart within bounds', () => {
      const resolved = resolveLabelCollisions(
        [
          {
id: 'a',
rawY: 100,
text: 'A',
color: '#111'
},
          {
id: 'b',
rawY: 102,
text: 'B',
color: '#222'
},
          {
id: 'c',
rawY: 104,
text: 'C',
color: '#333'
},
          {
id: 'd',
rawY: 196,
text: 'D',
color: '#444'
},
          {
id: 'e',
rawY: 198,
text: 'E',
color: '#555'
},
        ],
        20,
        200,
        16,
      )

      expect(resolved).toHaveLength(5)
      for (let i = 0; i < resolved.length; i++) {
        const cur = resolved[i]!
        expect(cur.slotY).toBeGreaterThanOrEqual(20)
        expect(cur.slotY).toBeLessThanOrEqual(200)
        if (i > 0) {
          expect(cur.slotY - resolved[i - 1]!.slotY).toBeGreaterThanOrEqual(16 - 1e-6)
        }
      }
    })

    it('compresses spacing evenly when total height is smaller than required minGap sum', () => {
      const resolved = resolveLabelCollisions(
        [
          {
id: 'a',
rawY: 10,
text: 'A',
color: '#111'
},
          {
id: 'b',
rawY: 12,
text: 'B',
color: '#222'
},
          {
id: 'c',
rawY: 14,
text: 'C',
color: '#333'
},
        ],
        0,
        20,
        16,
      )
      expect(resolved.map((r) => r.slotY)).toEqual([0, 10, 20])
    })
  })

  describe('computeSquarifiedTreemap', () => {
    it('allocates areas strictly proportional to weights and filters non-positive items', () => {
      const rect = {
x: 10,
y: 20,
width: 300,
height: 200
}
      const totalContainerArea = rect.width * rect.height
      const items = [
        { id: 'a', weight: 60 },
        { id: 'b', weight: 30 },
        { id: 'c', weight: 10 },
        { id: 'zero', weight: 0 },
        { id: 'neg', weight: -15 },
      ]

      const nodes = computeSquarifiedTreemap(items, rect)
      expect(nodes.map((n) => n.id).sort()).toEqual(['a', 'b', 'c'])

      const areaSum = nodes.reduce((acc, n) => acc + n.width * n.height, 0)
      expect(areaSum).toBeCloseTo(totalContainerArea, 5)

      const nodeA = nodes.find((n) => n.id === 'a')!
      const nodeB = nodes.find((n) => n.id === 'b')!
      const nodeC = nodes.find((n) => n.id === 'c')!
      expect(nodeA.width * nodeA.height).toBeCloseTo(totalContainerArea * 0.6, 5)
      expect(nodeB.width * nodeB.height).toBeCloseTo(totalContainerArea * 0.3, 5)
      expect(nodeC.width * nodeC.height).toBeCloseTo(totalContainerArea * 0.1, 5)

      for (const node of nodes) {
        expect(node.x).toBeGreaterThanOrEqual(rect.x - 1e-6)
        expect(node.y).toBeGreaterThanOrEqual(rect.y - 1e-6)
        expect(node.x + node.width).toBeLessThanOrEqual(rect.x + rect.width + 1e-6)
        expect(node.y + node.height).toBeLessThanOrEqual(rect.y + rect.height + 1e-6)
      }
    })

    it('marks tiny blocks with isTooSmall', () => {
      const nodes = computeSquarifiedTreemap(
        [
          { id: 'major', weight: 995 },
          { id: 'tiny', weight: 5 },
        ],
        { x: 0, y: 0, width: 200, height: 100 },
        0.01,
      )
      const tiny = nodes.find((n) => n.id === 'tiny')!
      const major = nodes.find((n) => n.id === 'major')!
      expect(tiny.isTooSmall).toBe(true)
      expect(major.isTooSmall).toBe(false)
    })
  })

  describe('deterministicAssetColor', () => {
    it('returns deterministic color for the same assetId and avoids collisions when usedColors is provided', () => {
      const c1 = deterministicAssetColor('csi300', 0)
      const c2 = deterministicAssetColor('csi300', 5)
      expect(c1).toBe(c2)

      const used = new Set<string>([c1])
      const fallback = deterministicAssetColor('csi300', 1, used)
      expect(fallback).not.toBe(c1)
    })
  })
})
