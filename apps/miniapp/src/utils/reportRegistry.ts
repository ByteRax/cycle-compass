import type {
  AnnualTrendResponse,
  AssetDetailResponse,
  ChartType,
  ComponentTrendResponse,
  PeriodMatrixResponse,
  TrendCurveResponse,
  TrendDashboardResponse,
  TrendDistributionResponse,
  TrendSeasonResponse,
  TrendTierResponse
} from '@/api/types/trend'
import { getTrendTempMeta } from '@/utils/trendDict'

export interface ReportDataBundle {
  chart: ChartType
  assetId: string
  scopeId: string
  currency: string
  range: string
  detail: AssetDetailResponse | null
  curve: TrendCurveResponse | null
  matrix: PeriodMatrixResponse | null
  tier: TrendTierResponse | null
  season: TrendSeasonResponse | null
  annual: AnnualTrendResponse | null
  components: ComponentTrendResponse | null
  dashboard: TrendDashboardResponse | null
  distribution: TrendDistributionResponse | null
}

export interface ReportPageSpec {
  pageIndex: number
  totalPages: number
  title: string
  subtitle: string
  kind: 'curve' | 'matrix' | 'tier' | 'season' | 'summary'
  matrixSlices?: Array<{
    groupName: string
    isContinuation: boolean
    rows: PeriodMatrixResponse['groups'][number]['rows']
  }>
  tierSlices?: Array<{
    label: string
    ruleText: string
    items: TrendTierResponse['tiers'][number]['items']
  }>
  seasonSlices?: Array<{
    label: string
    ruleDesc: string
    count: number
    members: TrendSeasonResponse['stages'][number]['members']
  }>
}

export interface ChartExporterMeta {
  key: ChartType
  label: string
  exportButtonLabel: string
  previewButtonLabel: string
}

export const REPORT_EXPORTER_REGISTRY: Record<ChartType, ChartExporterMeta> = {
  curve: {
    key: 'curve',
    label: '趋势曲线',
    exportButtonLabel: '保存趋势曲线图',
    previewButtonLabel: '预览趋势曲线'
  },
  matrix: {
    key: 'matrix',
    label: '周期矩阵',
    exportButtonLabel: '保存周期矩阵图',
    previewButtonLabel: '预览周期矩阵'
  },
  tier: {
    key: 'tier',
    label: '趋势梯队榜',
    exportButtonLabel: '保存梯队榜图',
    previewButtonLabel: '预览梯队榜'
  },
  season: {
    key: 'season',
    label: '趋势节气图',
    exportButtonLabel: '保存节气图',
    previewButtonLabel: '预览节气图'
  },
  annual: {
    key: 'annual',
    label: '年度同期走势',
    exportButtonLabel: '导出资产摘要',
    previewButtonLabel: '预览资产摘要'
  },
  components: {
    key: 'components',
    label: '品种成分趋势',
    exportButtonLabel: '导出资产摘要',
    previewButtonLabel: '预览资产摘要'
  },
  dashboard: {
    key: 'dashboard',
    label: '趋势仪表盘',
    exportButtonLabel: '导出资产摘要',
    previewButtonLabel: '预览资产摘要'
  },
  distribution: {
    key: 'distribution',
    label: '右侧分布图',
    exportButtonLabel: '导出资产摘要',
    previewButtonLabel: '预览资产摘要'
  }
}

export function getReportExporterMeta(chart: ChartType): ChartExporterMeta {
  return REPORT_EXPORTER_REGISTRY[chart] ?? REPORT_EXPORTER_REGISTRY.curve
}

export function resolveReportAssetName(bundle: ReportDataBundle): string {
  return (
    bundle.matrix?.pool_name
    || bundle.detail?.asset?.name
    || bundle.curve?.asset_name
    || bundle.tier?.asset_name
    || bundle.season?.asset_name
    || bundle.annual?.asset_name
    || bundle.components?.asset_name
    || bundle.dashboard?.asset_name
    || bundle.distribution?.asset_name
    || bundle.assetId
  )
}

export function resolveReportAsOf(bundle: ReportDataBundle): string {
  return (
    bundle.matrix?.snapshot?.as_of
    || bundle.detail?.snapshot?.as_of
    || bundle.curve?.snapshot?.as_of
    || bundle.tier?.snapshot?.as_of
    || bundle.season?.snapshot?.as_of
    || '—'
  )
}

