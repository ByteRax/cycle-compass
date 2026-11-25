<script lang="ts" setup>
import type {
  LabelCollisionTarget,
  ResolvedLabelSlot
} from '@/utils/chartGeometry'
import { computed, getCurrentInstance, nextTick, onMounted, ref, watch } from 'vue'
import { useFinanceStore } from '@/stores/finance'
import {
  buildTimeXMapper,
  computeNiceYRange,
  lttbDownsample,
  resolveLabelCollisions,
  splitSegmentsByGap
} from '@/utils/chartGeometry'

export interface TrendLineSeries {
  id: string
  name: string
  color: string
  values: Array<number | null>
  dashed?: boolean
  lineWidth?: number
  highlighted?: boolean
  dimmed?: boolean
  hidden?: boolean
  labelText?: string
}

export interface TrendStateInterval {
  startDate: string
  endDate: string
  side: 'right' | 'left' | 'neutral' | 'unknown'
}

export interface CrosshairReadoutItem {
  id: string
  name: string
  color: string
  value: number | null
  formatted: string
}

const props = withDefaults(defineProps<{
  canvasId?: string
  dates: string[]
  series: TrendLineSeries[]
  height?: number
  valueFormat?: 'pct' | 'price' | 'number'
  logScale?: boolean
  showZeroLine?: boolean
  baselineValue?: number | null
  showRightLabels?: boolean
  stateIntervals?: TrendStateInterval[]
  maxXLabels?: number
  unit?: string
}>(), {
  canvasId: 'trendLineCanvas',
  height: 220,
  valueFormat: 'pct',
  logScale: false,
  showZeroLine: true,
  baselineValue: null,
  showRightLabels: true,
  stateIntervals: () => [],
  maxXLabels: 5,
  unit: ''
})

const emit = defineEmits<{
  crosshairChange: [payload: { index: number | null, date: string, items: CrosshairReadoutItem[] }]
}>()

const instance = getCurrentInstance()
const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

const activeIndex = ref<number | null>(null)
const canvasRect = ref({
left: 0,
top: 0,
width: 320,
height: props.height
})

// 缓存每帧布局供触摸十字线查表
let cachedPlotLeft = 42
let cachedPlotRight = 308
let cachedXPositions: number[] = []
let drawTimer: ReturnType<typeof setTimeout> | null = null

function formatYValue(val: number | null): string {
  if (val === null || !Number.isFinite(val))
    return '—'
  if (props.valueFormat === 'pct') {
    const sign = val > 0
? '+'
: ''
    return `${sign}${val.toFixed(Math.abs(val) >= 100
? 0
: 1)}%`
  }
  if (props.valueFormat === 'price') {
    const abs = Math.abs(val)
    if (abs >= 1000)
      return val.toFixed(0)
    if (abs >= 100)
      return val.toFixed(1)
    if (abs >= 1)
      return val.toFixed(2)
    return val.toFixed(4)
  }
  return val.toFixed(2)
}

function formatAxisTick(val: number, isLogMode: boolean): string {
  const actual = isLogMode
? 10 ** val
: val
  if (props.valueFormat === 'pct') {
    const sign = actual > 0
? '+'
: ''
    return `${sign}${Math.round(actual * 10) / 10}%`
  }
  const abs = Math.abs(actual)
  if (abs >= 10000)
    return `${(actual / 1000).toFixed(1)}k`
  if (abs >= 100)
    return actual.toFixed(0)
  if (abs >= 1)
    return actual.toFixed(1)
  return actual.toFixed(2)
}

const visibleSeries = computed(() => props.series.filter(s => !s.hidden))

const activeDateLabel = computed(() => {
  if (activeIndex.value === null)
    return ''
  return props.dates[activeIndex.value] || ''
})

const activeReadoutItems = computed<CrosshairReadoutItem[]>(() => {
  if (activeIndex.value === null)
    return []
  const idx = activeIndex.value
  return visibleSeries.value.map((s) => {
    const raw = s.values[idx] ?? null
    return {
      id: s.id,
      name: s.name,
      color: s.color,
      value: raw,
      formatted: formatYValue(raw)
    }
  })
})

