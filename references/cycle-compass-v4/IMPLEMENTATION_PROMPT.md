# /goal 周期罗盘金融资产 V4 · 完整增量实施提示词

> 将本文件连同 `design/`、`docs/` 交给具备本地代码编辑和测试能力的开发 Agent。若宿主不支持 `/goal` 命令，直接以本文件作为完整任务说明。所有设计稿均为目标态，不是现有实现截图。

## 0. 你的角色与终极目标

你是 CycleCompass 的首席产品工程师、Go 后端架构师、量化数据工程师、UniApp 微信小程序资深开发者和 UI/UX 设计负责人。基于仓库 `https://github.com/ByteRax/cycle-compass` 的当前代码，逐阶段完整实现「金融资产」模块，借鉴趋势动物 Pro 的公开功能和用户提供的截图，但使用周期罗盘自主品牌、真实合法的数据、可解释的算法及原创设计。

目标不是制作另一个模拟行情页面，不是仅加几个图表，不是重写原项目。要将已有 `pages-sub/divergence/index.vue` 的六大趋势视角和七级温度体系升级为完整的真实数据驱动研究能力，形成「资产体系 → 历史行情及综合指数 → 趋势分析 → 七图研究 → 报告分享 → 收藏及设置」闭环，并保留已上线的策略、行情、产业和模拟账本。

任务完成的定义：P0 所有功能均实际接线，具备可复现数据、可验证指标、真实状态、深浅色主题和真机验收证据；P1/P2 按后述依赖逐步完成，未实现、无合法数据或待审核的能力必须明确记录，不可伪造成功。

基线提交：`master @ d1a322d8b1c7cea783a21d54db936da7a89cb308`，提交时间 2026-11-24 16:45:28 UTC。先查看 `docs/REPO_AUDIT.md`（随本包交付），然后以你实际检出的最新代码为事实源，若 HEAD 更新则重新比对差距。设计图及数据仅用于视觉与测试，不可当作线上数据。任何文件路径和接口新增方案均须与最新实际工程核对后落地。

## 1. 执行纪律与禁止事项

1. 首先 `git status --short`、`git rev-parse HEAD`、检查分支、已有修改、现有 CI、最新 `README`、各级 `AGENTS.md`、`CLAUDE.md`、`package.json`、`go.mod`、`pages.config.ts`、`manifest.config.ts`、`vite.config.ts`、`docs/DECISIONS.md`。先读源码，建立复用/替换/新增/延期矩阵。不覆盖用户未提交改动、不 reset --hard、不清理用户文件、不删除旧页面或策略算法。
2. 当前生产服务是 Go + chi + SQLite（WAL、单连接池），不是 Gin/PostgreSQL。沿用 `services/api-go` 和现有 Docker Compose；保留 `services/api` 及 `packages/*` 研究/校验职责。遵守实际 `go.mod` 的工具链（基线 1.26.0），不擅自替换数据库、重构成微服务、部署生产或改动真实账本数据。
3. 前端沿用 `apps/miniapp`：UniApp + Vue 3 + TypeScript + Vite + Pinia + UniKu + Tailwind CSS v4 + weapp-tailwindcss v5，pnpm@10.25.0（如 HEAD 已变，以实际锁文件为准）。复用 http、router、BaseLayout、CSafeArea、CTabs、反馈组件和既有图标风格。禁止重新引入 Wot Design Uni、UnoCSS、unocss-applet、z-paging、Oxlint、Oxfmt。保留 ESLint / Prettier / vue-tsc / Vitest。对过时 AGENTS 说明增量修正，不照过时文档回退依赖。
4. 禁止生产页面使用模拟时间序列、固定分数、随机行情、写死行业/日期、`Date.now()` 冒充行情时间、`|| 2` 将0改成非零、未知状态当中性、未接线的空按钮。允许固定合成 fixtures 仅用于测试/设计，必须标明「演示数据」且与生产构建隔离；离线真实快照须保留来源与原日期。
5. 不复用竞品 Logo、水印、二维码或专有代码；不尝试绕过付费权限或采集受限数据。指标、标签、报告为客观研究描述，不提供个性化证券推荐、保证收益、自动交易、代客理财。原有双策略模拟研究/账本作为独立业务，不被新温度指标覆盖。移除新趋势页面中的「果断建仓」「坚决斩仓」「反转做空」等指令式话术；原有策略规则不擅自改变。不要使用「上天台」等不必要标签。
6. 数据授权是上线 gate：核对每个数据源的合法获取、商用、缓存、衍生计算、公开展示、分享导出权限，核查微信小程序类目、金融信息服务和投资咨询边界。未验证权限的资产标记「暂未覆盖」/隐藏发布，不能以免责声明替代资质。BTC/虚拟货币、未上市公司、期权和复杂衍生品默认不公开上线，专项核查后再启用；本轮不做房产模块。
7. 不暴露 API Key、用户隐私、真实账户号或完整本地数据库；密钥在服务端环境变量；不把测试种子、竞品原图、UI PNG 放入小程序主包。不得擅自 push、合并或发布生产，最终提交本地可审查改动，只有明确授权才进行远程写入/生产发布。

## 2. 分阶段 Goal 与 10 个并行子代理

若宿主支持 subagent，使用 10 个角色；如不支持，用串行任务模拟同样的责任矩阵。协调者先冻结 DTO、算法口径、路由和设计 token，分配互不冲突的文件所有权；并行阶段只允许各角色在自己负责的文件中写入，必要时用 worktree/分支隔离，接口公共文件由协调者合并。禁止多个 Agent 同时修改同一个核心文件或 migration 编号。每个子代理必须给出证据（路径、测试结果、失败/风险）而非只输出建议。

| Agent | 责任与产物 | 依赖/冲突控制 |
|---|---|---|
| A1 架构审查/协调 | HEAD 基线、模块复用矩阵、ADR、阶段状态、DTO 冻结、汇总 PR 说明 | 唯一修改公共路由、接口合并及迁移编号 |
| A2 资产与数据 | 标的注册、分类/观察池、数据许可矩阵、Provider、交易日历、数据清洗 | 与 A3 冻结 series contract；不得伪造行情 |
| A3 量化引擎 | 复权/收益、指数、温度、方向状态机、节气、强度、梯队、事件 | 用固定 fixtures 和独立 Python 参考校验 |
| A4 持久化/API | SQLite 增量迁移、Go DTO/service/handler、查询分页/缓存、快照发布 | 由 A1 整合 main/router，禁止写入旧账本数据 |
| A5 资产页面 | 行情入口、全景、分类列表、资产详情、搜索选择器 | 复用现有布局与路由，消费真实 API |
| A6 时间序列图 | 曲线、年度同期、成分趋势及 Canvas/交互核心 | 统一 ChartContext / renderer，与 A7 对齐 |
| A7 结构图表 | 仪表盘、节气、右侧分布、梯队、状态分布 | 复用 A6 图表内核；无硬编码样本 |
| A8 设计/报告 | 主题、图标、海报排版、高清导出、分享/扫码还原 | 不复制现有六图组件、不引入字体文件；真实数据同快照 |
| A9 用户功能/质量 | 自选、图表收藏、偏好、错误态、可访问性、前端交互测试 | 先本地存储，用户隔离明确；不扩大权限 |
| A10 独立 Review/测试 | 代码/数据审计、回归、真机、性能、安全/合规与验收 | 不自行放行自己修改；对 P0 问题提出复现步骤 |