export function resolveReportSnapshotId(bundle: ReportDataBundle): string {
  return (
    bundle.matrix?.snapshot?.snapshot_id
    || bundle.detail?.snapshot?.snapshot_id
    || bundle.curve?.snapshot?.snapshot_id
    || bundle.tier?.snapshot?.snapshot_id
    || bundle.season?.snapshot?.snapshot_id
    || ''
  )
}

export function buildReportTitle(bundle: ReportDataBundle): string {
  const name = resolveReportAssetName(bundle)
  const meta = getReportExporterMeta(bundle.chart)
  if (bundle.chart === 'matrix')
    return `${name} · 周期收益矩阵`
  if (bundle.chart === 'curve')
    return `${name} · ${bundle.range}趋势曲线`
  if (bundle.chart === 'tier')
    return `${name} · 趋势梯队榜`
  if (bundle.chart === 'season')
    return `${name} · 趋势节气分布`
  return `${name} · ${meta.label}摘要`
}

export function buildReportPages(bundle: ReportDataBundle): ReportPageSpec[] {
  const title = buildReportTitle(bundle)
  const asOf = resolveReportAsOf(bundle)
  const algo = bundle.detail?.snapshot?.algorithm_version || bundle.matrix?.snapshot?.algorithm_version || 'trend-v1.0.0'
  const subtitle = `截止 ${asOf} · 计价 ${bundle.currency === 'ORIGINAL'
? '原币'
: bundle.currency} · ${algo}`

  if (bundle.chart === 'matrix' && bundle.matrix) {
    const maxRowsPerPage = 13
    const pages: Array<ReportPageSpec['matrixSlices']> = []
    let currentSlices: NonNullable<ReportPageSpec['matrixSlices']> = []
    let usedLines = 0

    for (const grp of bundle.matrix.groups || []) {
      const rows = grp.rows || []
      if (rows.length === 0)
        continue

      let offset = 0
      while (offset < rows.length) {
        const headerCost = 1
        if (usedLines + headerCost + 1 > maxRowsPerPage && currentSlices.length > 0) {
          pages.push(currentSlices)
          currentSlices = []
          usedLines = 0
        }
        const avail = Math.max(1, maxRowsPerPage - usedLines - headerCost)
        const chunk = rows.slice(offset, offset + avail)
        currentSlices.push({
          groupName: grp.group_name || '默认分组',
          isContinuation: offset > 0,
          rows: chunk
        })
        usedLines += headerCost + chunk.length
        offset += chunk.length
        if (offset < rows.length) {
          pages.push(currentSlices)
          currentSlices = []
          usedLines = 0
        }
      }
    }

    if (currentSlices.length > 0 || pages.length === 0)
      pages.push(currentSlices)

    const totalPages = pages.length
    return pages.map((slices, idx) => ({
      pageIndex: idx + 1,
      totalPages,
      title,
      subtitle,
      kind: 'matrix',
      matrixSlices: slices
    }))
  }

  if (bundle.chart === 'tier' && bundle.tier) {
    const slices = (bundle.tier.tiers || []).map(t => ({
      label: t.pro_label || t.tier,
      ruleText: t.rule_text,
      items: t.items || []
    }))
    return [
{
      pageIndex: 1,
      totalPages: 1,
      title,
      subtitle,
      kind: 'tier',
      tierSlices: slices
    }
]
  }

  if (bundle.chart === 'season' && bundle.season) {
    const slices = (bundle.season.stages || [])
      .filter(s => s.count > 0)
      .map(s => ({
        label: s.stage_label,
        ruleDesc: s.rule_desc,
        count: s.count,
        members: s.members || []
      }))
    return [
{
      pageIndex: 1,
      totalPages: 1,
      title,
      subtitle,
      kind: 'season',
      seasonSlices: slices
    }
]
  }

  return [
{
    pageIndex: 1,
    totalPages: 1,
    title,
    subtitle,
    kind: bundle.chart === 'curve'
? 'curve'
: 'summary'
  }
]
}

