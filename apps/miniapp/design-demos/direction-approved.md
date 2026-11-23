# Direction Approved · CycleCompass 微信小程序改版

- **展示的方案**：
  - `apps/miniapp/design-demos/dir-1.html`（方向一 · 精密航海罗盘 · 暗场仪器盘）
  - `apps/miniapp/design-demos/dir-2.html`（方向二 · 财经报刊研报 · 暖纸排印版）
  - `apps/miniapp/design-demos/dir-3.html`（方向三 · 等高线周期地貌 · 2-Tab 轻流程）
  - `apps/miniapp/design-demos/dir-4.html`（备选四 · 三环信号监测 · Apple 健康/Oura 颜色编码）
  - `apps/miniapp/design-demos/dir-5.html`（备选五 · 周一纪律清单 · Things 3 原生精修）
  - `apps/miniapp/design-demos/dir-6.html`（备选六 · 港式行情招牌墙 · East 06 实验版）
- **用户选择原话**：`组合方案 方向一和方向四`
- **混合方案落地方案**：
  - **视觉基座（取自方向一）**：墨曜石暗场底 `#090C0F` + 深炭仪器面板 `#12161B` + 1px 发丝线网格 + 铜金/琥珀刻度强调 `#E5B352` + 等宽数字 `tabular-nums` + 严格 A 股红涨（`#F05A50`）绿跌（`#2AD1A3`），去 emoji、去 32 张同质白卡、去紫蓝渐变。
  - **核心信号语言（取自方向四）**：三同心策略体征环（宏观置信度 / 1500 日轮动分位 / 红利四维评分）+ 区间色带 + 半透明状态胶囊（`中间区` / `门控通行` / `波动预警`），颜色直接回答「今天该不该出手」。
  - **工程落地目标**：直接改造 `apps/miniapp` 真实 UniApp (Vue 3 + Tailwind CSS v4 + weapp-tailwindcss v5) 源码，确保 100% 兼容微信小程序 WXML/WXSS（不使用 WXML 不支持的内联 `<svg>`），并通过 `pnpm type-check`、`pnpm test:run` 与 `pnpm build:mp-weixin`。