阶段：G0 基线与现有行为测试 → G1 数据与持久化 → G2 算法与 API → G3 入口/列表/详情 → G4 七图 → G5 报告/自选/设置 → G6 集成与真机回归 → G7 P1/P2 有条件增强。每阶段更新 `docs/GOAL_PROGRESS.md` 与 `docs/ISSUE_TRACKER.md`，记录任务、依赖、测试、阻塞、下一步。阶段完成必须满足质量门槛，不以“编译通过”替代数据验证；可并行的任务先并行。遇到不可获得合法数据、资质或真机时，先完成不依赖该阻塞的部分，记录 blocker，严禁伪造测试结果或将 mock 包装为真实功能。每轮 review → fix → verify；有 P0/P1 缺陷继续迭代，完成后提交最终可核验报告。

## 3. 产品架构及迁移（P0）

### 3.1 一级导航和新入口

保留 `src/pages/index/index.vue` 当前 `cycle/markets/sectors/portfolio` 四 Tab 及 `utils/swipe.ts`、现有默认策略页和滑动行为。`pages/index/components/TabMarkets.vue` 的 `openDivergence()` 原先固定进入 `?id=etf_macro`，保留旧快捷入口（或变成「趋势研究」次入口），在行情页顶端新增「金融资产」入口：全景、数据状态、自选快捷入口。不挤占旧市场子 Tab，不新增一套全局 TabBar。金融模块内用正常页面栈、返回和顶栏收藏/搜索，返回时恢复滚动位置、筛选和分组状态。保持现有深链与分享路径兼容。

建议增量页面（最终命名可随现有自动路由约定调整）：

| 路由/逻辑视图 | 职责 |
|---|---|
| `/pages/index/index`（行情 Tab） | 保留现有页面，新增金融资产入口 |
| `pages-sub/finance/index` | 金融资产精选观察池/市场目录，资产全景 |
| `pages-sub/finance/asset` | 统一资产/分类详情（id、poolId、path），内嵌曲线、成分列表 |
| `pages-sub/finance/search` | 搜索、历史记录、选择资产/基准/候选池 |
| **原** `pages-sub/divergence/index` | 保留 URL 并升级七图工作台，`assetId`、`chart`、`range`、`currency`、`snapshotId`、`poolId` 等 |
| `pages-sub/finance/favorites` | 收藏资产、收藏图表、默认主页、本地分组 |
| `pages-sub/finance/report` | 报告预览、导出、分享与恢复配置 |
| `pages-sub/finance/settings` / `guide` | 主题、币种、数据来源、指标方法、更新时间、研究偏好 |
| `pages-sub/finance/events`（P1） | 趋势事件历史、筛选和跳转 |
| `pages-sub/finance/universe`（P1） | 自定义观察池，资产选择与排序 |

页面自动发现沿用 `pages.config.ts` / `definePage()`，由构建生成 `src/pages.json`，禁止手工改生成文件作为事实源。选择器和配置面板可以是同一页面的弹层，非必要不新建路由。路由路径按真实项目分包配置校准。

### 3.2 兼容当前实现

- `divergence/index.vue` 目前 `currentViewIndex=2`，`assetOptions` 四个 ID `a_shares`、`dividend_etf`、`etf_macro`、`a_shares_long`，数字 `view=1..6`。旧链接保留原默认第三视图；旧 `id` 映射到真正 asset/pool 配置（超长周期为独立研究预设，不是重复资产）；旧数字顺序严格保留，第7张为 tier。新路由采用稳定字符串 `curve`/`annual`/`components`/`dashboard`/`season`/`distribution`/`tier`。从资产详情打开默认 curve。非法 ID / view / range 明确提示并选取合理回退；未知资产不静默变成 A股。分享配置校验、URL 编码、快照版本兼容。
- 现有 `src/data/divergence.ts` 七级字典保留枚举与颜色体系，生产数据从新 API 获取；迁移 `DIVERGENCE_DATASETS`/`makeSeries` 到测试 fixtures，移除运行时 `TODAY` 假数据日期。`getTempConfig` 可兼容旧 `fair` → `flat`，但 null/未知返回 unavailable，禁止默认 flat。真实历史长度不足时相应时间范围禁用并说明，不画假 27 年。
- `components/trend/` 六个既有组件增量重构，新增 `TrendTierView.vue`；`TrendMenuDropdown.vue` 统一使用 ChartRegistry，不重复写死六项；`DivergenceChart.vue` 复用轴与标签几何逻辑，修复实际日期和过冲；`TrendDashboardView.vue` 从固定温度仪表重构为数据报表，旧仪表迁为 `TrendTemperatureGauge` 可用于资产详情（不删除有用组件）；`SolarTermsView` 删除当前自然节气映射；`RightDistributionView` 删除硬编码 groups/totalCount 和假2%。

## 4. 统一资产数据体系

### 4.1 对象及分类

底层区分 `asset`、`category`、`membership`、`universe`（观察池）、`index_definition`（明确权重的组合/基准）。唯一资产 ID 不能仅用 symbol（市场、类型、交易所、合约区分），例如 `CN.SSE.510300`、`CN.FUT.SHFE.AU.CONT`、`US.INDEX.SPX`（示意命名）。同一资产可属于多个行业分类/专题/观察池，行情仅存一份。分类和观察池都可作为研究节点，但只有已定义合法综合指数/真实标准指数才具有「最新指数值」及趋势曲线；单纯观察池没有可计算组合指数时仅展示成分视图。

建议首发覆盖按合法数据和服务器能力分批启用：

| 资产域 | P0 基础覆盖目标 | 扩展（P1/P2） |
|---|---|---|
| A股 | 标准宽基、行业代表指数/代理，金融/能源/材料/科技/制造/消费/医药等 | 行业内个股、细分行业完整成分 |
| ETF基金 | 宽基、行业、策略、债、商品、跨境、货币、可转债ETF；区分主动ETF、LOF等实际证券类型 | 每只ETF及历史规模/净值、行业多级分类 |
| 美股、港股、全球股指 | 合法公开可展示的代表指数、部分行业 | 美股/港股成分及美国/香港ETF |
| 商品期货 | 黄金、白银、铜、原油、铁矿、农产品等合法日线代表品种 | 完整合约、连续合约编制、更多商品 |
| 债券/外汇/现金 | 可获得的全价/全收益债券指数、货币对、明确现金基准；收益率仅作为观测指标 | 国债期货、REITs、可转债及更多品种 |
| 自定义观察池 | 精选金融大类，不含房产 | 用户自定义及可追溯权重组合 |

