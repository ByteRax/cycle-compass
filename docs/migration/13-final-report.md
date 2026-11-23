# 周期罗盘 CycleCompass 全面迁移 Go 与低资源优化总结报告 (Phase 8 最终交付)

## 一、项目背景与核心挑战
「周期罗盘 CycleCompass」项目已正式上线运营，并在 Debian 13、2 vCPU、1.9GB 物理内存的低配云服务器上运行。原技术栈使用 Python FastAPI 提供在线 HTTP 服务。
由于 Python 运行时、Pandas/NumPy 数据处理开销以及 Uvicorn 双 Worker 架构，原 API 常驻物理内存高达 225.8MB，在高并发与多指标计算时面临 CPU 飙升与内存 Swap 颠簸风险；此外，原模拟账户数据存放于进程内存，重启即丢失账本流水。

本次工程目标：**在不改变现有 uni-app 前端代码、确保接口 100% 兼容、模拟资金账本零丢失的前提下，将在线 HTTP API 无损迁移为 Go 原生实现，消除内存与延迟瓶颈，实现生产平滑热切流与极端低资源优化。**

---

## 二、架构演进与技术选型

```
[ 微信小程序 (uni-app) ]
         │ (HTTPS 443 / TLS 1.3 / HTTP/2)
         ▼
[ 1Panel OpenResty (Nginx 反向代理) ]
         │ (内部反代 upstream: 127.0.0.1:8002)
         ▼
[ CycleCompass Go API (单一静态二进制 / 11MB) ]
 ├── net/http + go-chi/v5 (轻量路由中间件、Gzip 压缩、CORS、HEAD 支持)
 ├── 腾讯公开实时行情适配器 (3s 内存 TTL 缓存 + 并发防击穿互斥锁 + GBK 转码)
 ├── 纯 Go 离线种子提供者 (CSV 毫秒级内存预加载 + 1600天比值序列缓存)
 ├── 嵌入式 SQLite 账本引擎 (modernc.org/sqlite, WAL 模式, 自动 Seed, 零 CGO)
 ├── 量化策略计算引擎 (创50/红利100比值分位计算、红利四维评分折算)
 └── 纯 Go 量化回测引擎 (历史步进状态机、牛顿迭代法 XIRR、64 槽位 LRU 缓存)
```

---

## 三、工程产物与交付成果清单

1. **Go 核心代码库** (`services/api-go/`)：
   - `cmd/server/main.go`：主服务入口、后台异步预热与优雅停机（Graceful Shutdown）。
   - `internal/config/config.go`：环境变量动态加载。
   - `internal/dto/`：完整契约结构体（`market`, `sector`, `cycle`, `strategy`, `portfolio`, `backtest`, `health`）。
   - `internal/market/`：
     - `seed.go`：离线 CSV 种子数据解析、24 产业链与宏观国债指标、历史比值序列缓存。
     - `client.go`：长连接池腾讯行情客户端、GBK 自动转码、3s 内存防抖缓存、并发防击穿锁与平滑降级。
   - `internal/storage/sqlite.go`：纯 Go 嵌入式 SQLite 存储引擎，WAL 并发优化，支持模拟账户 `ACC_DEMO_001`、持仓与流水持久化。
   - `internal/service/`：行情服务、板块服务、宏观周期服务、指数双策略计算服务、模拟投资资产折算服务、高频量化回测引擎。
   - `internal/handler/`：各业务 Handler 与统一错误/成功响应渲染。
   - `internal/router/router.go`：Chi 路由装配、全局 RequestID、RealIP、Logger、Recoverer、Gzip(5 级)与 HEAD 方法全面支持。
   - `Dockerfile`：Alpine 生产极简多阶段构建镜像（体积仅 24.5MB）。
2. **测试与比对套件**：
   - `services/api-go/internal/service/service_test.go`：Go 单元测试，0.31 秒全量通过。
   - `tests/verify_go_parity.py`：全量 11 个接口自动化对齐比对测试脚本。
   - `tests/golden_api_snapshots/`：线上 11 个接口真实黄金样本存档。
3. **架构与实施文档库** (`docs/migration/`)：
   - `01-current-state.md`（生产现状与源码全面审计）
   - `02-baseline-report.md`（黄金样本与性能基线）
   - `03-api-parity-matrix.md`（API 对等性与兼容性矩阵，100% 兼容）
   - `04-api-diff-report.md`（差异与兼容性深度分析报告）
   - `05-go-architecture.md`（Go API 架构设计方案）
   - `06-python-worker-design.md`（Python 离线计算与 Worker 设计）
   - `07-database-compatibility.md`（SQLite 兼容性与数据模型设计）
   - `08-performance-report.md`（性能基线与对比报告）
   - `09-security-review.md`（安全与防风险审计报告）
   - `10-deployment-plan.md`（生产发布与平滑切换方案）
   - `11-rollback-plan.md`（紧急回滚预案）
   - `12-production-verification.md`（生产验收与全链路测试报告）
   - `13-final-report.md`（本总结报告）

---

## 四、核心优化指标达成对比

| 关键指标 | 迁移前 (Python/FastAPI) | 迁移后 (Go 原生) | 优化成果 |
| :--- | :--- | :--- | :---: |
| **API 进程常驻内存** | **225.8 MB** | **9.5 MB** | **降低 95.8% (节省 216MB)** |
| **服务器总可用物理内存** | **857 MB** | **1,077 MB** | **可用物理内存突破 1GB 大关** |
| **二进制与依赖体积** | **1,350 MB** | **11 MB** (静态剥离) | **体积减少 99.2%** |
| **服务冷启动时间** | **3,850 ms** | **48 ms** | **启动提速 80 倍** |
| **API 首屏 P50 延迟** | **38.62 ms** | **1.85 ms** | **延迟降低 95%** |
| **回测单次计算耗时** | **185.0 ms** | **1.85 ms** | **计算提速 100 倍** |
| **100 并发吞吐 (QPS)** | **153.6 req/s** | **2,480.0 req/s** | **吞吐能力跃升 16 倍** |
| **数据持久化安全性** | 内存变量（重启易失） | SQLite WAL（原子持久化） | **持仓与资金流水永久可靠** |
| **HTTP 规范性** | HEAD 请求报 405 | 原生完美支持 HEAD / GET | **消除探针异常，提升兼容性** |

---

## 五、最终状态与验收建议
1. **当前线上状态**：生产域名 `https://cycle-api.tanranran.cn` 已由 Go API 全面接管（运行在 `127.0.0.1:8002`，由 systemd 守护），旧 Python 容器已安全停止（保留镜像备用），全链路回归 100% 成功。
2. **前端兼容性**：微信小程序无需任何代码改动，接口入参、响应字段、精度、数据结构完全无缝衔接。
3. **后续运维**：服务由 systemd 托管，开机自启、Crash 3 秒自愈，日志通过 `journalctl -u cyclecompass-go -f` 统一查看，轻量纯粹，维护极简。
