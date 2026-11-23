<script lang="ts" setup>
defineOptions({
  options: {
    virtualHost: true
  }
})

interface Props {
  title?: string
  subtitle?: string
  extra?: string
  bordered?: boolean
  padding?: string
}

withDefaults(defineProps<Props>(), {
  title: '',
  subtitle: '',
  extra: '',
  bordered: true,
  padding: 'p-4'
})
</script>

<template>
  <view
    class="rounded-2xl bg-[#12161b] text-[#f1efea] transition-colors"
    :class="[bordered ? 'border border-white/[0.09]' : '', padding]"
  >
    <view v-if="title || $slots['header'] || extra" class="mb-3 flex items-center justify-between gap-2">
      <view class="min-w-0">
        <view class="flex items-center space-x-2">
          <slot name="prefix" />
          <text class="text-sm font-bold tracking-tight text-[#f1efea]">{{ title }}</text>
        </view>
        <text v-if="subtitle" class="mt-0.5 block text-[11px] text-[#8d99a4]">{{ subtitle }}</text>
      </view>
      <slot name="extra">
        <text v-if="extra" class="shrink-0 text-xs font-medium text-[#8d99a4]">{{ extra }}</text>
      </slot>
    </view>
    <slot />
  </view>
</template>