数据来源不能仅依据竞品截图或公开网页可访问即视为合法授权。先填写 `docs/DATA_SOURCE_MATRIX.md`：asset/pool、provider、标识映射、支持历史起始日期、更新节奏、数据单位/币种、许可范围、延迟、限流、失败降级、证据、商用审批状态。只有确认权利的市场能公开展示；若未覆盖显示覆盖范围与原因。数据源选择须按接口合同验证，不硬编码 Tushare 等未授权账户。现有 Tencent quote 客户端覆盖有限（映射函数仅列 5 个代码），不得将它宣称为全球历史数据源。日频历史独立 `HistoricalProvider`，复用已有 quote provider 接口/质量信息，必要时保留 Python 离线校验器，而非引入高成本运行依赖。

### 4.2 标准行情/收益

存原始报价及标准化序列。字段：`asset_id, market, symbol, measure_type, price_type, unit, currency, timezone, trading_date, as_of, available_at, source_name, quality_status, open/high/low/close, adjusted_close, total_return_index, volume, turnover, float_market_cap, fund_aum, open_interest_value, dividend_yield, yield_pct, adjustment_version`，可空字段使用 null，不能用0填缺失。价格、净值、收益率（百分比）、汇率（货币对）、合成指数分开。`return_type` 记录 price/total_return/continuous/FX-adjusted；`yield_pct` 变化用 bp 或百分点，不能按股价涨幅解释。对报价货币转换需按历史汇率和方向；只在数学可比资产上计算回报。

- 股票使用明确复权/全收益序列；ETF 市场价与净值分开；债券收益率与全价指数分开；期货连续合约须明确换月规则（不直接拼接原始价格造成假跳空），可能负值/零或合约特殊情况用明确的合约收益定义，否则不支持。汇率定义 base/quote，币种换算使用目标币种/原始币种当日或当时可用的汇率，禁止用最新汇率回填历史。
- 对可比正值财富指数 `W`，区间收益 `W(t)/W(t0)-1`；`W_target = W_local * FX_t`（以相同基期归一化）。现金基准 1 或计息现金序列必须标注，两者不可混用。
- 时间：`as_of` 实际源时间、`available_at` 可用时间、`trading_date` 当地交易日、`updated_at` 写入时间，全部可追溯；市场日历含节假日、半日、跨时区、期货夜盘归属。周/月收益按起点前一有效收盘到该期末；rolling 1M/1Y 使用历法回溯及明确的前一有效点，不等同本月/今年；表头清楚。展示截面允许“各品种最新值”并标注日期不齐；严谨排名/回测默认统一可用截止时点，禁止未来数据、未来成分、后视修订和错误填充。缺失点标 gap，连续作图不得虚构价格；受限 carry-forward 仅用于明确的对齐展示，不当作市场实际交易或新信号。
- 数据质量：延用 `ok/degraded/offline_seed/error`，可新增 `stale/insufficient/not_covered/market_closed` 独立业务状态，不混淆 provider status 与交易状态。离线缓存保留原来源/日期，失败可重试；网络收到 seed 不等于真实成功；不得自动以系统日期更新快照。元数据 `DataSourceMeta` 兼容旧字段，新增数据覆盖、staleness、license_flags。

### 4.3 综合指数/观察池

正式指数有合法供应商序列时优先使用，并明确指数代码/供应商。自研指数基期 100 或 1000，成分和权重有效期明确，采用 `I_t=I_(t-1)*(1+sum(w_(i,t-1)*r_i,t))`，成分收益为一致币种/收益类型，权重归一化，设定可用覆盖率（如初始>=90% 权重，需校准）。成员变更、暂停、再平衡、公司行为均追溯至当时已知信息。不得用绝对价格均值替代指数；不默认将重复覆盖的全球股指+美国+ETF 等观察池等权合成为金融指数。数据不足则 `null`，说明原因。

## 5. 自主量化模型 V1（可执行初始规则，不是原版算法）

**指标计算全在 Go 端完成，前端仅格式化与可视化。** 先按现有代码校验是否已有可复用的纯函数，新增与旧双策略分离的研究模块。所有窗口、阈值、分档、状态转换、标签 TTL 由 `algorithm_versions` 配置，版本号例如 `trend-v1.0.0`，可追溯、可测试、可回填。参数是研发初始值，不代表投资有效性；通过不同市场历史、样本外和敏感性验证后再发布。不要为追求与竞品相同数值进行过拟合。

### 5.1 输入与绝对趋势温度（七档）

- 仅使用连续可解释的正值财富/价格序列，以有效交易日为样本；数据清洗不通过时不可计算。初始默认完整窗口 121 个有效点、且覆盖相应市场最近 121 个预期交易日的至少 90%；窗口缺失过多、未上市足够久、负值、不适用均返回 `null + reason`。新资产可按第5.3独立较短排名窗口参与强度，不强行补出温度。
- 用 `r_n = ln(W_t/W_(t-n))`；`sigma60` 为过去60个单日 log return 样本标准差（仅可比序列），下限 epsilon=1e-6；零波动且零收益取0，避免0/0。`z_n = clamp(r_n/(max(sigma60,epsilon)*sqrt(n)), -4, 4)`；`slope20 = (EMA20(t)-EMA20(t-5))/max(abs(EMA20(t-5)),epsilon)`，`slope60` 同理（采用明确初始化 SMA）。令 `a = clamp(ln(EMA20/EMA60)/(max(sigma60,epsilon)*sqrt(20)),-4,4)`；`b = clamp(slope20/(max(sigma60,epsilon)*sqrt(5)),-4,4)`；`c = clamp(slope60/(max(sigma60,epsilon)*sqrt(5)),-4,4)`。初始 `Q = 0.25*z20 + 0.25*z60 + 0.20*z120 + 0.15*a + 0.10*b + 0.05*c`；`temperature_score = clamp(50+50*tanh(Q/2),0,100)`。用未四舍五入值分档，显示可取整数。温度独立于横截面强度。
- 沿用仓库现有 `TrendDashboardView` 阈值作为自研初始值（不是竞品验证）：冻 [0,15)，寒 [15,29)，凉 [29,43)，平 [43,58)，温 [58,72)，热 [72,86)，沸 [86,100]。缺失为 unavailable；允许配置化后续升级，UI 始终七级。仅客观描述“极弱、下行、偏弱、中性、偏强、强势、极强”，不宣称必然上涨/见顶。

### 5.2 方向、确认及天数状态机

