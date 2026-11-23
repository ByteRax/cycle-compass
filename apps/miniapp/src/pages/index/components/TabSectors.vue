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
  lastUpdatedTime: string
}>()

const sectorSubTabs = [
  { key: 'global', label: '全球 24 产业' },
  { key: 'ai', label: 'AI 算力链' }
] as const

type SectorSubTabKey = (typeof sectorSubTabs)[number]['key']

const sectorSubTab = ref<SectorSubTabKey>('global')
const loadedSubTabs = reactive<Record<SectorSubTabKey, boolean>>({
  global: true,
  ai: false
})

const switchSubTab = (key: SectorSubTabKey) => {
  sectorSubTab.value = key
  loadedSubTabs[key] = true
}
</script>

<template>
  <view class="space-y-3.5 tab-content-panel">
    <!-- 轻量级子分组切换条（使用原子类避免微信小程序 scoped 样式隔离失效） -->
    <view class="flex items-center justify-between border-b border-white/[0.08] pb-0.5">
      <view class="flex items-center space-x-5">
        <view
          v-for="sub in sectorSubTabs"
          :key="sub.key"
          class="pb-2 text-[13px] transition-all cursor-pointer border-b-2"
          :class="sectorSubTab === sub.key ? 'text-[#e5b352] font-bold border-[#e5b352]' : 'text-[#8d99a4] font-medium border-transparent'"
          @tap="switchSubTab(sub.key)"
        >
          {{ sub.label }}
        </view>
      </view>
      <text class="text-[10px] font-mono text-[#64707c] pb-2">{{ lastUpdatedTime.slice(5, 16) }}</text>
    </view>

    <!-- 子页：全球 24 条产业链景气（懒挂载 + v-show 保活） -->
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

    <!-- 子页：AI 算力链 -->
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

    <view class="text-center py-4">
      <text class="text-[11px] text-[#64707c]">本程序展示所有公开查询数据，仅供参考，不构成任何投资建议</text>
    </view>
  </view>
</template>
