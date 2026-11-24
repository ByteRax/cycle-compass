<script lang="ts" setup>
import { reactive, ref } from 'vue'
import SectorCard from '@/components/finance/SectorCard.vue'
import type { IndustryMetric } from '@/api/cycle'
import type { PriceItem } from '@/data/seeds'

defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps<{
  sectorList: IndustryMetric[]
  aiProducts: PriceItem[]
  aiEquipments: PriceItem[]
  robotDriveList?: PriceItem[]
  robotSenseList?: PriceItem[]
  semiEquipList?: PriceItem[]
  semiMaterialList?: PriceItem[]
  dividendCoreList?: PriceItem[]
  lastUpdatedTime: string
}>()

const sectorSubTabs = [
  { key: 'global', label: '全球 24 产业' },
  { key: 'ai', label: 'AI 算力链' },
  { key: 'robot', label: '具身机器人' },
  { key: 'semi', label: '芯片半导体' },
  { key: 'dividend', label: '红利防守' }
] as const

type SectorSubTabKey = (typeof sectorSubTabs)[number]['key']

const sectorSubTab = ref<SectorSubTabKey>('global')
const loadedSubTabs = reactive<Record<SectorSubTabKey, boolean>>({
  global: true,
  ai: false,
  robot: false,
  semi: false,
  dividend: false
})

const switchSubTab = (key: SectorSubTabKey) => {
  sectorSubTab.value = key
  loadedSubTabs[key] = true
}

const openDivergence = () => {
  uni.navigateTo({
    url: '/pages-sub/divergence/index?id=a_shares'
  })
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 同期涨跌分化快捷全景入口 -->
    <view
      class="flex items-center justify-between rounded-xl bg-gradient-to-r from-[#e5b352]/15 via-[#18202a] to-[#12181F] px-3 py-2 border border-[#e5b352]/30 active:opacity-80 transition-all cursor-pointer shadow-sm"
      @tap="openDivergence"
    >
      <view class="flex items-center space-x-2">
        <text class="text-sm">⚡</text>
        <view>
          <text class="text-xs font-bold text-[#f1efea] block">同期涨跌分化全景看板</text>
          <text class="text-[10px] text-[#8d99a4]">多资产/行业同期走势分化与截面热度监测</text>
        </view>
      </view>
      <view class="flex items-center space-x-0.5">
        <text class="text-[10px] text-[#e5b352] font-medium">进入看板</text>
        <text class="text-xs text-[#e5b352]">〉</text>
      </view>
    </view>

    <!-- 轻量级子分组切换条（使用原子类避免微信小程序 scoped 样式隔离失效） -->
    <view class="flex items-center justify-between border-b border-white/[0.08] pb-0.5">
      <view class="flex-1 flex items-center space-x-3 overflow-x-auto no-scrollbar pr-2">
        <view
          v-for="sub in sectorSubTabs"
          :key="sub.key"
          class="pb-2 text-[13px] whitespace-nowrap transition-all cursor-pointer border-b-2 shrink-0"
          :class="sectorSubTab === sub.key ? 'text-[#e5b352] font-bold border-[#e5b352]' : 'text-[#8d99a4] font-medium border-transparent'"
          @tap="switchSubTab(sub.key)"
        >
          {{ sub.label }}
        </view>
      </view>
      <text class="text-[10px] font-mono text-[#64707c] pb-2 shrink-0 pl-1.5">{{ lastUpdatedTime.slice(5, 16) }}</text>
    </view>

    <!-- 子页 1：全球 24 条产业链景气（懒挂载 + v-show 保活） -->
    <view v-if="loadedSubTabs.global" v-show="sectorSubTab === 'global'" class="space-y-2">
      <view class="flex items-center justify-between px-0.5">
        <text class="text-sm font-bold text-[#f1efea]">全球产业链景气</text>
        <text class="text-[11px] font-mono text-[#64707c]">{{ sectorList.length }} 条主线</text>
      </view>
      <view class="grid grid-cols-2 gap-2">
        <SectorCard
          v-for="item in sectorList"
          :key="item.sector_id"
          :name="item.name"
          :stage="item.stage"
          :change-percent="item.price_change_pct"
          :price="item.price_index"
        />
      </view>
    </view>

    <!-- 子页 2：AI 算力链 -->
    <view v-if="loadedSubTabs.ai" v-show="sectorSubTab === 'ai'" class="space-y-3.5">
      <view class="space-y-2">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">AI 产品价格</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in aiProducts"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">AI 硬件与能耗设备</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in aiEquipments"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>
    </view>

    <!-- 子页 3：具身机器人 -->
    <view v-if="loadedSubTabs.robot" v-show="sectorSubTab === 'robot'" class="space-y-3.5">
      <view class="space-y-2">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">核心传动与执行部件</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in (robotDriveList || [])"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">感知、灵巧手与材料</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in (robotSenseList || [])"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>
    </view>

    <!-- 子页 4：芯片半导体自主链 -->
    <view v-if="loadedSubTabs.semi" v-show="sectorSubTab === 'semi'" class="space-y-3.5">
      <view class="space-y-2">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">芯片核心制造设备</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in (semiEquipList || [])"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>

      <view class="space-y-2 pt-1">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">先进材料、封装与EDA</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in (semiMaterialList || [])"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>
    </view>

    <!-- 子页 5：红利防守链 -->
    <view v-if="loadedSubTabs.dividend" v-show="sectorSubTab === 'dividend'" class="space-y-3.5">
      <view class="space-y-2">
        <text class="text-sm font-bold text-[#f1efea] px-0.5 block">高股息防御核心主线</text>
        <view class="grid grid-cols-2 gap-2">
          <SectorCard
            v-for="item in (dividendCoreList || [])"
            :key="item.name"
            :name="item.name"
            :change-percent="item.change_pct"
            :price="item.price"
          />
        </view>
      </view>
    </view>

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">本程序展示所有公开查询数据，仅供参考，不构成任何投资建议</text>
    </view>
  </view>
</template>
