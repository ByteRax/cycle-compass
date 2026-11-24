<script lang="ts" setup>
import { computed } from 'vue'
import type { DivergenceItem } from '@/data/divergence'

interface Props {
  timeline: string[]
  items: DivergenceItem[]
  benchmarkName?: string
  benchmarkVal?: number
}

const props = withDefaults(defineProps<Props>(), {
  timeline: () => ['2026/09', '2026/10', '2026/11', '2026/11/25'],
  items: () => [],
  benchmarkName: '',
  benchmarkVal: 0
})

// SVG 画布几何参数（保持黄金比例与高清晰度）
const svgWidth = 660
const svgHeight = 310
const padLeft = 60
const padRight = 148 // 为右侧品种名与涨跌幅预留充裕空间，绝不换行重叠
const padTop = 28
const padBottom = 38

const plotWidth = svgWidth - padLeft - padRight
const plotHeight = svgHeight - padTop - padBottom

// 动态计算 Y 轴边界与网格步长（自适应 -50% ~ +20% 或 0% ~ +10000% 超大跨度）
const yRange = computed(() => {
  let min = 0
  let max = 0
  for (const it of props.items) {
    for (const v of it.series) {
      if (v < min) min = v
      if (v > max) max = v
    }
    if (it.total_change_pct < min) min = it.total_change_pct
    if (it.total_change_pct > max) max = it.total_change_pct
  }
  if (props.benchmarkVal !== undefined) {
    if (props.benchmarkVal < min) min = props.benchmarkVal
    if (props.benchmarkVal > max) max = props.benchmarkVal
  }

  const span = Math.max(8, max - min)
  let step = 5
  if (span > 5000) step = 5000
  else if (span > 2000) step = 1000
  else if (span > 800) step = 500
  else if (span > 300) step = 100
  else if (span > 100) step = 50
  else if (span > 40) step = 10
  else step = 5

  const yMax = Math.ceil((max + span * 0.08) / step) * step
  const yMin = Math.floor((min - span * 0.08) / step) * step

  return { yMin, yMax, step }
})

// 坐标转换计算
const getX = (idx: number, total: number) => {
  if (total <= 1) return padLeft
  return padLeft + (idx / (total - 1)) * plotWidth
}

const getY = (val: number) => {
  const { yMin, yMax } = yRange.value
  const diff = yMax - yMin || 1
  return padTop + ((yMax - val) / diff) * plotHeight
}

// 0% 与基准线 Y 坐标
const zeroY = computed(() => getY(0))
const benchmarkY = computed(() => getY(props.benchmarkVal ?? 0))

// 动态计算网格线刻度
const gridTicks = computed(() => {
  const { yMin, yMax, step } = yRange.value
  const ticks: number[] = [0]
  for (let v = step; v <= yMax; v += step) ticks.push(v)
  for (let v = -step; v >= yMin; v -= step) ticks.push(v)
  return ticks.sort((a, b) => b - a)
})

// 生成平滑贝塞尔曲线路径
const buildSmoothPath = (series: number[]) => {
  if (!series || series.length < 2) return ''
  const points = series.map((val, idx) => ({
    x: getX(idx, series.length),
    y: getY(val)
  }))

  let path = `M ${points[0].x.toFixed(1)} ${points[0].y.toFixed(1)}`
  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i === 0 ? 0 : i - 1]
    const p1 = points[i]
    const p2 = points[i + 1]
    const p3 = points[i + 2 < points.length ? i + 2 : i + 1]

    const cp1x = p1.x + (p2.x - p0.x) / 6
    const cp1y = p1.y + (p2.y - p0.y) / 6
    const cp2x = p2.x - (p3.x - p1.x) / 6
    const cp2y = p2.y - (p3.y - p1.y) / 6

    path += ` C ${cp1x.toFixed(1)} ${cp1y.toFixed(1)}, ${cp2x.toFixed(1)} ${cp2y.toFixed(1)}, ${p2.x.toFixed(1)} ${p2.y.toFixed(1)}`
  }
  return path
}

// 避让算法：对右侧标签执行垂直微调，防止多条紧邻曲线文本重叠
const labelSlots = computed(() => {
  const sorted = [...props.items]
    .map((item) => ({
      item,
      targetY: getY(item.total_change_pct),
      lastX: getX(item.series.length - 1, item.series.length)
    }))
    .sort((a, b) => a.targetY - b.targetY)

  const minGap = 16
  for (let i = 1; i < sorted.length; i++) {
    if (sorted[i].targetY - sorted[i - 1].targetY < minGap) {
      sorted[i].targetY = sorted[i - 1].targetY + minGap
    }
  }

  // 二次向上拉扯超下界项
  const bottomLimit = svgHeight - padBottom + 4
  for (let i = sorted.length - 1; i >= 0; i--) {
    if (sorted[i].targetY > bottomLimit) {
      sorted[i].targetY = bottomLimit - (sorted.length - 1 - i) * minGap
    }
  }

  return sorted
})

