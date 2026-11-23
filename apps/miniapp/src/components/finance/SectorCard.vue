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
</script>

<template>
  <view
    class="flex items-center justify-between rounded-xl bg-[#12161b] px-3 py-2.5 border border-white/[0.08] active:bg-[#181e25] transition-colors"
    @click="emit('click')"
  >
    <view class="flex items-center space-x-2 min-w-0 pr-1">
      <!-- 语义色竖线替代 emoji 图标 -->
      <view
        class="w-0.5 h-4 rounded-full shrink-0"
        :class="[isUp ? 'bg-[#f05a50]' : isDown ? 'bg-[#2ad1a3]' : 'bg-[#64707c]']"
      />
      <view class="min-w-0">
        <text class="text-[13px] font-semibold text-[#f1efea] block truncate">{{ name }}</text>
        <text v-if="stage" class="text-[10px] text-[#64707c] block truncate">{{ stage }}</text>
      </view>
    </view>

    <view class="inline-flex items-center space-x-1.5 shrink-0">
      <text
        v-if="displayPrice"
        class="font-mono text-xs font-semibold text-[#f1efea] tracking-tight cc-tabular"
      >
        {{ displayPrice }}
      </text>
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
