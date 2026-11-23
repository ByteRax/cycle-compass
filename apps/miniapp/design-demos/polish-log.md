# Polish Log · CycleCompass 小程序三方向初稿

> 运行环境说明：当前沙箱未安装 Python 版 `playwright` 且受限无法启动本机无头 Chrome（`spawn EPERM`），因此本轮初稿未生成自动截图，改为按 `references/polish-loop.md` 与 `references/app-prototype.md` 进行逐项几何坐标、可读性底线与交互状态审查，并已修复以下 3 处结构问题。

## Round 1 · 三方向几何与规范审查

### 发现的最伤三处
1. **`dir-2.html` 右侧三台次级手机缩放过小（违反 ≥300px 可读底线）**：原 `scale: 0.64` 导致机身宽 `252.8px < 300px`，窄屏表格文字易偏小。
   - **已修复**：将右侧三台手机统一上调至 `scale: 0.76`（`300.2px`），右栏 `left` 调整为 `1084px`，画布总高扩展至 `2220px`，三台手机垂直间距 `36px` 无挤压。
2. **`dir-1.html` 1440px 展示板右侧留白失衡**：原右侧 `1132px~1440px` 存在 `308px` 空白，且转发卡片压在主屏正下方。
   - **已修复**：重组为均衡四栏（主屏 `56px` / 次屏列 `476px` / 产业屏 `812px` / 转发卡与色彩推导栏 `1144px`），整体视觉重心收束在 `56px~1400px`。
3. **跨三版外链依赖与离线可靠性**：原计划使用 `unpkg.com` React+Babel CDN 在受限网络下可能白屏或编译延迟。
   - **已修复**：三版均重构为零外链自包含 HTML/CSS/SVG/JS 单文件，双击 `file://` 秒开且所有 tab 切换、子板块筛选、一键周定投打卡、回测展开均可真实点击交互。

### 一致性核验
- 三版骨架、色温、字体嗓音、导航层级与展示板构图（`hero` / `closeup` / `flow`）均完全不同。
- 全部使用 `apps/miniapp/src/pages/index/index.vue` 真实数据，零 emoji 图标、零紫蓝 AI 渐变、零 `&lt;` 乱码与标注重叠。
- 待用户选定方向并记录 `direction-approved.md` 后，将直接把选定方向的视觉系统与组件落地改造到 `apps/miniapp` 真小程序代码中，并跑通 `pnpm type-check`、`pnpm test:run` 与 `pnpm build:mp-weixin`。
