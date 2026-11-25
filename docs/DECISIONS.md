# 周期罗盘 CycleCompass 技术决策记录 (Architecture Decision Records)

## ADR-001：Go Web 框架与路由器选型
- **决策**：选用 Go 官方标准库 `net/http` + `github.com/go-chi/chi/v5` 路由器。
- **理由**：Chi 100% 兼容 Go 官方 `http.Handler` 标准，零过度抽象，支持精准的中间件编排（Gzip、Recoverer、RealIP），内存与 CPU 开销微乎其微。

## ADR-002：嵌入式数据库选型
- **决策**：选用纯 Go 实现的 `modernc.org/sqlite`，关闭 CGO 编译。
- **理由**：
  1. 纯 Go 实现无需安装 gcc/musl，跨平台编译（Mac -> Linux amd64）零外部 C 库依赖，完全消除动态链接库丢失风险。
  2. 启用 WAL 模式与 5000ms BusyTimeout，读写性能达纳秒级，彻底解决原 Python 进程内存持仓在重启时数据易失的问题。

## ADR-003：生产切流机制
- **决策**：采用独立端口 `8002` + 1Panel OpenResty 热重载切流。
- **理由**：
  1. 不直接覆写或停机旧服务，新旧服务在服务器内部并存（`8000` 与 `8002`）。
  2. 在本地通过完整 11 个接口回归后，修改 Nginx upstream 并执行 `nginx -s reload`，实现毫秒级零丢包热切流。
  3. 保留旧容器镜像在磁盘，具备 30 秒内秒级一键回滚能力。

## ADR-004：生产进程托管
- **决策**：采用 systemd 原生托管单文件纯静态 Linux 二进制。
- **理由**：进一步省去 Docker 守护进程与 bridge 虚拟网卡的额外内存开销，服务常驻物理内存仅 9.5MB，开机自启且具备 3 秒 Crash 自愈能力。


---

## V4 金融资产模块架构决策

## ADR-005：金融资产分包路由与旧入口兼容
- **决策**：保留 `pages/index/index.vue` 四 Tab（策略/行情/产业/账本）及 `TabMarkets.vue` 原有四大子页；在 `TabMarkets.vue` 顶部新增「金融资产」全景入口，路由指向 `pages-sub/finance/index`；保留 `pages-sub/divergence/index` 原 URL，支持旧数字 `view=1..6` 及新增 `view=7`/`chart=tier` 和新字符串 `chart` 参数，旧 `id`（`a_shares`/`dividend_etf`/`etf_macro`/`a_shares_long`）映射到新资产/观察池 ID。
- **理由**：不破坏已上线导航与深链，避免主包体积膨胀。

## ADR-006：SQLite 增量迁移与不可变计算快照
- **决策**：在 `services/api-go/internal/storage` 引入版本化 `schema_migrations` 增量迁移，新增资产、分类、历史日线、算法版本、趋势快照、每日指标、相对强度排名、事件与标签表；绝不改动现有 `accounts`/`positions`/`transactions` 表。每次趋势计算生成唯一 `snapshot_id`，原子更新 `is_current` 指针，确保列表、七图、报告读取同一不可变快照。
- **理由**：保证账本零回归，避免前端在刷新跨天时混用不同批次指标。

## ADR-007：自研可解释量化模型 `trend-v1.0.0`
- **决策**：在 `services/api-go/internal/trend` 实现纯 Go 量化计算引擎，与现有双策略（轮动/四维打分）完全隔离。包含 7 档绝对趋势温度（`50+50*tanh(Q/2)`）、3 日确认/2 日退出状态机、横截面相对强度 `RS=100*(N-rank)/(N-1)`、趋势节气生命周期映射、5 级趋势梯队（S/A/B/C/D）、O(n) 极值回撤及自然年周次（53 桶）年度同期。
- **理由**：全部计算可追溯、可幂等重算、可离线通过 Python 独立参考实现交叉验证。

## ADR-008：金融模块作用域深浅主题
- **决策**：保持 `app.css` 暗色墨曜石 `#090C0F` 为全站默认；扩展 `stores/theme.ts` 与 `stores/finance.ts` 支持金融研究子页面独立选择 `dark` / `light` / `system`，通过作用域 class（如 `.cc-finance-light`）覆盖 CSS 变量，不波及策略/行情/产业/账本主 Tab。
- **理由**：满足 V4 浅色金融阅读主题需求，同时确保已有页面视觉零回归。
