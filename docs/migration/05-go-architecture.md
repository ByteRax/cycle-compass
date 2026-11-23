# Phase 2: Go API 架构设计规范

## 一、 设计目标与选型原则

1. **极轻量运行**：无重型框架与多余依赖，单二进制打包，运行内存控制在 15MB~30MB 内。
2. **零运行时外部依赖**：使用纯 Go 原生实现与纯 Go SQLite 驱动，无 CGO 依赖，交叉编译与容器分发极其敏捷。
3. **100% 契约保真**：严格依据 `tests/golden_api_snapshots/` 结构体标签定义 DTO，保持字段命名、空值与精度一致。

---

## 二、 目录结构设计 (`services/api-go`)

```text
services/api-go/
├── cmd/
│   └── server/
│       └── main.go                  # 应用入口、依赖组装、优雅停机
├── internal/
│   ├── config/
│   │   └── config.go                # 环境变量配置
│   ├── dto/
│   │   ├── response.go              # 通用 ApiResponse[T] 包装
│   │   ├── market.go                # 市场行情 DTO
│   │   ├── sector.go                # 产业景气度 DTO
│   │   ├── cycle.go                 # 宏观周期罗盘 DTO
│   │   ├── strategy.go              # 双策略信号与打分 DTO
│   │   ├── backtest.go              # 回测请求与回测输出 DTO
│   │   └── portfolio.go             # 模拟账户持仓与流水 DTO
│   ├── handler/
│   │   ├── health.go                # /health, /
│   │   ├── market.go                # /api/v1/markets/*
│   │   ├── sector.go                # /api/v1/sectors/*
│   │   ├── cycle.go                 # /api/v1/cycles/*
│   │   ├── strategy.go              # /api/v1/strategies/*
│   │   ├── backtest.go              # /api/v1/backtests/*
│   │   └── portfolio.go             # /api/v1/portfolios/*
│   ├── market/
│   │   ├── client.go                # 腾讯批量行情拉取 + 3s TTL + SingleFlight
│   │   └── seed.go                  # 本地 CSV 离线种子快照读取与缓存
│   ├── middleware/
│   │   ├── recovery.go              # Panic 恢复与 500 兜底
│   │   ├── logger.go                # slog 结构化访问日志
│   │   ├── cors.go                  # CORS 跨域支持
│   │   └── gzip.go                  # 响应自动压缩
│   ├── router/
│   │   └── router.go                # 路由挂载与中间件组装
│   └── storage/
│       └── sqlite.go                # SQLite 嵌入式数据库（账户持久化）
├── go.mod
├── go.sum
└── Dockerfile                       # 多阶段构建生产镜像
```

---

## 三、 核心并发与稳定性机制

1. **行情批量并发与 SingleFlight 防击穿**：
   - 沿用成熟的 SingleFlight 模式（Go 原生 `golang.org/x/sync/singleflight`），无论多少个用户并发请求行情，向腾讯发起的实际请求始终为 1 次，其余并发协程等待复用结果。
2. **内存极短 TTL 缓存 (3.0s)**：
   - 内存字典存储最新行情，读写锁（`sync.RWMutex`）保护，3 秒内查询 0 毫秒立即返回。
3. **优雅停机 (Graceful Shutdown)**：
   - 监听 `SIGINT` 与 `SIGTERM`，给在途请求预留 5 秒超时窗口，释放资源后安全退出。