export function canExportReportBundle(bundle: ReportDataBundle): { ok: boolean, reason: string } {
  const snap = bundle.matrix?.snapshot || bundle.detail?.snapshot || bundle.curve?.snapshot || bundle.tier?.snapshot || bundle.season?.snapshot
  if (!snap || !snap.snapshot_id)
    return { ok: false, reason: '缺少有效快照元数据' }
  if (snap.is_expired)
    return { ok: false, reason: '快照已过期，请刷新后重试' }

  if (bundle.chart === 'matrix') {
    const rowCount = (bundle.matrix?.groups || []).reduce((acc, g) => acc + g.rows.length, 0)
    if (!bundle.matrix || rowCount === 0)
      return { ok: false, reason: '周期矩阵暂无可导出数据' }
    return { ok: true, reason: '' }
  }

  if (bundle.chart === 'tier') {
    if (!bundle.tier || bundle.tier.valid_count === 0)
      return { ok: false, reason: '梯队榜暂无有效成员可导出' }
    return { ok: true, reason: '' }
  }

  if (bundle.chart === 'season') {
    if (!bundle.season || bundle.season.valid_count === 0)
      return { ok: false, reason: '趋势节气图暂无有效成员可导出' }
    return { ok: true, reason: '' }
  }

  if (!bundle.detail?.asset && !bundle.curve)
    return { ok: false, reason: '资产详情或曲线数据尚未加载' }

  if (bundle.chart === 'curve' && (!bundle.curve?.points || bundle.curve.points.length < 2))
    return { ok: false, reason: '当前区间曲线点位不足，暂不可生成曲线图报告' }

  return { ok: true, reason: '' }
}

