<script lang="ts" setup>
import type { AssetDetailResponse, ChartType, TrendCurveResponse } from '@/api/types/trend'
import { onLoad, onShareAppMessage } from '@dcloudio/uni-app'
import { computed, getCurrentInstance, ref } from 'vue'
import { getAssetDetail, getTrendCurve } from '@/api/trend'
import CSafeArea from '@/components/ui/CSafeArea.vue'
import { useFinanceStore } from '@/stores/finance'
import { getTrendTempMeta } from '@/utils/trendDict'

definePage({
  style: {
    navigationBarTitleText: '研究报告',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    navigationStyle: 'custom'
  }
})

const financeStore = useFinanceStore()
const instance = getCurrentInstance()

const assetId = ref('a_shares')
const chartType = ref<ChartType>('curve')
const rangeParam = ref('1Y')
const currencyParam = ref('CNY')
const snapshotIdParam = ref('')
const reportTheme = ref<'light' | 'dark'>(financeStore.reportThemePref)

const loading = ref(false)
const exporting = ref(false)
const assetDetail = ref<AssetDetailResponse | null>(null)
const curveData = ref<TrendCurveResponse | null>(null)
const exportedTempPath = ref('')

const chartLabelMap: Record<ChartType, string> = {
  curve: '趋势曲线',
  annual: '年度价格走势',
  components: '品种成分趋势',
  dashboard: '趋势仪表盘',
  season: '趋势节气图',
  distribution: '右侧分布图',
  tier: '趋势梯队榜'
}

const displayAssetName = computed(() => assetDetail.value?.asset?.name || curveData.value?.asset_name || '金融研究报告')
const displayAsOf = computed(() => assetDetail.value?.snapshot?.as_of || curveData.value?.snapshot?.as_of || '—')
const displayAlgo = computed(() => assetDetail.value?.snapshot?.algorithm_version || 'v1')
const displayLatestValue = computed(() => {
  const v = assetDetail.value?.asset?.latest_value
  return v !== null && v !== undefined
? v.toFixed(2)
: '—'
})
const displayTempMeta = computed(() => getTrendTempMeta(assetDetail.value?.asset?.metrics?.temperature))
const displayStrength = computed(() => assetDetail.value?.asset?.metrics?.strength?.score ?? '—')
const displayReturnPct = computed(() => {
  const pts = curveData.value?.points || []
  for (let i = pts.length - 1; i >= 0; i--) {
    const pt = pts[i]
    if (pt && pt.return_pct !== null && pt.return_pct !== undefined)
      return pt.return_pct
  }
  return null
})
const displayBenchmarkReturnPct = computed(() => {
  const pts = curveData.value?.benchmark_points || []
  for (let i = pts.length - 1; i >= 0; i--) {
    const pt = pts[i]
    if (pt && pt.return_pct !== null && pt.return_pct !== undefined)
      return pt.return_pct
  }
  return null
})

async function loadReportData() {
  if (!assetId.value)
    return
  loading.value = true
  try {
    const [detailRes, curveRes] = await Promise.allSettled([
      getAssetDetail(assetId.value, {
        currency: currencyParam.value,
        snapshot_id: snapshotIdParam.value || undefined
      }),
      getTrendCurve(assetId.value, {
        range: rangeParam.value,
        currency: currencyParam.value,
        snapshot_id: snapshotIdParam.value || undefined
      })
    ])
    if (detailRes.status === 'fulfilled')
      assetDetail.value = detailRes.value.data ?? null
    if (curveRes.status === 'fulfilled')
      curveData.value = curveRes.value.data ?? null
  }
  finally {
    loading.value = false
  }
}

onLoad((query: any) => {
  if (query?.asset_id)
    assetId.value = decodeURIComponent(query.asset_id)
  if (query?.chart && query.chart in chartLabelMap)
    chartType.value = query.chart as ChartType
  if (query?.range)
    rangeParam.value = decodeURIComponent(query.range)
  if (query?.currency)
    currencyParam.value = decodeURIComponent(query.currency)
  if (query?.snapshot_id)
    snapshotIdParam.value = decodeURIComponent(query.snapshot_id)
  if (query?.theme === 'light' || query?.theme === 'dark')
    reportTheme.value = query.theme
  loadReportData()
})