- 独立字段 `direction = up|down|neutral|unknown`、`confirmation = forming|confirmed|none|unknown`、`side = right|left|neutral|unknown`（本产品明确将右侧定义为确认上行、左侧定义为确认下行；与其他交易术语不混用），`started_at`、`confirmed_at`、`duration_trading_days`、`duration_calendar_days`、`state_version`。事件和天数由历史状态推导，不能从温度等级硬映射。
- 初始确认：UP 条件 `W>EMA60 && EMA20>EMA60 && slope20>0 && Q>=0.35`；DOWN 对称 `W<EMA60 && EMA20<EMA60 && slope20<0 && Q<=-0.35`。连续3个有效交易日满足则确认，形成阶段单独显示。确认后只要没有退出条件保持状态，UP 退出条件 `W<EMA60 || Q<0.10` 连续2个有效交易日；DOWN 对称 `W>EMA60 || Q>-0.10` 连续2日。退出到 neutral/forming，反向确认需满足3日；若退出与反向确认同一日，明确优先处理退出、再进入且只生成一次事件；可将反向确认计数从退出日重新计（定死并测试）。确认日=第1天，按各市场有效交易日递增，休市不增加，缺失不生成新信号、状态标 stale/unknown；当覆盖恢复重新评估。显示右侧+N、下行-N（列表）与完整文字（详情），neutral 为 `—`；中性持续不混入左右侧天数。
- 保存每日状态快照及转换事件；幂等重算、历史修订产生新版本，不覆盖已发布的审计快照。趋势背景区间来自实际确认状态序列。

### 5.3 相对强度（0～100，按候选池）

- 定义研究原始因子而非温度分数：初始 `M=0.55*z20+0.45*z60`（用 `sigma20` 替代 `sigma60` 也可作为版本化另一算法，但不要混用；本版使用 `sigma60`，至少61有效点）。风险调整比较需要一致收益定义、币种、截止时间、可比范围；收益率/不可比序列不纳入。历史不足61点则 unavailable，不为满足截图效果虚构次新品种分数。零波动稳定序列取0，明确现金基准是否可纳入，单独测试。
- 固定 universe membership + `snapshot_id + currency + algorithm_version`。按 M 降序，等值取平均名次，`RS=100*(N-rank)/(N-1)`，N>=2；N<2 或无有效因子 -> null。存 raw_factor、rank（可能小数）、candidate_count、valid_count、display_score（四舍五入规则明确）、source date、scope label。排名和梯队使用未四舍五入分数；若边界显示整数可能误解，边界展示1位小数或解释。0 与100是真实分数。
- 分别计算全局精选池、同市场/大类池、行业/ETF/期货池、自定义观察池；「页面内强度」=当前固定候选集合，不按分页或滚动可见区域重新计算。排序、筛选默认不改原范围；用户显式「按筛选集合重算」时生成新 snapshot/scope，所有图表用相同范围。强度变化按同池同币种同算法同可比成员条件比较，成员或方法变化须注明不可比；初始 Δ>=5 单箭头、>=15 双箭头，反向对称，其他不显示。

### 5.4 趋势节气（生命周期）

保留前后截图已出现的上行「谷雨/立夏/夏至/小暑」、下行「寒露/立冬/冬至/小寒/大寒」，上行末期用自研「大暑」。独立于自然日期，采用以下初始可解释状态映射：

| 阶段 | 自研初始规则（方向先确认） |
|---|---|
| 谷雨 / 寒露 | 上行/下行 forming，未确认 |
| 立夏 / 立冬 | 已确认 1–10 有效交易日 |
| 夏至 / 冬至 | 已确认 11–40 日 |
| 小暑 / 小寒 | 已确认 41–90 日 |
| 大暑 / 大寒 | 已确认 >90 日 |
| 震荡 / 未知 | neutral / unknown（不强制分配自然节气） |

此版本只用方向、确认及持续时间，属于透明可验证的最小可用模型；后续在独立版本中加入动量加速度、标准化幅度、风险或状态回退条件，先做回测和校准。不得宣称原版的阶段“末期”预示即将反转。曲线是示意状态轴不是未来价格，图上应明确。阶段可以提前退出或重入，状态机版本切换要可解释；日历节气与金融节气完全分离。

### 5.5 五级趋势梯队（截图明确的展示规则）

用户截图规则：≥90且右侧 → 最高级；≥90但非右侧 → 第二级；≥70 → 第二级；≥30 → 第三级；≥10 → 第四级；<10且左侧 → 第五级；<10非左侧 → 第四级。未知 side 本版标为待确认/不入榜并列示原因；neutral 与 unknown 不混用。无有效 RS 不入榜，单独列“暂无强度”。专业标签 S领跑/A强势/B中坚/C偏弱/D落后；趣味别名 夯/顶级/人上人/NPC/拉完了 可作为用户主动切换的可选展示。按 raw score 降序，同分用稳定 assetId，候选直属成分或固定观察池，可配置最大样本数且解释规模/权重筛选。展示总数、有效数、未入选数、分档规则、池范围和日期。禁止按市值/价格直接赋予强度。

### 5.6 事件与资产标签

转换事件：`right_started/right_ended/down_started/down_ended/temperature_changed/phase_changed/tier_changed/strength_changed/new_high/new_low/drawdown_expanded`。包含 `event_id, asset_id, event_type, old_value, new_value, as_of, detected_at, snapshot_id, algorithm_version, scope_id, dedupe_key`，同一状态/批次只生成一次，重算幂等。列表显示最近一有效期/用户配置有效期，历史仍可查询；「进入右侧」与持续右侧背景分开。资产特征 `liquidity_leader/size_leader/trend_leader/new_listing` 按明确范围条件持续有效（初始前10%或 Top3，范围样本不足则不赋标，阈值版本化）；历史新高必须有可核验全量历史，否则写「可用历史新高」或阶段新高。危险标签用客观名称「异常波动/回撤扩大」及阈值解释，勿将未知风险当投资建议。

### 5.7 通用统计

正值财富指数：最大回撤 `min_t(W_t/max_{s<=t}W_s-1)`，最大上涨 `max_t(W_t/min_{s<=t}W_s-1)`，线性 O(n) 扫描并返回起止日期/值，不能用全局高低点乱配。当前回撤与区间最大回撤分开。最高最低点、收益、成交额、标题从完整原始序列计算，不用降采样曲线算。对数坐标对正值财富指数取 log 后标为收益百分数；不能对负收益取对数。日成交额/周月成交额按实际有效数据聚合，不把缺失当0。波动率、夏普等仅在收益类型、年化系数及无风险利率口径定义明确时提供。

年度同期：前一年度最后有效收盘作为年初基准；自然年周次（明确 ISO 或自然年锚定方案；建议自1月1日起每7天，最多53桶），每桶取截至桶末且可用的最后真实收盘，当前年截止 as_of；53周、节假日、跨年、上市不足、当前周未完必须正确。历史同期比较用相同周次截止点，历史全年值另列；历史分位、中位数、均值、上涨比例明确样本数、排除不完整年份，不作未来预测。统一图表层使用真实时间坐标，不用简单点序号冒充时间；数据 gap 断线。

## 6. 存储、调度、API 及前后端契约

### 6.1 SQLite（增量迁移，不破坏账本）