function fmtSignedPct(val: number | null | undefined, digits = 2): string {
  if (val === null || val === undefined || !Number.isFinite(val))
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(digits)}%`
}

export function drawReportPageToCanvas(
  ctx: any,
  width: number,
  height: number,
  bundle: ReportDataBundle,
  page: ReportPageSpec,
  reportTheme: 'light' | 'dark',
  cnRedUp = true
) {
  const isLight = reportTheme === 'light'
  const bg = isLight ? '#F4F3EF' : '#090C0F'
  const cardBg = isLight ? '#FFFFFF' : '#12161B'
  const headerBg = isLight ? '#EDECE6' : '#181E25'
  const border = isLight ? '#DCDDD8' : '#29313A'
  const textPrimary = isLight ? '#22292E' : '#F1EFEA'
  const textSecondary = isLight ? '#515F6A' : '#A9B3BD'
  const textMuted = isLight ? '#626F77' : '#8D99A4'
  const brand = isLight ? '#8A641F' : '#E5B352'
  const riseColor = isLight ? '#B93842' : '#F05A50'
  const fallColor = isLight ? '#14745E' : '#2AD1A3'

  const getDirColor = (val: number | null | undefined) => {
    if (val === null || val === undefined || !Number.isFinite(val) || val === 0)
      return textMuted
    const isUp = val > 0
    return (cnRedUp
? isUp
: !isUp)
? riseColor
: fallColor
  }

  ctx.fillStyle = bg
  ctx.fillRect(0, 0, width, height)

  // 顶部品牌条
  ctx.fillStyle = brand
  ctx.font = 'bold 34px sans-serif'
  ctx.fillText('周期罗盘 CycleCompass', 64, 92)

  ctx.fillStyle = textMuted
  ctx.font = '26px sans-serif'
  const pageTag = page.totalPages > 1
? ` · 第 ${page.pageIndex}/${page.totalPages} 页`
: ''
  ctx.fillText(`${getReportExporterMeta(bundle.chart).label}${pageTag}`, 640, 92)

  // 标题与副标题
  ctx.fillStyle = textPrimary
  ctx.font = 'bold 46px sans-serif'
  ctx.fillText(page.title.slice(0, 22), 64, 174)

  ctx.fillStyle = textSecondary
  ctx.font = '24px sans-serif'
  ctx.fillText(page.subtitle, 64, 222)

  if (page.kind === 'matrix' && bundle.matrix) {
    drawMatrixPage(ctx, width, height, bundle.matrix, page, {
      cardBg,
      headerBg,
      border,
      textPrimary,
      textSecondary,
      textMuted,
      brand,
      getDirColor
    })
  }
  else if (page.kind === 'tier' && page.tierSlices) {
    drawTierPage(ctx, width, height, bundle.tier!, page, {
      cardBg,
      headerBg,
      border,
      textPrimary,
      textSecondary,
      textMuted,
      brand
    })
  }
  else if (page.kind === 'season' && page.seasonSlices) {
    drawSeasonPage(ctx, width, height, bundle.season!, page, {
      cardBg,
      headerBg,
      border,
      textPrimary,
      textSecondary,
      textMuted,
      brand
    })
  }
  else {
    drawCurveOrSummaryPage(ctx, width, height, bundle, {
      isLight,
      cardBg,
      border,
      textPrimary,
      textMuted,
      brand,
      riseColor,
      fallColor,
      getDirColor
    })
  }

  // 底部免责与页码
  ctx.fillStyle = border
  ctx.fillRect(64, height - 180, width - 128, 2)
  ctx.fillStyle = brand
  ctx.font = 'bold 28px sans-serif'
  ctx.fillText('周期罗盘 CycleCompass', 64, height - 118)
  ctx.fillStyle = textMuted
  ctx.font = '22px sans-serif'
  ctx.fillText('历史表现不代表未来收益 · 仅供量化研究参考', 64, height - 76)
  ctx.fillText(`第 ${page.pageIndex} / ${page.totalPages} 页`, width - 210, height - 118)
}

function drawMatrixPage(
  ctx: any,
  width: number,
  _height: number,
  matrix: PeriodMatrixResponse,
  page: ReportPageSpec,
  palette: {
    cardBg: string
    headerBg: string
    border: string
    textPrimary: string
    textSecondary: string
    textMuted: string
    brand: string
    getDirColor: (v: number | null | undefined) => string
  }
) {
  const cols = (matrix.columns || []).slice(0, 5)
  const tableX = 64
  const tableW = width - 128
  const nameW = 240
  const colW = cols.length > 0
? Math.floor((tableW - nameW) / cols.length)
: 140

  let curY = 264
  // 重复列头
  ctx.fillStyle = palette.headerBg
  ctx.fillRect(tableX, curY, tableW, 56)
  ctx.strokeStyle = palette.border
  ctx.lineWidth = 2
  ctx.strokeRect(tableX, curY, tableW, 56)

  ctx.fillStyle = palette.textSecondary
  ctx.font = 'bold 24px sans-serif'
  ctx.fillText('资产名称', tableX + 18, curY + 36)
  cols.forEach((col, cIdx) => {
    const cx = tableX + nameW + cIdx * colW + colW - 14
    ctx.textAlign = 'right'
    ctx.fillText(col.label, cx, curY + 36)
  })
  ctx.textAlign = 'left'
  curY += 56

  for (const slice of page.matrixSlices || []) {
    // 分组头
    ctx.fillStyle = palette.headerBg
    ctx.fillRect(tableX, curY, tableW, 48)
    ctx.fillStyle = palette.brand
    ctx.font = 'bold 24px sans-serif'
    ctx.fillText(`${slice.groupName}${slice.isContinuation ? '（续）' : ''}`, tableX + 18, curY + 32)
    curY += 48

    for (const row of slice.rows) {
      ctx.fillStyle = palette.cardBg
      ctx.fillRect(tableX, curY, tableW, 58)
      ctx.strokeStyle = palette.border
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.moveTo(tableX, curY + 58)
      ctx.lineTo(tableX + tableW, curY + 58)
      ctx.stroke()

      ctx.fillStyle = palette.textPrimary
      ctx.font = 'bold 24px sans-serif'
      ctx.fillText((row.asset.short_name || row.asset.name).slice(0, 9), tableX + 18, curY + 36)

      cols.forEach((col, cIdx) => {
        const cell = row.cells[col.key]
        const cx = tableX + nameW + cIdx * colW + colW - 14
        ctx.textAlign = 'right'
        ctx.font = 'bold 24px monospace'
        if (cell && cell.bp_change !== null && cell.bp_change !== undefined) {
          ctx.fillStyle = palette.getDirColor(cell.bp_change)
          const sign = cell.bp_change > 0
? '+'
: ''
          ctx.fillText(`${sign}${cell.bp_change.toFixed(1)}bp`, cx, curY + 36)
        }
        else if (cell && cell.return_pct !== null && cell.return_pct !== undefined) {
          ctx.fillStyle = palette.getDirColor(cell.return_pct)
          ctx.fillText(fmtSignedPct(cell.return_pct, 2), cx, curY + 36)
        }
        else {
          ctx.fillStyle = palette.textMuted
          ctx.fillText('—', cx, curY + 36)
        }
      })
      ctx.textAlign = 'left'
      curY += 58
    }
  }
}

function drawTierPage(
  ctx: any,
  width: number,
  _height: number,
  tier: TrendTierResponse,
  page: ReportPageSpec,
  palette: {
    cardBg: string
    headerBg: string
    border: string
    textPrimary: string
    textSecondary: string
    textMuted: string
    brand: string
  }
) {
  let curY = 264
  ctx.fillStyle = palette.textSecondary
  ctx.font = '24px sans-serif'
  ctx.fillText(`候选 ${tier.candidate_count} · 有效 ${tier.valid_count} · 排除 ${tier.excluded_count}`, 64, curY)
  curY += 32

  for (const bucket of page.tierSlices || []) {
    if (curY > 1160)
      break
    ctx.fillStyle = palette.headerBg
    ctx.fillRect(64, curY, width - 128, 52)
    ctx.fillStyle = palette.brand
    ctx.font = 'bold 26px sans-serif'
    ctx.fillText(`${bucket.label} (${bucket.items.length})`, 84, curY + 34)
    ctx.fillStyle = palette.textMuted
    ctx.font = '20px sans-serif'
    ctx.fillText(bucket.ruleText, 320, curY + 34)
    curY += 56

    const membersText = bucket.items.length > 0
      ? bucket.items.slice(0, 8).map(it => `${it.name}(${Math.round(it.score)})`).join(' · ')
      : '本梯队暂无标的'
    ctx.fillStyle = palette.cardBg
    ctx.fillRect(64, curY, width - 128, 72)
    ctx.strokeStyle = palette.border
    ctx.strokeRect(64, curY, width - 128, 72)
    ctx.fillStyle = palette.textPrimary
    ctx.font = '22px sans-serif'
    ctx.fillText(membersText.slice(0, 44), 84, curY + 44)
    curY += 88
  }
}

function drawSeasonPage(
  ctx: any,
  width: number,
  _height: number,
  season: TrendSeasonResponse,
  page: ReportPageSpec,
  palette: {
    cardBg: string
    headerBg: string
    border: string
    textPrimary: string
    textSecondary: string
    textMuted: string
    brand: string
  }
) {
  let curY = 264
  ctx.fillStyle = palette.textSecondary
  ctx.font = '24px sans-serif'
  ctx.fillText(`上行 ${season.up_count} · 下行 ${season.down_count} · 震荡 ${season.other_count} · 有效 ${season.valid_count}`, 64, curY)
  curY += 34

  for (const st of page.seasonSlices || []) {
    if (curY > 1160)
      break
    ctx.fillStyle = palette.headerBg
    ctx.fillRect(64, curY, width - 128, 50)
    ctx.fillStyle = palette.brand
    ctx.font = 'bold 26px sans-serif'
    ctx.fillText(`${st.label} · ${st.count} 只`, 84, curY + 34)
    ctx.fillStyle = palette.textMuted
    ctx.font = '20px sans-serif'
    ctx.fillText(st.ruleDesc, 280, curY + 34)
    curY += 54

    const names = st.members.slice(0, 6).map(m => m.name).join('、') || '暂无成员'
    ctx.fillStyle = palette.cardBg
    ctx.fillRect(64, curY, width - 128, 64)
    ctx.strokeStyle = palette.border
    ctx.strokeRect(64, curY, width - 128, 64)
    ctx.fillStyle = palette.textPrimary
    ctx.font = '22px sans-serif'
    ctx.fillText(names.slice(0, 42), 84, curY + 40)
    curY += 78
  }
}

function drawCurveOrSummaryPage(
  ctx: any,
  width: number,
  _height: number,
  bundle: ReportDataBundle,
  palette: {
    isLight: boolean
    cardBg: string
    border: string
    textPrimary: string
    textMuted: string
    brand: string
    riseColor: string
    fallColor: string
    getDirColor: (v: number | null | undefined) => string
  }
) {
  const detail = bundle.detail
  const curve = bundle.curve
  const tempMeta = getTrendTempMeta(detail?.asset?.metrics?.temperature)

  ctx.fillStyle = palette.cardBg
  ctx.fillRect(64, 276, width - 128, 210)
  ctx.strokeStyle = palette.border
  ctx.lineWidth = 2
  ctx.strokeRect(64, 276, width - 128, 210)

  const latestVal = detail?.asset?.latest_value
  const kpis = [
    {
      label: '最新值',
      value: latestVal !== null && latestVal !== undefined
? latestVal.toFixed(2)
: '—',
      sub: bundle.currency === 'ORIGINAL'
? (detail?.asset?.currency || '原币')
: bundle.currency,
      color: palette.brand
    },
    {
      label: '趋势温度',
      value: tempMeta.label,
      sub: tempMeta.shortDesc,
      color: palette.isLight ? tempMeta.lightColor : tempMeta.color
    },
    {
      label: '相对强度',
      value: String(detail?.asset?.metrics?.strength?.score !== null && detail?.asset?.metrics?.strength?.score !== undefined
? Math.round(detail.asset.metrics.strength.score)
: '—'),
      sub: detail?.asset?.metrics?.strength?.rank
? `第 ${detail.asset.metrics.strength.rank}/${detail.asset.metrics.strength.valid_count} 名`
: '观察池百分位',
      color: palette.brand
    }
  ]

  kpis.forEach((k, idx) => {
    const x = 104 + idx * 316
    ctx.fillStyle = palette.textMuted
    ctx.font = '26px sans-serif'
    ctx.fillText(k.label, x, 334)
    ctx.fillStyle = k.color
    ctx.font = 'bold 54px sans-serif'
    ctx.fillText(k.value, x, 410)
    ctx.fillStyle = palette.textMuted
    ctx.font = '24px sans-serif'
    ctx.fillText(k.sub, x, 456)
  })

  const chartX = 64
  const chartY = 530
  const chartW = width - 128
  const chartH = 480
  ctx.fillStyle = palette.cardBg
  ctx.fillRect(chartX, chartY, chartW, chartH)
  ctx.strokeStyle = palette.border
  ctx.strokeRect(chartX, chartY, chartW, chartH)

  const pts = (curve?.points || []).filter(p => p.return_pct !== null)
  if (pts.length >= 2) {
    const vals = pts.map(p => p.return_pct as number)
    const minV = Math.min(...vals, 0)
    const maxV = Math.max(...vals, 0)
    const span = Math.max(maxV - minV, 1)

    const zeroY = chartY + chartH - 48 - ((0 - minV) / span) * (chartH - 96)
    ctx.strokeStyle = palette.border
    ctx.beginPath()
    ctx.moveTo(chartX + 36, zeroY)
    ctx.lineTo(chartX + chartW - 36, zeroY)
    ctx.stroke()

    ctx.strokeStyle = palette.brand
    ctx.lineWidth = 5
    ctx.beginPath()
    pts.forEach((p, idx) => {
      const px = chartX + 36 + (idx / (pts.length - 1)) * (chartW - 72)
      const py = chartY + chartH - 48 - (((p.return_pct as number) - minV) / span) * (chartH - 96)
      if (idx === 0)
        ctx.moveTo(px, py)
      else
        ctx.lineTo(px, py)
    })
    ctx.stroke()
  }
  else {
    ctx.fillStyle = palette.textMuted
    ctx.font = '28px sans-serif'
    ctx.fillText('资产核心摘要与状态快照（无足够曲线点位）', chartX + 180, chartY + chartH / 2)
  }

  let latestRet: number | null = null
  for (let i = (curve?.points?.length || 0) - 1; i >= 0; i--) {
    const v = curve?.points?.[i]?.return_pct
    if (v !== null && v !== undefined) {
      latestRet = v
      break
    }
  }
  let benchRet: number | null = null
  for (let i = (curve?.benchmark_points?.length || 0) - 1; i >= 0; i--) {
    const v = curve?.benchmark_points?.[i]?.return_pct
    if (v !== null && v !== undefined) {
      benchRet = v
      break
    }
  }

  const ext = curve?.extremum
  const statsY = 1064
  const bottomStats = [
    {
label: '当前累计收益',
value: fmtSignedPct(latestRet, 1),
color: palette.getDirColor(latestRet)
},
    {
label: '比较基准',
value: fmtSignedPct(benchRet, 1),
color: palette.getDirColor(benchRet)
},
    {
label: '最大回撤',
value: fmtSignedPct(ext?.max_drawdown_pct, 1),
color: palette.fallColor
},
    {
label: '最大上涨',
value: fmtSignedPct(ext?.max_gain_pct, 1),
color: palette.riseColor
}
  ]
  bottomStats.forEach((s, idx) => {
    const x = 64 + idx * 244
    ctx.fillStyle = palette.textMuted
    ctx.font = '24px sans-serif'
    ctx.fillText(s.label, x, statsY)
    ctx.fillStyle = s.color
    ctx.font = 'bold 36px sans-serif'
    ctx.fillText(s.value, x, statsY + 52)
  })
}
