export interface NiceYRangeResult {
  yMin: number
  yMax: number
  step: number
  ticks: number[]
}

export interface TimeXTick {
  index: number
  label: string
  x: number
}

export interface TimeXMapperResult {
  mapX: (dateStr: string) => number
  ticks: TimeXTick[]
  minTs: number
  maxTs: number
}

export interface LabelCollisionTarget {
  id: string
  rawY: number
  text: string
  color: string
}

export interface ResolvedLabelSlot extends LabelCollisionTarget {
  slotY: number
}

export interface TreemapInputItem {
  id: string
  weight: number
}

export interface TreemapRect {
  x: number
  y: number
  width: number
  height: number
}

export interface TreemapNode extends TreemapRect {
  id: string
  weight: number
  areaRatio: number
  isTooSmall: boolean
}

const NICE_MULTIPLIERS = [1, 2, 2.5, 5, 10] as const

function cleanFloat(value: number, step: number): number {
  if (value === 0) 
return 0
  const absStep = Math.abs(step)
  if (absStep === 0 || !Number.isFinite(absStep)) 
return value
  const decimals = Math.max(0, Math.min(10, -Math.floor(Math.log10(absStep)) + 2))
  const rounded = Number(value.toFixed(decimals))
  return rounded === 0
? 0
: rounded
}

function pickNiceStep(rawSpan: number, targetTickCount = 5): number {
  const safeSpan = rawSpan > 0 && Number.isFinite(rawSpan)
? rawSpan
: 1
  const roughStep = safeSpan / Math.max(1, targetTickCount - 1)
  const magnitude = 10 ** Math.floor(Math.log10(roughStep))
  const normalized = roughStep / magnitude

  for (const mult of NICE_MULTIPLIERS) {
    if (normalized <= mult) {
      return cleanFloat(mult * magnitude, mult * magnitude)
    }
  }
  return cleanFloat(10 * magnitude, 10 * magnitude)
}

/**
 * 自适应计算 Y 轴范围、步长与刻度数组（支持全 0、空数组、负值与超大/极小跨度）。
 */
export function computeNiceYRange(values: number[], includeZero = true): NiceYRangeResult {
  const finite = values.filter((v) => Number.isFinite(v))

  if (finite.length === 0) {
    return {
yMin: 0,
yMax: 10,
step: 5,
ticks: [0, 5, 10]
}
  }

  let min = finite[0] ?? 0
  let max = min
  for (let i = 1; i < finite.length; i++) {
    const val = finite[i]
    if (val === undefined) 
continue
    if (val < min) 
min = val
    if (val > max) 
max = val
  }

  if (includeZero) {
    if (min > 0) 
min = 0
    if (max < 0) 
max = 0
  }

  if (min === 0 && max === 0) {
    return {
yMin: 0,
yMax: 10,
step: 5,
ticks: [0, 5, 10]
}
  }

  if (min === max) {
    const pad = Math.abs(min) * 0.1 || 1
    min -= pad
    max += pad
  }

  const rawSpan = max - min
  const step = pickNiceStep(rawSpan, 5)
  const yMin = cleanFloat(Math.floor(min / step) * step, step)
  let yMax = cleanFloat(Math.ceil(max / step) * step, step)

  if (yMax <= yMin) {
    yMax = cleanFloat(yMin + step, step)
  }

  const ticks: number[] = []
  const maxSteps = 20
  for (let i = 0; i <= maxSteps; i++) {
    const tick = cleanFloat(yMin + i * step, step)
    ticks.push(tick)
    if (tick >= yMax - step * 1e-6) 
break
  }

  return { yMin, yMax, step, ticks }
}

function parseDateTimestamp(dateStr: string): number {
  const trimmed = dateStr.trim()
  if (!trimmed) 
return Number.NaN

  const match = /^(\d{4})[-/](\d{1,2})(?:[-/](\d{1,2}))?/.exec(trimmed)
  if (match) {
    const year = Number(match[1])
    const month = Number(match[2])
    const day = match[3] !== undefined
? Number(match[3])
: 1
    if (
      Number.isInteger(year)
      && Number.isInteger(month)
      && Number.isInteger(day)
      && month >= 1
      && month <= 12
      && day >= 1
      && day <= 31
    ) {
      return Date.UTC(year, month - 1, day)
    }
  }

  return Date.parse(trimmed)
}

