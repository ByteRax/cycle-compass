<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useFinanceStore } from '@/stores/finance'

export type FinanceNavTab = 'overview' | 'workbench' | 'favorites' | 'settings'

const props = withDefaults(defineProps<{
  title: string
  activeTab?: FinanceNavTab
  assetId?: string
  showBack?: boolean
  showBottomNav?: boolean
}>(), {
  activeTab: 'overview',
  assetId: '',
  showBack: true,
  showBottomNav: true
})

const emit = defineEmits<{
  back: []
}>()

const financeStore = useFinanceStore()
const isDark = computed(() => financeStore.isDarkEffective)

// 动态测量微信胶囊与安全区
const statusBarHeight = ref(24)
const navContentHeight = ref(44)
const capsuleRightPadding = ref(12)
const safeBottomPx = ref(0)

function measureSafeArea () {
  try {
    const sys = uni.getSystemInfoSync()
    statusBarHeight.value = sys.statusBarHeight || 24
    safeBottomPx.value = sys.safeAreaInsets?.bottom ?? 0

    // #ifdef MP-WEIXIN
    if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
      const rect = uni.getMenuButtonBoundingClientRect()
      if (rect && rect.top && rect.height) {
        const gap = Math.max(4, rect.top - statusBarHeight.value)
        navContentHeight.value = rect.height + gap * 2
        const winW = sys.windowWidth || 375
        capsuleRightPadding.value = Math.max(12, winW - rect.left + 8)
      }
    }
    // #endif
  }
  catch {
    statusBarHeight.value = 24
    navContentHeight.value = 44
  }
}

onMounted(() => {
  measureSafeArea()
})

function handleBack () {
  emit('back')
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  })
}

interface NavItem {
  key: FinanceNavTab
  label: string
  url: string
  iconOff: string
  iconOn: string
}

