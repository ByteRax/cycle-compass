<script lang="ts" setup>
import type { MarketQuote } from '@/api/cycle'
import type { FxQuoteItem, PriceItem, SymbolPriceItem } from '@/data/seeds'
import { reactive, ref } from 'vue'
import MarketCard from '@/components/finance/MarketCard.vue'
import SectorCard from '@/components/finance/SectorCard.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps<{
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
  lastUpdatedTime: string
}>()

const marketSubTabs = [
  { key: 'global', label: '全球宏观' },
  { key: 'hk_us', label: '港美中概' },
  { key: 'apac', label: '亚太日韩台' },
  { key: 'metals', label: '有色金属' }
] as const

type MarketSubTabKey = (typeof marketSubTabs)[number]['key']

const marketSubTab = ref<MarketSubTabKey>('global')
const loadedSubTabs = reactive<Record<MarketSubTabKey, boolean>>({
  global: true,
  hk_us: false,
  apac: false,
  metals: false
})

function switchSubTab (key: MarketSubTabKey) {
  marketSubTab.value = key
  loadedSubTabs[key] = true
}

function openDivergence () {
  uni.navigateTo({
    url: '/pages-sub/divergence/index?id=etf_macro'
  })
}

function openFinanceOverview () {
  uni.navigateTo({
    url: '/pages-sub/finance/index'
  })
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 金融资产核心入口卡片（设计稿 01） -->
    <view
      class="rounded-2xl bg-[#12161b] border border-[#e5b352]/30 p-3.5 shadow-sm"
    >
      <view class="flex items-start justify-between">
        <view>
          <text class="text-[19px] font-bold text-[#e5b352] tracking-tight block">金融资产</text>
          <text class="text-[12px] text-[#a9b3bd] mt-0.5 block">跨市场趋势 · 统一强弱比较</text>
        </view>
        <view class="w-8 h-8 rounded-full bg-[#e5b352]/14 border border-[#e5b352]/40 flex items-center justify-center shrink-0">
          <view class="w-4 h-4 rounded-full border-2 border-[#e5b352] flex items-center justify-center">
            <view class="w-1.5 h-1.5 rounded-full bg-[#e5b352]" />
          </view>
        </view>
      </view>

      <view class="grid grid-cols-3 gap-2 mt-3 pt-2.5 border-t border-white/[0.08]">
        <view>
          <text class="text-[12px] font-bold text-[#e5b352] block">七档温度</text>
          <text class="text-[11px] text-[#8d99a4] mt-0.5 block">趋势识别</text>
        </view>
        <view>
          <text class="text-[12px] font-bold text-[#e5b352] block">七类图表</text>
          <text class="text-[11px] text-[#8d99a4] mt-0.5 block">深度研究</text>
        </view>
        <view>
          <text class="text-[12px] font-bold text-[#e5b352] block">每日快照</text>
          <text class="text-[11px] text-[#8d99a4] mt-0.5 block">可追溯数据</text>
        </view>
      </view>

      <view
        class="mt-3 min-h-[44px] rounded-xl bg-[#e5b352] flex items-center justify-center active:opacity-85 transition-opacity cursor-pointer"
        @tap="openFinanceOverview"
      >
        <text class="text-[14px] font-bold text-[#090c0f]">进入金融资产全景 →</text>
      </view>
    </view>

    <!-- 同期涨跌分化快捷全景次入口 -->
    <view
      class="flex items-center justify-between rounded-xl bg-[#12161b] px-3 py-2.5 border border-white/[0.08] active:opacity-80 transition-all cursor-pointer min-h-[44px]"
      @tap="openDivergence"
    >
      <view class="flex items-center space-x-2">
        <text class="text-xs font-bold text-[#e5b352]">⚡</text>
        <view>
          <text class="text-xs font-bold text-[#f1efea] block">同期涨跌分化全景看板</text>
          <text class="text-[11px] text-[#8d99a4]">多资产/行业同期走势分化与截面热度监测</text>
        </view>
      </view>
      <view class="flex items-center space-x-0.5 shrink-0">
        <text class="text-[11px] text-[#e5b352] font-medium">趋势研究</text>
        <text class="text-xs text-[#e5b352]">〉</text>
      </view>
    </view>

    <!-- 轻量级子分组切换条（使用原子类避免微信小程序 scoped 样式隔离失效） -->
    <view class="flex items-center justify-between border-b border-white/[0.08] pb-0.5">
      <view class="flex items-center space-x-4">
        <view
          v-for="sub in marketSubTabs"
          :key="sub.key"
          class="pb-2 text-[13px] transition-all cursor-pointer border-b-2"
          :class="marketSubTab === sub.key ? 'text-[#e5b352] font-bold border-[#e5b352]' : 'text-[#8d99a4] font-medium border-transparent'"
          @tap="switchSubTab(sub.key)"
        >
          {{ sub.label }}
        </view>
      </view>
      <text class="text-[10px] font-mono text-[#64707c] pb-2">{{ lastUpdatedTime.slice(5, 16) }}</text>
    </view>

    <!-- 子页 1：全球宏观（懒挂载 + v-show 保活） -->
    <view v-if="loadedSubTabs.global" v-show="marketSubTab === 'global'" class="space-y-3.5">
      <view
        v-if="vixQuote"
        class="rounded-xl bg-[#f05a50]/14 border border-[#f05a50]/35 px-3.5 py-3 flex items-center justify-between"
      >
        <view>
          <view class="inline-flex items-center rounded-full bg-[#f05a50]/22 px-2 py-0.5 text-[10px] font-bold text-[#ff736a]">
            波动率预警
          </view>
          <text class="text-sm font-bold text-[#f1efea] block mt-1">恐慌指数 VIX {{ vixQuote.price.toFixed(2) }}</text>
        </view>
        <view class="text-right">
          <text class="text-lg font-bold font-mono cc-tabular text-[#ff736a]">
            {{ vixQuote.change_pct >= 0 ? '+' : '' }}{{ vixQuote.change_pct.toFixed(2) }}%
          </text>
          <text class="text-[10px] font-mono text-[#8d99a4] block">{{ lastUpdatedTime.slice(5, 16) }}</text>
        </view>
      </view>

      <view class="space-y-2">
        <view class="flex items-center justify-between px-0.5">
          <text class="text-sm font-bold text-[#f1efea]">全球宏观核心指标</text>
          <text class="text-[11px] font-mono text-[#64707c]">{{ globalEconomicData.length }} 项</text>
        </view>
        <view class="grid grid-cols-2 gap-2">
          <MarketCard
            v-for="item in globalEconomicData"
            :key="item.symbol"
            :symbol="item.symbol"
            :name="item.name"
            :price="item.price.toString()"
            :change-percent="item.change_pct"
            :unit="item.currency"
          />
        </view>
      </view>

      <view v-if="cryptoAgriList && cryptoAgriList.length" class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">流动性与大宗风向标</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in cryptoAgriList"
            :key="item.symbol"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>
    </view>

    <!-- 子页 2：港美中概 -->
    <view v-if="loadedSubTabs.hk_us" v-show="marketSubTab === 'hk_us'" class="space-y-3.5">
      <view class="space-y-2">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">港股与离岸中国核心</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in (hkUsMacroList || []).slice(0, 5)"
            :key="item.symbol"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">美股三大指数与基准利率</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in (hkUsMacroList || []).slice(5)"
            :key="item.symbol"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>
    </view>

    <!-- 子页 3：亚太日韩台 -->
    <view v-if="loadedSubTabs.apac" v-show="marketSubTab === 'apac'" class="space-y-3.5">
      <view v-if="twMacroList && twMacroList.length" class="space-y-2">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">台湾半导体核心市场</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard v-for="item in twMacroList" :key="item.symbol" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">韩国综合与核心产业</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard v-for="item in krMacroList" :key="item.symbol" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
          <SectorCard v-for="item in krSectorList" :key="item.name" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">日本综合与核心产业</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard v-for="item in jpMacroList" :key="item.symbol" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
          <SectorCard v-for="item in jpSectorList" :key="item.name" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">亚洲综合与汇率</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard v-for="item in asiaMacroList" :key="item.symbol" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
          <MarketCard
            v-for="item in fxList"
            :key="item.symbol"
            :symbol="item.symbol"
            :name="item.name"
            :price="item.price"
            :change-percent="item.change_pct"
            :unit="item.currency"
          />
        </view>
      </view>
    </view>

    <!-- 子页 4：有色金属 -->
    <view v-if="loadedSubTabs.metals" v-show="marketSubTab === 'metals'" class="space-y-3.5">
      <view class="space-y-2">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">金银贵金属</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard v-for="item in preciousMetals" :key="item.name" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">工业金属</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard v-for="item in industrialMetals" :key="item.name" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">小金属与稀有金属</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard v-for="item in otherMetals" :key="item.name" :name="item.name" :change-percent="item.change_pct" :price="item.price" />
        </view>
      </view>
    </view>

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">本程序展示所有公开查询数据，仅供参考，不构成任何投资建议</text>
    </view>
  </view>
</template>
