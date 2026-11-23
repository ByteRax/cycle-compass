<script lang="ts" setup>
const { theme, isDark } = useTheme()
onLaunch(() => {
  // console.log('Root onLaunch')
})

onLoad(() => {
  // console.log('Root onLoad')
})

onShow(() => {
  // console.log('Root onShow')
})

onReady(() => {
  // console.log('Root onReady')
})
</script>

<template>
  <view :class="['page-wrapper', theme, isDark ? 'dark' : '']">
    <ku-root-view />
    <global-loading />
    <global-toast />
    <global-message />
    <!-- #ifdef MP-WEIXIN -->
    <privacy-popup />
    <!-- #endif -->
  </view>
</template>
