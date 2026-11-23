# Phase 0: 生产现状与源码全面审计报告

## 一、 当前生产实际部署拓扑

```
[微信小程序 (uni-app / WeApp)]
        │ HTTPS (TLS 1.2/1.3 + HTTP/2)
        ▼
[110.40.230.69:443 - 1Panel OpenResty (Nginx 1.31)]
        │ Upstream proxy_pass: http://127.0.0.1:8000 (Keep-Alive 32)
        ▼
[cycle-compass-api (Docker 容器)]
  ├── 架构: Linux Debian 13 (x86_64), Python 3.15.2
  ├── ASGI 服务器: Uvicorn (2 Workers, uvloop, httptools)
  ├── 框架: FastAPI 0.165 + Pydantic v2
  ├── 内存配额: 512MB 限制 (当前常驻 223.9 MB)
  └── 数据交互:
        ├── 在线实时行情: 腾讯财经 (qt.gtimg.cn)
        ├── 历史时序: 本地 CSV 快照 (references/.../data/*.csv)
        └── 模拟账户: 内存单例 (_MOCK_ACCOUNT, LedgerEntry)
```

---

## 二、 生产服务器硬件与系统指标审计

- **操作系统**：Debian GNU/Linux 13 (trixie)，内核版本 `6.12.113+deb13-amd64`
- **CPU**：AMD EPYC 7K62 64-Core，分配 2 vCPU
- **物理内存**：总计 1.9 GiB，已使用 1.0 GiB，可用 923 MiB，缓存 911 MiB
- **虚拟内存（Swap）**：总计 2.0 GiB，已使用 405 MiB，剩余 1.6 GiB
- **磁盘存储**：根分区 `/` 容量 49 GB，已用 11 GB，可用 39 GB（使用率 22%）
- **现有容器分布**：
  1. `cycle-compass-api`：常驻内存 **223.9 MiB**（当前 Python 双进程）
  2. `1Panel-qinglong-VgwJ`：内存 216.5 MiB
  3. `1Panel-nps-0QYC`：内存 20.7 MiB
  4. `1Panel-openresty-lRvy`：内存 11.8 MiB
  5. `starrynightrider`：内存 11.8 MiB
  6. `quadrantforce`：内存 9.7 MiB

---

## 三、 全量 HTTP API 接口清单与业务契约

| 序号 | Method | 路由路径 | 参数/Body | 响应核心数据结构 | 数据来源 | 关键逻辑 |
| :---: | :---: | :--- | :--- | :--- | :--- | :--- |
| 1 | `GET` | `/` | 无 | `{name, version, docs, health}` | 静态元数据 | 根路径信息与服务版本展示 |
| 2 | `GET` | `/api/v1/health` | 无 | `{status, timestamp, services, monitored_sources}` | 核心标的在线探活 | 健康检查，若核心标的异常返回 degraded |
| 3 | `GET` | `/api/v1/markets/quotes` | `category` (query, 可选)<br>`symbols` (query, 逗号分隔) | `list[MarketQuote]` | 腾讯公开行情源 + 本地 CSV 快照降级 | 批量查询，带 3s TTL 内存防抖缓存 |
| 4 | `GET` | `/api/v1/markets/quotes/{symbol}` | `symbol` (path) | `MarketQuote` | 同上 | 单标的实时/快照行情 |
| 5 | `GET` | `/api/v1/sectors/metrics` | 无 | `list[IndustryMetric]` | 静态行业景气度种子配置 | 24条重点产业链评分看板，带 15s 缓存头 |
| 6 | `GET` | `/api/v1/cycles/overview` | 无 | `CycleCompassState` | 实时行情 + 宏观国债收益率 | 计算当前股债利差与比值，给出宏观罗盘状态 |
| 7 | `GET` | `/api/v1/strategies/rotation/signal` | 无 | `RotationSignal` | 创业板50与红利低波实时比值 + 1600天历史分位 | 三区间轮动（低区/高区/中间区）买卖信号计算 |
| 8 | `GET` | `/api/v1/strategies/scoring/detail` | `base_amount` (query, 默认200) | `ScoreResult` | 估值分位、股息率、均线偏离、风格偏离 | 红利低波100四维评分（50/25/15/10）及定投金额建议 |
| 9 | `GET` | `/api/v1/strategies/combined` | 无 | `CombinedStrategySignal` | 轮动信号 + 四维打分 | 双策略联合动作指导与执行文案 |
| 10 | `POST` | `/api/v1/backtests/run` | `{"low_line":0.16, "high_line":0.28, "weekly_amount":1000.0, "days":1500}` | `BacktestOutput` | 1600日历史日K线数据 (CSV) | 运行全历史定投轮动回测（XIRR、最大回撤、时序走势） |
| 11 | `GET` | `/api/v1/portfolios/overview` | 无 | `PortfolioOverview` | 模拟账户持仓 + 实时现价折算 | 资金守恒记账流水、持仓盈亏、现金池总览 |

---

## 四、 数据库与持久化现状

1. **目前无独立数据库服务**：未运行独立 PostgreSQL 或 MySQL 容器。
2. **时序与行情基准数据**：本地 CSV 存储（`dividend_930955.csv`、`growth_399673.csv`、`bond10y.csv`），只读挂载。
3. **模拟投资账户**：内存单例对象 `_MOCK_ACCOUNT`，进程重启即丢失用户新增流水（存在业务缺陷，Go 迁移中应顺势升级为轻量持久化）。

---

## 五、 迁移痛点与风险分析

1. **内存压力**：Python API 即使优化后仍常驻 223.9MB，占去服务器超 11% 内存，容易因突发计算触碰 512MB 限制。
2. **CPU 计算与 I/O 阻塞**：回测算法涉及大量时序逐日推演与 XIRR 牛顿迭代法，在 Python 单线程事件循环中易产生争抢。
3. **API 兼容性红线**：前端微信小程序已深度绑定 JSON 响应结构（`code: 0, message: "success", data: ...`），Go 必须保证每一个字段的命名（小写下划线）、数据类型（浮点精度、时间格式）100% 契约一致。