const navItems = computed<NavItem[]>(() => {
  const strokeOff = isDark.value ? '%238D99A4' : '%23626F77'
  const strokeOn = isDark.value ? '%23E5B352' : '%238A641F'
  const wbUrl = props.assetId
    ? `/pages-sub/finance/workbench?id=${encodeURIComponent(props.assetId)}`
    : '/pages-sub/finance/workbench'

  return [
    {
      key: 'overview',
      label: '全景',
      url: '/pages-sub/finance/index',
      iconOff: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOff}" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M3 10.5L12 3l9 7.5V20a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1v-9.5z"/></svg>`,
      iconOn: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOn}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 10.5L12 3l9 7.5V20a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1v-9.5z"/></svg>`
    },
    {
      key: 'workbench',
      label: '趋势图',
      url: wbUrl,
      iconOff: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOff}" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="M7 15l4-4 3 3 6-6"/></svg>`,
      iconOn: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOn}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="M7 15l4-4 3 3 6-6"/></svg>`
    },
    {
      key: 'favorites',
      label: '收藏',
      url: '/pages-sub/finance/favorites',
      iconOff: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOff}" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>`,
      iconOn: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOn}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>`
    },
    {
      key: 'settings',
      label: '设置',
      url: '/pages-sub/finance/settings',
      iconOff: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOff}" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="7" x2="20" y2="7"/><circle cx="9" cy="7" r="2"/><line x1="4" y1="12" x2="20" y2="12"/><circle cx="15" cy="12" r="2"/><line x1="4" y1="17" x2="20" y2="17"/><circle cx="10" cy="17" r="2"/></svg>`,
      iconOn: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${strokeOn}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="7" x2="20" y2="7"/><circle cx="9" cy="7" r="2"/><line x1="4" y1="12" x2="20" y2="12"/><circle cx="15" cy="12" r="2"/><line x1="4" y1="17" x2="20" y2="17"/><circle cx="10" cy="17" r="2"/></svg>`
    }
  ]
})

function switchNav (item: NavItem) {
  if (item.key === props.activeTab)
    return
  uni.redirectTo({
    url: item.url,
    fail: () => {
      uni.navigateTo({ url: item.url })
    }
  })
}
</script>

<template>
  <view
    class="min-h-screen flex flex-col"
    :class="isDark ? 'bg-[#090C0F] text-[#F1EFEA]' : 'cc-finance-light bg-[#F4F3EF] text-[#22292E]'"
  >
    <!-- 顶部安全区 + 胶囊适配导航栏 -->
    <view
      class="sticky top-0 z-30 border-b"
      :class="isDark ? 'bg-[#090C0F]/95 border-[#29313A]' : 'bg-[#F4F3EF]/95 border-[#DCDDD8]'"
      :style="{ paddingTop: `${statusBarHeight}px` }"
    >
      <view
        class="flex items-center justify-between px-2"
        :style="{ height: `${navContentHeight}px`, paddingRight: `${capsuleRightPadding}px` }"
      >
        <view class="flex items-center min-w-0 flex-1">
          <view
            v-if="showBack"
            class="min-w-[44px] min-h-[44px] flex items-center justify-center shrink-0 active:opacity-70"
            @tap="handleBack"
          >
            <text
              class="text-[20px] font-medium leading-none"
              :class="isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]'"
            >
              ‹
            </text>
          </view>
          <text
            class="text-[17px] font-bold truncate"
            :class="[isDark ? 'text-[#F1EFEA]' : 'text-[#22292E]', showBack ? '' : 'pl-2']"
          >
            {{ title }}
          </text>
        </view>
        <view class="flex items-center shrink-0">
          <slot name="header-right" />
        </view>
      </view>
      <slot name="sub-header" />
    </view>

    <!-- 页面主体内容 -->
    <view
      class="flex-1"
      :style="{ paddingBottom: showBottomNav ? `${64 + safeBottomPx}px` : `${16 + safeBottomPx}px` }"
    >
      <slot />
    </view>

    <!-- 底部 4 项金融子导航栏 -->
    <view
      v-if="showBottomNav"
      class="fixed bottom-0 left-0 right-0 z-30 border-t"
      :class="isDark ? 'bg-[#090C0F]/95 border-[#29313A]' : 'bg-[#FFFFFF]/95 border-[#DCDDD8]'"
      :style="{ paddingBottom: `${safeBottomPx}px` }"
    >
      <view class="grid grid-cols-4 h-[56px]">
        <view
          v-for="item in navItems"
          :key="item.key"
          class="flex flex-col items-center justify-center min-h-[44px] active:opacity-75"
          @tap="switchNav(item)"
        >
          <image
            :src="activeTab === item.key ? item.iconOn : item.iconOff"
            class="w-[20px] h-[20px]"
            mode="aspectFit"
          />
          <text
            class="text-[11px] mt-0.5 leading-none"
            :class="activeTab === item.key
              ? (isDark ? 'text-[#E5B352] font-bold' : 'text-[#8A641F] font-bold')
              : (isDark ? 'text-[#8D99A4]' : 'text-[#626F77]')"
          >
            {{ item.label }}
          </text>
        </view>
      </view>
    </view>
  </view>
</template>

<style>
.cc-finance-light {
  --cc-brand-primary: #8a641f;
  --cc-brand-secondary: #14745e;
  --cc-up: #b93842;
  --cc-up-subtle: rgba(185, 56, 66, 0.12);
  --cc-down: #14745e;
  --cc-down-subtle: rgba(20, 116, 94, 0.12);
  --cc-neutral: #64707c;
  --cc-bg-page: #f4f3ef;
  --cc-bg-card: #ffffff;
  --cc-bg-elevated: #edece6;
  --cc-text-primary: #22292e;
  --cc-text-secondary: #515f6a;
  --cc-text-muted: #626f77;
  --cc-border: #dcddd8;
  --cc-border-strong: #c5c6c0;
}
</style>