function canUseLogScale(): boolean {
  if (!props.logScale)
    return false
  for (const s of visibleSeries.value) {
    for (const v of s.values) {
      if (v !== null && Number.isFinite(v) && v <= 0)
        return false
    }
  }
  return true
}

function toPlotSpace(v: number, isLogMode: boolean): number {
  return isLogMode
? Math.log10(v)
: v
}

function scheduleDraw() {
  if (drawTimer)
    clearTimeout(drawTimer)
  drawTimer = setTimeout(() => {
    nextTick(() => {
      drawChart()
    })
  }, 16)
}

function drawChart() {
  const dates = props.dates
  const list = visibleSeries.value
  if (!dates.length || !list.length)
    return

  const query = uni.createSelectorQuery().in(instance?.proxy)
  query
    .select(`#${props.canvasId}`)
    .fields({ node: true, size: true, rect: true }, () => {})
    .exec((res: any) => {
      const target = res?.[0]
      if (!target || !target.node)
        return

      const canvas = target.node
      const ctx = canvas.getContext('2d') as CanvasRenderingContext2D | null
      if (!ctx)
        return

      const w = Math.max(200, target.width || 320)
      const h = Math.max(120, target.height || props.height)
      canvasRect.value = {
        left: target.left || 0,
        top: target.top || 0,
        width: w,
        height: h
      }

      const dpr = uni.getSystemInfoSync().pixelRatio || 2
      canvas.width = Math.round(w * dpr)
      canvas.height = Math.round(h * dpr)
      if (canvas.style) {
        canvas.style.width = `${w}px`
        canvas.style.height = `${h}px`
      }
      ctx.save()
      ctx.scale(dpr, dpr)
      ctx.clearRect(0, 0, w, h)

      const isLogMode = canUseLogScale()
      const padL = 44
      const padR = props.showRightLabels ? 68 : 14
      const padT = 16
      const padB = 28
      const plotW = Math.max(40, w - padL - padR)
      const plotH = Math.max(40, h - padT - padB)
      cachedPlotLeft = padL
      cachedPlotRight = padL + plotW

      // 收集 Y 轴数值范围
      const allSpaceValues: number[] = []
      for (const s of list) {
        for (const v of s.values) {
          if (v !== null && Number.isFinite(v))
            allSpaceValues.push(toPlotSpace(v, isLogMode))
        }
      }
      if (props.baselineValue !== null && props.baselineValue !== undefined && Number.isFinite(props.baselineValue)) {
        if (!isLogMode || props.baselineValue > 0)
          allSpaceValues.push(toPlotSpace(props.baselineValue, isLogMode))
      }

      const includeZero = props.valueFormat === 'pct' && !isLogMode
      const yRange = computeNiceYRange(allSpaceValues, includeZero)
      const ySpan = Math.max(1e-9, yRange.yMax - yRange.yMin)
      const mapY = (spaceVal: number) =>
        padT + ((yRange.yMax - spaceVal) / ySpan) * plotH

      // X 轴时间映射
      const xMapper = buildTimeXMapper(dates, padL, plotW, props.maxXLabels)
      cachedXPositions = dates.map(d => xMapper.mapX(d))

      // 1. 状态区间背景着色
      if (props.stateIntervals.length > 0) {
        for (const seg of props.stateIntervals) {
          if (seg.side !== 'right' && seg.side !== 'left')
            continue
          const x1 = xMapper.mapX(seg.startDate)
          const x2 = xMapper.mapX(seg.endDate)
          const segW = Math.max(1, x2 - x1)
          if (seg.side === 'right') {
            ctx.fillStyle = isDark.value
              ? 'rgba(229, 179, 82, 0.08)'
              : 'rgba(138, 100, 31, 0.07)'
          }
          else {
            ctx.fillStyle = isDark.value
              ? 'rgba(56, 189, 248, 0.07)'
              : 'rgba(20, 116, 94, 0.06)'
          }
          ctx.fillRect(x1, padT, segW, plotH)
        }
      }

      // 2. 网格线 & Y 轴刻度
      const gridColor = isDark.value ? 'rgba(241, 239, 234, 0.07)' : 'rgba(34, 41, 46, 0.08)'
      const textMuted = isDark.value ? '#8D99A4' : '#626F77'
      ctx.font = '10px monospace'
      ctx.textAlign = 'right'
      ctx.textBaseline = 'middle'

      for (const tick of yRange.ticks) {
        const gy = mapY(tick)
        if (gy < padT - 2 || gy > padT + plotH + 2)
          continue
        ctx.strokeStyle = gridColor
        ctx.lineWidth = 1
        ctx.setLineDash([2, 3])
        ctx.beginPath()
        ctx.moveTo(padL, gy)
        ctx.lineTo(padL + plotW, gy)
        ctx.stroke()
        ctx.setLineDash([])

        ctx.fillStyle = textMuted
        ctx.fillText(formatAxisTick(tick, isLogMode), padL - 5, gy)
      }

      // 3. 零线 / 基准线
      const refVal = props.baselineValue ?? (
        props.showZeroLine && !isLogMode && yRange.yMin <= 0 && yRange.yMax >= 0
          ? 0
          : null
      )

      if (refVal !== null && (!isLogMode || refVal > 0)) {
        const refY = mapY(toPlotSpace(refVal, isLogMode))
        if (refY >= padT && refY <= padT + plotH) {
          ctx.strokeStyle = isDark.value ? 'rgba(241, 239, 234, 0.28)' : 'rgba(34, 41, 46, 0.28)'
          ctx.lineWidth = 1
          ctx.setLineDash([4, 3])
          ctx.beginPath()
          ctx.moveTo(padL, refY)
          ctx.lineTo(padL + plotW, refY)
          ctx.stroke()
          ctx.setLineDash([])
        }
      }

      // 4. X 轴时间刻度
      ctx.textAlign = 'center'
      ctx.textBaseline = 'top'
      ctx.fillStyle = textMuted
      ctx.font = '10px monospace'
      const lastTickIdx = xMapper.ticks.length - 1
      xMapper.ticks.forEach((t, idx) => {
        if (idx === 0)
          ctx.textAlign = 'left'
        else if (idx === lastTickIdx && lastTickIdx > 0)
          ctx.textAlign = 'right'
        else
          ctx.textAlign = 'center'
        const shortLabel = t.label.length >= 10
? t.label.slice(2, 10)
: t.label
        ctx.fillText(shortLabel, t.x, padT + plotH + 6)
      })

      // 5. 绘制多条折线（按 gap 断开 + LTTB 降采样）
      const labelTargets: LabelCollisionTarget[] = []
      const lttbThreshold = Math.max(80, Math.round(plotW * 1.1))

      // 先画非高亮线，最后画高亮线确保置顶
      const orderedSeries = [...list].sort((a, b) => {
        const aScore = a.highlighted
? 2
: a.dimmed ? 0 : 1
        const bScore = b.highlighted
? 2
: b.dimmed ? 0 : 1
        return aScore - bScore
      })

      for (const s of orderedSeries) {
        const rawPoints = dates.map((d, i) => ({
          x: cachedXPositions[i] ?? padL,
          value: s.values[i] ?? null
        }))
        const segments = splitSegmentsByGap(rawPoints)
        if (!segments.length)
          continue

        ctx.save()
        ctx.strokeStyle = s.color
        ctx.lineWidth = s.lineWidth ?? (s.highlighted ? 2.6 : 1.7)
        ctx.lineJoin = 'round'
        ctx.lineCap = 'round'
        ctx.globalAlpha = s.dimmed ? 0.25 : 1
        if (s.dashed)
          ctx.setLineDash([5, 3])

        let lastEndPoint: { x: number, y: number, value: number } | null = null

        for (const seg of segments) {
          const mappedSeg = seg.map(pt => ({
            x: pt.x,
            y: mapY(toPlotSpace(pt.value as number, isLogMode)),
            value: pt.value as number
          }))
          const sampled = mappedSeg.length > lttbThreshold
            ? lttbDownsample(mappedSeg, lttbThreshold)
            : mappedSeg

          if (sampled.length === 1) {
            const p = sampled[0]
            if (p) {
              ctx.fillStyle = s.color
              ctx.beginPath()
              ctx.arc(p.x, p.y, 2.5, 0, Math.PI * 2)
              ctx.fill()
              lastEndPoint = p
            }
            continue
          }

          ctx.beginPath()
          sampled.forEach((p, idx) => {
            if (idx === 0)
              ctx.moveTo(p.x, p.y)
            else
              ctx.lineTo(p.x, p.y)
          })
          ctx.stroke()
          lastEndPoint = sampled[sampled.length - 1] ?? lastEndPoint
        }

        ctx.setLineDash([])

        // 末端圆点与右侧标签目标
        if (lastEndPoint && !s.dimmed) {
          ctx.fillStyle = s.color
          ctx.beginPath()
          ctx.arc(lastEndPoint.x, lastEndPoint.y, s.highlighted ? 3.5 : 2.5, 0, Math.PI * 2)
          ctx.fill()

          if (props.showRightLabels) {
            labelTargets.push({
              id: s.id,
              rawY: lastEndPoint.y,
              text: s.labelText || `${s.name.slice(0, 4)} ${formatYValue(lastEndPoint.value)}`,
              color: s.color
            })
          }
        }
        ctx.restore()
      }

      // 6. 右侧避让标签
      if (props.showRightLabels && labelTargets.length > 0) {
        const resolved: ResolvedLabelSlot[] = resolveLabelCollisions(
          labelTargets,
          padT + 6,
          padT + plotH - 6,
          13
        )
        ctx.font = '10px monospace'
        ctx.textAlign = 'left'
        ctx.textBaseline = 'middle'
        for (const item of resolved) {
          ctx.fillStyle = item.color
          ctx.fillText(item.text, padL + plotW + 6, item.slotY)
        }
      }

      // 7. 十字线与高亮圆点
      if (activeIndex.value !== null && activeIndex.value >= 0 && activeIndex.value < dates.length) {
        const cx = cachedXPositions[activeIndex.value] ?? padL
        ctx.strokeStyle = isDark.value ? 'rgba(229, 179, 82, 0.65)' : 'rgba(138, 100, 31, 0.65)'
        ctx.lineWidth = 1
        ctx.setLineDash([3, 3])
        ctx.beginPath()
        ctx.moveTo(cx, padT)
        ctx.lineTo(cx, padT + plotH)
        ctx.stroke()
        ctx.setLineDash([])

        for (const s of list) {
          if (s.dimmed)
            continue
          const v = s.values[activeIndex.value]
          if (v === null || v === undefined || !Number.isFinite(v))
            continue
          if (isLogMode && v <= 0)
            continue
          const cy = mapY(toPlotSpace(v, isLogMode))
          ctx.fillStyle = s.color
          ctx.beginPath()
          ctx.arc(cx, cy, 4, 0, Math.PI * 2)
          ctx.fill()
          ctx.strokeStyle = isDark.value ? '#090C0F' : '#FFFFFF'
          ctx.lineWidth = 1.5
          ctx.stroke()
        }
      }

      ctx.restore()
    })
}

