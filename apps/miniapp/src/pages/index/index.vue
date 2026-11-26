<script lang="ts" setup>
import type {
  BacktestOutput,
  CombinedSignal,
  CycleOverview,
  IndustryMetric,
  PortfolioOverview as IPortfolioOverview,
  MarketQuote
} from '@/api/cycle'
import type { BottomTabKey } from '@/utils/swipe'
import { onHide, onLoad, onPageScroll, onPullDownRefresh, onShow } from '@dcloudio/uni-app'
import { computed, nextTick, onMounted, onUnmounted, ref, shallowRef } from 'vue'
import {
  getAdaptiveRefreshIntervalMs,
  getCombinedStrategySignal,
  getCycleOverview,
  getIndustryMetrics,
  getMarketQuotes,
  getPortfolioOverview,
  runBacktest
} from '@/api/cycle'
import { usePortfolio } from '@/composables/usePortfolio'
import {
  FALLBACK_BACKTEST_DATA,
  SEED_AI_EQUIPMENTS,
  SEED_AI_PRODUCTS,
  SEED_ASIA_MACRO,
  SEED_CRYPTO_AGRI,
  SEED_CYCLE_OVERVIEW,
  SEED_DEMO_PORTFOLIO,
  SEED_DIVIDEND_CORE,
  SEED_FX_LIST,
  SEED_GLOBAL_ECONOMIC_DATA,
  SEED_HK_US_MACRO,
  SEED_INDUSTRIAL_METALS,
  SEED_JP_MACRO,
  SEED_JP_SECTORS,
  SEED_KR_MACRO,
  SEED_KR_SECTORS,
  SEED_OTHER_METALS,
  SEED_PRECIOUS_METALS,
  SEED_ROBOT_DRIVE,
  SEED_ROBOT_SENSE,
  SEED_SECTOR_LIST,
  SEED_SEMI_EQUIP,
  SEED_SEMI_MATERIAL,
  SEED_STRATEGY_SIGNAL,
  SEED_TW_MACRO
} from '@/data/seeds'
import { BOTTOM_TAB_KEYS, evaluateTabSwipe, resolveBottomTabKey } from '@/utils/swipe'
import TabMarketHub from './components/TabMarketHub.vue'
import TabOverview from './components/TabOverview.vue'
import TabPortfolio from './components/TabPortfolio.vue'
import TabResearch from './components/TabResearch.vue'

definePage({
  type: 'home',
  style: {
    navigationBarTitleText: '周期罗盘',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    enablePullDownRefresh: true,
    h5: {
      titleNView: false
    }
  }
})

// 底部 4 Tab 导航（总览 / 市场 / 研究 / 账本）
const tabs: ReadonlyArray<{
  key: BottomTabKey
  label: string
  iconOff: string
  iconOn: string
}> = [
  {
    key: 'overview',
    label: '总览',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxjaXJjbGUgY3g9IjEyIiBjeT0iMTIiIHI9IjkiLz48Y2lyY2xlIGN4PSIxMiIgY3k9IjEyIiByPSI0LjUiLz48cGF0aCBkPSJNMTIgNy41VjEybDIuNSAyIi8+PC9zdmc+',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxjaXJjbGUgY3g9IjEyIiBjeT0iMTIiIHI9IjkiLz48Y2lyY2xlIGN4PSIxMiIgY3k9IjEyIiByPSI0LjUiLz48cGF0aCBkPSJNMTIgNy41VjEybDIuNSAyIi8+PC9zdmc+'
  },
  {
    key: 'markets',
    label: '市场',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0zIDE2bDUtNSA0IDQgOS05Ii8+PHBhdGggZD0iTTE2IDZoNXY1Ii8+PC9zdmc+',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0zIDE2bDUtNSA0IDQgOS05Ii8+PHBhdGggZD0iTTE2IDZoNXY1Ii8+PC9zdmc+'
  },
  {
    key: 'research',
    label: '研究',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMy41IiB5PSIxMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMTMuNSIgd2lkdGg9IjciIGhlaWdodD0iNyIgcng9IjEuNSIvPjwvc3ZnPg==',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMy41IiB5PSIxMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMTMuNSIgd2lkdGg9IjciIGhlaWdodD0iNyIgcng9IjEuNSIvPjwvc3ZnPg=='
  },
  {
    key: 'portfolio',
    label: '账本',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjQiIHk9IjMiIHdpZHRoPSIxNiIgaGVpZ2h0PSIxOCIgcng9IjIuNSIvPjxsaW5lIHgxPSI4IiB5MT0iOSIgeDI9IjE2IiB5Mj0iOSIvPjxsaW5lIHgxPSI4IiB5MT0iMTMuNSIgeDI9IjE0IiB5Mj0iMTMuNSIvPjxsaW5lIHgxPSI4IiB5MT0iMTcuNSIgeDI9IjExLjUiIHkyPSIxNy41Ii8+PC9zdmc+',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjQiIHk9IjMiIHdpZHRoPSIxNiIgaGVpZ2h0PSIxOCIgcng9IjIuNSIvPjxsaW5lIHgxPSI4IiB5MT0iOSIgeDI9IjE2IiB5Mj0iOSIvPjxsaW5lIHgxPSI4IiB5MT0iMTMuNSIgeDI9IjE0IiB5Mj0iMTMuNSIvPjxsaW5lIHgxPSI4IiB5MT0iMTcuNSIgeDI9IjExLjUiIHkyPSIxNy41Ii8+PC9zdmc+'
  }
]

