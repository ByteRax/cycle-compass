<script lang="ts" setup>
import type { IndustryMetric, MarketQuote } from '@/api/cycle'
import type { FxQuoteItem, PriceItem, SymbolPriceItem } from '@/data/seeds'
import { computed, reactive, ref, watch } from 'vue'
import MarketCard from '@/components/finance/MarketCard.vue'
import SectorCard from '@/components/finance/SectorCard.vue'
import { buildFinanceUrl } from '@/utils/financeRoute'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = withDefaults(defineProps<{
  initialSegment?: 'quotes' | 'industries'
  globalEconomicData: MarketQuote[]
  vixQuote?: MarketQuote
  hkUsMacroList?: SymbolPriceItem[]
  twMacroList?: SymbolPriceItem[]
  cryptoAgriList?: SymbolPriceItem[]
  krMacroList: SymbolPriceItem[]
  krSectorList: PriceItem[]
  jpMacroList: SymbolPriceItem[]
  jpSectorList: PriceItem[]
  asiaMacroList: SymbolPriceItem[]
  fxList: FxQuoteItem[]
  preciousMetals: PriceItem[]
  industrialMetals: PriceItem[]
  otherMetals: PriceItem[]
  sectorList: IndustryMetric[]
  aiProducts: PriceItem[]
  aiEquipments: PriceItem[]
  robotDriveList?: PriceItem[]
  robotSenseList?: PriceItem[]
  semiEquipList?: PriceItem[]
  semiMaterialList?: PriceItem[]
  dividendCoreList?: PriceItem[]
  marketsFetchedAt: string
  marketsError: string
  marketsUsingSeed: boolean
  sectorsFetchedAt: string
  sectorsError: string
  sectorsUsingSeed: boolean
}>(), {
  initialSegment: 'quotes',
  vixQuote: undefined,
  hkUsMacroList: () => [],
  twMacroList: () => [],
  cryptoAgriList: () => [],
  robotDriveList: () => [],
  robotSenseList: () => [],
  semiEquipList: () => [],
  semiMaterialList: () => [],
  dividendCoreList: () => []
})

const emit = defineEmits<{
  (e: 'retryMarkets'): void
  (e: 'retrySectors'): void
}>()

const showVixGlossary = ref(false)
const vixLevelText = computed(() => {
  const val = props.vixQuote?.price
  if (val === undefined || val === null)
    return ''
  if (val < 15)
    return '当前 < 15：市场很淡定，大家普遍偏乐观'
  if (val < 25)
    return '当前 15 ~ 25：正常波动区间，情绪比较平稳'
  if (val < 35)
    return '当前 25 ~ 35：有点紧张，市场短期波动变大'
  return '当前 ≥ 35：极度恐慌，暴涨暴跌常见但也可能孕育机会'
})

const activeSegment = ref<'quotes' | 'industries'>(props.initialSegment || 'quotes')
watch(() => props.initialSegment, (val) => {
  if (val)
    activeSegment.value = val
})

const marketSubTabs = [
  { key: 'global', label: '全球宏观' },
  { key: 'hk_us', label: '港美中概' },
  { key: 'apac', label: '亚太日韩台' },
  { key: 'metals', label: '有色金属' }
] as const

type MarketSubTabKey = (typeof marketSubTabs)[number]['key']
const marketSubTab = ref<MarketSubTabKey>('global')
const loadedMarketSubTabs = reactive<Record<MarketSubTabKey, boolean>>({
  global: true,
  hk_us: false,
  apac: false,
  metals: false
})

function switchMarketSubTab(key: MarketSubTabKey) {
  marketSubTab.value = key
  loadedMarketSubTabs[key] = true
}

const sectorSubTabs = [
  { key: 'global', label: '全球 24 产业' },
  { key: 'ai', label: 'AI 算力链' },
  { key: 'robot', label: '具身机器人' },
  { key: 'semi', label: '芯片半导体' },
  { key: 'dividend', label: '红利防守' }
] as const

type SectorSubTabKey = (typeof sectorSubTabs)[number]['key']
const sectorSubTab = ref<SectorSubTabKey>('global')
const loadedSectorSubTabs = reactive<Record<SectorSubTabKey, boolean>>({
  global: true,
  ai: false,
  robot: false,
  semi: false,
  dividend: false
})

function switchSectorSubTab(key: SectorSubTabKey) {
  sectorSubTab.value = key
  loadedSectorSubTabs[key] = true
}

