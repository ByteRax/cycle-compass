<script lang="ts" setup>
import { computed } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

interface Props {
  symbol?: string
  name: string
  price: number | string
  changePercent: number
  unit?: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'click'): void
}>()

const numericChangePercent = computed(() => {
  const val = typeof props.changePercent === 'number' ? props.changePercent : parseFloat(props.changePercent as any)
  return Number.isFinite(val) ? val : null
})

const isUp = computed(() => (numericChangePercent.value ?? 0) > 0)
const isDown = computed(() => (numericChangePercent.value ?? 0) < 0)

const displayPrice = computed(() => {
  const num = typeof props.price === 'number' ? props.price : parseFloat(props.price as string)
  if (Number.isFinite(num)) {
    const abs = Math.abs(num)
    return abs > 0 && abs < 0.1 ? num.toFixed(3) : num.toFixed(2)
  }
  return props.price || '--'
})

const formattedPercent = computed(() => {
  if (numericChangePercent.value === null) return '--'
  const sign = numericChangePercent.value > 0 ? '+' : ''
  return `${sign}${numericChangePercent.value.toFixed(2)}%`
})
</script>

<template>
  <view
    class="flex flex-col justify-between rounded-xl bg-[#12161b] px-3 py-2.5 border border-white/[0.08] active:bg-[#181e25] transition-colors"
    @click="emit('click')"
  >
    <view class="flex items-center justify-between gap-1">
      <text class="text-xs text-[#9aa5b1] font-medium truncate">{{ name }}</text>
      <text v-if="symbol" class="text-[10px] font-mono text-[#64707c] shrink-0">{{ symbol }}</text>
    </view>

    <view class="mt-2 flex items-baseline justify-between gap-1">
      <view class="flex items-baseline gap-1 min-w-0">
        <text class="text-base font-bold text-[#f1efea] tracking-tight font-mono cc-tabular">
          {{ displayPrice }}
        </text>
        <text v-if="unit" class="text-[10px] text-[#64707c] font-mono truncate">{{ unit }}</text>
      </view>

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
