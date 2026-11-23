# UniApp 模板优化执行记录

日期：2026-10-25。参考项目全程只读，未提交、推送或上传。

## 范围与实现

| 优先级 | 证据与问题                                                                           | 本轮处理                                                                           | 验证入口                                                           |
| ------ | ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| 第一批 | 导航 options 覆盖内部 complete，锁可残留；旧 complete 还可能释放后续导航的锁         | 回调组合、幂等解锁、异常释放；守卫在加锁前预检                                     | `tests/runtime/router.test.ts`、`navigation-integration.test.ts`   |
| 第一批 | `getAllPages()` 已带前导斜杠，存在性检查又加一次；登录页引用缺失的自动导入符号       | 修正路径比较；登录页改用真实 Store、解码函数与首页兜底                             | `tests/runtime/route-guard.test.ts`，H5 登录回跳                   |
| 第一批 | 同步重试结果被后续 loading 覆盖                                                      | 保存回调，先切 loading 再重试；不增加声明式布局 API                                | `tests/components/base-layout.test.ts`                             |
| 第一批 | Token 属性不检查到期，资料 Store 调用方法不存在，旧 logout 可清理新会话              | getter ref 每次读取检查时钟；对齐现有资料 action；按会话对象隔离清理               | `tests/runtime/session.test.ts`                                    |
| 第一批 | 旧请求完成误删同键新请求，取消缺少回调时 Promise 挂起，拦截器异常无法正常结算        | 请求入口统一收尾、显式取消结算、清理钩子、Map 身份检查，支持拦截器返回替代 options | `tests/runtime/http.test.ts`                                       |
| 第一批 | 并发／迟到 401 与远端登出相互影响                                                    | 401 只清理对应本地会话，一次跳登录；主动退出仍调用远端                             | `tests/runtime/http.test.ts`                                       |
| 第二批 | Toast Store 依赖组件动画清理；旧关闭回调可能影响新提示                               | Store 管理单个计时器，支持替换、关闭和销毁；UI 使用 duration 0；回调校验对象身份   | `tests/runtime/toast.test.ts`、`tests/components/feedback.test.ts` |
| 第二批 | Loading／Message 需要明确卸载及完成行为                                              | 保留页面归属隔离，卸载清理，Message 结果清 Store；外部关闭保持 Wot 语义            | `tests/components/feedback.test.ts`                                |
| 第三批 | 无行为回归入口                                                                       | 接入独立 Vitest 配置与 Vue SFC 编译，新增 `test`／`test:run`                       | `vitest.config.ts`、`tests/`                                       |
| 第三批 | ESLint preset 键不存在，配置组合器被当作普通数组元素                                 | 使用真实 UnoCSS flat 配置，并由 uni 配置组合器接入                                 | `pnpm lint`                                                        |
| 第三批 | Vite 整文件跳过类型检查，包含当前插件不支持的 dts 参数和无效 build.server            | 移除全局跳过与无效配置，使用 bundle optimizer 的公开 client 类型；保持严格选项     | `pnpm type-check` 中项目源码诊断                                   |
| 第三批 | mode 解析不支持等号形式或缺失值校验，页面元数据使用 any                              | 类型化解析与页面处理，分包显式名称保留；不新增预生成文件或接入闲置生成器           | `tests/tooling/config.test.ts`                                     |
| 第三批 | Node 原生加载 UnoCSS 配置时扩展名缺失；平台模块有未导出消费者                        | 补 `.ts` 扩展名，导出已被实际导入的 `isH5`                                         | H5 开发启动与构建、类型检查                                        |
| 第三批 | 手工分块后的 H5 发布页报循环初始化错误，开发环境无法暴露                             | 移除宽泛的手工 vendor 分块，保留 Rollup 默认分块与现有原生端优化插件               | H5 发布产物内置浏览器渲染、导航与登录回跳                          |
| 第三批 | UnoCSS 扫描 Wot 注释时间戳生成非法 WXSS；全局 `*::before`／`*::after` 也被 WXSS 拒绝 | 排除第三方内容扫描，保留依赖自身样式；伪元素改为等价的 `::before`／`::after`       | 微信官方 `wcsc` 校验与模拟器截图                                   |
| 第三批 | manifest 初始化阶段缺少 CLI 状态，H5 HTML 插件又缓存旧生成文件的名称                 | 统一构建类型推断与 watch 语义，HTML post hook 从 manifest 配置源取标题             | 配置回归、H5 HTML 的 release 标题、App manifest                    |

## 保留与暂缓

