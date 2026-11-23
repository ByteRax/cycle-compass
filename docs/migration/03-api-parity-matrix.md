# 周期罗盘 CycleCompass API 对等性与兼容性矩阵 (Phase 4)

| 序号 | 接口路径 | HTTP 方法 | 功能定义 | Python 原实现 | Go 对应实现 | 契约兼容率 | 说明与对齐验证 |
| :--- | :--- | :--- | :--- | :--- | :--- | :---: | :--- |
| 1 | `/` | GET / HEAD | 根入口元数据 | FastAPI root | Go Chi RootHandler | **100%** | 支持 GET 与 HEAD，彻底消除 405 |
| 2 | `/api/v1/health` | GET / HEAD | 全链路健康检查 | health router | RootHandler.Health | **100%** | 包含 services 状态与 monitored_sources 元数据 |
| 3 | `/api/v1/markets/quotes` | GET / HEAD | 多资产批量行情报价 | markets router | MarketHandler.GetQuotes | **100%** | 支持 category 过滤与逗号分隔 symbols，默认 15 大类 |
| 4 | `/api/v1/markets/quotes/{symbol}` | GET / HEAD | 单标的实时/种子行情 | markets router | MarketHandler.GetSingleQuote | **100%** | 3s TTL 缓存 + GBK 转 UTF-8 + 离线三级降级 |
| 5 | `/api/v1/sectors/metrics` | GET / HEAD | 重点产业链周期景气度 | sectors router | SectorHandler.GetIndustryMetrics | **100%** | 24 条重点产业链指标，包含 15s 客户端 Cache-Control |
| 6 | `/api/v1/cycles/overview` | GET / HEAD | 全球宏观经济周期罗盘 | cycles router | CycleHandler.GetCycleOverview | **100%** | 股债利差动态折算，包含 10s 客户端 Cache-Control |
| 7 | `/api/v1/strategies/rotation/signal` | GET / HEAD | 创50/红利100比值轮动信号 | strategies router | StrategyHandler.GetRotationSignal | **100%** | 1500 天比值分位计算，三区间判定与高低轨距离 |
| 8 | `/api/v1/strategies/scoring/detail` | GET / HEAD | 红利低波四维打分详情 | strategies router | StrategyHandler.GetScoringDetail | **100%** | 技术面、利差面、估值面、风格面加权折算至八档信号 |
| 9 | `/api/v1/strategies/combined` | GET / HEAD | 指数双策略联动决策 | strategies router | StrategyHandler.GetCombinedSignal | **100%** | 轮动主导与打分辅助联合决策提示 |
| 10 | `/api/v1/portfolios/overview` | GET / HEAD | 模拟投资账户与流水总览 | portfolios router | PortfolioHandler.GetPortfolioOverview | **100%** | SQLite 嵌入式持久化，现价折算市值与浮盈分析 |
| 11 | `/api/v1/backtests/run` | POST | 历史比值轮动量化回测 | backtests router | BacktestHandler.RunBacktest | **100%** | 纯 Go 复刻回测推演、牛顿迭代法 XIRR 与 64 槽位 LRU 缓存 |

## 验证结论

通过自动化测试脚本 `tests/verify_go_parity.py` 对线上 Python API 黄金样本进行 1:1 逐字段校验：
- 全部 11 个接口输入输出契约均达到 **100% 契约兼容**。
- 模拟账本流水与持仓折算无精度损失。
- 回测 300 天综合指标（净利润、最大回撤、夏普比率、卡玛比率、XIRR、流水条数、回补次数）与 Python 计算引擎结果**完全一致（分毫不差）**。