沿用 `services/api-go/internal/storage/sqlite.go`，新增独立 schema migration 目录/runner。建议：`schema_migrations`、`assets`、`asset_categories`、`asset_memberships(valid_from,valid_to)`、`universes`、`universe_members`、`provider_symbols`、`daily_bars`、`corporate_actions`、`fx_rates`、`trading_calendars`、`index_definitions`、`index_components`、`index_daily_values`、`algorithm_versions`、`trend_snapshots`、`trend_daily_metrics`、`strength_daily_rankings`、`trend_events`、`asset_feature_tags`、`data_sync_jobs`。保持现有 accounts/positions/transactions schema、数据及访问语义。按真实用量合并表也可，但不能丢失历史成分和版本。主键/唯一约束/索引覆盖 `(asset_id, trading_date, provider, revision)`、`(snapshot_id, asset_id)`、`(snapshot_id, scope_id, currency, asset_id)`、event dedupe，外键、migration 版本和向后兼容。数据库备份、恢复演练、事务失败不留下半成品；WAL + busy_timeout 保留，批量小事务、单写协调、避免查询 N+1；不在事务内发 HTTP 或做大计算。历史回填按日期、市场分批，进度可恢复、有限并发，避免阻塞 2.5s 行情刷新。按日频和需求保留历史，先容量评估再扩大到全市场。不得触碰生产真实文件，仅用临时副本测试。

### 6.2 更新与快照发布

新增独立盘后任务与手动管理命令（受保护，不公开可触发的写 API）。按市场时区/净值/夜盘实际可用时间触发，幂等锁、有限退避、限流、超时、取消、指标监控；不在每次页面请求同步抓全历史。流程：获取增量 → 校验/异常隔离 → 标准收益/指数 → 指标 → 排名/节气/梯队 → 事件 → 生成统一 `snapshot_id` → 原子切换 current pointer。输入失败保持上一个合法快照并显示 stale，发布不能部分成功伪装为已更新；可按市场独立快照并提供研究统一截面，记录批次 source dates 与覆盖。API 缓存键包含 snapshotId/asset/pool/range/currency/chart/benchmark/config/method；利用现有 ETag 与 gzip，防用户数据跨账号泄露。只读快照可缓存，用户设置和报告权限敏感数据不可共享缓存。旧 `/api/v1/markets/*` 等接口契约不变。

### 6.3 建议接口（沿用 chi + `{code,message,data}`）

统一在 `services/api-go/internal/router/router.go` 注入新 handler（main.go 构造），沿用 handler/render.go `writeSuccess/writeFail`，旧接口 JSON 不变。前端新建 `src/api/trend.ts` + `src/api/types/trend.ts`（按工程规范），复用 `http`，不复制请求拦截器。所有路径是建议契约，G0 冻结后以文档为准。GET 支持明确参数、白名单、游标/页长（最大100或200）、错误码、ETag；过长自定义组合用有限 POST 只读分析请求而非无限 URL。不要允许客户端指定任意上游 URL（SSRF）。

| 方法 & 路径 | 输出/关键参数 |
|---|---|
| GET `/api/v1/assets/catalog` | 市场/分类目录、节点能力、覆盖状态、snapshot meta |
| GET `/api/v1/assets/universes` | 公开观察池配置、候选规则、日期 |
| GET `/api/v1/assets/search?q=&type=&cursor=` | 名称/代码/别名/拼音索引，分页 |
| GET `/api/v1/assets/{id}` | 资产元数据、可用图表/币种/历史范围、快照指标 |
| GET `/api/v1/assets/{id}/members?scope_id=&sort=&order=&cursor=` | 直属成员及分组、趋势、行情、规模、成交额；排序稳定；也支持 universe 节点 ID |
| GET `/api/v1/trends/snapshots/latest?scope_id=&currency=&mode=` | 快照版本、as_of、各数据实际日期、覆盖率、状态 |
| GET `/api/v1/trends/{id}/curve` | range/start/end/freq/currency/benchmark_id/snapshot_id；实际日期收益与价格、基准、成交额、极值、事件、状态区间 |
| GET `/api/v1/trends/{id}/annual` | years/week_mode/currency/snapshot_id；周序列、same_week、full_year、样本、分位 |
| GET `/api/v1/trends/{id}/components` | scope_id/range/members/top_n/sort；同基期成分多曲线、截面字段、缺失原因 |
| GET `/api/v1/trends/{id}/dashboard` | change_period/spark_range/sort/scope_id/group/cursor；分组明细、真实 sparkline |
| GET `/api/v1/trends/{id}/season` | direction/snapshot_id；阶段定义、各阶段计数、成员及进度（按模型版本） |
| GET `/api/v1/trends/{id}/distribution` | weight_by/color_by/level/snapshot_id；层级树、权重、valid/right/unknown counts、加权占比 |
| GET `/api/v1/trends/{id}/tier` | scope_id/snapshot_id；五级成员、分档规则、候选数/有效数/排除项 |
| GET `/api/v1/trends/{id}/events` | types/since/until/cursor/snapshot_id；事件历史 |
| GET `/api/v1/trends/methodology` | 算法版本、阈值、数据含义及限制 |
| GET `/api/v1/trends/data-status` | 数据来源、实际日期、市场状态、覆盖、降级 |
| POST `/api/v1/reports/prepare`（有必要时） | 校验报告配置、不可篡改快照标识、公开可分享权限；限流、有鉴权要求时先复用真实认证 |

P0 收藏、图表收藏、主题、主页先使用 Pinia 持久化 + 版本化本地 storage，匿名用户也可使用；不可假称已云同步。P1 若已有合法用户身份体系且确认需要云同步，再增加 user-scoped CRUD，鉴权、权限、冲突处理和迁移；不能把模拟账户ID当真实用户身份。收藏资产主键 assetId；图表收藏存 chartType、range、currency、benchmark、scope、filters、snapshotMode，动态收藏恢复最新、固定报告恢复原快照（无法恢复时明确提示）。

### 6.4 DTO 示例（不是固定返回值）

```ts
export type Temp = 'boiling'|'hot'|'warm'|'flat'|'cool'|'cold'|'freezing'
export type Direction = 'up'|'down'|'neutral'|'unknown'
export type Side = 'right'|'left'|'neutral'|'unknown'
export type ChartType = 'curve'|'annual'|'components'|'dashboard'|'season'|'distribution'|'tier'
interface SnapshotMeta {
  snapshot_id: string; as_of: string; computed_at: string;
  algorithm_version: string; currency: string; scope_id: string;
  scope_label: string; candidate_count: number; valid_count: number;
  coverage: number|null; mode: 'latest_available'|'aligned';
  status: 'ok'|'degraded'|'stale'|'offline_seed'|'error'|'insufficient'|'not_covered';
  source_dates: Array<{market:string;as_of:string;source:string}>;
  warnings: string[];
}
interface Strength { scope_id:string; scope_label:string; score:number|null; rank:number|null;
  raw_factor:number|null; valid_count:number; delta:number|null; reason?:string }
interface TrendMetrics { temperature:Temp|null; temperature_score:number|null;
  direction:Direction; confirmation:'forming'|'confirmed'|'none'|'unknown'; side:Side;
  started_at:string|null; confirmed_at:string|null; duration_trading_days:number|null;
  duration_calendar_days:number|null; phase:string|null; strength:Strength[];
  tier:'S'|'A'|'B'|'C'|'D'|null; events:TrendEvent[]; reason?:string }
interface SeriesPoint { ts:string; trading_date:string; value:number|null;
  return_pct:number|null; turnover:number|null; status:string }
interface ChartContext { asset_id:string; pool_id:string; snapshot_id:string;
  chart:ChartType; currency:string; range:string; benchmark_ids:string[];
  change_period?:string; spark_range?:string; weight_by?:string; color_by?:string;
  selected_members?:string[]; sort?:string; order?:'asc'|'desc' }
```

