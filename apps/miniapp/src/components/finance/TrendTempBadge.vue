<script lang="ts" setup>
import { computed } from 'vue'
import { useFinanceStore } from '@/stores/finance'
import { getTrendTempMeta } from '@/utils/trendDict'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = withDefaults(defineProps<{
  temperature?: string | null
  size?: 'sm' | 'md'
  showDesc?: boolean
}>(), {
  temperature: 'unavailable',
  size: 'sm',
  showDesc: false
})

const financeStore = useFinanceStore()
const meta = computed(() => getTrendTempMeta(props.temperature))
const isDark = computed(() => financeStore.isDarkEffective)

const textColor = computed(() => (isDark.value ? meta.value.color : meta.value.lightColor))
const bgColor = computed(() => {
  if (!isDark.value) {
    if (meta.value.key === 'unavailable')
      return 'rgba(98, 111, 119, 0.10)'
    return meta.value.bg
  }
  return meta.value.bg
})
</script>

<template>
  <view
    class="inline-flex items-center justify-center rounded border shrink-0"
    :class="size === 'md' ? 'px-2 py-0.5 gap-1' : 'min-w-[26px] h-[24px] px-1'"
    :style="{
      backgroundColor: bgColor,
      borderColor: meta.border
    }"
  >
    <text
      class="font-bold leading-none"
      :class="size === 'md' ? 'text-[13px]' : 'text-[12px]'"
      :style="{ color: textColor }"
    >
      {{ meta.label }}
    </text>
    <text
      v-if="showDesc"
      class="text-[11px] font-medium leading-none"
      :style="{ color: textColor }"
    >
      · {{ meta.shortDesc }}
    </text>
  </view>
</template>
