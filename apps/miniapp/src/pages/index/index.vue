<script lang="ts" setup>
import { ref, shallowRef, computed, onMounted } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import TabCycle from './components/TabCycle.vue'
import TabMarkets from './components/TabMarkets.vue'
import TabSectors from './components/TabSectors.vue'
import TabPortfolio from './components/TabPortfolio.vue'
import {
  getMarketQuotes,
  getIndustryMetrics,
  getCycleOverview,
  getCombinedStrategySignal,
  runBacktest,
  getPortfolioOverview,
  type MarketQuote,
  type IndustryMetric,
  type CycleOverview,
  type CombinedSignal,
  type BacktestOutput,
  type PortfolioOverview as IPortfolioOverview
} from '@/api/cycle'
import {
  FALLBACK_BACKTEST_DATA,
  SEED_AI_EQUIPMENTS,
  SEED_AI_PRODUCTS,
  SEED_ASIA_MACRO,
  SEED_CYCLE_OVERVIEW,
  SEED_DEMO_PORTFOLIO,
  SEED_FX_LIST,
  SEED_GLOBAL_ECONOMIC_DATA,
  SEED_INDUSTRIAL_METALS,
  SEED_JP_MACRO,
  SEED_JP_SECTORS,
  SEED_KR_MACRO,
  SEED_KR_SECTORS,
  SEED_OTHER_METALS,
  SEED_PRECIOUS_METALS,
  SEED_SECTOR_LIST,
  SEED_STRATEGY_SIGNAL
} from '@/data/seeds'
import { BOTTOM_TAB_KEYS, evaluateTabSwipe, type BottomTabKey } from '@/utils/swipe'
import { usePortfolio } from '@/composables/usePortfolio'

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

// 底部 4 Tab 导航（融合方向一「策略 / 行情 / 产业 / 账本」与方向四同心环视觉图标，兼容小程序 WXML）
const tabs = [
  {
    key: 'cycle',
    label: '策略',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxjaXJjbGUgY3g9IjEyIiBjeT0iMTIiIHI9IjkiLz48Y2lyY2xlIGN4PSIxMiIgY3k9IjEyIiByPSI0LjUiLz48cGF0aCBkPSJNMTIgNy41VjEybDIuNSAyIi8+PC9zdmc+',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxjaXJjbGUgY3g9IjEyIiBjeT0iMTIiIHI9IjkiLz48Y2lyY2xlIGN4PSIxMiIgY3k9IjEyIiByPSI0LjUiLz48cGF0aCBkPSJNMTIgNy41VjEybDIuNSAyIi8+PC9zdmc+'
  },
  {
    key: 'markets',
    label: '行情',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0zIDE2bDUtNSA0IDQgOS05Ii8+PHBhdGggZD0iTTE2IDZoNXY1Ii8+PC9zdmc+',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxwYXRoIGQ9Ik0zIDE2bDUtNSA0IDQgOS05Ii8+PHBhdGggZD0iTTE2IDZoNXY1Ii8+PC9zdmc+'
  },
  {
    key: 'sectors',
    label: '产业',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMy41IiB5PSIxMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMTMuNSIgd2lkdGg9IjciIGhlaWdodD0iNyIgcng9IjEuNSIvPjwvc3ZnPg==',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMy41IiB5PSIxMy41IiB3aWR0aD0iNyIgaGVpZ2h0PSI3IiByeD0iMS41Ii8+PHJlY3QgeD0iMTMuNSIgeT0iMTMuNSIgd2lkdGg9IjciIGhlaWdodD0iNyIgcng9IjEuNSIvPjwvc3ZnPg=='
  },
  {
    key: 'portfolio',
    label: '账本',
    iconOff: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiM2RTdCODYiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjQiIHk9IjMiIHdpZHRoPSIxNiIgaGVpZ2h0PSIxOCIgcng9IjIuNSIvPjxsaW5lIHgxPSI4IiB5MT0iOSIgeDI9IjE2IiB5Mj0iOSIvPjxsaW5lIHgxPSI4IiB5MT0iMTMuNSIgeDI9IjE0IiB5Mj0iMTMuNSIvPjxsaW5lIHgxPSI4IiB5MT0iMTcuNSIgeDI9IjExLjUiIHkyPSIxNy41Ii8+PC9zdmc+',
    iconOn: 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIiBzdHJva2U9IiNFNUIzNTIiIHN0cm9rZS13aWR0aD0iMS45IiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjxyZWN0IHg9IjQiIHk9IjMiIHdpZHRoPSIxNiIgaGVpZ2h0PSIxOCIgcng9IjIuNSIvPjxsaW5lIHgxPSI4IiB5MT0iOSIgeDI9IjE2IiB5Mj0iOSIvPjxsaW5lIHgxPSI4IiB5MT0iMTMuNSIgeDI9IjE0IiB5Mj0iMTMuNSIvPjxsaW5lIHgxPSI4IiB5MT0iMTcuNSIgeDI9IjExLjUiIHkyPSIxNy41Ii8+PC9zdmc+'
  }
] as const
const activeTab = ref('cycle')
// 记录已经挂载激活过的 Tab（懒渲染 + v-show 保留 DOM，杜绝切换时反复重绘）
interface TabState {
  cycle: boolean
  markets: boolean
  sectors: boolean
  portfolio: boolean
  [key: string]: boolean
}

