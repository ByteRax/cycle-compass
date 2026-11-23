# 周期罗盘 CycleCompass

> **洞察全球市场，把握经济周期。**
> 全球经济与产业周期观察、金融行情监测、指数策略研究和模拟投资平台。

---

## 一、 产品愿景与定位

建立一套能够长期维护、扩展和持续运行的金融数据及投资研究平台，打通：
$$\text{全球行情} \longrightarrow \text{产业景气} \longrightarrow \text{周期分析} \longrightarrow \text{量化策略} \longrightarrow \text{回测研究} \longrightarrow \text{模拟投资}$$
的完整闭环。

- **产品首期形态**：微信小程序（mp-weixin）为主，预留 H5 / Web / App 多端能力。
- **业务边界**：首期不接入真实交易账户，不实现真实自动下单，不代客理财，不提供收益保证。所有策略与模拟持仓均为严谨的规则量化研究与可解释的记账账本。

---

## 二、 核心架构与技术栈

### 1. 前端技术栈 (`apps/miniapp`)
- **基座框架**：基于 [ByteRax/Uniapp-Template-Cli](https://github.com/ByteRax/Uniapp-Template-Cli) 构建，全面复用 Vue 3 + TypeScript + Vite + Pinia 持久化 + UniKu 根组件能力。
- **样式体系**：**Tailwind CSS v4** + **weapp-tailwindcss v5**，全面对齐 CSS-First 架构与 `@theme` 金融颜色变量体系（红涨绿跌、主色系、`rem2rpx` 适配）。
- **组件库**：**Compass UI** 自研组件系统（基础 UI、金融卡片、周期罗盘、轮动指示器、四维评分仪表、回测概览、持仓账本等）。
- **质量保障**：ESLint + Prettier + vue-tsc (0 error) + Vitest (48/48 测试全部通过)。
- **严禁依赖治理**：彻底移除 Wot Design Uni、UnoCSS、unocss-applet、Oxlint、Oxfmt、z-paging，采用原生 `scroll-view` 与响应式数据流。

### 2. 后端技术栈（`services/api-go` 主服务 & `services/api` 研究服务）
- **Go 生产主服务 (`services/api-go`)**：Go 1.23+，内置 SQLite 持久化、2.5s 主动行情刷新、优雅停机、非 root 多阶段 Alpine 镜像与本机 `pprof` 探针，默认监听 `8002`。
- **Python 量化研究服务 (`services/api`)**：Python 3.12+ FastAPI + Pydantic v2 + Uvicorn + uv，与 `packages/*` 算法核心直连，用于离线研究与回测校验，默认监听 `8000`。
- **RESTful API**：
  - `/api/v1/health`：系统健康检查与多源数据质量监控
  - `/api/v1/markets/*`：全球指数、汇率、大宗商品、亚太行情与国债收益率
  - `/api/v1/sectors/*`：24条重点产业链景气度、AI Token/算力/存储价格看板
  - `/api/v1/cycles/*`：宏观周期罗盘阶段、美林时钟偏好、股债收益利差
  - `/api/v1/strategies/*`：创业板50/红利低波100三区间轮动、红利四维评分、双策略联合决策
  - `/api/v1/backtests/*`：全历史回测在线计算器（XIRR、TWR 时间加权最大回撤、夏普比率）
  - `/api/v1/portfolios/*`：实盘模拟投资账户、虚拟持仓、资金守恒审计流水

### 3. 量化策略与行情引擎 (`packages/*`)
- **`packages/strategy_engine`**：指数双策略算法引擎。修复现金池回补与当周定投单日多事件分离记账，实现纯函数式状态推进与资金守恒记账流水（LedgerEntry）。
- **`packages/market_data`**：多源聚合行情服务。建立三级数据质量状态（`OK` 实时主源 / `DEGRADED` 备用源 / `OFFLINE_SEED` 离线种子快照），杜绝静态假数据。
- **`packages/backtesting`**：量化回测与绩效分析框架（牛顿迭代 XIRR、最大回撤、夏普比率、卡玛比率）。

---

## 三、 目录结构

```
cycle-compass/
├── apps/
│   └── miniapp/             # 微信小程序 / H5 前端项目
│       ├── src/
│       │   ├── api/         # 接口定义与请求封装
│       │   ├── components/  # Compass UI (ui, finance, charts, layout)
│       │   ├── pages/       # 主包页面 (index 底部4Tab全链路: 策略/行情/产业/账本)
│       │   ├── pages-sub/   # 分包页面 (AI助手, 登录, 状态页等)
│       │   ├── data/        # 离线兜底种子数据
│       │   ├── http/        # HTTP 请求拦截器与取消机制
│       │   └── app.css      # Tailwind v4 CSS-First 入口与 @theme 变量
│       ├── vite.config.ts   # WeappTailwindcss 插件配置
│       └── package.json
├── packages/
│   ├── strategy_engine/     # 指数双策略算法引擎 (轮动、四维评分、执行器)
│   ├── market_data/         # 多源行情数据服务 (在线抓取与种子快照)
│   └── backtesting/         # 回测框架 (XIRR、TWR最大回撤、夏普比率)
├── services/
│   ├── api-go/              # Go 生产主 API 服务 (SQLite 持久化 + 2.5s 主动刷新)
│   └── api/                 # FastAPI 量化研究与回测参考服务
├── infra/
│   └── docker/              # Dockerfile.api & docker-compose.yml
├── tests/                   # 黄金快照 Golden Tests、回测测试与 API 集成测试
├── references/              # 原始双策略交付包与设计图参考
├── docs/                    # 架构、接口、测试报告与追踪文档
├── pyproject.toml           # Python 工作区与依赖定义 (uv 管理)
└── .github/workflows/       # GitHub Actions 自动化 CI 流程
```

---

## 四、 快速上手

### 1. 后端与算法服务启动
```bash
# 1. 创建并激活虚拟环境 (推荐使用 uv)
uv venv
source .venv/bin/activate

# 2. 安装项目依赖
uv pip install -e ".[dev]"

# 3. 运行全量后端与算法测试 (14 项全部通过)
pytest tests/ -v

# 4. 启动 FastAPI 服务
uvicorn services.api.main:app --host 0.0.0.0 --port 8000 --reload
# 访问 Swagger API 文档: http://localhost:8000/docs
```

### 2. 前端小程序开发与构建
```bash
cd apps/miniapp

# 1. 安装依赖
pnpm install

# 2. 类型检查 (0 错误)
pnpm type-check

# 3. 运行前端自动化测试 (61 项全部通过)
pnpm test:run

# 4. 启动 H5 开发环境
pnpm dev:h5

# 5. 编译微信小程序产物
pnpm build:mp-weixin
# 产物输出至 apps/miniapp/dist/build/mp-weixin，在微信开发者工具中导入即可预览
```

### 3. Docker 容器化一键部署
```bash
cd infra/docker
docker compose up -d
```

---

## 五、 测试与质量验收

| 测试层级 | 命令 | 测试套件/用例数 | 结果 |
|---|---|---|---|
| **后端算法回归 (Golden Tests)** | `pytest tests/test_golden_20261120.py` | 4 用例 | ✅ 100% 通过（2026-11-20 快照浮点误差 < 1e-4） |
| **回测与度量测试** | `pytest tests/test_backtesting.py` | 3 用例 | ✅ 100% 通过（XIRR、MDD、回测引擎全通过） |
| **后端 API 集成测试** | `pytest tests/test_api_integration.py` | 7 用例 | ✅ 100% 通过（健康检查、行情、产业、策略等） |
| **前端类型检查** | `pnpm type-check` | vue-tsc | ✅ 0 error |
| **前端自动化测试** | `pnpm test:run` | 11 套件 / 61 用例 | ✅ 100% 通过 (耗时 394ms) |
| **微信小程序编译** | `pnpm build:mp-weixin` | Vite + WeappTailwindcss | ✅ 成功生成微信开发者工具产物 |

---

## 六、 开源与免责声明

本项目提供的量化策略、回测模型与模拟账本仅供学术研究、规则验证和量化模拟投资交流使用，不构成任何实质性投资建议或收益保证。市场有风险，投资需谨慎。