const activeTab = ref<BottomTabKey>('overview')
const marketSegment = ref<'quotes' | 'industries'>('quotes')
const overviewRef = ref<InstanceType<typeof TabOverview> | null>(null)

const loadedTabs = ref<Record<BottomTabKey, boolean>>({
  overview: true,
  markets: false,
  research: false,
  portfolio: false
})

// 按 Tab 独立保存滚动位置，避免切换时强行重置顶部（AUDIT 20）
const tabScrollTops = ref<Record<BottomTabKey, number>>({
  overview: 0,
  markets: 0,
  research: 0,
  portfolio: 0
})
let currentScrollTop = 0

onPageScroll((e) => {
  currentScrollTop = e?.scrollTop ?? 0
  tabScrollTops.value[activeTab.value] = currentScrollTop
})

// 各板块响应式状态（离线种子仅作初始结构占位，独立标记 UsingSeed 与 Error）
const globalEconomicData = shallowRef<MarketQuote[]>([...SEED_GLOBAL_ECONOMIC_DATA])
const sectorList = shallowRef<IndustryMetric[]>([...SEED_SECTOR_LIST])
const hkUsMacroList = shallowRef([...SEED_HK_US_MACRO])
const twMacroList = shallowRef([...SEED_TW_MACRO])
const cryptoAgriList = shallowRef([...SEED_CRYPTO_AGRI])
const krMacroList = shallowRef([...SEED_KR_MACRO])
const krSectorList = shallowRef([...SEED_KR_SECTORS])
const jpMacroList = shallowRef([...SEED_JP_MACRO])
const jpSectorList = shallowRef([...SEED_JP_SECTORS])
const asiaMacroList = shallowRef([...SEED_ASIA_MACRO])
const fxList = shallowRef([...SEED_FX_LIST])
const preciousMetals = shallowRef([...SEED_PRECIOUS_METALS])
const industrialMetals = shallowRef([...SEED_INDUSTRIAL_METALS])
const otherMetals = shallowRef([...SEED_OTHER_METALS])
const aiProducts = shallowRef([...SEED_AI_PRODUCTS])
const aiEquipments = shallowRef([...SEED_AI_EQUIPMENTS])
const robotDriveList = shallowRef([...SEED_ROBOT_DRIVE])
const robotSenseList = shallowRef([...SEED_ROBOT_SENSE])
const semiEquipList = shallowRef([...SEED_SEMI_EQUIP])
const semiMaterialList = shallowRef([...SEED_SEMI_MATERIAL])
const dividendCoreList = shallowRef([...SEED_DIVIDEND_CORE])
const cycleData = shallowRef<CycleOverview>({ ...SEED_CYCLE_OVERVIEW })
const strategyData = shallowRef<CombinedSignal>({ ...SEED_STRATEGY_SIGNAL })
const portfolioData = ref<IPortfolioOverview>({ ...SEED_DEMO_PORTFOLIO })