DTO 值用 number|null，不把 `亿元` 或 `+6.1%` 混为数据字段；百分比用清楚统一单位（`*_pct`=百分数，比如 2.4 代表2.4%，内部计算收益 fraction 仅在转换边界乘100）；`right_ratio` 明确 0~1 还是百分数，建议 `right_ratio_pct`，同时返回 right_count、valid_count、unknown_count、weight_coverage、weighted_ratio_pct；数值和图表日期同快照。所有未定义字段通过契约测试。禁止将前端路由参数作为可信权限或 SQL 排序字符串直接拼接。

## 7. UI/UX 设计落地（对照 design/screens 和 DESIGN_SPEC）

### 7.1 品牌视觉及布局

延续仓库批准的「精密航海仪器盘 × 三环信号颜色编码」，主背景 `#090C0F`，卡片 `#12161B`，突出层 `#181E25`，品牌金 `#E5B352`，红涨 `#F05A50`、绿跌 `#2AD1A3`，文字 `#F1EFEA`/`#9AA5B1`。保留品牌罗盘视觉、严格等宽数字、细边线、窄边距、克制卡片，不采用竞品原样复制或大面积渐变。浅色金融阅读主题（纸白背景、深文字、暖金强调）与暗色统一语义，不能依靠全局硬编码导致其他 Tab 破坏。参考 `design/tokens.json`，由 `app.css` @theme/CSS variables 统一；主题管理扩展系统主题与金融模块手动偏好，避免主题影响旧页面，确保字体对比度。正文≥12px，核心数字≥14px，标签尽量≥11px，触控目标≥44×44逻辑像素（紧凑图标的透明触控区域同样达标），手机375×812/750rpx为设计基准，适配320/360/390/414/大字体/横屏。顶部微信原生胶囊区域、底部安全区使用真实 `uni.getMenuButtonBoundingClientRect` 等平台 API、CSafeArea/BaseLayout，不硬编码状态栏高度。使用现有图标组件或原创 SVG 静态资源、data URI 或 Canvas；**小程序 WXML 不直接嵌内联 `<svg>`**，HTML原型的 SVG 只用于设计展示。避免 emoji 作为核心图标。

### 7.2 金融资产首页 / 分类列表 / 详情

- 首页标题「金融资产」，提供市场目录、全局搜索、收藏、自选默认主页、数据状态；保留高密度表格，顶部可折叠市场概况（右侧品种数/有效数、增强/减弱、数据覆盖），默认首屏可看到至少5～7行（视设备）。精选观察池非互斥分类，入口支持全部/股票/ETF/债券/期货/汇率等，用户切换类别时保留合法筛选。进入「大类资产」是观察池，不冒充独立指数。选择市场分类可钻取行业或具体资产。列表组头显示成员数量/展开折叠；全部折叠/展开、排序（名称/温度/天数/最新值/涨跌/规模/成交额/强度）、升降箭头、筛选（状态、强度、标签、市场）、币种、列配置。 sticky 表头、列表分页、返回状态恢复；横向滑动表格不触发底部 Tab 切换。
- AssetRow：温度+持续天数；资产名称、代码/成分数、右侧占比；最新值+规模；日涨幅+成交额；强度条+范围或变化箭头；收藏。温度缺失 `—`，0分有效；价格缺失 `—`，0涨跌为 `0.0%` 中性色；标签独立第二/第三行，最多两行，更多折叠；右侧状态用金色浅背景与文字共同表达，不用颜色作为唯一状态。宽度小优先保留名称/温度/价格/涨跌/强度，次要规模成交额可缩进第二行或列设置/横向查看；避免字体缩到9px。每种资产字段适配 `measure_type`（ETF 净值≠市场价，债券收益率≠价格，期货规模不叫流通市值）。
- 统一资产详情：面包屑（可点击上级）、资产ID/单位、更新时间、状态/事件、2×3或3×2核心指标区（最新值、趋势状态/天数、温度、相对强度/比较范围、区间变化、适用股息率/右侧占比），内嵌主资产+基准累计收益曲线、成交额、极值标记，时间范围（近3月/6月/1年/3年/6年/全部 按能力），更多趋势图入口，事件、下属成员列表、关联资产。叶节点无成员则隐藏，不能显示空表。详情支持收藏、设为默认金融主页、分享；旧宏观行情和账本功能保持。
- 搜索选择器：搜索中文/英文/证券代码/拼音/别名；历史搜索、推荐分类、结果按市场/资产类型分组；节流、防乱序响应、空态、无权限提示、分页；选择资产后跳转对应上下文，不误改基准。跨资产切换要取消旧请求、清空不适用图表范围、恢复该资产已保存配置。

### 7.3 七图统一研究工作台

复用 `pages-sub/divergence/index.vue`，将标题栏改为返回+资产名/可选资产、上/下品种（按进入的固定列表顺序）、数据日期、收藏/配置。两行紧凑图表 Tab 或横向滚动（选中自动居中，显示当前序号）；可滑动切换，图表拖动/双指/图例/表格触摸时不触发 swiper，按需锁定手势。不要固定 `min-h-[780px]` 造成空白和裁切。每个视图独立高度/单一垂直滚动容器，底部安全区域给工具栏，长表格完整可达；非当前图表懒加载、卸载 Canvas/监听、避免七图同时占内存。切换资产/时间/币种/基准时以最后请求为准，骨架/缓存态可保留旧数据但标旧快照，禁止混合。

