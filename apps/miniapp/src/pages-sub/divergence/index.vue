<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import TrendMenuDropdown from '@/components/trend/TrendMenuDropdown.vue'
import TrendCurveView from '@/components/trend/TrendCurveView.vue'
import AnnualTrendView from '@/components/trend/AnnualTrendView.vue'
import ComponentTrendView from '@/components/trend/ComponentTrendView.vue'
import TrendDashboardView from '@/components/trend/TrendDashboardView.vue'
import SolarTermsView from '@/components/trend/SolarTermsView.vue'
import RightDistributionView from '@/components/trend/RightDistributionView.vue'
import {
  DIVERGENCE_DATASETS,
  getTodayDateStr,
  type DivergenceDataset
} from '@/data/divergence'

definePage({
  style: {
    navigationBarTitleText: '趋势分析看板',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    enablePullDownRefresh: false,
    navigationStyle: 'custom'
  }
})

// 当前资产数据集（默认 A股 或 红利ETF）
const currentDatasetId = ref<string>('a_shares')

// 动态获取真实最新日期
const latestDateStr = ref<string>(getTodayDateStr())

// 6大趋势视角状态 (0: 趋势曲线, 1: 年度价格走势, 2: 品种成分趋势, 3: 趋势仪表盘, 4: 趋势节气图, 5: 右侧分布图)
const currentViewIndex = ref<number>(2) // 默认 3: 品种成分趋势

// 顶部下拉菜单开关
const showDropdown = ref<boolean>(false)

const viewTitles = [
  '趋势曲线',
  '年度价格走势',
  '品种成分趋势',
  '趋势仪表盘',
  '趋势节气图',
  '右侧分布图'
]

const currentViewTitle = computed(() => {
  return viewTitles[currentViewIndex.value] || '品种成分趋势'
})

const dataset = computed<DivergenceDataset>(() => {
  return DIVERGENCE_DATASETS[currentDatasetId.value] || DIVERGENCE_DATASETS.a_shares
})

onLoad((query: any) => {
  if (query?.id && DIVERGENCE_DATASETS[query.id]) {
    currentDatasetId.value = query.id
  }
  if (query?.view !== undefined) {
    const v = parseInt(query.view, 10)
    if (!isNaN(v) && v >= 1 && v <= 6) {
      currentViewIndex.value = v - 1
    }
  }
})

// 滑动切换回调
const onSwiperChange = (e: any) => {
  currentViewIndex.value = e.detail.current
}

const handleSelectView = (idx: number) => {
  currentViewIndex.value = idx
}