/**
 * 按真实日期时间戳映射 X 坐标，并生成不重叠的稀疏刻度。
 */
export function buildTimeXMapper(
  dates: string[],
  padLeft: number,
  plotWidth: number,
  maxTicks = 5,
): TimeXMapperResult {
  const safeWidth = Math.max(0, plotWidth)

  if (dates.length === 0) {
    return {
      mapX: () => padLeft,
      ticks: [],
      minTs: 0,
      maxTs: 0,
    }
  }

  const parsedTs = dates.map((d) => parseDateTimestamp(d))
  const validTs = parsedTs.filter((ts) => Number.isFinite(ts))

  if (validTs.length === 0) {
    const count = dates.length
    const mapByIndex = (dateStr: string): number => {
      const idx = dates.indexOf(dateStr)
      if (idx < 0 || count <= 1) 
return padLeft
      return padLeft + (idx / (count - 1)) * safeWidth
    }
    const ticks = pickSparseIndices(count, maxTicks).map((index) => ({
      index,
      label: dates[index] ?? '',
      x: count <= 1
? padLeft
: padLeft + (index / (count - 1)) * safeWidth,
    }))
    return {
mapX: mapByIndex,
ticks,
minTs: 0,
maxTs: 0
}
  }

  let minTs = validTs[0] ?? 0
  let maxTs = minTs
  for (let i = 1; i < validTs.length; i++) {
    const ts = validTs[i]
    if (ts === undefined) 
continue
    if (ts < minTs) 
minTs = ts
    if (ts > maxTs) 
maxTs = ts
  }

  const tsSpan = maxTs - minTs
  const tsToX = (ts: number): number => {
    if (!Number.isFinite(ts) || tsSpan <= 0) 
return padLeft
    return padLeft + ((ts - minTs) / tsSpan) * safeWidth
  }

  const mapX = (dateStr: string): number => {
    const ts = parseDateTimestamp(dateStr)
    if (!Number.isFinite(ts)) {
      const fallbackIdx = dates.indexOf(dateStr)
      if (fallbackIdx < 0 || dates.length <= 1) 
return padLeft
      return padLeft + (fallbackIdx / (dates.length - 1)) * safeWidth
    }
    return tsToX(ts)
  }

  const ticks = buildTimeTicks(dates, parsedTs, minTs, maxTs, tsToX, safeWidth, maxTicks)

  return { mapX, ticks, minTs, maxTs }
}

function pickSparseIndices(length: number, maxTicks: number): number[] {
  if (length <= 0) 
return []
  if (length === 1 || maxTicks <= 1) 
return [0]
  if (length <= maxTicks) {
    return Array.from({ length }, (_, i) => i)
  }

  const indices = new Set<number>([0, length - 1])
  const segments = maxTicks - 1
  for (let i = 1; i < segments; i++) {
    const idx = Math.round((i / segments) * (length - 1))
    indices.add(idx)
  }
  return Array.from(indices).sort((a, b) => a - b)
}