- ChartRegistry 统一能力：curve/annual 需可用收益序列；components/dashboard/distribution/tier 需成员；season 叶节点显示历史阶段而非空分布；无指数的观察池支持成员图但禁用自身曲线；非法视图明确说明。按能力展示，数字旧深链仍可识别。默认组件可切换上一/下一图；每种图的独立配置存储。`activeChart`/range/asset/pool/currency/snapshotId 的唯一数据源由组合式函数管理，不能多个组件互相覆盖状态。
- **图1 `TrendCurveView`**：顶部当前收益及价格/温度/强度，比较基准父级或用户选择，区间收益/最高低点/最大上涨/最大回撤、实际日期、成交额柱（独立轴、单位）、右侧/下行背景、零线/终值线/事件。当前价格模式/累计收益/相对收益（用明确的比率或差额口径）；线性/可适用对数；日周月频独立于区间。长按十字线、显示日期和数值、缩放后局部指标明确是选定区间还是全局；颜色固定。标题从计算结果生成，不生成买卖预测。
- **图2 `AnnualTrendView`**：当前年粗线、历史近3年默认可选、当前截止周垂直线、同周点、历史全年终点分别标识，统一自然年周次；显示今年同期、同周均值/中位数/排名/样本数；历史月份上涨比例 P1 用真实数据统计（不是预测概率）。跨年、当前未满年和53周无数据时留空，不延伸当前曲线。支持选择更多年份但限数量，图例可隐藏。
- **图3 `ComponentTrendView`**：复用现有排序和温度指南，主基准粗线，成分多曲线、同基期/币种/日期；固定资产色映射（资产ID哈希稳定、冲突解决），默认6–8个，TopN或手动选择，图例点击隐藏/突出，终点标签避让并画引线；截面表温度、品种、最新值、区间变动、日/月成交额、累计收益，排序与选中联动，基准置顶。长周期按真实覆盖，保留旧27年预设但不编造。`DivergenceChart` 的 yRange/labels 可复用，x 以 ts 映射，直线为默认，单调插值才允许，防贝塞尔过冲。缺失断线，基准真实序列不能误画成只有末值的水平线。
- **图4 `TrendDashboardView`**：分组数据仪表盘（不是固定78分半圆温度盘）；列：温度/天数、品种/价格、变动、迷你图、成交额、相对强度+背景条。每日(日涨幅/近3月)、每周(周涨幅/近6月)、每月(月变动/近1年)报告预设；`change_period` 和 `spark_range` 独立控制；支持规模/强度等排序、全/指定分组、筛选、表头吸顶、长表格分页。迷你图标真实高低/最新点、可选基准，极值由全数据算；列宽不能在不同分组变动。完整导出与当前可视列表分离，报告显示候选范围。
- **图5 `SolarTermsView`**：上行/下行/全部选择，按自研阶段映射绘制阶段带/示意曲线与资产点；阶段标题、解释、成员数、占比、点选/阶段列表，典型标注限量且不重叠；无趋势/未知单列；叶节点显示状态时间线。禁止使用当前自然节气或预设 `current:true`；曲线明确非价格预测，数量取实际有效成员，分母与统计范围一致。
- **图6 `RightDistributionView`**：真实嵌套 treemap，面积按流通市值/成交额/等权（必须适用，混合不同单位的规模不聚合），颜色按趋势状态/温度/节气/区间涨跌；默认状态（上行确认浅暖、震荡中性、下行冷色深浅），图例、数量占比+加权占比、样本数及权重覆盖。树图面积严格按权重，层级标题和选中状态，过小色块聚合“其他”且可下钻/列表查看；不写死A股，0%正确、不把unknown计为neutral/非右侧而不说明。点击色块可查看指标、进详情或钻取。
- **图7 `TrendTierView`**：五梯队分档逻辑见第5.5；每组左侧等级+数量，右侧网格/列表的名称、分数、天数、变化；可切换专业/趣味命名、固定候选范围、同分排序、筛选及完整长图。高分非右侧和低分非左侧不能进极端梯队；显示未计算及入选规则。P1 历史梯队迁移及回放。

### 7.4 图表底层与端上实现

先核查 `components/charts/CChart.vue` 当前仅简单 view 柱条，不可当完整引擎。抽象纯 geometry/scale/tick/line/spark/treemap/label-layout modules（TS），H5 可用 Canvas/SVG适配，微信优先 Canvas 2D/平台 API，禁止未经验证的 Web-only DOM/SVG 直接进入 WXML。复用 `uni.createSelectorQuery().in(instance)` 及可见时初始化，处理 DPR、宽高变化、安全区域、清理 listener/canvas。触摸命中坐标考虑滚动与 DPR。可用数据量限制：每条曲线渲染按画布像素降采样（LTTB/保留关键点，业务指标仍基于原始序列）；绘制最多约8–12条可见曲线，长表分组与分页，避免大量 Canvas 并行，迷你图批量/共享绘制或轻量静态预生成。为屏幕阅读器/无障碍提供文本表格和数据摘要，不让图表只依赖颜色。若使用第三方图表库，先验证小程序兼容、包大小、许可证和真机性能，再记录 ADR；不为图表重构全站。

### 7.5 收藏、偏好、状态与报告

- 收藏列表分资产/图表；资产支持分组、排序、批量移除和空态，星标立即反馈，持久化失败回滚；设为金融模块默认页，不劫持整个 App 默认策略 Tab。图表收藏保存配置/快照语义，分享复现。金融设置：深色/浅色/跟随系统，默认CNY（若有原币/其他目标币种）、涨跌色约定、密度、默认区间/主页、报告主题。提供指标说明（温度、强度、右侧、节气、梯队、来源、模型版本及限制），解释0、null、排名范围；数据状态页逐市场展示实际日期、覆盖、休市/延迟/失败/未授权。
- 状态矩阵：初次 loading 骨架、刷新保留数据标旧日期、无数据/样本不足/不适用、空筛选、网络错误可重试、stale缓存、休市、许可不可用、部分图表可用、图片导出失败/权限拒绝、快照过期。禁止显示假实时绿点或用当天日期顶替。请求序列 token/取消、异常和回退都可测试。按对照 `design/screens` 实现，不把占位交互变成死按钮。
- 海报报告：阅读版与海报独立模板；七图全部可导出，仪表盘、梯队可分页/长图。数据来自同一个 ChartContext/不可变 snapshot，页眉品牌、标题、实际数据日期、币种/范围/口径、图表主体、结论仅客观描述、来源、算法版本、限制/样本数，页尾合法小程序码（真实生成，不能用设计稿占位码上线）。1080px宽或 3×DPR，正文大小不因长表变得不可读；超过平台 Canvas 最大尺寸/内存时分页导出并提示，重试/取消/授权流程完整。适配微信 `onShareAppMessage`/分享路径/保存相册（用户触发、权限正确），5:4 分享卡片，链接恢复 asset/chart/range/currency/scope/benchmark/snapshotId。不得泄露用户私有观察池；默认报告隐藏个人信息，校验公开分享权限。若原快照不可用，提示后选择最新，不静默变更历史。严禁截取滚动屏幕当完整报告。前端 Canvas 或后端受控渲染择一/混合，保持轻量和授权。先实现客户端海报，服务端长图按能力后补；不承诺尚未验证的微信分享限制。

## 8. 实现文件定位与复用（建议）

