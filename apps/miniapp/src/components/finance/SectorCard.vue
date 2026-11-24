<script lang="ts" setup>
import { computed } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

interface Props {
  icon?: string
  name: string
  changePercent: number
  price?: number | string
  category?: string
  typeTag?: string
  stage?: string
  prosperityScore?: number
}

const props = withDefaults(defineProps<Props>(), {
  icon: '',
  changePercent: 0,
  price: undefined,
  category: '',
  typeTag: '',
  stage: '',
  prosperityScore: undefined
})

const emit = defineEmits<{
  (e: 'click'): void
}>()

const isUp = computed(() => props.changePercent > 0)
const isDown = computed(() => props.changePercent < 0)

const displayPrice = computed(() => {
  if (props.price === undefined || props.price === null || props.price === '') {
    return ''
  }
  if (typeof props.price === 'string' && isNaN(Number(props.price))) {
    return props.price
  }
  const num = typeof props.price === 'number' ? props.price : parseFloat(props.price as string)
  if (isNaN(num)) return ''
  if (num > 0 && num < 0.1) {
    return num.toFixed(3)
  }
  return num.toFixed(2)
})

const formattedPercent = computed(() => {
  const sign = props.changePercent > 0 ? '+' : ''
  return `${sign}${props.changePercent.toFixed(2)}%`
})
const handleClick = () => {
  emit('click')
  let targetId = 'a_shares'
  const n = props.name || ''
  if (n.includes('红利') || n.includes('煤炭') || n.includes('水电')) {
    targetId = 'dividend_etf'
  } else if (n.includes('ETF') || n.includes('300') || n.includes('500') || n.includes('恒生') || n.includes('标普') || n.includes('纳斯达克')) {
    targetId = 'etf_macro'
  } else {
    targetId = 'a_shares'
  }
  uni.navigateTo({
    url: `/pages-sub/divergence/index?id=${targetId}`
  })
}
</script>

<template>
  <view
    class="flex flex-col justify-between rounded-xl bg-[#12161b] px-3 py-2.5 border border-white/[0.08] active:bg-[#181e25] transition-colors gap-1.5 cursor-pointer"
    @click="handleClick"
  >
    <!-- 第一行：行业名称 + 周期阶段 -->
    <view class="flex items-center justify-between min-w-0">
      <view class="flex items-center space-x-1.5 min-w-0 pr-1">
        <!-- 语义色指示条 -->
        <view
          class="w-1 h-3 rounded-full shrink-0"
          :class="[isUp ? 'bg-[#f05a50]' : isDown ? 'bg-[#2ad1a3]' : 'bg-[#64707c]']"
        />
        <text class="text-[13px] font-semibold text-[#f1efea] truncate">{{ name }}</text>
      </view>
      <text
        v-if="stage"
        class="text-[10px] text-[#8d99a4] bg-white/[0.06] px-1.5 py-0.5 rounded shrink-0 font-medium"
      >
        {{ stage }}
      </text>
    </view>

    <!-- 第二行：景气点位 + 涨跌幅 -->
    <view class="flex items-baseline justify-between min-w-0 pt-0.5">
      <text
        v-if="displayPrice"
        class="font-mono text-sm font-bold text-[#f1efea] tracking-tight cc-tabular"
      >
        {{ displayPrice }}
      </text>
      <view v-else />
      <text
        class="inline-flex items-center rounded-full px-1.5 py-0.5 text-[11px] font-bold font-mono cc-tabular shrink-0"
        :class="[
          isUp ? 'bg-[#f05a50]/16 text-[#ff736a]' : isDown ? 'bg-[#2ad1a3]/14 text-[#2ad1a3]' : 'bg-white/[0.06] text-[#8d99a4]'
        ]"
      >
        {{ formattedPercent }}
      </text>
    </view>
  </view>
</template>