function buildTimeTicks(
  dates: string[],
  parsedTs: number[],
  minTs: number,
  maxTs: number,
  tsToX: (ts: number) => number,
  safeWidth: number,
  maxTicks: number,
): TimeXTick[] {
  const count = dates.length
  if (count === 0) 
return []
  if (count === 1 || maxTicks <= 1) {
    const ts = parsedTs[0]
    return [
      {
        index: 0,
        label: dates[0] ?? '',
        x: ts !== undefined && Number.isFinite(ts)
? tsToX(ts)
: tsToX(minTs),
      },
    ]
  }

  const tsSpan = maxTs - minTs
  if (tsSpan <= 0) {
    return [
      {
index: 0,
label: dates[0] ?? '',
x: tsToX(minTs)
},
    ]
  }

  const targetCount = Math.min(count, Math.max(2, maxTicks))
  const chosenIndices = new Set<number>([0, count - 1])

  for (let k = 1; k < targetCount - 1; k++) {
    const targetTs = minTs + (k / (targetCount - 1)) * tsSpan
    let bestIdx = -1
    let bestDist = Number.POSITIVE_INFINITY
    for (let i = 1; i < count - 1; i++) {
      if (chosenIndices.has(i)) 
continue
      const ts = parsedTs[i]
      if (ts === undefined || !Number.isFinite(ts)) 
continue
      const dist = Math.abs(ts - targetTs)
      if (dist < bestDist) {
        bestDist = dist
        bestIdx = i
      }
    }
    if (bestIdx >= 0) {
      chosenIndices.add(bestIdx)
    }
  }

  const minPixelGap = safeWidth > 0
? Math.min(48, safeWidth / (targetCount * 1.4))
: 0
  const sorted = Array.from(chosenIndices).sort((a, b) => a - b)
  const result: TimeXTick[] = []

  for (const idx of sorted) {
    const ts = parsedTs[idx]
    const x = ts !== undefined && Number.isFinite(ts)
? tsToX(ts)
: tsToX(minTs)
    const label = dates[idx] ?? ''
    const isLast = idx === count - 1
    const prev = result[result.length - 1]

    if (prev && minPixelGap > 0 && x - prev.x < minPixelGap) {
      // 保留终点刻度，必要时替换前一个中间刻度
      if (isLast && result.length > 1) {
        result.pop()
        result.push({ index: idx, label, x })
      }
      continue
    }
    result.push({ index: idx, label, x })
  }

  return result
}

/**
 * 当遇到 `value === null`（或非有限数值）时拆分为多段连续折线，绝不将缺失点连线或当 0。
 */
export function splitSegmentsByGap<T extends { value: number | null }>(points: T[]): T[][] {
  const segments: T[][] = []
  let current: T[] = []

  for (const pt of points) {
    if (pt.value !== null && Number.isFinite(pt.value)) {
      current.push(pt)
    }
    else if (current.length > 0) {
      segments.push(current)
      current = []
    }
  }

  if (current.length > 0) {
    segments.push(current)
  }

  return segments
}

/**
 * Largest-Triangle-Three-Buckets (LTTB) 降采样，保留首尾与视觉特征极值点。
 */
export function lttbDownsample<T extends { x: number; y: number }>(
  points: T[],
  threshold: number,
): T[] {
  const len = points.length
  if (threshold <= 0 || len === 0) 
return []
  if (len <= threshold) 
return [...points]
  if (threshold === 1) {
    const first = points[0]
    return first
? [first]
: []
  }
  if (threshold === 2) {
    const first = points[0]
    const last = points[len - 1]
    if (!first) 
return []
    return last && len > 1
? [first, last]
: [first]
  }

  const sampled: T[] = []
  const firstPoint = points[0]
  if (!firstPoint) 
return []
  sampled.push(firstPoint)

  const bucketSize = (len - 2) / (threshold - 2)
  let prevSelected = firstPoint

  for (let i = 0; i < threshold - 2; i++) {
    const avgRangeStart = Math.floor((i + 1) * bucketSize) + 1
    const avgRangeEnd = Math.min(len, Math.floor((i + 2) * bucketSize) + 1)

    let avgX = 0
    let avgY = 0
    let avgCount = 0
    for (let j = avgRangeStart; j < avgRangeEnd; j++) {
      const pt = points[j]
      if (!pt) 
continue
      avgX += pt.x
      avgY += pt.y
      avgCount++
    }

    if (avgCount > 0) {
      avgX /= avgCount
      avgY /= avgCount
    }
    else {
      const last = points[len - 1]
      if (last) {
        avgX = last.x
        avgY = last.y
      }
    }

    const rangeStart = Math.floor(i * bucketSize) + 1
    const rangeEnd = Math.min(len - 1, Math.floor((i + 1) * bucketSize) + 1)

    let maxArea = -1
    let maxPoint = points[rangeStart] ?? prevSelected

    for (let j = rangeStart; j < rangeEnd; j++) {
      const pt = points[j]
      if (!pt) 
continue
      const area = Math.abs(
        (prevSelected.x - avgX) * (pt.y - prevSelected.y)
        - (prevSelected.x - pt.x) * (avgY - prevSelected.y),
      ) * 0.5
      if (area > maxArea) {
        maxArea = area
        maxPoint = pt
      }
    }

    sampled.push(maxPoint)
    prevSelected = maxPoint
  }

  const lastPoint = points[len - 1]
  if (lastPoint) {
    sampled.push(lastPoint)
  }

  return sampled
}