| 路径 | 工作内容 |
|---|---|
| `apps/miniapp/src/pages/index/components/TabMarkets.vue` | 金融入口、保留旧趋势看板 |
| `apps/miniapp/src/pages-sub/divergence/index.vue` | 工作台、旧参数适配、单一上下文、懒加载 |
| `apps/miniapp/src/components/trend/{TrendCurveView,AnnualTrendView,ComponentTrendView,TrendDashboardView,SolarTermsView,RightDistributionView,TrendMenuDropdown}.vue` | 逐个真实数据接线，新增 TrendTierView，统一 props/emits |
| `apps/miniapp/src/components/finance/DivergenceChart.vue` | 复用坐标/标签逻辑，修复时间序列与 MP Canvas |
| `apps/miniapp/src/data/divergence.ts` | 字典提取到 domain；演示数据移到 tests/fixtures，生产移除 |
| `apps/miniapp/src/pages-sub/finance/*` | 新增资产全景、详情、搜索、自选、报告、设置/说明 |
| `apps/miniapp/src/components/finance/*` | 复用 MarketCard 等或新增 AssetList/AssetRow/AssetMetricGrid/AssetSelector/TrendTemp/TrendEventTag 等，避免重复 |
| `apps/miniapp/src/api/trend.ts`, `src/api/types/trend.ts`, `src/composables/useTrendResearch.ts`, `src/stores/finance.ts` | 新 API、类型、请求/缓存、用户偏好（命名以工程约定为准） |
| `apps/miniapp/src/app.css`, `src/stores/theme.ts`, `src/composables/useTheme.ts` | token、作用域主题、清除过时 Wot 示例，不破坏旧页面 |
| `services/api-go/internal/market` | 历史 provider adapter、许可及真实数据 |
| `services/api-go/internal/storage` | migration、查询、快照、幂等数据存储 |
| `services/api-go/internal/trend`（建议新包） | 可复用纯量化函数、状态机、ranking、phase、tier、event |
| `services/api-go/internal/{dto,service,handler,router}` | 契约、服务编排与新路由；保留旧 API |
| `services/api-go/cmd/server/main.go` | 组装任务、依赖注入、关闭生命周期 |
| `packages/*`、`tests/` | 保留原研究核心，新增独立黄金样本/交叉验证测试，不改变双策略定义 |
| `docs/`、`.github/workflows/ci.yml` | 方法文档、许可、迁移、接口、测试报告、CI，不降低 gate |

## 9. P1/P2 差异化（先保证 P0）

P1：趋势事件中心（按时间/市场/自选/类型筛选）；自定义观察池（不默认产生指数，组合权重需明确）；市场广度（数量/市值加权分布及历史占比）；周期罗盘相对轮动（强弱 × 动量变化四象限，与宏观美林时钟和节气模型区分；说明基准及计算）；历史状态/梯队/轮动回放（使用历史快照）；收益日历、相关性矩阵（共同样本/收益类型/币种/窗口）；自选提醒（仅用户主动订阅，合规与微信能力允许，非投资建议）；更多合法资产覆盖。P2：事件后历史收益分布（含样本数、置信区间和选择偏差）、量化实验室（历史可用信息、防幸存者偏差、交易成本、样本外；不改原双策略）、AI客观数据摘要（只引用已计算数据、日期和来源，不生成未证实事实/个性化建议）、会员（先核查类目、资质、支付及权益，不默认收费）。P1/P2 若阻塞则文档标记，不渲染虚假功能。

## 10. 测试、性能、安全与验收要求

详见 `docs/ACCEPTANCE.md`（本包）。必须先跑原测试基线，记录失败不伪造通过。保持 Python pytest、Go test/build、前端 type-check/test/build；增加趋势算法 golden fixtures、正负/缺失边界、API契约、迁移回滚、跨币种、跨市场、报告一致性、图表手势、微信真机测试。所有新增测试可离线运行，网络集成测试单列。基准库与实现使用独立计算方式验证，不能用同一个错误函数生成期望值。

建议实际运行（以当前 package scripts / CI 为准）：

```bash
# 仓库根目录（Python 研究/历史验证依赖按 pyproject 安装）
pytest tests/ -v
# Go
cd services/api-go
gofmt -w <本次修改的Go文件>
go test ./...
go vet ./...
CGO_ENABLED=0 go build -o /tmp/cyclecompass-go ./cmd/server
# 如条件允许，单独运行 go test -race ./...（需CGO）
# 前端，从仓库根进入 apps/miniapp
pnpm install --frozen-lockfile
pnpm type-check
pnpm test:run
pnpm lint
pnpm format:check
pnpm build:mp-weixin
pnpm build:h5
```

不要全仓格式化无关文件；格式检查若已有历史失败，先记录基线再保证本次改动通过并逐步修复。对照 CI 的真实工作目录。H5截图不能替代微信真机/开发者工具截图；至少检查 iOS/Android、320/375/390/414px、暗/浅主题、长名称、100+行、多曲线、超长历史、休市/断网/慢网、导出权限拒绝、图表切换手势、安全区。性能目标为建议预算，先基线压测再调整：2核2G环境日频预计算与行情刷新共存，热缓存列表/首屏图 API p95 <300ms（本机或明确网络条件、典型100资产，测量实际环境）、冷请求 p95 <1.5s（不含外部回填）、列表首屏不依赖全历史；手机端避免明显掉帧，长表滚动、切图及导出不 OOM；新增主包增长受控，图表在分包，核对当前微信实际包体限制。记录样本量、设备、网络、CPU/RSS、SQLite容量、构建包体和测试方法；未测量不得写“达标”。

安全：SQL 参数化、sort/filters allowlist、限制输入长度/页长/图表范围/并发/报告尺寸，禁止 SSRF、日志脱敏、公开报告权限、缓存隔离、依赖漏洞检查。历史数据许可、真实市场日期和投资信息边界为上线阻断项。生产变更需独立备份和人工授权。

## 11. 交付及停止条件

交付：实际增量源码、迁移脚本与回滚说明、`docs/FINANCE_V4_ARCHITECTURE.md`、`docs/TREND_METHODOLOGY.md`、`docs/FINANCE_API.md`（请求/响应/错误）、更新 `DATA_SOURCE_MATRIX.md`、`DECISIONS.md`、`GOAL_PROGRESS.md`、`ISSUE_TRACKER.md`、`TEST_REPORT.md`、真机截图与验收对照（注明真实/演示）、性能/安全/合规报告、未覆盖/剩余风险。可用的报告功能必须有导出样图（真实授权样本或明确演示）。提供变更文件清单、复用矩阵、测试命令与真实输出、下一步。若代码中发现新的缺陷，逐轮 review/fix/verify，保留既有产品功能与 API。P0 未接线、无数据却显示真实行情、算法不一致、崩溃、迁移损坏、权限/合规阻塞不得标为完成。无法执行的测试列原因及可复现步骤。P1/P2 未实施应如实列为延期，不以 UI 占位充数。

最终不要自动合并或部署；待用户确认后才执行相关远程操作。开始 G0，读取本包设计和审查资料，先完成基线、冻结契约，再持续推进至全部可执行任务完成。
