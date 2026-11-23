import type { ToastOptions } from '@/composables/useGlobalToast'
import type { LoadingOptions } from '@/composables/useGlobalLoading'
import { isShowLoadingOptions, isToastOptions } from '@/utils/is.ts'
import { getValue } from '@/utils/object.ts'

export type ToastIcon = 'success' | 'loading' | 'error' | 'none' | 'fail' | 'exception'
export type ToastOption = UniApp.ShowToastOptions | ToastOptions | string
export type ShowLoadingOption = UniApp.ShowLoadingOptions | LoadingOptions | string

export default {
  /**
   * 显示消息提示框
   * @param option 提示内容或配置项
   */
  showToast(option: ToastOption, icon?: ToastIcon) {
    if (isToastOptions(option)) {
      uni.showToast(option as UniApp.ShowToastOptions).then()
    } else if (typeof option === 'string') {
      uni.showToast({ title: option, icon: (icon as any) || 'none' }).then()
    } else {
      useGlobalToast().show(option as ToastOptions)
    }
  },

  /**
   * 显示加载提示框
   * @param option 加载提示内容或配置项
   */
  showLoading(option: ShowLoadingOption) {
    if (isShowLoadingOptions(option)) {
      uni.showLoading(option as UniApp.ShowLoadingOptions).then()
    } else if (typeof option === 'string') {
      uni.showLoading({ title: option }).then()
    } else {
      useGlobalLoading().loading(option as LoadingOptions)
    }
  },

  hideLoading(isSystem: boolean = true) {
    if (isSystem) {
      if (typeof uni !== 'undefined' && typeof uni.hideLoading === 'function') {
        uni.hideLoading()
      }
    } else {
      useGlobalLoading().close()
    }
  },

  showAlert(title: string, msg: string) {
    useGlobalMessage().alert({
      title,
      msg
    })
  },

  /**
   * 兼容微信小程序端获取系统信息的方法
   */
  getSystemInfo(): UniApp.GetSystemInfoResult {
    let systemInfo: any = {
      platform: 'devtools',
      osName: 'devtools',
      browserName: '',
      browserVersion: '',
      osVersion: '',
      ua: '',
      uniCompileVersion: '',
      uniPlatform: 'mp-weixin',
      uniRuntimeVersion: ''
    }
    try {
      if (typeof uni !== 'undefined') {
        if (typeof uni.getDeviceInfo === 'function') {
          const deviceInfo = uni.getDeviceInfo()
          const windowInfo = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : {}
          const appBaseInfo = typeof uni.getAppBaseInfo === 'function' ? uni.getAppBaseInfo() : {}
          systemInfo = {
            ...systemInfo,
            ...deviceInfo,
            ...windowInfo,
            ...appBaseInfo,
            osName: deviceInfo?.platform || 'devtools',
            osVersion: getValue(deviceInfo, 'osVersion', '')
          }
        } else if (typeof uni.getSystemInfoSync === 'function') {
          systemInfo = uni.getSystemInfoSync()
        }
      }
    } catch (error) {
      console.warn('获取系统信息失败:', error)
    }
    return systemInfo
  }
}