function updateCrosshairFromClientX(clientX: number) {
  if (!cachedXPositions.length)
    return
  const relX = clientX - canvasRect.value.left
  const clampedX = Math.max(cachedPlotLeft, Math.min(cachedPlotRight, relX))
  let bestIdx = 0
  let bestDist = Number.POSITIVE_INFINITY
  for (let i = 0; i < cachedXPositions.length; i++) {
    const d = Math.abs((cachedXPositions[i] ?? 0) - clampedX)
    if (d < bestDist) {
      bestDist = d
      bestIdx = i
    }
  }
  if (activeIndex.value !== bestIdx) {
    activeIndex.value = bestIdx
    emit('crosshairChange', {
      index: bestIdx,
      date: props.dates[bestIdx] || '',
      items: activeReadoutItems.value
    })
    scheduleDraw()
  }
}

function extractTouchPoint(e: any): { x: number | null, y: number | null } {
  const touch = e?.touches?.[0] || e?.changedTouches?.[0]
  if (touch) {
    const x = typeof touch.clientX === 'number'
      ? touch.clientX
      : (typeof touch.x === 'number'
? canvasRect.value.left + touch.x
: null)
    const y = typeof touch.clientY === 'number'
      ? touch.clientY
      : (typeof touch.y === 'number'
? canvasRect.value.top + touch.y
: null)
    return { x, y }
  }
  const dx = typeof e?.detail?.x === 'number'
? e.detail.x
: (typeof e?.clientX === 'number'
? e.clientX
: null)
  const dy = typeof e?.detail?.y === 'number'
? e.detail.y
: (typeof e?.clientY === 'number'
? e.clientY
: null)
  return { x: dx, y: dy }
}

