import { defineUniPages } from '@uni-helper/vite-plugin-uni-pages'
import { env } from './getEnv'

const { VITE_APP_TITLE } = env
export default defineUniPages({
  globalStyle: {
    navigationStyle: 'default',
    navigationBarBackgroundColor: '@navBgColor',
    navigationBarTextStyle: '@navTxtStyle',
    navigationBarTitleText: `${VITE_APP_TITLE || '周期罗盘'}`,
    backgroundColor: '@bgColor',
    backgroundTextStyle: '@bgTxtStyle',
    backgroundColorTop: '@bgColorTop',
    backgroundColorBottom: '@bgColorBottom',
    enablePullDownRefresh: false,
    onReachBottomDistance: 50
  },
  easycom: {
    autoscan: true,
    custom: {
      '^C(.*)': '@/components/ui/C$1.vue'
    }
  }
})