onShareAppMessage(() => {
  const params = [
    `asset_id=${encodeURIComponent(assetId.value)}`,
    `chart=${encodeURIComponent(chartType.value)}`,
    `range=${encodeURIComponent(rangeParam.value)}`,
    `currency=${encodeURIComponent(currencyParam.value)}`,
    `theme=${encodeURIComponent(reportTheme.value)}`
  ]
  if (snapshotIdParam.value)
    params.push(`snapshot_id=${encodeURIComponent(snapshotIdParam.value)}`)
  return {
    title: `${displayAssetName.value} · ${chartLabelMap[chartType.value]}研究报告`,
    path: `/pages-sub/finance/report?${params.join('&')}`
  }
})

function handleBack() {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function setReportTheme(theme: 'light' | 'dark') {
  reportTheme.value = theme
  financeStore.reportThemePref = theme
  exportedTempPath.value = ''
}

function formatPct(val: number | null | undefined): string {
  if (val === null || val === undefined)
    return '—'
  const sign = val > 0
? '+'
: ''
  return `${sign}${val.toFixed(1)}%`
}

function drawReportCanvas(ctx: any, width: number, height: number) {
  const isLight = reportTheme.value === 'light'
  const bg = isLight ? '#F4F3EF' : '#090C0F'
  const cardBg = isLight ? '#FFFFFF' : '#12161B'
  const border = isLight ? '#DCDDD8' : '#29313A'
  const textPrimary = isLight ? '#22292E' : '#F1EFEA'
  const textSecondary = isLight ? '#515F6A' : '#A9B3BD'
  const textMuted = isLight ? '#626F77' : '#8D99A4'
  const brand = isLight ? '#8A641F' : '#E5B352'
  const upColor = isLight ? '#B93842' : '#F05A50'
  const downColor = isLight ? '#14745E' : '#2AD1A3'

  // 背景
  ctx.fillStyle = bg
  ctx.fillRect(0, 0, width, height)

  // 头部品牌条
  ctx.fillStyle = brand
  ctx.font = 'bold 34px sans-serif'
  ctx.fillText('周期罗盘 CycleCompass', 64, 96)

  ctx.fillStyle = textMuted
  ctx.font = '26px sans-serif'
  ctx.fillText(`研究报告 / ${chartLabelMap[chartType.value]}`, 680, 96)

  // 标题
  ctx.fillStyle = textPrimary
  ctx.font = 'bold 54px sans-serif'
  ctx.fillText(`${displayAssetName.value} · ${rangeParam.value}趋势`, 64, 188)

  ctx.fillStyle = textSecondary
  ctx.font = '28px sans-serif'
  ctx.fillText(`快照日期 ${displayAsOf.value} · 计价 ${currencyParam.value} · 算法 ${displayAlgo.value}`, 64, 240)

  // 三项 KPI 卡片
  ctx.fillStyle = cardBg
  ctx.fillRect(64, 286, width - 128, 210)
  ctx.strokeStyle = border
  ctx.lineWidth = 2
  ctx.strokeRect(64, 286, width - 128, 210)

  const kpis = [
    {
label: '最新值',
value: displayLatestValue.value,
sub: currencyParam.value,
color: brand
},
    {
label: '趋势温度',
value: displayTempMeta.value.label,
sub: displayTempMeta.value.shortDesc,
color: isLight ? displayTempMeta.value.lightColor : displayTempMeta.value.color
},
    {
label: '相对强度',
value: String(displayStrength.value),
sub: '观察池百分位',
color: brand
}
  ]
  kpis.forEach((k, idx) => {
    const x = 104 + idx * 316
    ctx.fillStyle = textMuted
    ctx.font = '26px sans-serif'
    ctx.fillText(k.label, x, 344)
    ctx.fillStyle = k.color
    ctx.font = 'bold 56px sans-serif'
    ctx.fillText(k.value, x, 420)
    ctx.fillStyle = textMuted
    ctx.font = '24px sans-serif'
    ctx.fillText(k.sub, x, 466)
  })

  // 趋势曲线绘图区
  const chartX = 64
  const chartY = 540
  const chartW = width - 128
  const chartH = 480
  ctx.fillStyle = cardBg
  ctx.fillRect(chartX, chartY, chartW, chartH)
  ctx.strokeStyle = border
  ctx.strokeRect(chartX, chartY, chartW, chartH)

  const pts = (curveData.value?.points || []).filter(p => p.return_pct !== null)
  if (pts.length >= 2) {
    const vals = pts.map(p => p.return_pct as number)
    const minV = Math.min(...vals, 0)
    const maxV = Math.max(...vals, 0)
    const span = Math.max(maxV - minV, 1)

    // 0% 基准线
    const zeroY = chartY + chartH - 48 - ((0 - minV) / span) * (chartH - 96)
    ctx.strokeStyle = border
    ctx.beginPath()
    ctx.moveTo(chartX + 36, zeroY)
    ctx.lineTo(chartX + chartW - 36, zeroY)
    ctx.stroke()

    ctx.strokeStyle = brand
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
    ctx.fillStyle = textMuted
    ctx.font = '28px sans-serif'
    ctx.fillText('当前快照暂无足够曲线点位，已保留核心指标摘要', chartX + 160, chartY + chartH / 2)
  }

  // 底部四项极值指标
  const ext = curveData.value?.extremum
  const statsY = 1074
  const bottomStats = [
    {
label: '当前累计收益',
value: formatPct(displayReturnPct.value),
color: (displayReturnPct.value ?? 0) >= 0
? upColor
: downColor
},
    {
label: '比较基准',
value: formatPct(displayBenchmarkReturnPct.value),
color: (displayBenchmarkReturnPct.value ?? 0) >= 0
? upColor
: downColor
},
    {
label: '最大回撤',
value: formatPct(ext?.max_drawdown_pct),
color: downColor
},
    {
label: '最大上涨',
value: formatPct(ext?.max_gain_pct),
color: upColor
}
  ]
  bottomStats.forEach((s, idx) => {
    const x = 64 + idx * 244
    ctx.fillStyle = textMuted
    ctx.font = '24px sans-serif'
    ctx.fillText(s.label, x, statsY)
    ctx.fillStyle = s.color
    ctx.font = 'bold 38px sans-serif'
    ctx.fillText(s.value, x, statsY + 52)
  })

  // 底部免责与二维码占位
  ctx.fillStyle = border
  ctx.fillRect(64, 1180, width - 128, 2)
  ctx.fillStyle = brand
  ctx.font = 'bold 30px sans-serif'
  ctx.fillText('周期罗盘 CycleCompass', 64, 1250)
  ctx.fillStyle = textMuted
  ctx.font = '22px sans-serif'
  ctx.fillText('历史表现不代表未来收益 · 指标口径见小程序内说明', 64, 1296)

  ctx.strokeStyle = border
  ctx.strokeRect(width - 204, 1216, 140, 140)
  ctx.fillStyle = textMuted
  ctx.font = '22px sans-serif'
  ctx.fillText('小程序码', width - 178, 1294)
}

function generateReportImage(): Promise<string> {
  return new Promise((resolve, reject) => {
    const query = uni.createSelectorQuery().in(instance?.proxy)
    query.select('#reportCanvas').fields({ node: true, size: true }, () => {}).exec((res) => {
      const node = res?.[0]?.node
      if (!node) {
        reject(new Error('当前环境不支持 Canvas 2D 高清导出'))
        return
      }
      const width = 1080
      const height = 1440
      node.width = width
      node.height = height
      const ctx = node.getContext('2d')
      drawReportCanvas(ctx, width, height)
      uni.canvasToTempFilePath({
        canvasId: 'reportCanvas',
        canvas: node,
        destWidth: width,
        destHeight: height,
        fileType: 'png',
        quality: 1,
        success: (out) => {
          exportedTempPath.value = out.tempFilePath
          resolve(out.tempFilePath)
        },
        fail: err => reject(err)
      } as UniApp.CanvasToTempFilePathOptions)
    })
  })
}

async function handlePreviewImage() {
  if (exporting.value)
    return
  exporting.value = true
  uni.showLoading({ title: '生成高清图...' })
  try {
    const filePath = exportedTempPath.value || await generateReportImage()
    uni.hideLoading()
    uni.previewImage({
urls: [filePath]
})
  }
  catch (err: any) {
    uni.hideLoading()
    uni.showToast({ title: err?.message || '生成失败', icon: 'none' })
  }
  finally {
    exporting.value = false
  }
}

async function handleSaveToAlbum() {
  if (exporting.value)
    return
  exporting.value = true
  uni.showLoading({ title: '正在导出 1080px 高清图...' })
  try {
    const filePath = exportedTempPath.value || await generateReportImage()
    uni.saveImageToPhotosAlbum({
      filePath,
      success: () => {
        uni.hideLoading()
        uni.showToast({ title: '已保存到相册', icon: 'success' })
      },
      fail: (err) => {
        uni.hideLoading()
        const msg = err?.errMsg || ''
        if (msg.includes('auth deny') || msg.includes('authorize')) {
          uni.showModal({
            title: '需要相册权限',
            content: '请在设置中允许保存图片到相册，以便导出高清研究报告。',
            confirmText: '去设置',
            success: (modalRes) => {
              if (modalRes.confirm)
                uni.openSetting({})
            }
          })
        }
        else {
          uni.showToast({ title: '保存失败，请长按预览图保存', icon: 'none' })
        }
      }
    })
  }
  catch (err: any) {
    uni.hideLoading()
    uni.showToast({ title: err?.message || '导出失败', icon: 'none' })
  }
  finally {
    exporting.value = false
  }
}
</script>

<template>
  <view
    class="min-h-screen box-border flex flex-col"
    :class="{ 'cc-finance-light': !financeStore.isDarkEffective }"
    style="background-color: var(--cc-bg-page); color: var(--cc-text-primary);"
  >
    <CSafeArea position="top" />

    <!-- 顶部导航 -->
    <view
      class="flex items-center justify-between px-3.5 py-2.5 border-b"
      style="border-color: var(--cc-border);"
    >
      <view
        class="min-w-[44px] min-h-[44px] flex items-center gap-2 cursor-pointer"
        @tap="handleBack"
      >
        <text class="text-lg font-bold cc-brand">‹</text>
        <text class="text-sm font-semibold cc-text-primary">研究报告</text>
      </view>
    </view>

    <view class="px-3.5 pt-3.5 pb-6 flex-1">
      <!-- 标题栏 -->
      <view class="flex items-center justify-between mb-1">
        <text class="text-[22px] font-bold tracking-tight cc-text-primary">分享预览</text>
        <view
          class="px-2.5 py-1 rounded-md border text-xs font-medium cc-brand"
          style="border-color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);"
        >
          高清报告
        </view>
      </view>
      <text class="text-xs cc-text-secondary block mb-3">
        {{ displayAssetName }} · {{ chartLabelMap[chartType] }} / {{ displayAsOf }}
      </text>

      <!-- 报告预览画布卡片（独立于页面主题，跟随 reportTheme） -->
      <view
        class="p-4 rounded-2xl border mb-4 transition-colors"
        :style="reportTheme === 'light'
          ? 'background-color: #F4F3EF; border-color: #DCDDD8; color: #22292E;'
          : 'background-color: #12161B; border-color: #29313A; color: #F1EFEA;'"
      >
        <view class="flex items-center justify-between mb-3">
          <text
            class="text-xs font-bold"
            :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }"
          >
            ◎ 周期罗盘
          </text>
          <text
            class="text-[11px]"
            :style="{ color: reportTheme === 'light' ? '#626F77' : '#8D99A4' }"
          >
            研究报告 / {{ chartLabelMap[chartType] }}
          </text>
        </view>

        <text class="text-lg font-bold block mb-1">
          {{ displayAssetName }} · {{ rangeParam }}趋势对比
        </text>
        <text
          class="text-[11px] block mb-3"
          :style="{ color: reportTheme === 'light' ? '#515F6A' : '#A9B3BD' }"
        >
          {{ currencyParam }} · {{ displayAsOf }} · 算法 {{ displayAlgo }}
        </text>

        <view v-if="loading" class="py-8 text-center text-xs opacity-70">
          正在加载报告快照...
        </view>

        <template v-else>
          <!-- 3栏 KPI -->
          <view
            class="grid grid-cols-3 gap-2 p-3 rounded-xl border mb-3"
            :style="reportTheme === 'light'
              ? 'background-color: #FFFFFF; border-color: #DCDDD8;'
              : 'background-color: #181E25; border-color: #29313A;'"
          >
            <view>
              <text class="text-[10px] opacity-70 block">最新值</text>
              <text
                class="text-lg font-bold cc-tabular block"
                :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }"
              >
                {{ displayLatestValue }}
              </text>
              <text class="text-[10px] opacity-60 block">{{ currencyParam }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">趋势温度</text>
              <text
                class="text-lg font-bold block"
                :style="{ color: reportTheme === 'light' ? displayTempMeta.lightColor : displayTempMeta.color }"
              >
                {{ displayTempMeta.label }}
              </text>
              <text class="text-[10px] opacity-60 block">{{ displayTempMeta.shortDesc }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">相对强度</text>
              <text
                class="text-lg font-bold cc-tabular block"
                :style="{ color: reportTheme === 'light' ? '#8A641F' : '#E5B352' }"
              >
                {{ displayStrength }}
              </text>
              <text class="text-[10px] opacity-60 block">百分位</text>
            </view>
          </view>

          <!-- 底部极值 -->
          <view class="grid grid-cols-4 gap-2 mb-3 pt-2 border-t" :style="{ borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }">
            <view>
              <text class="text-[10px] opacity-70 block">累计收益</text>
              <text class="text-xs font-bold cc-tabular">{{ formatPct(displayReturnPct) }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">比较基准</text>
              <text class="text-xs font-bold cc-tabular">{{ formatPct(displayBenchmarkReturnPct) }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">最大回撤</text>
              <text class="text-xs font-bold cc-tabular">{{ formatPct(curveData?.extremum?.max_drawdown_pct) }}</text>
            </view>
            <view>
              <text class="text-[10px] opacity-70 block">最大上涨</text>
              <text class="text-xs font-bold cc-tabular">{{ formatPct(curveData?.extremum?.max_gain_pct) }}</text>
            </view>
          </view>
        </template>

        <view
          class="pt-2.5 border-t flex items-center justify-between text-[10px] opacity-70"
          :style="{ borderColor: reportTheme === 'light' ? '#DCDDD8' : '#29313A' }"
        >
          <text>周期罗盘 CycleCompass · 历史表现不代表未来</text>
          <text>1080px 高清</text>
        </view>
      </view>

      <!-- 报告主题与规格切换 -->
      <view class="flex items-center gap-2 mb-3 flex-wrap">
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium border cursor-pointer"
          :style="reportTheme === 'light'
            ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);'
            : 'border-color: var(--cc-border); color: var(--cc-text-secondary); background-color: var(--cc-bg-card);'"
          @tap="setReportTheme('light')"
        >
          浅色报告
        </view>
        <view
          class="px-3 py-1.5 rounded-lg text-xs font-medium border cursor-pointer"
          :style="reportTheme === 'dark'
            ? 'border-color: var(--cc-brand-primary); color: var(--cc-brand-primary); background-color: var(--cc-bg-elevated);'
            : 'border-color: var(--cc-border); color: var(--cc-text-secondary); background-color: var(--cc-bg-card);'"
          @tap="setReportTheme('dark')"
        >
          深色报告
        </view>
        <view class="px-3 py-1.5 rounded-lg text-xs cc-elevated cc-text-muted">
          1080px 高清
        </view>
      </view>

      <text class="text-[11px] cc-text-muted block mb-3">
        图片由 Canvas 2D 独立渲染，非页面截图；导出后可直接保存相册。
      </text>

      <!-- 操作按钮 -->
      <view class="grid grid-cols-2 gap-2.5 mb-3">
        <view
          class="py-3 rounded-xl cc-panel text-center text-xs font-semibold cc-text-primary cursor-pointer"
          @tap="handlePreviewImage"
        >
          查看完整图
        </view>
        <view
          class="py-3 rounded-xl text-center text-xs font-semibold cursor-pointer"
          style="background-color: var(--cc-brand-primary); color: #090C0F;"
          @tap="handleSaveToAlbum"
        >
          保存相册
        </view>
      </view>

      <!-- 微信转发按钮 -->
      <button
        open-type="share"
        class="w-full py-2.5 rounded-xl cc-elevated cc-text-primary text-xs font-medium border"
        style="border-color: var(--cc-border);"
      >
        转发给微信好友（还原当前报告参数）
      </button>
    </view>

    <!-- 离屏 1080px Canvas 2D 节点 -->
    <canvas
      id="reportCanvas"
      type="2d"
      style="width: 1080px; height: 1440px; position: fixed; left: -9999px; top: -9999px;"
    />

    <CSafeArea position="bottom" />
  </view>
</template>
