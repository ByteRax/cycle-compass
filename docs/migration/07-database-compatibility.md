# Phase 2: 数据库兼容性与无损持久化方案

## 一、 当前存储现状分析

当前 Python FastAPI 服务端并未启用独立外部数据库（如 PostgreSQL / MySQL），而是：
1. **历史 K 线与行情种子**：直接以只读方式加载 `references/strategy_original/index_rotation/data/*.csv`。
2. **模拟持仓与资金流水**：存放在 `_MOCK_ACCOUNT` 内存单例中。

---

## 二、 Go 迁移中的存储升级策略（无损 + 持久化）

为了彻底解决“服务端重启后模拟账户新增资金流水丢失”的痛点，同时兼顾 2GB 服务器无多余内存运行 Postgres 的现实约束，Go 服务端采用：

### 1. 嵌入式 SQLite 3 存储
- **驱动选型**：`modernc.org/sqlite`（纯 Go 语言转译版本，无 CGO 依赖，跨平台编译零门槛）。
- **库文件位置**：`/opt/cycle-compass/data/cycle_compass.db`。
- **并发控制**：开启 `WAL`（Write-Ahead Logging）模式，读写完全并发互不阻塞；设置 `busy_timeout = 5000ms`。

### 2. Schema 结构设计（与原有数据结构精确映射）
- **`positions` 表**：
  - `account_id TEXT`, `symbol TEXT`, `name TEXT`, `shares REAL`, `cost_price REAL`, `created_at TEXT`, `updated_at TEXT`
- **`transactions` 表 (LedgerEntry)**：
  - `id INTEGER PRIMARY KEY AUTOINCREMENT`, `account_id TEXT`, `date TEXT`, `event_type TEXT`, `symbol TEXT`, `target_name TEXT`, `shares REAL`, `price REAL`, `amount REAL`, `cash_inflow REAL`, `cash_outflow REAL`, `cash_pool_delta REAL`, `cash_pool_after REAL`, `description TEXT`, `created_at TEXT`
- **`accounts` 表**：
  - `account_id TEXT PRIMARY KEY`, `available_cash REAL`, `cash_pool REAL`, `updated_at TEXT`

### 3. 数据无损初始化（Seed Initialization）
Go API 服务启动时，如果检测到 SQLite 数据库为空，自动将原本 Python 中内嵌的默认演示账户数据（`ACC_DEMO_001`、红利低波 12500 股、创业板50 8000 股、两条定投审计流水）无损导入数据库，保证与旧版数据 **100% 精确一致**。