/**
 * 垂直标签避让算法：保证标签在 `[topBound, bottomBound]` 范围内且不重叠。
 */
export function resolveLabelCollisions(
  targets: LabelCollisionTarget[],
  topBound: number,
  bottomBound: number,
  minGap = 16,
): ResolvedLabelSlot[] {
  if (targets.length === 0) 
return []

  const top = Math.min(topBound, bottomBound)
  const bottom = Math.max(topBound, bottomBound)
  const safeGap = Math.max(0, minGap)

  const sorted: ResolvedLabelSlot[] = [...targets]
    .sort((a, b) => a.rawY - b.rawY)
    .map((item) => ({
      ...item,
      slotY: Math.min(bottom, Math.max(top, item.rawY)),
    }))

  const count = sorted.length
  const availableSpan = bottom - top

  // 物理高度不足以容纳全部 minGap 时，均匀压缩间距以确保不越界且不交叉
  if (count > 1 && availableSpan < (count - 1) * safeGap) {
    const compressedGap = availableSpan / (count - 1)
    for (let i = 0; i < count; i++) {
      const slot = sorted[i]
      if (slot) {
        slot.slotY = top + i * compressedGap
      }
    }
    return sorted
  }

  for (let i = 1; i < count; i++) {
    const prev = sorted[i - 1]
    const curr = sorted[i]
    if (!prev || !curr) 
continue
    if (curr.slotY - prev.slotY < safeGap) {
      curr.slotY = prev.slotY + safeGap
    }
  }

  const last = sorted[count - 1]
  if (last && last.slotY > bottom) {
    last.slotY = bottom
    for (let i = count - 2; i >= 0; i--) {
      const curr = sorted[i]
      const next = sorted[i + 1]
      if (!curr || !next) 
continue
      if (next.slotY - curr.slotY < safeGap) {
        curr.slotY = next.slotY - safeGap
      }
    }
  }

  const first = sorted[0]
  if (first && first.slotY < top) {
    first.slotY = top
    for (let i = 1; i < count; i++) {
      const prev = sorted[i - 1]
      const curr = sorted[i]
      if (!prev || !curr) 
continue
      if (curr.slotY - prev.slotY < safeGap) {
        curr.slotY = prev.slotY + safeGap
      }
    }
  }

  return sorted
}

interface WeightedAreaItem {
  id: string
  weight: number
  areaRatio: number
  area: number
}

function worstAspectRatio(areas: number[], shortSide: number): number {
  if (areas.length === 0 || shortSide <= 0) 
return Number.POSITIVE_INFINITY
  let sum = 0
  let minArea = Number.POSITIVE_INFINITY
  let maxArea = 0
  for (const a of areas) {
    sum += a
    if (a < minArea) 
minArea = a
    if (a > maxArea) 
maxArea = a
  }
  if (sum <= 0 || minArea <= 0) 
return Number.POSITIVE_INFINITY
  const w2 = shortSide * shortSide
  const s2 = sum * sum
  return Math.max((w2 * maxArea) / s2, s2 / (w2 * minArea))
}

function layoutTreemapRow(
  row: WeightedAreaItem[],
  rem: TreemapRect,
  minAreaRatio: number,
  out: TreemapNode[],
): TreemapRect {
  if (row.length === 0) 
return rem
  let rowArea = 0
  for (const item of row) rowArea += item.area

  const horizontalSlice = rem.width >= rem.height
  if (horizontalSlice) {
    const colWidth = rem.height > 0
? rowArea / rem.height
: 0
    let curY = rem.y
    for (const item of row) {
      const h = colWidth > 0
? item.area / colWidth
: 0
      const isTooSmall = item.areaRatio < minAreaRatio || colWidth < 18 || h < 14
      out.push({
        id: item.id,
        weight: item.weight,
        areaRatio: item.areaRatio,
        x: rem.x,
        y: curY,
        width: colWidth,
        height: h,
        isTooSmall,
      })
      curY += h
    }
    return {
      x: rem.x + colWidth,
      y: rem.y,
      width: Math.max(0, rem.width - colWidth),
      height: rem.height,
    }
  }

  const rowHeight = rem.width > 0
? rowArea / rem.width
: 0
  let curX = rem.x
  for (const item of row) {
    const w = rowHeight > 0
? item.area / rowHeight
: 0
    const isTooSmall = item.areaRatio < minAreaRatio || w < 18 || rowHeight < 14
    out.push({
      id: item.id,
      weight: item.weight,
      areaRatio: item.areaRatio,
      x: curX,
      y: rem.y,
      width: w,
      height: rowHeight,
      isTooSmall,
    })
    curX += w
  }
  return {
    x: rem.x,
    y: rem.y + rowHeight,
    width: rem.width,
    height: Math.max(0, rem.height - rowHeight),
  }
}