- Token API 当前只有 `token`／`expiresIn`，没有 refresh token 契约；不新增刷新接口或提前一小时刷新策略。
- 保留同方法、URL、data 请求取消前一个的策略，包含写请求。服务端写入是否撤销不由客户端取消保证；本轮不改变该既有交互契约。
- 保留现有主题与 `themeVars`。`useTheme` 当前已经正确复用监听回调，旧报告的泄漏判断不适用。
- `App.ku.vue` 由 uni-ku 插件注入每个页面，保留反馈实例页面快照；不按常驻单根假设整体重写。
- Wot 1 的 Toast `closed` 是函数参数；Message 外部 `close()` 不产生用户确认或取消结果，保持这一库契约。
- 未发现必须增加平台页面预生成步骤的复现证据，未增加额外运行时页面元数据或路由常量文件。
- Uni CLI 的构建类型与 `VITE_USER_NODE_ENV` 设置链已核对，保留现有发布版本策略。验证构建会触发 AutoVersion 写盘，已通过快照恢复仅本轮产生的版本与 manifest 改动。
- 保留框架及现有工具链版本；Vitest 和 Vue 测试编译插件是本轮新增的开发依赖。安装时 lockfile 清除了 manifest 已不存在的 `vue-i18n` importer 项，未作为依赖清理另行删包。
- 不仅凭直接引用缺失删除上传、Sass、开发工具等依赖；插件间接使用与现有类型声明仍需考虑。未搬入金融、埋点、AI、设备协议或业务组件。
- 当前两个演示页都有 `type: 'home'`，本轮保留最终生成结果，不替产品决定默认首页；README 已明确普通演示入口。

## 验证状态

| 检查                    | 基线                              | 本轮结果                                                                                                                 |
| ----------------------- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Vitest                  | 无入口                            | 50 项通过，覆盖真实模块与 SFC 状态／生命周期                                                                             |
| `pnpm lint`             | ESLint 加载失败，无法检查 Vue     | 已恢复并通过；保留 Oxlint 与 Vue ESLint 路线                                                                             |
| `pnpm type-check`       | 544 条诊断，其中 6 条属于项目代码 | 项目代码诊断已消除；仍有 538 条 Wot 源文件内部诊断，全局检查失败，未降低严格选项                                         |
| 修改文件的 Oxfmt 检查   | 包含既有格式问题                  | 已通过定向检查                                                                                                           |
| `pnpm format:check`     | 14 个文件失败                     | 剩余 8 个既有文件失败，未格式化无关文件                                                                                  |
| `pnpm build:h5`         | 本轮前无构建基线                  | 最终构建成功；HTML 初始标题为 `uniapp-template-cli-release`                                                              |
| `pnpm build:mp-weixin`  | 本轮前无构建基线                  | 最终构建成功；非法时间戳规则及 CSS 语法警告已消失，微信官方 `wcsc -lc app.wxss` 通过                                     |
| `pnpm build:custom app` | 无专用 `build:app` 脚本           | 使用现有 custom 脚本，App 构建成功；未做 HBuilderX／真机运行                                                             |
| H5 交互                 | 未验收                            | 内置浏览器验证开发版和发布版普通导航、未登录守卫、模拟登录回跳及 Toast；最终发布页未捕获运行错误，深层时序由行为测试覆盖 |
| 微信模拟器              | 未验收                            | 修复后读取到 `pages/index/index`，截图显示首页正常；嵌套组件选择器未命中，未声称完成微信端点击验收                       |

类型与全仓格式检查保留原始失败状态。原始基线日志位于 `/tmp/uniapp-baseline-typecheck.log`、`/tmp/uniapp-baseline-lint.log`、`/tmp/uniapp-baseline-format.log`，后续日志位于同目录的 `uniapp-*.log`；这些是本机临时证据，不是仓库持续集成产物。

剩余格式失败文件：`.oxlintrc.json`、`index.html`、`src/components/business/PrivacyPopup.vue`、`src/composables/useTheme.ts`、`src/pages-sub/styles/index.vue`、`src/pages/index/paging.vue`、`src/uni.scss`、`vite-plugins/vite-plugin-auto-version.ts`。

独立审查后的导航递归与请求替代 options 假设已补集成回归核对；同时补齐替代 options 的取消结算。未把审查中的条件性推测直接当作已确认缺陷。最后一轮构建修复也完成只读独立核验。

## 交付限制

本地模拟登录、测试替身和构建不能证明真实后端登录／微信授权已验收；本轮未运行真实账号登录、远端登出、上传或真机测试。依赖严格类型诊断、未改文件的格式问题与框架 peer 版本警告仍需后续独立处理。

微信端原生界面点击降级尝试受 macOS 锁屏限制，未绕过锁屏；本轮微信运行证据来自开发者工具的页面信息与模拟器截图。

运行截图：[H5 发布版 Toast](/Users/hmey/.codex/visualizations/2026/10/25/01a17147-2d1c-7d21-b943-04f4cd07580b/uniapp-template-release-toast.png) 、[微信模拟器首页](/Users/hmey/.codex/visualizations/2026/10/25/01a17147-2d1c-7d21-b943-04f4cd07580b/uniapp-template-weixin.jpg) 。
