# 依赖更新记录

更新日期：2026-10-25。

## 已更新

- `@dcloudio/*` 运行时与构建工具成组升到 Vue 3 `3.0.0-5030320261023001` 预发布线；该线 peer 要求 Vite `5.2.8`，因此项目 Vite 从 `5.4.21` 调整为 `5.2.8`，并保留 DCloud 固定的 `@vitejs/plugin-vue@5.2.4`。
- Vue 升到 `3.5.44`，Pinia 升到 `4.1.0`，dayjs 升到 `1.11.23`。合并最新 `main` 时保留 Wot UI V2，`@wot-ui/ui` 从 `^2.0.5` 锁定到 `2.3.2`；较早工作分支中的 `wot-design-uni@1.14.0` 已移除。
- Uni Helper 页面、组件、manifest、类型与 UnoCSS 插件按新版 peer 组合更新；UnoCSS/Applet 固定在 preset 要求的 `66.10.x` 范围。
- TypeScript 升到 `6.0.3`，配套更新 `vue-tsc`、`@vue/tsconfig` 和 Uni Helper 类型；CLI／lint／格式化工具、样式工具、Commitlint、代码检查器及 Rollup 一并更新到与当前工具链兼容的版本。Vitest 保留在与 Vite 5 相容的 `3.2.7`。
- Node 最低版本提高到 `22.22.2`，与 Pages 插件的引擎要求一致；当前 `.nvmrc` 的 `v26.9.0` 满足该要求。pnpm 仍锁定 `10.25.0`。
- 为新 peer 约束添加直接声明：Rollup 4、Vue DevTools API 8、Uni UI/Cloud 类型、Antfu ESLint 配置和 format 插件。这些包升级前已有传递依赖，但 pnpm 需在根依赖中明确选择兼容 peer 版本。
- 已核对新增 peer 包的 registry 版本、仓库地址与上游提交时间。GitHub API 请求触及 IP rate limit，未确认最近 issue／PR 的响应时间；未据此声称所有维护者均活跃。

## 保留较旧版本的原因

`pnpm outdated --format json` 仍报告以下 11 项。它们分别受当前 DCloud 预发布工具链的精确 peer、Uni Helper／UnoCSS peer 范围，或错误 dist-tag 限制；继续追 registry 的最大版本会破坏当前组合。

| 依赖                       | 当前锁定                 | 保留原因                                                              |
| -------------------------- | ------------------------ | --------------------------------------------------------------------- |
| `@dcloudio/types`          | `3.4.31`                 | 当前 Uni App peer 精确要求 `3.4.31`。                                 |
| `unocss`                   | `66.10.5`                | Uni preset/Applet 要求 `~66.10.1`；`66.10.5` 是该范围内最新兼容版本。 |
| `@antfu/eslint-config`     | `9.5.1`                  | Uni Helper ESLint 配置的 peer 范围最高为 9。                          |
| `@vitejs/plugin-vue`       | `5.2.4`                  | 当前 DCloud 插件固定依赖该版本。                                      |
| `typescript`               | `6.0.3`                  | Uni Pages 依赖 `^6.0.3`；Uni Types 1.4 支持 TypeScript 5／6。         |
| `vite`                     | `5.2.8`                  | DCloud Vite 插件声明精确 peer `5.2.8`。                               |
| `vite-plugin-vue-devtools` | `7.7.10`                 | 此版本支持 Vite 5；8.x 要求 Vite 6 及以上。                           |
| `vitest`                   | `3.2.7`                  | 此版本支持 Vite 5；5.x 要求 Vite 6.4 及以上。                         |
| `uni-app-harmony`          | `3.0.0-5030320261023001` | registry 的 `latest` 指向较旧的 alpha 版本。                          |
| `uni-components`           | `3.0.0-5030320261023001` | registry 的 `latest` 指向较旧的 alpha 版本。                          |
| `vite-plugin-uni`          | `3.0.0-5030320261023001` | registry 的 `latest` 指向旧版 Vue 2 alpha；本项目使用 Vue 3 线。      |

## 验证

- `pnpm install --frozen-lockfile` 通过，锁文件与依赖清单一致。
- `pnpm lint` 通过。
- 合并最新 `main` 后，`pnpm build:h5`、`pnpm build:mp-weixin`、`SKIP_OPEN_DEVTOOLS=true pnpm build:custom app` 均通过。构建触发的 AutoVersion 与 release manifest 变化已恢复；构建生成的类型声明更新已保留。
- 内置浏览器验证 H5 首页、导航到测试页和 Wot Toast，控制台无错误。微信模拟器页面为 `pages/index/index`，截图可见首页，`get_simulator_console` 未匹配错误；App 构建完成，未在 HBuilderX／真机运行。
- 实际交互截图：[H5 Toast](/Users/hmey/.codex/visualizations/2026/10/25/01a17147-2d1c-7d21-b943-04f4cd07580b/uniapp-deps-final-h5-toast.png) 、[微信首页](/Users/hmey/.codex/visualizations/2026/10/25/01a17147-2d1c-7d21-b943-04f4cd07580b/uniapp-deps-weixin.jpg) 。
- 合并后 `pnpm type-check` 仍因 `@wot-ui/ui@2.3.2` 内部源码的 24 条类型诊断失败，项目源码诊断为 0；没有关闭严格类型检查。
- `pnpm format:check` 在 42 个未修改文件失败；没有批量格式化这些无关文件。
- 按要求未运行测试命令。

安装日志仍含 peer 警告：DCloud Server Renderer 固定 Vue `3.4.21` 而应用使用 `3.5.44`；UnoCSS 可选 Inspector 对 Vite 有较高版本 peer；`@binbinji/unplugin-compress-json` 声明 Rollup 3 peer、微信 CI 内部 Babel ESLint Parser 声明 ESLint 7／8 peer。三个目标平台构建、H5 页面和微信模拟器均已验证，尚未验证 HBuilderX／真机运行。