/**
 * 标准 Squarified Treemap 布局算法，面积严格按 `weight` 比例分配，自动过滤非正权重项并标记过小色块。
 */
export function computeSquarifiedTreemap(
  items: TreemapInputItem[],
  rect: TreemapRect,
  minAreaRatio = 0.008,
): TreemapNode[] {
  if (rect.width <= 0 || rect.height <= 0) 
return []

  const positive = items
    .filter((it) => Number.isFinite(it.weight) && it.weight > 0)
    .sort((a, b) => b.weight - a.weight)

  if (positive.length === 0) 
return []

  let totalWeight = 0
  for (const it of positive) totalWeight += it.weight
  if (totalWeight <= 0) 
return []

  const totalArea = rect.width * rect.height
  const queue: WeightedAreaItem[] = positive.map((it) => {
    const areaRatio = it.weight / totalWeight
    return {
      id: it.id,
      weight: it.weight,
      areaRatio,
      area: areaRatio * totalArea,
    }
  })

  const nodes: TreemapNode[] = []
  let rem: TreemapRect = { ...rect }
  let row: WeightedAreaItem[] = []

  for (const item of queue) {
    const shortSide = Math.min(rem.width, rem.height)
    if (row.length === 0) {
      row.push(item)
      continue
    }

    const currentAreas = row.map((r) => r.area)
    const nextAreas = [...currentAreas, item.area]
    if (worstAspectRatio(nextAreas, shortSide) <= worstAspectRatio(currentAreas, shortSide)) {
      row.push(item)
    }
    else {
      rem = layoutTreemapRow(row, rem, minAreaRatio, nodes)
      row = [item]
    }
  }

  if (row.length > 0) {
    layoutTreemapRow(row, rem, minAreaRatio, nodes)
  }

  return nodes
}

const ASSET_COLOR_PALETTE = [
  '#E5B352',
  '#4EA8DE',
  '#FF736A',
  '#2EC4B6',
  '#A78BFA',
  '#F59E0B',
  '#34D399',
  '#F472B6',
  '#60A5FA',
  '#FB7185',
  '#C084FC',
  '#38BDF8',
  '#FBBF24',
  '#4ADE80',
  '#FB923C',
  '#818CF8',
] as const

function fnv1a32(input: string): number {
  let hash = 0x811c9dc5
  for (let i = 0; i < input.length; i++) {
    hash ^= input.charCodeAt(i)
    hash = Math.imul(hash, 0x01000193)
  }
  return hash >>> 0
}

/**
 * 基于 `assetId` 的稳定哈希调色板映射；传入 `usedColors` 时自动探测规避颜色冲突。
 */
export function deterministicAssetColor(
  assetId: string,
  index: number,
  usedColors?: ReadonlySet<string>,
): string {
  const paletteLen = ASSET_COLOR_PALETTE.length
  const normalizedId = assetId.trim()
  const baseIdx = normalizedId
    ? fnv1a32(normalizedId) % paletteLen
    : Math.abs(Math.trunc(index)) % paletteLen

  const primary = ASSET_COLOR_PALETTE[baseIdx] ?? '#E5B352'
  if (!usedColors || !usedColors.has(primary)) {
    return primary
  }

  for (let offset = 1; offset < paletteLen; offset++) {
    const candidate = ASSET_COLOR_PALETTE[(baseIdx + offset) % paletteLen]
    if (candidate && !usedColors.has(candidate)) {
      return candidate
    }
  }

  const hue = (fnv1a32(`${normalizedId}:${index}`) % 360)
  return `hsl(${hue}, 68%, 62%)`
}
