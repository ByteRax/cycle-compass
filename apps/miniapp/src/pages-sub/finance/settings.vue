<script lang="ts" setup>
import type { ColorSchemePref, DensityPref, FinanceThemePref } from '@/stores/finance'
import { computed } from 'vue'
import CSafeArea from '@/components/ui/CSafeArea.vue'
import { useFinanceStore } from '@/stores/finance'

definePage({
  style: {
    navigationBarTitleText: '金融研究设置',
    navigationBarBackgroundColor: '#090C0F',
    navigationBarTextStyle: 'white',
    backgroundColor: '#090C0F',
    navigationStyle: 'custom'
  }
})

const financeStore = useFinanceStore()

const themeOptions: Array<{ value: FinanceThemePref, label: string }> = [
  { value: 'dark', label: '墨曜石深色' },
  { value: 'light', label: '暖纸浅色' },
  { value: 'system', label: '跟随系统' }
]

const colorSchemeOptions: Array<{ value: ColorSchemePref, label: string }> = [
  { value: 'cn_red_up', label: '红涨绿跌' },
  { value: 'intl_green_up', label: '绿涨红跌' }
]

const densityOptions: Array<{ value: DensityPref, label: string }> = [
  { value: 'compact', label: '紧凑' },
  { value: 'standard', label: '标准' }
]

const currencyOptions: Array<{ value: 'CNY' | 'USD' | 'HKD' | 'RAW', label: string }> = [
  { value: 'CNY', label: 'CNY' },
  { value: 'USD', label: 'USD' },
  { value: 'HKD', label: 'HKD' },
  { value: 'RAW', label: '原币种' }
]

const rangeOptions: Array<{ value: '3M' | '6M' | '1Y' | '3Y', label: string }> = [
  { value: '3M', label: '近3月' },
  { value: '6M', label: '近6月' },
  { value: '1Y', label: '近1年' },
  { value: '3Y', label: '近3年' }
]

const tierModeOptions: Array<{ value: 'pro' | 'fun', label: string }> = [
  { value: 'pro', label: '专业 (S~D)' },
  { value: 'fun', label: '趣味 (夯~拉)' }
]

const reportThemeOptions: Array<{ value: 'light' | 'dark', label: string }> = [
  { value: 'light', label: '暖纸浅色' },
  { value: 'dark', label: '墨曜石深色' }
]

const currentThemeLabel = computed(() => themeOptions.find(o => o.value === financeStore.themePref)?.label || '墨曜石深色')
const currentColorLabel = computed(() => colorSchemeOptions.find(o => o.value === financeStore.colorScheme)?.label || '红涨绿跌')
const currentDensityLabel = computed(() => densityOptions.find(o => o.value === financeStore.density)?.label || '紧凑')
const currentRangeLabel = computed(() => rangeOptions.find(o => o.value === financeStore.defaultRange)?.label || '近1年')
const currentTierModeLabel = computed(() => tierModeOptions.find(o => o.value === financeStore.tierLabelMode)?.label || '专业 (S~D)')
const currentReportThemeLabel = computed(() => reportThemeOptions.find(o => o.value === financeStore.reportThemePref)?.label || '暖纸浅色')