const backtestResult = shallowRef<BacktestOutput | null>(null)
const isBacktesting = ref(false)
const backtestError = ref('')
const isSampleBacktestResult = ref(false)

// 分模块独立时间戳与报错态（严禁全局时间戳掩盖局部失败，AUDIT 16/18）
function formatTimestamp (d = new Date()) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}:${ss}`
}

const overviewFetchedAt = ref('')
const overviewError = ref('')
const overviewUsingSeed = ref(true)

const marketsFetchedAt = ref('')
const marketsError = ref('')
const marketsUsingSeed = ref(true)

const sectorsFetchedAt = ref('')
const sectorsError = ref('')
const sectorsUsingSeed = ref(true)

let lastActiveFetchMs = 0

// 用户专属自定义模拟账本系统
const {
  portfolioMode,
  userPortfolio,
  lastPersistError,
  currentDisplayPortfolio,
  cashPoolTicks,
  currentWeekKey,
  currentWeekCheckIn,
  rebuyRuleStatus,
  dividendValuation,
  growthValuation,
  loadUserPortfolio,
  handleSavePortfolio,
  handleResetPortfolio,
  handleWeeklyCheckIn,
  handleRebuyCashPool
} = usePortfolio({
  portfolioData,
  strategyData,
  formatTimestamp
})

async function loadCycleTabData(force = false): Promise<boolean> {
  overviewError.value = ''
  let updated = false
  try {
    const [cRes, sRes] = await Promise.allSettled([
      getCycleOverview(force),
      getCombinedStrategySignal(force)
    ])
    if (cRes.status === 'fulfilled' && cRes.value?.data) {
      cycleData.value = cRes.value.data
      updated = true
    }
    if (sRes.status === 'fulfilled' && sRes.value?.data) {
      strategyData.value = sRes.value.data
      updated = true
    }
    if (updated) {
      overviewFetchedAt.value = formatTimestamp()
      overviewUsingSeed.value = false
    }
    else {
      overviewError.value = '周期与策略接口未返回有效数据'
    }
  }
  catch (err: any) {
    overviewError.value = err?.message || '周期与策略接口连接失败'
  }
  return updated
}

async function loadMarketsTabData(force = false): Promise<boolean> {
  marketsError.value = ''
  try {
    const mRes = await getMarketQuotes(undefined, force)
    if (mRes?.data?.length) {
      const quoteMap = new Map(mRes.data.map(q => [
q.symbol.toUpperCase(),
q
]))
      globalEconomicData.value = globalEconomicData.value.map(item => quoteMap.get(item.symbol.toUpperCase()) ?? item)
      hkUsMacroList.value = hkUsMacroList.value.map((item) => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      twMacroList.value = twMacroList.value.map((item) => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      cryptoAgriList.value = cryptoAgriList.value.map((item) => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      krMacroList.value = krMacroList.value.map((item) => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      jpMacroList.value = jpMacroList.value.map((item) => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      asiaMacroList.value = asiaMacroList.value.map((item) => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      fxList.value = fxList.value.map((item) => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q
? {
...item,
price: q.price.toFixed(2),
change_pct: q.change_pct
}
: item
      })
      preciousMetals.value = preciousMetals.value.map((item) => {
        const sym = item.name === '黄金'
? 'GOLD_OZ'
: item.name === '白银'
? 'SILVER_OZ'
: null
        const q = sym
? quoteMap.get(sym)
: null
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      industrialMetals.value = industrialMetals.value.map((item) => {
        const sym = item.name === '铜'
? 'COPPER_HG'
: null
        const q = sym
? quoteMap.get(sym)
: null
        return q
? { ...item, price: q.price, change_pct: q.change_pct }
: item
      })
      marketsFetchedAt.value = formatTimestamp()
      marketsUsingSeed.value = false
      return true
    }
    marketsError.value = '全球行情接口暂无数据返回'
  }
  catch (err: any) {
    marketsError.value = err?.message || '全球行情拉取失败'
  }
  return false
}

async function loadSectorsTabData(force = false): Promise<boolean> {
  sectorsError.value = ''
  try {
    const secRes = await getIndustryMetrics(force)
    if (secRes?.data?.length) {
      sectorList.value = secRes.data
      sectorsFetchedAt.value = formatTimestamp()
      sectorsUsingSeed.value = false
      return true
    }
    sectorsError.value = '产业指标接口暂无返回'
  }
  catch (err: any) {
    sectorsError.value = err?.message || '产业指标拉取失败'
  }
  return false
}

async function loadPortfolioTabData(force = false): Promise<boolean> {
  try {
    const pRes = await getPortfolioOverview(force)
    if (pRes?.data) {
      portfolioData.value = pRes.data
      return true
    }
  }
  catch {
    // 示范账本失败保留初始模板
  }
  return false
}

async function ensureTabData(tabKey: BottomTabKey, force = false) {
  lastActiveFetchMs = Date.now()
  if (tabKey === 'overview' || tabKey === 'research') {
    await loadCycleTabData(force)
  }
  else if (tabKey === 'markets') {
    await Promise.all([
      loadMarketsTabData(force),
      loadSectorsTabData(force)
    ])
  }
  else if (tabKey === 'portfolio') {
    await Promise.all([
      loadCycleTabData(force),
      loadPortfolioTabData(force)
    ])
  }
}

async function loadAllData(force = false) {
  lastActiveFetchMs = Date.now()
  await Promise.all([
    loadCycleTabData(force),
    overviewRef.value?.refreshOverviewSubModules?.(),
    loadedTabs.value.markets
? loadMarketsTabData(force)
: Promise.resolve(false),
    loadedTabs.value.markets
? loadSectorsTabData(force)
: Promise.resolve(false),
    loadedTabs.value.portfolio
? loadPortfolioTabData(force)
: Promise.resolve(false)
  ])
}

async function handleRunBacktest(params: { low_line: number, high_line: number, weekly_amount: number, days: number }) {
  isBacktesting.value = true
  backtestError.value = ''
  isSampleBacktestResult.value = false
  try {
    const res = await runBacktest(params)
    if (res?.data?.summary && (res.code === 0 || res.code === undefined)) {
      backtestResult.value = res.data
      uni.showToast({ title: `${res.data.summary.trading_days}交易日回测已完成`, icon: 'success' })
    }
    else {
      backtestResult.value = null
      backtestError.value = (res as any)?.msg || (res as any)?.message || '回测接口返回异常响应'
      uni.showToast({ title: '回测计算失败', icon: 'none' })
    }
  }
  catch (err: any) {
    backtestResult.value = null
    backtestError.value = err?.message || '回测服务连接失败（已拒绝使用样本数据冒充实时运行）'
    uni.showToast({ title: '回测服务未连通', icon: 'none' })
  }
  finally {
    isBacktesting.value = false
  }
}

function handleShowSampleBacktest() {
  backtestError.value = ''
  isSampleBacktestResult.value = true
  backtestResult.value = FALLBACK_BACKTEST_DATA
  uni.showToast({ title: '已载入历史样例（2023-08~2026-11）', icon: 'none' })
}

let autoPollTimer: ReturnType<typeof setTimeout> | null = null
let isPageVisible = true

function stopPolling() {
  if (autoPollTimer) {
    clearTimeout(autoPollTimer)
    autoPollTimer = null
  }
}

function scheduleNextPoll() {
  stopPolling()
  if (!isPageVisible)
    return
  const delayMs = getAdaptiveRefreshIntervalMs()
  autoPollTimer = setTimeout(async () => {
    if (isPageVisible) {
      await ensureTabData(activeTab.value, true)
      if (activeTab.value === 'overview')
        overviewRef.value?.refreshOverviewSubModules?.()
    }
    scheduleNextPoll()
  }, delayMs)
}

function startPolling() {
  scheduleNextPoll()
}

onLoad((query: any) => {
  const resolved = resolveBottomTabKey(query?.tab)
  activeTab.value = resolved.tab
  loadedTabs.value[resolved.tab] = true
  if (resolved.marketSegment)
    marketSegment.value = resolved.marketSegment

  ensureTabData(activeTab.value)
  loadUserPortfolio()
})

onMounted(() => {
  if (!loadedTabs.value[activeTab.value]) {
    loadedTabs.value[activeTab.value] = true
    ensureTabData(activeTab.value)
  }
  startPolling()
})

onShow(() => {
  isPageVisible = true
  const stale = Date.now() - lastActiveFetchMs > getAdaptiveRefreshIntervalMs()
  if (stale) {
    ensureTabData(activeTab.value, true)
    if (activeTab.value === 'overview')
      overviewRef.value?.refreshOverviewSubModules?.()
  }
  startPolling()
})

onHide(() => {
  isPageVisible = false
  stopPolling()
})

onUnmounted(() => {
  isPageVisible = false
  stopPolling()
})

onPullDownRefresh(async () => {
  await loadAllData(true)
  uni.stopPullDownRefresh()
})

const vixQuote = computed(() => globalEconomicData.value.find(item => item.symbol === 'VIX'))

function switchTab(rawKey: string) {
  const resolved = resolveBottomTabKey(rawKey)
  const targetTab = resolved.tab
  if (resolved.marketSegment)
    marketSegment.value = resolved.marketSegment

  if (targetTab === activeTab.value)
    return

  tabScrollTops.value[activeTab.value] = currentScrollTop
  activeTab.value = targetTab

  if (!loadedTabs.value[targetTab]) {
    loadedTabs.value[targetTab] = true
    ensureTabData(targetTab)
  }

  uni.vibrateShort?.({ type: 'light' })
  nextTick(() => {
    const savedY = tabScrollTops.value[targetTab] ?? 0
    currentScrollTop = savedY
    uni.pageScrollTo({ scrollTop: savedY, duration: 0 })
  })
}

// ==================== 左右手势滑动支持（touchcancel 严格忽略，防误切页） ====================
let touchStartX = 0
let touchStartY = 0
let touchStartTime = 0
let touchActive = false

function onTouchStart(e: any) {
  const touch = e.touches?.[0] || e.changedTouches?.[0]
  if (touch) {
    touchActive = true
    touchStartX = touch.clientX
    touchStartY = touch.clientY
    touchStartTime = Date.now()
  }
}

function onTouchCancel() {
  touchActive = false
}

function onTouchEnd(e: any) {
  if (!touchActive)
    return
  touchActive = false
  const touch = e.changedTouches?.[0] || e.touches?.[0]
  if (!touch)
    return

  const nextTab = evaluateTabSwipe(
    activeTab.value,
    BOTTOM_TAB_KEYS,
    touch.clientX - touchStartX,
    touch.clientY - touchStartY,
    Date.now() - touchStartTime
  )
  if (nextTab !== activeTab.value) {
    switchTab(nextTab)
  }
}
</script>

<template>
  <base-layout>
    <view
      class="cc-page-shell px-3.5 pt-3 space-y-3.5 bg-[#090c0f] text-[#f1efea] min-h-screen"
      @touchstart="onTouchStart"
      @touchend="onTouchEnd"
      @touchcancel="onTouchCancel"
    >
      <!-- ==================== 1. 总览 Tab ==================== -->
      <view v-if="loadedTabs.overview" v-show="activeTab === 'overview'">
        <TabOverview
          ref="overviewRef"
          :cycle-data="cycleData"
          :strategy-data="strategyData"
          :overview-fetched-at="overviewFetchedAt"
          :overview-error="overviewError"
          :overview-using-seed="overviewUsingSeed"
          @switch-tab="switchTab"
          @retry-overview="ensureTabData('overview', true)"
        />
      </view>

      <!-- ==================== 2. 市场 Tab ==================== -->
      <view v-if="loadedTabs.markets" v-show="activeTab === 'markets'">
        <TabMarketHub
          :initial-segment="marketSegment"
          :global-economic-data="globalEconomicData"
          :vix-quote="vixQuote"
          :hk-us-macro-list="hkUsMacroList"
          :tw-macro-list="twMacroList"
          :crypto-agri-list="cryptoAgriList"
          :kr-macro-list="krMacroList"
          :kr-sector-list="krSectorList"
          :jp-macro-list="jpMacroList"
          :jp-sector-list="jpSectorList"
          :asia-macro-list="asiaMacroList"
          :fx-list="fxList"
          :precious-metals="preciousMetals"
          :industrial-metals="industrialMetals"
          :other-metals="otherMetals"
          :sector-list="sectorList"
          :ai-products="aiProducts"
          :ai-equipments="aiEquipments"
          :robot-drive-list="robotDriveList"
          :robot-sense-list="robotSenseList"
          :semi-equip-list="semiEquipList"
          :semi-material-list="semiMaterialList"
          :dividend-core-list="dividendCoreList"
          :markets-fetched-at="marketsFetchedAt"
          :markets-error="marketsError"
          :markets-using-seed="marketsUsingSeed"
          :sectors-fetched-at="sectorsFetchedAt"
          :sectors-error="sectorsError"
          :sectors-using-seed="sectorsUsingSeed"
          @retry-markets="loadMarketsTabData(true)"
          @retry-sectors="loadSectorsTabData(true)"
        />
      </view>

      <!-- ==================== 3. 研究 Tab ==================== -->
      <view v-if="loadedTabs.research" v-show="activeTab === 'research'">
        <TabResearch
          :strategy-data="strategyData"
          :is-backtesting="isBacktesting"
          :backtest-result="backtestResult"
          :backtest-error="backtestError"
          :is-sample-result="isSampleBacktestResult"
          @run-backtest="handleRunBacktest"
          @show-sample-backtest="handleShowSampleBacktest"
        />
      </view>

      <!-- ==================== 4. 账本 Tab ==================== -->
      <view v-if="loadedTabs.portfolio" v-show="activeTab === 'portfolio'">
        <TabPortfolio
          v-model:portfolio-mode="portfolioMode"
          :current-display-portfolio="currentDisplayPortfolio"
          :cash-pool-ticks="cashPoolTicks"
          :strategy-data="strategyData"
          :user-portfolio="userPortfolio"
          :current-week-key="currentWeekKey"
          :current-week-check-in="currentWeekCheckIn"
          :last-persist-error="lastPersistError"
          :rebuy-rule-status="rebuyRuleStatus"
          :dividend-valuation="dividendValuation"
          :growth-valuation="growthValuation"
          @weekly-check-in="handleWeeklyCheckIn"
          @rebuy-cash-pool="handleRebuyCashPool"
          @save-portfolio="handleSavePortfolio"
          @reset-portfolio="handleResetPortfolio"
        />
      </view>

      <!-- 底部 4 Tab 主导航：总览 / 市场 / 研究 / 账本 -->
      <view class="cc-bottom-tabbar fixed bottom-0 left-0 right-0 z-40">
        <view class="grid grid-cols-4 min-h-[56px] py-1">
          <view
            v-for="item in tabs"
            :key="item.key"
            class="flex flex-col items-center justify-center relative cursor-pointer active:opacity-80"
            @tap="switchTab(item.key)"
          >
            <view
              class="w-5 h-[2px] rounded-full mb-1 transition-all"
              :class="activeTab === item.key ? 'bg-[#e5b352]' : 'bg-transparent'"
            />
            <image
              :src="activeTab === item.key ? item.iconOn : item.iconOff"
              class="w-[21px] h-[21px] shrink-0"
              mode="aspectFit"
            />
            <text
              class="text-[10px] leading-tight mt-1 whitespace-nowrap"
              :class="activeTab === item.key ? 'text-[#e5b352] font-bold' : 'text-[#6e7b86] font-medium'"
            >
              {{ item.label }}
            </text>
          </view>
        </view>
      </view>
    </view>
  </base-layout>
</template>

<style scoped>
.cc-page-shell {
  padding-bottom: calc(82px + env(safe-area-inset-bottom));
}

.cc-bottom-tabbar {
  background: rgba(11, 14, 17, 0.96);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: env(safe-area-inset-bottom);
}
</style>
