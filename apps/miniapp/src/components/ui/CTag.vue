<script lang="ts" setup>
interface Props {
  variant?: 'up' | 'down' | 'neutral' | 'primary' | 'warning'
  size?: 'xs' | 'sm' | 'md'
}

withDefaults(defineProps<Props>(), {
  variant: 'neutral',
  size: 'sm'
})
</script>

<template>
  <view
    class="inline-flex items-center justify-center rounded-full font-semibold whitespace-nowrap shrink-0"
    :class="[
      size === 'xs' ? 'px-2 py-0.5 text-[10px]' : '',
      size === 'sm' ? 'px-2.5 py-0.5 text-[11px]' : '',
      size === 'md' ? 'px-3 py-1 text-xs' : '',
      variant === 'up' ? 'bg-[#f05a50]/18 text-[#ff736a]' : '',
      variant === 'down' ? 'bg-[#2ad1a3]/16 text-[#2ad1a3]' : '',
      variant === 'neutral' ? 'bg-white/[0.08] text-[#9aa5b1]' : '',
      variant === 'primary' ? 'bg-[#e5b352]/18 text-[#e5b352]' : '',
      variant === 'warning' ? 'bg-[#f5b731]/18 text-[#f5b731]' : ''
    ]"
  >
    <slot />
  </view>
</template>
