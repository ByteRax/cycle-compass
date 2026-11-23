<script lang="ts" setup>
import { reactive, ref } from 'vue'
import MarketCard from '@/components/finance/MarketCard.vue'
import SectorCard from '@/components/finance/SectorCard.vue'
import type { MarketQuote } from '@/api/cycle'
import type { FxQuoteItem, PriceItem, SymbolPriceItem } from '@/data/seeds'

defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps<{
  globalEconomicData: MarketQuote[]
  vixQuote?: MarketQuote
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
  { key: 'apac', label: '日韩亚太' },
  { key: 'metals', label: '有色金属' }
] as const

type MarketSubTabKey = (typeof marketSubTabs)[number]['key']

const marketSubTab = ref<MarketSubTabKey>('global')
const loadedSubTabs = reactive<Record<MarketSubTabKey, boolean>>({
  global: true,
  apac: false,
  metals: false
})

const switchSubTab = (key: MarketSubTabKey) => {
  marketSubTab.value = key
  loadedSubTabs[key] = true
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 轻量级子分组切换条（使用原子类避免微信小程序 scoped 样式隔离失效） -->
    <view class="flex items-center justify-between border-b border-white/[0.08] pb-0.5">
      <view class="flex items-center space-x-5">
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

    <!-- 子页：全球宏观（懒挂载 + v-show 保活，避免切换销毁重建卡片） -->
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
          <text class="text-sm font-bold text-[#f1efea]">全球宏观指标</text>
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
    </view>

    <!-- 子页：日韩亚太 -->
    <view v-if="loadedSubTabs.apac" v-show="marketSubTab === 'apac'" class="space-y-3.5">
      <view class="space-y-2">
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

    <!-- 子页：有色金属 -->
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