// 生成完整的原生 SVG 字符串
const svgMarkup = computed(() => {
  // 1. 背景横向刻度线与刻度值
  const ticksMarkup = gridTicks.value
    .map((tick) => {
      const isZero = tick === 0
      const stroke = isZero ? 'rgba(255,255,255,0.30)' : 'rgba(255,255,255,0.06)'
      const dash = isZero ? '3,3' : '2,2'
      const label = tick > 0 ? `+${tick}%` : `${tick}%`
      const fill = isZero ? '#f1efea' : '#64707c'
      return `
        <line x1="${padLeft}" y1="${getY(tick)}" x2="${svgWidth - padRight}" y2="${getY(tick)}" stroke="${stroke}" stroke-dasharray="${dash}" stroke-width="1" />
        <text x="${padLeft - 6}" y="${getY(tick) + 3.5}" text-anchor="end" fill="${fill}" font-size="10" font-family="monospace" font-weight="500">${label}</text>
      `
    })
    .join('')

  // 2. 基准水平虚线与左侧高亮黑金胶囊标签（若基准不为0，绘制专属基准线）
  let benchmarkMarkup = ''
  if (props.benchmarkVal !== undefined) {
    const bY = benchmarkY.value
    const isDistinctFromZero = Math.abs(bY - zeroY.value) > 6
    const line = isDistinctFromZero
      ? `<line x1="${padLeft}" y1="${bY}" x2="${svgWidth - padRight}" y2="${bY}" stroke="rgba(229,179,82,0.6)" stroke-dasharray="4,3" stroke-width="1.2" />`
      : ''
    const valText = props.benchmarkVal > 0 ? `+${props.benchmarkVal}%` : `${props.benchmarkVal}%`
    benchmarkMarkup = `
      ${line}
      <g class="benchmark-badge">
        <rect x="4" y="${bY - 9}" width="50" height="18" rx="3.5" fill="#1E2732" stroke="#e5b352" stroke-width="1.2" />
        <text x="29" y="${bY + 3.5}" text-anchor="middle" fill="#e5b352" font-size="9.5" font-family="monospace" font-weight="bold">
          ${valText}
        </text>
      </g>
    `
  }

  // 3. 所有品种平滑走势曲线
  const pathsMarkup = props.items
    .map((it) => {
      const d = buildSmoothPath(it.series)
      const sw = it.is_benchmark ? 2.4 : 1.4
      const dash = it.is_benchmark ? 'stroke-dasharray="5,2"' : ''
      const op = it.is_benchmark ? 1 : 0.85
      return `<path d="${d}" fill="none" stroke="${it.color}" stroke-width="${sw}" ${dash} stroke-linecap="round" stroke-linejoin="round" opacity="${op}" />`
    })
    .join('')

  // 4. 曲线末端圆点、避让引线与指示标签
  const labelsMarkup = labelSlots.value
    .map((slot) => {
      const origY = getY(slot.item.total_change_pct)
      const hasOffset = Math.abs(origY - slot.targetY) > 2
      const line = hasOffset
        ? `<line x1="${slot.lastX}" y1="${origY}" x2="${slot.lastX + 6}" y2="${slot.targetY}" stroke="${slot.item.color}" stroke-width="0.8" stroke-opacity="0.6" />`
        : ''
      const pctStr =
        slot.item.total_change_pct > 0
          ? `+${slot.item.total_change_pct}%`
          : `${slot.item.total_change_pct}%`
      const pctColor = slot.item.total_change_pct >= 0 ? '#ff736a' : '#2ec4b6'
      const displayName = slot.item.short_name || slot.item.name

      return `
        <circle cx="${slot.lastX}" cy="${origY}" r="2.5" fill="${slot.item.color}" />
        ${line}
        <circle cx="${slot.lastX + 8}" cy="${slot.targetY}" r="2" fill="${slot.item.color}" />
        <text x="${slot.lastX + 13}" y="${slot.targetY + 3.5}" font-size="9.5" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-weight="bold">
          <tspan fill="#8d99a4" font-weight="normal">${displayName}</tspan>
          <tspan fill="${pctColor}" dx="3" font-family="monospace">${pctStr}</tspan>
        </text>
      `
    })
    .join('')

  // 5. X 轴时间刻度与终点高亮金色胶囊
  const timelineMarkup = props.timeline
    .map((t, i) => {
      const x = getX(i, props.timeline.length)
      if (i === props.timeline.length - 1) {
        return `
          <rect x="${x - 38}" y="${svgHeight - padBottom + 12}" width="76" height="18" rx="3.5" fill="#e5b352" />
          <text x="${x}" y="${svgHeight - padBottom + 24.5}" text-anchor="middle" fill="#090c0f" font-size="10" font-family="monospace" font-weight="bold">${t}</text>
        `
      }
      return `
        <text x="${x}" y="${svgHeight - padBottom + 24}" text-anchor="middle" fill="#64707c" font-size="10" font-family="monospace">${t}</text>
      `
    })
    .join('')

  return `
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${svgWidth} ${svgHeight}" class="w-full h-auto block" preserveAspectRatio="xMidYMid meet" style="width:100%;height:auto;display:block;">
      <g class="grid-lines">${ticksMarkup}</g>
      ${benchmarkMarkup}
      <g class="trend-lines">${pathsMarkup}</g>
      <g class="endpoint-labels">${labelsMarkup}</g>
      <g class="x-axis">${timelineMarkup}</g>
    </svg>
  `
})
</script>

<template>
  <view class="divergence-chart-wrapper w-full rounded-2xl bg-[#12181F] border border-white/[0.08] p-3.5 shadow-lg relative overflow-hidden">
    <!-- 图表上方图例条 (分2列紧凑排布) -->
    <view class="grid grid-cols-2 gap-x-3 gap-y-1 pb-3 px-1 border-b border-white/[0.05]">
      <view
        v-for="it in items"
        :key="it.id"
        class="flex items-center space-x-1.5 overflow-hidden"
      >
        <view class="w-3 h-0.5 rounded-full shrink-0" :style="{ backgroundColor: it.color }" />
        <text class="text-[10px] text-[#8d99a4] truncate">{{ it.name }}</text>
      </view>
    </view>

    <!-- 原生 SVG 走势图 (使用 v-html 确保 SVG 内 text 标签被原生解析渲染) -->
    <view class="w-full relative mt-2" v-html="svgMarkup" />
  </view>
</template>

<style scoped>
.divergence-chart-wrapper {
  background: linear-gradient(180deg, #131A22 0%, #0D1217 100%);
}
</style>