const touchStartPoint = ref<{ x: number, y: number } | null>(null)
const gestureAxis = ref<'none' | 'horizontal' | 'vertical'>('none')

function clearCrosshair() {
  if (activeIndex.value !== null) {
    activeIndex.value = null
    emit('crosshairChange', { index: null, date: '', items: [] })
    scheduleDraw()
  }
}

function handleTouchStart(e: any) {
  const pt = extractTouchPoint(e)
  touchStartPoint.value = pt.x !== null && pt.y !== null
? { x: pt.x, y: pt.y }
: null
  gestureAxis.value = 'none'
}

function handleLongPress(e: any) {
  gestureAxis.value = 'horizontal'
  const pt = extractTouchPoint(e)
  if (pt.x !== null)
    updateCrosshairFromClientX(pt.x)
}

function handleTouchMove(e: any) {
  const pt = extractTouchPoint(e)
  if (pt.x === null)
    return

  if (gestureAxis.value === 'vertical')
    return

  if (gestureAxis.value === 'none' && touchStartPoint.value && pt.y !== null) {
    const dx = Math.abs(pt.x - touchStartPoint.value.x)
    const dy = Math.abs(pt.y - touchStartPoint.value.y)
    if (Math.max(dx, dy) < 6)
      return
    if (dy > dx * 1.1) {
      gestureAxis.value = 'vertical'
      clearCrosshair()
      return
    }
    gestureAxis.value = 'horizontal'
  }

  if (gestureAxis.value === 'horizontal') {
    e?.preventDefault?.()
    e?.stopPropagation?.()
    updateCrosshairFromClientX(pt.x)
  }
}

