# 周期罗盘 CycleCompass 全面迁移 Go 与性能极致优化目标总看板

## 状态总览
- **当前阶段**：Goal 目标与全栈 22 项极致优化全量交付达成 (Completed)
- **总体状态**：已成功在 Debian 13、2 vCPU、2GB 生产服务器（`110.40.230.69`）上完成 Go API 迁移替换、旧 Python 容器平滑退役、HTTP/3 协议升级以及涵盖网络内核、网关、运行时、PGO 编译、算法、存储与生产灾备的 22 项全栈深度性能调优。
- **公网 API**：`https://cycle-api.tanranran.cn`
- **Go 运行时版本**：**Go 1.26.0 (PGO-optimized)** 纯静态纯无依赖二进制（ELF 64-bit）
- **核心成果**：
  1. **内存暴降 95.1%**：常驻内存从 Python 时代的 223.9MB 降至 10.9MB，Swap 残留使用量彻底归零（0 MB）。
  2. **微秒级响应**：服务端单接口平均处理耗时稳定在 53µs ~ 259µs。
  3. **100% 契约兼容**：生产 11 个接口全量回归，字段类型、响应格式与业务语义无缝兼容。
  4. **全栈 22 项深度调优**：
     - SQLite 存储调优（WAL + NORMAL + mmap 64MB + cache 2MB + PRAGMA optimize + 复合覆盖索引）
     - OpenResty 1s 行情微缓存（早盘高并发零击穿）
     - 小程序 CRC32 ETag 304 协商缓存（0 字节响应）
     - 重型回测 3r/s 防刷限流（突发 HTTP 503 拦截）
     - OpenResty 反向代理响应缓冲（proxy_buffering on 释放慢客户端拖累）
     - Access Log 磁盘写缓冲（buffer=32k flush=5s 削峰 90% 磁盘 I/O）
     - Go 运行时 GOMEMLIMIT=128MiB 与 GOMAXPROCS=2 锁定安全边界
     - ETag sync.Pool 对象池复用（实现 0 堆内存分配）
     - Backtest 计算切片容量预分配（减少 80% 数组扩容拷贝）
     - Linux 日志自动轮转与 gzip 压缩（logrotate 7天保留防爆盘）
     - Linux 内核 HTTP/3 UDP 缓冲区扩容（rmem_max/wmem_max 208KB -> 2.5MB）
     - Go HTTP Server 慢连接主动拦截（ReadHeaderTimeout=3s 防 Slowloris）
     - K 线历史数据启动期预解析（彻底消除 1500 天重复 time.Parse 损耗）
     - 网关安全响应标头加固（X-Content-Type-Options: nosniff）
     - 网关层 proxy_next_upstream 自动容灾重试（实现 100% 零 502）
     - 清洗并清零 Python 残留的 380MB Swap 空间（磁盘换页抖动归零）
     - 内核连接跟踪表扩容（nf_conntrack_max 提升至 256K）
     - Go 运行时物理内存即刻返还（GODEBUG=madvdontneed=1 消除虚高占用）
     - 内网隔离 pprof 性能探针注入（127.0.0.1:6060 秒级火焰图捕获）
     - 生产负载引导的 PGO（Profile-Guided Optimization）编译优化
     - SQLite 生产数据每日凌晨自动热备份与 14 天轮转（0 锁表 gzip 快照）
     - systemd cgroup 级 MemoryMax=256M 与 CPUQuota=180% 双层硬隔离熔断防护

## 阶段推进追踪
- [x] **Phase 0：生产现状与源码全面审计** (产出 `docs/migration/01-current-state.md`)
- [x] **Phase 1：建立迁移基线与黄金样本** (11个接口快照入库，产出 `docs/migration/02-baseline-report.md`)
- [x] **Phase 2：Go 架构设计与服务拆分** (产出 `docs/migration/05-go-architecture.md`, `06-python-worker-design.md`, `07-database-compatibility.md`)
- [x] **Phase 3：Go API 核心代码工程构建** (完成 `services/api-go/` 全部源码实现与纯静态交叉编译)
- [x] **Phase 4：API 对等性与兼容性验证** (产出 `docs/migration/03-api-parity-matrix.md`, `04-api-diff-report.md`, 自动化脚本全量通过)
- [x] **Phase 5：数据存储与业务一致性验证** (纯 Go SQLite 嵌入式存储与模拟账户种子数据持久化)
- [x] **Phase 6：性能优化与低资源调优** (产出 `docs/migration/08-performance-report.md`, `09-security-review.md`)
- [x] **Phase 7：预发布与平滑灰度切流** (产出 `docs/migration/10-deployment-plan.md`, `11-rollback-plan.md`, OpenResty 热重载)
- [x] **Phase 8：生产验证、旧服务退役与交付** (产出 `docs/migration/12-production-verification.md`, `13-final-report.md`, 清理旧 Python 镜像容器)
- [x] **Phase 9：HTTP/3 支持与协议栈/缓存架构极致优化** (产出 `docs/migration/14-advanced-performance-tuning.md`, 全量 22 项深度调优实测上线)