const loadedTabs = ref<TabState>({
  cycle: true,
  markets: false,
  sectors: false,
  portfolio: false
})



// ==================== 各板块响应式状态（离线种子已抽离至 @/data/seeds，保留 shallowRef 浅响应式） ====================
const globalEconomicData = shallowRef<MarketQuote[]>([...SEED_GLOBAL_ECONOMIC_DATA])
const sectorList = shallowRef<IndustryMetric[]>([...SEED_SECTOR_LIST])
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
const cycleData = shallowRef<CycleOverview>({ ...SEED_CYCLE_OVERVIEW })
const strategyData = shallowRef<CombinedSignal>({ ...SEED_STRATEGY_SIGNAL })
const portfolioData = ref<IPortfolioOverview>({ ...SEED_DEMO_PORTFOLIO })

const backtestResult = shallowRef<BacktestOutput | null>(null)
const isBacktesting = ref(false)

// 时间格式化函数（精确到秒）
const formatTimestamp = (d = new Date()) => {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${day} ${hh}:${mm}:${ss}`
}

const lastUpdatedTime = ref(formatTimestamp())

// ==================== 用户专属自定义模拟账本系统（已抽离至 usePortfolio） ====================
const {
  portfolioMode,
  userPortfolio,
  currentDisplayPortfolio,
  cashPoolTicks,
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

// 分模块按需加载数据（仅在真实命中后端数据时返回 true 并刷新时间戳）
const loadCycleTabData = async (force = false): Promise<boolean> => {
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
  } catch {
    // 优雅离线降级
  }
  return updated
}

const loadMarketsTabData = async (force = false): Promise<boolean> => {
  try {
    const mRes = await getMarketQuotes(undefined, force)
    if (mRes?.data?.length) {
      const quoteMap = new Map(mRes.data.map(q => [q.symbol.toUpperCase(), q]))
      globalEconomicData.value = globalEconomicData.value.map(item => quoteMap.get(item.symbol.toUpperCase()) ?? item)
      krMacroList.value = krMacroList.value.map(item => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q ? { ...item, price: q.price, change_pct: q.change_pct } : item
      })
      jpMacroList.value = jpMacroList.value.map(item => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q ? { ...item, price: q.price, change_pct: q.change_pct } : item
      })
      asiaMacroList.value = asiaMacroList.value.map(item => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q ? { ...item, price: q.price, change_pct: q.change_pct } : item
      })
      fxList.value = fxList.value.map(item => {
        const q = quoteMap.get(item.symbol.toUpperCase())
        return q ? { ...item, price: q.price.toFixed(2), change_pct: q.change_pct } : item
      })
      return true
    }
  } catch {
    // 优雅离线降级
  }
  return false
}

const loadSectorsTabData = async (force = false): Promise<boolean> => {
  try {
    const secRes = await getIndustryMetrics(force)
    if (secRes?.data?.length) {
      sectorList.value = secRes.data
      return true
    }
  } catch {
    // 优雅离线降级
  }
  return false
}

const loadPortfolioTabData = async (force = false): Promise<boolean> => {
  try {
    const pRes = await getPortfolioOverview(force)
    if (pRes?.data) {
      portfolioData.value = pRes.data
      return true
    }
  } catch {
    // 优雅离线降级
  }
  return false
}

// 调度加载指定 Tab 数据（按需触发）
const ensureTabData = async (tabKey: string, force = false) => {
  let updated = false
  if (tabKey === 'cycle') {
    updated = await loadCycleTabData(force)
  } else if (tabKey === 'markets') {
    updated = await loadMarketsTabData(force)
  } else if (tabKey === 'sectors') {
    updated = await loadSectorsTabData(force)
  } else if (tabKey === 'portfolio') {
    updated = await loadPortfolioTabData(force)
  }
  if (updated) {
    lastUpdatedTime.value = formatTimestamp()
  }
}

// 刷新当前活动 Tab，或全量重刷
const loadAllData = async (force = false) => {
  const results = await Promise.all([
    loadCycleTabData(force),
    loadedTabs.value.markets ? loadMarketsTabData(force) : Promise.resolve(false),
    loadedTabs.value.sectors ? loadSectorsTabData(force) : Promise.resolve(false),
    loadedTabs.value.portfolio ? loadPortfolioTabData(force) : Promise.resolve(false)
  ])
  if (results.some(Boolean)) {
    lastUpdatedTime.value = formatTimestamp()
  }
}

const handleRunBacktest = async () => {
  isBacktesting.value = true
  try {
    const res = await runBacktest({
      low_line: 0.16,
      high_line: 0.28,
      weekly_amount: 1000.0,
      days: 800
    })
    if (res?.data && (res.code === 0 || res.code === undefined)) {
      backtestResult.value = res.data
      uni.showToast({ title: '800交易日回测已完成', icon: 'success' })
    } else {
      backtestResult.value = FALLBACK_BACKTEST_DATA
      uni.showToast({ title: '已呈现离线回测快照', icon: 'none' })
    }
  } catch {
    backtestResult.value = FALLBACK_BACKTEST_DATA
    uni.showToast({ title: '接口未连通，已展示离线回测快照', icon: 'none' })
  } finally {
    isBacktesting.value = false
  }
}

onLoad(() => {
  // onLoad 阶段比 onMounted 早 50~150ms 触发，使网络拉取与 DOM 初始挂载完全并行
  ensureTabData(activeTab.value)
  loadUserPortfolio()
})

onMounted(() => {
  // 兜底保障
  if (!loadedTabs.value[activeTab.value]) {
    ensureTabData(activeTab.value)
  }
})

onPullDownRefresh(async () => {
  await loadAllData(true)
  uni.stopPullDownRefresh()
})

const vixQuote = computed(() => globalEconomicData.value.find(item => item.symbol === 'VIX'))

const switchTab = (key: string) => {
  activeTab.value = key
  if (!loadedTabs.value[key]) {
    loadedTabs.value[key] = true
    ensureTabData(key)
  }
  uni.vibrateShort?.({ type: 'light' })
  uni.pageScrollTo({ scrollTop: 0, duration: 150 })
}

// ==================== 左右手势滑动支持 ====================
let touchStartX = 0
let touchStartY = 0
let touchStartTime = 0

const onTouchStart = (e: any) => {
  const touch = e.touches?.[0] || e.changedTouches?.[0]
  if (touch) {
    touchStartX = touch.clientX
    touchStartY = touch.clientY
    touchStartTime = Date.now()
  }
}

const onTouchEnd = (e: any) => {
  const touch = e.changedTouches?.[0] || e.touches?.[0]
  if (!touch) return

  const nextTab = evaluateTabSwipe(
    activeTab.value as BottomTabKey,
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
      @touchcancel="onTouchEnd"
    >
      <!-- ==================== 1. 策略 Tab ==================== -->
      <view v-if="loadedTabs.cycle" v-show="activeTab === 'cycle'">
        <TabCycle
          :cycle-data="cycleData"
          :strategy-data="strategyData"
          :last-updated-time="lastUpdatedTime"
          :is-backtesting="isBacktesting"
          :backtest-result="backtestResult"
          @switch-tab="switchTab"
          @run-backtest="handleRunBacktest"
        />
      </view>

      <!-- ==================== 2. 行情 Tab ==================== -->
      <view v-if="loadedTabs.markets" v-show="activeTab === 'markets'">
        <TabMarkets
          :global-economic-data="globalEconomicData"
          :vix-quote="vixQuote"
          :kr-macro-list="krMacroList"
          :kr-sector-list="krSectorList"
          :jp-macro-list="jpMacroList"
          :jp-sector-list="jpSectorList"
          :asia-macro-list="asiaMacroList"
          :fx-list="fxList"
          :precious-metals="preciousMetals"
          :industrial-metals="industrialMetals"
          :other-metals="otherMetals"
          :last-updated-time="lastUpdatedTime"
        />
      </view>

      <!-- ==================== 3. 产业 Tab ==================== -->
      <view v-if="loadedTabs.sectors" v-show="activeTab === 'sectors'">
        <TabSectors
          :sector-list="sectorList"
          :ai-products="aiProducts"
          :ai-equipments="aiEquipments"
          :last-updated-time="lastUpdatedTime"
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
          @weekly-check-in="handleWeeklyCheckIn"
          @rebuy-cash-pool="handleRebuyCashPool"
          @save-portfolio="handleSavePortfolio"
          @reset-portfolio="handleResetPortfolio"
        />
      </view>

      <!-- 底部暗场仪器 TabBar（参考 dir-1 与 dir-4 命名及图标） -->
      <view class="cc-bottom-tabbar fixed bottom-0 left-0 right-0 z-40">
        <view class="grid grid-cols-4 h-[56px]">
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
              class="w-[21px] h-[21px]"
              mode="aspectFit"
            />
            <text
              class="text-[10px] leading-none mt-1"
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