function handleBack() {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

function pickOption<T extends string>(
  options: Array<{ value: T, label: string }>,
  onSelect: (val: T) => void
) {
  uni.showActionSheet({
    itemList: options.map(o => o.label),
    success: (res) => {
      const chosen = options[res.tapIndex]
      if (chosen) {
        onSelect(chosen.value)
        uni.showToast({ title: `已切换：${chosen.label}`, icon: 'none' })
      }
    }
  })
}

function navigateTo(url: string) {
  uni.navigateTo({ url })
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
        <text class="text-sm font-semibold cc-text-primary">金融研究设置</text>
      </view>
    </view>

    <view class="px-3.5 pt-3.5 pb-6 flex-1">
      <!-- 标题 -->
      <view class="flex items-center justify-between mb-4">
        <text class="text-[22px] font-bold tracking-tight cc-text-primary">研究偏好</text>
        <text class="text-xs cc-text-muted">本地设置</text>
      </view>

      <!-- 界面与显示 -->
      <text class="text-xs font-semibold cc-text-secondary block mb-2">界面与显示</text>
      <view class="rounded-xl cc-panel overflow-hidden mb-5">
        <view
          class="px-3.5 py-3.5 flex items-center justify-between border-b cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="pickOption(themeOptions, v => { financeStore.themePref = v })"
        >
          <text class="text-sm cc-text-primary">主题模式</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">{{ currentThemeLabel }}</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>

        <view
          class="px-3.5 py-3.5 flex items-center justify-between border-b cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="pickOption(colorSchemeOptions, v => { financeStore.colorScheme = v })"
        >
          <text class="text-sm cc-text-primary">涨跌配色</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">{{ currentColorLabel }}</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>

        <view
          class="px-3.5 py-3.5 flex items-center justify-between border-b cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="pickOption(densityOptions, v => { financeStore.density = v })"
        >
          <text class="text-sm cc-text-primary">列表密度</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">{{ currentDensityLabel }}</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>

        <view
          class="px-3.5 py-3.5 flex items-center justify-between border-b cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="pickOption(currencyOptions, v => { financeStore.defaultCurrency = v })"
        >
          <text class="text-sm cc-text-primary">默认计价币种</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-tabular cc-text-secondary">{{ financeStore.defaultCurrency }}</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>

        <view
          class="px-3.5 py-3.5 flex items-center justify-between cursor-pointer"
          @tap="pickOption(rangeOptions, v => { financeStore.defaultRange = v })"
        >
          <text class="text-sm cc-text-primary">默认研究区间</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">{{ currentRangeLabel }}</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>
      </view>

      <!-- 研究配置 -->
      <text class="text-xs font-semibold cc-text-secondary block mb-2">研究配置</text>
      <view class="rounded-xl cc-panel overflow-hidden mb-5">
        <view
          class="px-3.5 py-3.5 flex items-center justify-between border-b cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="pickOption(tierModeOptions, v => { financeStore.tierLabelMode = v })"
        >
          <text class="text-sm cc-text-primary">梯队标签模式</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">{{ currentTierModeLabel }}</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>

        <view
          class="px-3.5 py-3.5 flex items-center justify-between border-b cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="pickOption(reportThemeOptions, v => { financeStore.reportThemePref = v })"
        >
          <text class="text-sm cc-text-primary">图表报告主题</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">{{ currentReportThemeLabel }}</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>

        <view
          class="px-3.5 py-3.5 flex items-center justify-between cursor-pointer"
          @tap="navigateTo('/pages-sub/finance/favorites')"
        >
          <text class="text-sm cc-text-primary">我的收藏与默认主页</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">{{ financeStore.favoriteAssetIds.length }} 项</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>
      </view>

      <!-- 数据与方法 -->
      <text class="text-xs font-semibold cc-text-secondary block mb-2">数据与方法</text>
      <view class="rounded-xl cc-panel overflow-hidden mb-5">
        <view
          class="px-3.5 py-3.5 flex items-center justify-between border-b cursor-pointer"
          style="border-color: var(--cc-border);"
          @tap="navigateTo('/pages-sub/finance/methodology')"
        >
          <text class="text-sm cc-text-primary">指标说明</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">口径与版本</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>

        <view
          class="px-3.5 py-3.5 flex items-center justify-between cursor-pointer"
          @tap="navigateTo('/pages-sub/finance/data-status')"
        >
          <text class="text-sm cc-text-primary">数据来源与覆盖</text>
          <view class="flex items-center gap-1.5">
            <text class="text-xs cc-text-secondary">查看状态</text>
            <text class="text-xs cc-text-muted">›</text>
          </view>
        </view>
      </view>

      <!-- 底部说明 -->
      <view
        class="p-3 rounded-xl border text-xs cc-text-secondary leading-relaxed"
        style="background-color: var(--cc-bg-elevated); border-color: var(--cc-border);"
      >
        金融模块主题不强制改变既有策略、产业及账本页面。
      </view>
    </view>

    <CSafeArea position="bottom" />
  </view>
</template>