const handleGoBack = () => {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

// 独立资产页面快速切换（红利ETF、ETF大类、A股短周期、A股27年领跑）
const assetOptions = [
  { id: 'a_shares', label: 'A股行业' },
  { id: 'dividend_etf', label: '红利ETF' },
  { id: 'etf_macro', label: 'ETF大类' },
  { id: 'a_shares_long', label: 'A股27年' }
]
</script>

<template>
  <view class="min-h-screen bg-[#090C0F] text-[#f1efea] px-3.5 py-2.5 pb-10 overflow-x-hidden max-w-full box-border">
    <!-- 顶部导航栏 (100% 像素级复刻 media_1795534784283_9de3c953.png 顶部菜单触发器) -->
    <view class="flex items-center justify-between py-2 border-b border-white/[0.08] mb-2.5 relative z-40">
      <!-- 左侧返回按钮与资产路径 -->
      <view class="flex items-center space-x-2" @tap="handleGoBack">
        <view class="w-7 h-7 rounded-full bg-white/[0.06] flex items-center justify-center cursor-pointer active:scale-95 transition-all">
          <text class="text-xs text-[#e5b352] font-bold">〈</text>
        </view>
        <view class="flex flex-col">
          <text class="text-[9px] font-mono text-[#8d99a4]">{{ dataset.category_path }}</text>
          <text class="text-xs font-bold text-[#f1efea] leading-tight">返回宏观仪表盘</text>
        </view>
      </view>

      <!-- 中间可点击下拉菜单触发器: [◎ 当前趋势视角名 ⌄] (100% 对齐原图) -->
      <view
        class="flex items-center space-x-1.5 px-3 py-1 rounded-full bg-white/[0.06] border border-white/[0.12] cursor-pointer active:scale-95 transition-all"
        @tap="showDropdown = !showDropdown"
      >
        <text class="text-xs text-[#e5b352] font-black">◎</text>
        <text class="text-xs font-bold text-[#f1efea]">{{ currentViewTitle }}</text>
        <text class="text-[10px] text-[#8d99a4] transform transition-transform" :class="showDropdown ? 'rotate-180' : ''">⌄</text>
      </view>

      <!-- 右侧日期高亮胶囊 (动态最新真实日期) -->
      <text class="text-[10px] font-mono text-[#e5b352] bg-[#e5b352]/10 px-2 py-0.5 rounded border border-[#e5b352]/20">
        {{ latestDateStr }}
      </text>
    </view>

    <!-- 下拉弹出层 (100% 复刻原图) -->
    <TrendMenuDropdown
      :visible="showDropdown"
      :active-index="currentViewIndex"
      @select="handleSelectView"
      @close="showDropdown = false"
    />

    <!-- 顶部资产与视角快捷状态胶囊条 -->
    <view class="flex items-center justify-between gap-1 mb-3">
      <!-- 独立资产切换器 -->
      <view class="flex items-center space-x-1 bg-[#12181F] p-1 rounded-lg border border-white/[0.06]">
        <text
          v-for="opt in assetOptions"
          :key="opt.id"
          class="px-2 py-0.5 text-center text-[10px] rounded transition-all cursor-pointer select-none"
          :class="currentDatasetId === opt.id ? 'bg-[#e5b352] text-[#090C0F] font-bold shadow-xs' : 'text-[#8d99a4]'"
          @tap="currentDatasetId = opt.id"
        >
          {{ opt.label }}
        </text>
      </view>

      <!-- 6 视角页码圆点指示器 (支持左右滑动) -->
      <view class="flex items-center space-x-1 bg-[#12181F] px-2 py-1.5 rounded-lg border border-white/[0.06]">
        <text class="text-[9px] text-[#8d99a4] mr-0.5">滑动切视角</text>
        <view
          v-for="(t, i) in viewTitles"
          :key="i"
          class="w-1.5 h-1.5 rounded-full transition-all cursor-pointer"
          :class="currentViewIndex === i ? 'w-3 bg-[#e5b352]' : 'bg-white/[0.18]'"
          @tap="currentViewIndex = i"
        />
      </view>
    </view>

    <!-- 核心滑动容器：左右滑动无缝切换 6 大趋势视角 -->
    <swiper
      class="w-full min-h-[780px]"
      :current="currentViewIndex"
      :duration="260"
      @change="onSwiperChange"
    >
      <!-- 视角 1: 趋势曲线 -->
      <swiper-item class="overflow-y-auto">
        <TrendCurveView
          :dataset-title="dataset.title"
          :as-of-date="latestDateStr"
        />
      </swiper-item>

      <!-- 视角 2: 年度价格走势 -->
      <swiper-item class="overflow-y-auto">
        <AnnualTrendView
          :dataset-title="dataset.title"
        />
      </swiper-item>

      <!-- 视角 3: 品种成分趋势 (已完美复刻的贝塞尔平滑分化曲线 + 截面多维排行表) -->
      <swiper-item class="overflow-y-auto">
        <ComponentTrendView
          :dataset="dataset"
          :as-of-date="latestDateStr"
        />
      </swiper-item>

      <!-- 视角 4: 趋势仪表盘 -->
      <swiper-item class="overflow-y-auto">
        <TrendDashboardView
          :dataset-title="dataset.title"
        />
      </swiper-item>

      <!-- 视角 5: 趋势节气图 -->
      <swiper-item class="overflow-y-auto">
        <SolarTermsView
          :dataset-title="dataset.title"
        />
      </swiper-item>

      <!-- 视角 6: 右侧分布图 (100% 像素级复刻 media_1795534905662_70f9934e.jpg Treemap 树图) -->
      <swiper-item class="overflow-y-auto">
        <RightDistributionView
          :dataset-id="currentDatasetId"
          :as-of-date="latestDateStr"
        />
      </swiper-item>
    </swiper>
  </view>
</template>

<style scoped>
/* 隐藏原生滚动条 */
::-webkit-scrollbar {
  display: none;
  width: 0 !important;
  height: 0 !important;
  -webkit-appearance: none;
  background: transparent;
}
</style>
