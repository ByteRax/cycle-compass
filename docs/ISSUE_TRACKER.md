# 周期罗盘 CycleCompass 迁移问题追踪与闭环清单 (Issue Tracker)

| Issue ID | 发现阶段 | 问题现象与根因 | 修复方案与闭环状态 |
| :---: | :---: | :--- | :--- |
| **ISSUE-001** | Phase 0 | Python 镜像构建因 `orjson` 缺少预编译 wheel 而编译失败 | 移除 `orjson`，FastAPI 0.165+ 走 Pydantic v2 高性能直出字节流。(已闭环) |
| **ISSUE-002** | Phase 1 | `curl -I` 请求 FastAPI 接口返回 `405 Method Not Allowed` | Go Router 对所有 GET 路由同步注册 HEAD 处理支持。(已闭环) |
| **ISSUE-003** | Phase 4 | 腾讯行情接口返回 GBK 编码导致中文标的名称在 Go 中乱码 | 引入 `golang.org/x/text` 进行 GB18030 自动转码，结合 `KnownAssets` 字典精准兜底校正。(已闭环) |
| **ISSUE-004** | Phase 4 | 模拟投资流水 SQLite 查询排序与黄金样本逆序 | 修改 SQL 排序规则为 `ORDER BY id ASC`，与黄金样本严格对齐。(已闭环) |
| **ISSUE-005** | Phase 4 | 回测周定投事件描述与黄金样本存在前缀差异 | 去除定投事件描述中的多余前缀，统一格式为 `本周定投 {amount} 元买入 {target}`。(已闭环) |
| **ISSUE-006** | Phase 5 | 模拟投资账本流水在服务重启时丢失 | 引入 SQLite WAL 数据库持久化存储，首次启动自动 Seed。(已闭环) |
| **ISSUE-007** | Phase 6 | 2GB 服务器物理内存吃紧，常驻内存高达 225.8MB | 迁移 Go 纯静态二进制，内存降至 9.5MB，释放 220MB 物理内存。(已闭环) |


---

## V4 金融资产模块基线与实施问题追踪

| Issue ID | 发现阶段 | 严重度 | 问题现象与根因 | 处置状态 |
|---|---|---|---|---|
| **V4-001** | G0 | P1 | `apps/miniapp` 基线 `pnpm type-check` (`vue-tsc`) 存在 234 个类型错误（`DivergenceChart.vue` 22, `ComponentTrendView.vue` 1, `divergence.ts` 176, `seeds.ts` 33, `divergence/index.vue` 2），因开启 `noUncheckedIndexedAccess` 及 `seeds.ts` 使用未定义状态 `'online_sync'` | `OPEN`（将在 G3/G4 重构相关文件时一并修复，并修复 `seeds.ts` 类型） |
| **V4-002** | G0 | P2 | `apps/miniapp` 基线 `pnpm lint` 存在 923 个历史遗留问题（904 errors, 19 warnings），主要在配置文件和旧组件 | `OPEN`（遵循局部原则，不批量改无关旧文件；所有 V4 新增/修改文件必须 0 lint error） |
| **V4-003** | G0 | P0 | `src/data/divergence.ts` 使用 `getTodayDateStr()` 冒充行情日期、`makeSeries` 生成模拟数据；`TrendDashboardView` 写死 `score=78`；`SolarTermsView` 用自然节气；`RightDistributionView` 写死 A 股分组和 `totalCount=50` | `OPEN`（G3/G4 全部替换为真实 API 与算法计算） |
| **V4-004** | G0 | P1 | `services/api-go/internal/market/client.go` 仅映射 5 个腾讯代码，无历史日线数据底座与资产分类体系 | `IN_PROGRESS`（G1 建设历史行情与资产底座） |
| **V4-005** | G6 | P0 | `finance/index.vue` 与 `TrendDashboardView.vue` 请求 `sort=strength` 时 Go `allowedSortFields` 仅允许 `strength_score` 导致 HTTP 400 | `FIXED`（`trend_handler.go` 与 `trend_service.go` 同步支持 `strength`/`temperature`/`days` 别名） |
| **V4-006** | G6 | P0 | `asset.vue` 手写 Canvas 缺少 `.in(instance)` 导致详情曲线空白，且 `openWorkbench` 指向不存在的 `/pages-sub/finance/workbench` | `FIXED`（复用 `TrendLineCanvas.vue`，修正为 `/pages-sub/divergence/index`） |
| **V4-007** | G6 | P1 | `TrendDashboardView.vue` 空数据时标题重复 `趋势仪表盘 · 趋势仪表盘`；`trend_events` 重启插入时 `event_id` 主键冲突 | `FIXED`（修正标题拼接与 `ON CONFLICT DO NOTHING`） |
| **V4-008** | G8 | 外部阻塞 | 微信小程序真机/开发者工具真机截图与商用行情正式授权（中证/交易所）尚未在当前无真机环境中完成 | `BLOCKED`（已构建 `dist/build/mp-weixin`，待微信真机与法务授权确认后上线） |
