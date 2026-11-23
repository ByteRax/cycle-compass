import { createSSRApp } from 'vue'
import { setupHttp } from '@/http'
import { setupRoute } from '@/router'
import { setupStore } from '@/stores'
import App from './App.vue'
import './app.css'
import '@/styles/common.scss'
import '@/styles/variable.scss'

export function createApp() {
  const app = createSSRApp(App)
  // 挂载状态管理
  setupStore(app)
  // 路由拦截器
  setupRoute(app)
  // 网络请求
  setupHttp(app)
  return {
    app
  }
}
