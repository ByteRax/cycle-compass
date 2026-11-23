<script lang="ts" setup>
interface Props {
  position?: 'top' | 'bottom'
}

withDefaults(defineProps<Props>(), {
  position: 'bottom'
})
</script>

<template>
  <view
    :class="[
      position === 'bottom'
        ? 'pb-[constant(safe-area-inset-bottom)] pb-[env(safe-area-inset-bottom)]'
        : 'pt-[constant(safe-area-inset-top)] pt-[env(safe-area-inset-top)]'
    ]"
  />
</template>