const industryRefDate = computed(() => {
  const first = props.sectorList?.[0]
  return first?.data_date || first?.as_of || '2026-11-20'
})

function openSearch() {
  uni.navigateTo({ url: '/pages-sub/finance/search' })
}

function openFinanceOverview(poolId = 'universe_macro') {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/index', { scope_id: poolId })
  })
}

function openDivergence(poolId = 'universe_macro') {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/divergence/index', {
      asset_id: poolId,
      scope_id: poolId
    })
  })
}

function openIndustryDetail(item: IndustryMetric) {
  uni.navigateTo({
    url: buildFinanceUrl('/pages-sub/finance/sector-detail', {
      asset_id: `sector:${item.sector_id}`,
      title: item.name
    })
  })
}

function openQuotePoolOrSearch(name: string, symbol?: string) {
  const s = (symbol || '').toUpperCase()
  const knownAssets: Record<string, string> = {
    'CN10Y': 'BOND.CN.10Y',
    'US10Y': 'BOND.US.10Y',
    'HSI': 'INDEX.HK.HSI',
    'SPX': 'INDEX.US.SPX',
    'NDX': 'INDEX.US.NDX',
    'GOLD_OZ': 'COMM.GLOBAL.XAU',
    'USD/CNH': 'FX.USDCNH'
  }
  if (knownAssets[s]) {
    uni.navigateTo({
      url: buildFinanceUrl('/pages-sub/finance/asset', {
        asset_id: knownAssets[s],
        scope_id: 'universe_macro'
      })
    })
    return
  }
  if (name.includes('红利')) {
    uni.navigateTo({
      url: buildFinanceUrl('/pages-sub/finance/asset', {
        asset_id: 'INDEX.CN.930955',
        scope_id: 'universe_macro'
      })
    })
    return
  }
  openSearch()
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 统一搜索入口 -->
    <view
      class="flex items-center gap-2 rounded-xl bg-[#12161b] border border-white/[0.1] px-3 min-h-[42px] active:opacity-80 cursor-pointer"
      @tap="openSearch"
    >
      <text class="text-sm text-[#8d99a4]">🔍</text>
      <text class="text-xs text-[#8d99a4] flex-1">搜索观察池、行业专题、ETF 或代码...</text>
      <text class="text-[11px] text-[#e5b352] font-medium">去搜索 ›</text>
    </view>

    <!-- 金融资产全景入口卡 -->
    <view class="rounded-2xl bg-[#12161b] border border-[#e5b352]/30 p-3.5 shadow-sm">
      <view class="flex items-start justify-between">
        <view>
          <text class="text-[19px] font-bold text-[#e5b352] tracking-tight block">金融资产全景</text>
          <text class="text-[12px] text-[#a9b3bd] mt-0.5 block">跨市场温度 · 多周期矩阵 · 统一强弱比较</text>
        </view>
        <view
          class="rounded-xl bg-[#e5b352] px-3 py-2 text-xs font-bold text-[#090c0f] active:opacity-85 cursor-pointer shrink-0"
          @tap="openFinanceOverview('universe_macro')"
        >
          进入全景 →
        </view>
      </view>
    </view>

    <!-- 分段切换：资产行情 | 产业专题 -->
    <view class="grid grid-cols-2 p-1 rounded-xl bg-[#12161b] border border-white/[0.08]">
      <view
        class="py-2 rounded-lg text-center text-xs font-bold transition-colors cursor-pointer"
        :class="activeSegment === 'quotes' ? 'bg-[#e5b352] text-[#090c0f]' : 'text-[#8d99a4]'"
        @tap="activeSegment = 'quotes'"
      >
        资产行情
      </view>
      <view
        class="py-2 rounded-lg text-center text-xs font-bold transition-colors cursor-pointer"
        :class="activeSegment === 'industries' ? 'bg-[#e5b352] text-[#090c0f]' : 'text-[#8d99a4]'"
        @tap="activeSegment = 'industries'"
      >
        产业专题
      </view>
    </view>

    <!-- ==================== 子视图 A：资产行情 ==================== -->
    <view v-show="activeSegment === 'quotes'" class="space-y-3.5">
      <!-- 独立报错 / 离线说明 -->
      <view
        v-if="marketsError || marketsUsingSeed"
        class="rounded-xl border px-3 py-2 flex items-center justify-between gap-2"
        :class="marketsError ? 'bg-[#f05a50]/12 border-[#f05a50]/35' : 'bg-[#e5b352]/12 border-[#e5b352]/35'"
      >
        <text class="text-[11px] flex-1" :class="marketsError ? 'text-[#ff736a]' : 'text-[#e5b352]'">
          {{ marketsError || '全球行情模块当前展示本地参考数据' }}
        </text>
        <text class="text-[11px] font-bold shrink-0" :class="marketsError ? 'text-[#ff736a]' : 'text-[#e5b352]'" @tap="emit('retryMarkets')">重试</text>
      </view>

      <view
        class="flex items-center justify-between rounded-xl bg-[#12161b] px-3 py-2.5 border border-white/[0.08] active:opacity-80 cursor-pointer min-h-[44px]"
        @tap="openDivergence('universe_macro')"
      >
        <view class="flex items-center space-x-2">
          <text class="text-xs font-bold text-[#e5b352]">⚡</text>
          <view>
            <text class="text-xs font-bold text-[#f1efea] block">全球大类走势与强弱看板</text>
            <text class="text-[10px] text-[#8d99a4]">同起点归一曲线与全收益分解</text>
          </view>
        </view>
        <text class="text-xs text-[#e5b352]">工作台 ›</text>
      </view>

      <!-- 地区分组切换 -->
      <view class="flex items-center justify-between border-b border-white/[0.08] pb-0.5">
        <view class="flex items-center space-x-4 overflow-x-auto no-scrollbar">
          <view
            v-for="sub in marketSubTabs"
            :key="sub.key"
            class="pb-2 text-[13px] whitespace-nowrap cursor-pointer border-b-2 shrink-0"
            :class="marketSubTab === sub.key ? 'text-[#e5b352] font-bold border-[#e5b352]' : 'text-[#8d99a4] font-medium border-transparent'"
            @tap="switchMarketSubTab(sub.key)"
          >
            {{ sub.label }}
          </view>
        </view>
        <text class="text-[10px] font-mono text-[#64707c] pb-2 shrink-0 pl-2">
          {{ marketsFetchedAt ? `同步 ${marketsFetchedAt.slice(11, 16)}` : '参考值' }}
        </text>
      </view>

      <view v-if="loadedMarketSubTabs.global" v-show="marketSubTab === 'global'" class="space-y-3">
        <view
          v-if="vixQuote"
          class="rounded-xl bg-[#f05a50]/14 border border-[#f05a50]/35 px-3.5 py-2.5 flex items-center justify-between cursor-pointer active:opacity-80"
          @tap="showVixGlossary = true"
        >
          <view>
            <view class="flex items-center gap-1.5">
              <text class="text-[10px] font-bold text-[#ff736a]">波动率指标</text>
              <text class="text-[10px] text-[#e5b352] bg-[#e5b352]/15 px-1.5 py-0.5 rounded-full font-medium">💡小白释义</text>
            </view>
            <text class="text-sm font-bold text-[#f1efea] block mt-0.5">恐慌指数 VIX {{ vixQuote.price.toFixed(2) }}</text>
          </view>
          <view class="text-right">
            <text class="text-base font-bold font-mono cc-tabular text-[#ff736a]">
              {{ vixQuote.change_pct >= 0 ? '+' : '' }}{{ vixQuote.change_pct.toFixed(2) }}%
            </text>
            <text class="text-[10px] font-mono text-[#8d99a4] block">{{ vixQuote.as_of || '—' }}</text>
          </view>
        </view>

        <view class="grid grid-cols-2 gap-2">
          <MarketCard
            v-for="item in globalEconomicData"
            :key="item.symbol"
            :symbol="item.symbol"
            :name="item.name"
            :price="item.price"
            :change-percent="item.change_pct"
            :unit="item.currency"
            :disable-auto-nav="true"
            @click="openQuotePoolOrSearch(item.name, item.symbol)"
          />
        </view>

        <view v-if="cryptoAgriList.length" class="space-y-1.5 pt-1">
          <text class="text-xs font-bold text-[#a9b3bd] block">流动性与大宗风向标</text>
          <view class="grid grid-cols-2 gap-2">
            <SectorCard
              v-for="item in cryptoAgriList"
              :key="item.symbol"
              :name="item.name"
              :change-percent="item.change_pct"
              :price="item.price"
              :disable-auto-nav="true"
              @click="openQuotePoolOrSearch(item.name, item.symbol)"
            />
          </view>
        </view>
      </view>

      <view v-if="loadedMarketSubTabs.hk_us" v-show="marketSubTab === 'hk_us'" class="space-y-2.5">
        <text class="text-xs font-bold text-[#a9b3bd] block">港美及离岸核心指数</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in hkUsMacroList"
            :key="item.symbol"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
            :disable-auto-nav="true"
            @click="openQuotePoolOrSearch(item.name, item.symbol)"
          />
        </view>
      </view>

      <view v-if="loadedMarketSubTabs.apac" v-show="marketSubTab === 'apac'" class="space-y-2.5">
        <text class="text-xs font-bold text-[#a9b3bd] block">亚太核心市场与汇率</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in [...twMacroList, ...krMacroList, ...jpMacroList, ...asiaMacroList]"
            :key="item.symbol || item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
            :disable-auto-nav="true"
            @click="openQuotePoolOrSearch(item.name, item.symbol)"
          />
        </view>
      </view>

      <view v-if="loadedMarketSubTabs.metals" v-show="marketSubTab === 'metals'" class="space-y-2.5">
        <text class="text-xs font-bold text-[#a9b3bd] block">贵金属、工业金属与外汇</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in [...preciousMetals, ...industrialMetals, ...otherMetals]"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
            :disable-auto-nav="true"
            @click="openQuotePoolOrSearch(item.name)"
          />
        </view>
      </view>
    </view>

    <!-- ==================== 子视图 B：产业专题 ==================== -->
    <view v-show="activeSegment === 'industries'" class="space-y-3.5">
      <view
        v-if="sectorsError"
        class="rounded-xl border border-[#f05a50]/35 bg-[#f05a50]/12 px-3 py-2 flex items-center justify-between"
      >
        <text class="text-[11px] text-[#ff736a]">{{ sectorsError }}</text>
        <text class="text-[11px] font-bold text-[#ff736a]" @tap="emit('retrySectors')">重试</text>
      </view>

      <!-- 静态参考口径说明（AUDIT 06） -->
      <view class="rounded-xl bg-[#e5b352]/10 border border-[#e5b352]/30 px-3 py-2 flex items-center justify-between text-[11px] text-[#e5b352]">
        <text>产业参考基准日：{{ industryRefDate }}（参考景气指标，非实时逐笔行情）</text>
      </view>

      <view
        class="flex items-center justify-between rounded-xl bg-[#12161b] px-3 py-2.5 border border-[#e5b352]/30 active:opacity-80 cursor-pointer"
        @tap="openFinanceOverview('universe_a_sector')"
      >
        <view>
          <text class="text-xs font-bold text-[#f1efea] block">A股行业全视野观察池</text>
          <text class="text-[10px] text-[#8d99a4]">点击进入可交易行业指数横截面排名</text>
        </view>
        <text class="text-xs text-[#e5b352] font-bold">查看排名 ›</text>
      </view>

      <view class="flex items-center justify-between border-b border-white/[0.08] pb-0.5">
        <view class="flex-1 flex items-center space-x-3 overflow-x-auto no-scrollbar pr-2">
          <view
            v-for="sub in sectorSubTabs"
            :key="sub.key"
            class="pb-2 text-[13px] whitespace-nowrap cursor-pointer border-b-2 shrink-0"
            :class="sectorSubTab === sub.key ? 'text-[#e5b352] font-bold border-[#e5b352]' : 'text-[#8d99a4] font-medium border-transparent'"
            @tap="switchSectorSubTab(sub.key)"
          >
            {{ sub.label }}
          </view>
        </view>
        <text class="text-[10px] font-mono text-[#64707c] pb-2 shrink-0">{{ sectorList.length }}条</text>
      </view>

      <view v-if="loadedSectorSubTabs.global" v-show="sectorSubTab === 'global'" class="space-y-2">
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in sectorList"
            :key="item.sector_id"
            :sector-id="item.sector_id"
            :name="item.name"
            :stage="item.stage"
            :change-percent="item.price_change_pct"
            :price="item.price_index"
            :disable-auto-nav="true"
            @click="openIndustryDetail(item)"
          />
        </view>
      </view>

      <view v-if="loadedSectorSubTabs.ai" v-show="sectorSubTab === 'ai'" class="space-y-2.5">
        <text class="text-xs font-bold text-[#a9b3bd] block">AI 算力产品与设备参考价格</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in [...aiProducts, ...aiEquipments]"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
            :disable-auto-nav="true"
            @click="openSearch"
          />
        </view>
      </view>

      <view v-if="loadedSectorSubTabs.robot" v-show="sectorSubTab === 'robot'" class="space-y-2.5">
        <text class="text-xs font-bold text-[#a9b3bd] block">具身机器人核心链条</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in [...robotDriveList, ...robotSenseList]"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
            :disable-auto-nav="true"
            @click="openSearch"
          />
        </view>
      </view>

      <view v-if="loadedSectorSubTabs.semi" v-show="sectorSubTab === 'semi'" class="space-y-2.5">
        <text class="text-xs font-bold text-[#a9b3bd] block">半导体设备、材料与 EDA</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in [...semiEquipList, ...semiMaterialList]"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
            :disable-auto-nav="true"
            @click="openSearch"
          />
        </view>
      </view>

      <view v-if="loadedSectorSubTabs.dividend" v-show="sectorSubTab === 'dividend'" class="space-y-2.5">
        <text class="text-xs font-bold text-[#a9b3bd] block">高股息防御核心主线</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in dividendCoreList"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
            :disable-auto-nav="true"
            @click="openQuotePoolOrSearch(item.name)"
          />
        </view>
      </view>
    </view>

    <!-- 恐慌指数 VIX 小白释义弹窗 -->
    <view
      v-if="showVixGlossary"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4"
      @tap="showVixGlossary = false"
    >
      <view
        class="w-full max-w-sm bg-[#151a21] border border-white/[0.14] rounded-2xl p-4 space-y-3 text-[#f1efea]"
        @tap.stop
      >
        <view class="flex items-center justify-between border-b border-white/[0.08] pb-2">
          <text class="text-sm font-bold text-[#e5b352]">恐慌指数 VIX · 小白释义</text>
          <text class="text-[#8d99a4] px-2 py-0.5 cursor-pointer" @tap="showVixGlossary = false">✕</text>
        </view>
        <view v-if="vixLevelText" class="px-3 py-2 rounded-xl bg-[#f05a50]/12 border border-[#f05a50]/30">
          <text class="text-xs font-semibold text-[#ff736a] block">{{ vixLevelText }}</text>
        </view>
        <view class="grid grid-cols-3 gap-2 text-center">
          <view class="p-2 rounded-xl bg-[#181e25] border border-white/[0.07]">
            <text class="text-[10px] text-[#8d99a4] block">历史最低</text>
            <text class="text-sm font-bold font-mono text-[#2ad1a3] block mt-0.5">8.56</text>
            <text class="text-[10px] text-[#64707c] block">2017 大牛市</text>
          </view>
          <view class="p-2 rounded-xl bg-[#181e25] border border-white/[0.07]">
            <text class="text-[10px] text-[#8d99a4] block">平时正常范围</text>
            <text class="text-sm font-bold font-mono text-[#e5b352] block mt-0.5">12 ~ 20</text>
            <text class="text-[10px] text-[#64707c] block">大多数交易日</text>
          </view>
          <view class="p-2 rounded-xl bg-[#181e25] border border-white/[0.07]">
            <text class="text-[10px] text-[#8d99a4] block">历史最高</text>
            <text class="text-sm font-bold font-mono text-[#ff736a] block mt-0.5">89.53</text>
            <text class="text-[10px] text-[#64707c] block">08金融海啸/20熔断</text>
          </view>
        </view>
        <view class="space-y-2 text-xs text-[#d1d5db] leading-relaxed">
          <text class="block"><text class="font-bold text-[#f1efea]">一句话看懂：</text>VIX 就像股市的“天气预警等级”，专门衡量大家对未来 30 天美股大涨大跌有多害怕。</text>
          <text class="block"><text class="font-bold text-[#f1efea]">四档范围速查：</text>&lt;15 极度淡定 · 15~25 正常波动 · 25~35 偏紧张 · ≥35 极度恐慌（超 40 往往是历史大震荡）。</text>
          <text class="block"><text class="font-bold text-[#f1efea]">平时怎么用：</text>恐慌指数突然飙高时，别急着追涨杀跌；等它冲到高位再慢慢回落时，往往意味着最慌的时候快过去了。</text>
        </view>
        <view
          class="py-2 text-center rounded-xl bg-[#e5b352] text-[#090c0f] font-bold text-xs cursor-pointer"
          @tap="showVixGlossary = false"
        >
          我知道了
        </view>
      </view>
    </view>

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">本程序展示所有公开查询数据，仅供参考，不构成任何投资建议</text>
    </view>
  </view>
</template>
