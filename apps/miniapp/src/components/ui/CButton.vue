<script lang="ts" setup>
interface Props {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  loading?: boolean
  block?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  disabled: false,
  loading: false,
  block: false
})

const emit = defineEmits<{
  (e: 'click', event: any): void
}>()

function handleClick(e: any) {
  if (props.disabled || props.loading) return
  emit('click', e)
}
</script>

<template>
  <button
    class="relative inline-flex items-center justify-center font-bold transition-all select-none rounded-xl border-0"
    :class="[
      block ? 'w-full' : 'inline-flex',
      size === 'sm' ? 'px-3 py-1.5 text-xs' : '',
      size === 'md' ? 'px-4 py-2.5 text-sm' : '',
      size === 'lg' ? 'px-6 py-3 text-base' : '',
      variant === 'primary' ? 'bg-[#e5b352] text-[#090c0f] active:opacity-85' : '',
      variant === 'secondary' ? 'bg-[#181e25] text-[#f1efea] border border-white/[0.14] active:bg-[#222a33]' : '',
      variant === 'danger' ? 'bg-[#f05a50] text-white active:opacity-85' : '',
      variant === 'ghost' ? 'bg-transparent text-[#9aa5b1] active:bg-white/[0.06]' : '',
      disabled || loading ? 'opacity-50 pointer-events-none' : ''
    ]"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <view
      v-if="loading"
      class="mr-2 h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent"
    />
    <slot />
  </button>
</template>