function handleTouchEnd() {
  touchStartPoint.value = null
  gestureAxis.value = 'none'
  clearCrosshair()
}

onMounted(() => {
  scheduleDraw()
})

watch(
  () => [
props.dates,
props.series,
props.logScale,
props.valueFormat,
props.stateIntervals,
isDark.value
],
  () => {
    scheduleDraw()
  },
  { deep: true }
)
</script>

<template>
  <view class="w-full">
    <!-- 十字线读数条 -->
    <view
      v-if="activeIndex !== null && activeReadoutItems.length > 0"
      class="flex flex-wrap items-center gap-x-3 gap-y-1 px-2.5 py-1.5 mb-1.5 rounded-lg text-[11px] font-mono tabular-nums border"
      :class="isDark ? 'bg-[#090C0F]/90 border-[#29313A] text-[#F1EFEA]' : 'bg-[#EDECE6] border-[#DCDDD8] text-[#22292E]'"
    >
      <text class="font-bold" :class="isDark ? 'text-[#E5B352]' : 'text-[#8A641F]'">
        {{ activeDateLabel }}
      </text>
      <view
        v-for="item in activeReadoutItems.slice(0, 4)"
        :key="item.id"
        class="flex items-center space-x-1"
      >
        <view class="w-1.5 h-1.5 rounded-full shrink-0" :style="{ backgroundColor: item.color }" />
        <text class="truncate max-w-[72px]" :class="isDark ? 'text-[#8D99A4]' : 'text-[#515F6A]'">
          {{ item.name }}
        </text>
        <text class="font-bold">
          {{ item.formatted }}{{ unit && valueFormat !== 'pct' ? unit : '' }}
        </text>
      </view>
    </view>

    <view class="w-full relative" :style="{ height: `${height}px` }">
      <canvas
        :id="canvasId"
        :canvas-id="canvasId"
        type="2d"
        class="w-full h-full block"
        :style="{ width: '100%', height: `${height}px`, touchAction: 'pan-y' }"
        @touchstart="handleTouchStart"
        @touchmove="handleTouchMove"
        @longpress="handleLongPress"
        @touchend="handleTouchEnd"
        @touchcancel="handleTouchEnd"
      />
    </view>
  </view>
</template>
